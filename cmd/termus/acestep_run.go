package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mrbrutti/termus/internal/acestep"
	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
	"github.com/mrbrutti/termus/internal/track"
	"github.com/mrbrutti/termus/internal/tui"
)

// acestepOptions is the subset of the CLI flags that govern ACE-Step playback.
// Keeping this in one struct keeps main.go's switch clean.
type acestepOptions struct {
	trackPath    string
	serviceURL   string
	crossfadeSec float64
	queueDepth   int
	maxTracks    int
	outputDir    string
	renderTO     time.Duration
	autoStart    bool // when true, manage the daemon ourselves
	port         int
	serviceDir   string
	noTUI        bool // when true, skip the TUI and stream plain stderr updates
	initialVol   int
}

// messageSink is the minimum surface acestep streaming needs in order to push
// status updates somewhere. In TUI mode it is a *tea.Program; in headless mode
// it is a small adapter that writes phase changes to stderr.
type messageSink interface {
	Send(msg tea.Msg)
}

// stderrSink turns tea.Msg events into one-line stderr prints. Used by the
// headless (non-TTY or --no-tui) code path so scripted callers still see
// progress, just as before SP23.
type stderrSink struct {
	w         io.Writer
	mu        struct{ lastPhase string }
	lastRender atomic.Int64
}

func newStderrSink(w io.Writer) *stderrSink {
	if w == nil {
		w = os.Stderr
	}
	return &stderrSink{w: w}
}

func (s *stderrSink) Send(msg tea.Msg) {
	switch ev := msg.(type) {
	case tui.ACEStepInstallProgressMsg:
		if ev.Err != nil {
			fmt.Fprintf(s.w, "preparing AI engine: %s: ERROR: %v\n", ev.Phase, ev.Err)
			return
		}
		if ev.Phase != s.mu.lastPhase {
			s.mu.lastPhase = ev.Phase
			fmt.Fprintf(s.w, "preparing AI engine: %s: %s\n", ev.Phase, ev.Detail)
		}
	case tui.ACEStepStatusMsg:
		if ev.Err != nil {
			fmt.Fprintf(s.w, "preparing AI engine: %s: ERROR: %v\n", ev.Phase, ev.Err)
			return
		}
		if ev.Phase != s.mu.lastPhase {
			s.mu.lastPhase = ev.Phase
			fmt.Fprintf(s.w, "preparing AI engine: %s: %s\n", ev.Phase, ev.Detail)
		}
	case tui.ACEStepRenderingMsg:
		if ev.Done {
			return
		}
		// Throttle render-progress lines so we don't spam stderr.
		now := time.Now().UnixMilli()
		last := s.lastRender.Load()
		if now-last < 2000 {
			return
		}
		s.lastRender.Store(now)
		if ev.Err != nil {
			fmt.Fprintf(s.w, "termus: render seq=%d: %v\n", ev.Seq, ev.Err)
			return
		}
		detail := ev.Detail
		if detail == "" {
			detail = fmt.Sprintf("generating track %d", ev.Seq+1)
		}
		fmt.Fprintf(s.w, "termus: %s\n", detail)
	case tui.ACEStepReadyMsg:
		if ev.Detail != "" {
			fmt.Fprintf(s.w, "termus: %s\n", ev.Detail)
		} else {
			fmt.Fprintln(s.w, "termus: AI engine ready")
		}
	}
}

// runACEStep is the top-level entry. It chooses between TUI mode and the
// legacy stderr-only headless mode and dispatches accordingly.
func runACEStep(ctx context.Context, opts acestepOptions) error {
	if opts.noTUI || !isTerminal(os.Stderr) {
		return runACEStepHeadless(ctx, opts)
	}
	return runACEStepWithTUI(ctx, opts)
}

// runACEStepHeadless preserves the pre-SP23 stderr-printing behaviour for
// non-interactive callers (scripts, redirected stderr, --no-tui).
func runACEStepHeadless(ctx context.Context, opts acestepOptions) error {
	sink := newStderrSink(os.Stderr)
	file, err := loadACEStepTrack(opts.trackPath)
	if err != nil {
		return err
	}
	if opts.outputDir != "" {
		if err := os.MkdirAll(opts.outputDir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", opts.outputDir, err)
		}
	}
	client, manager, err := acquireACEStepClient(ctx, opts, sink)
	if err != nil {
		return err
	}
	prod := newRenderingProducer(client, file, opts.trackPath, opts.outputDir, sink)
	s := audio.NewStreamer(audio.StreamerConfig{
		Producer:     prod,
		QueueDepth:   opts.queueDepth,
		CrossfadeSec: opts.crossfadeSec,
		MaxTracks:    opts.maxTracks,
		Logger:       os.Stderr,
	})
	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	go func() {
		select {
		case sig := <-sigCh:
			fmt.Fprintf(os.Stderr, "termus: received %v, stopping...\n", sig)
			cancel()
		case <-streamCtx.Done():
		}
	}()
	if err := s.Start(streamCtx); err != nil {
		return fmt.Errorf("start streamer: %w", err)
	}
	statusDone := make(chan struct{})
	go acestepStatusHeartbeat(streamCtx, s, statusDone, os.Stderr)
	<-streamCtx.Done()
	fmt.Fprintln(os.Stderr, "termus: shutting down...")
	s.Stop()
	<-statusDone
	if manager != nil {
		_ = manager.Shutdown(context.Background())
	}
	if last := s.Status().LastError; last != nil && !errors.Is(last, context.Canceled) {
		return fmt.Errorf("streamer ended with error: %w", last)
	}
	fmt.Fprintln(os.Stderr, "termus: done.")
	return nil
}

// runACEStepWithTUI launches bubbletea immediately and drives the engine
// bootstrap + first-track render in a goroutine, surfacing every status as a
// tea.Msg the model already knows how to render.
func runACEStepWithTUI(ctx context.Context, opts acestepOptions) error {
	file, err := loadACEStepTrack(opts.trackPath)
	if err != nil {
		return err
	}
	if opts.outputDir != "" {
		if err := os.MkdirAll(opts.outputDir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", opts.outputDir, err)
		}
	}
	vol := opts.initialVol
	if vol <= 0 {
		vol = 70
	}
	title := acestepTrackNameFromPath(opts.trackPath)
	ring := scope.NewRing(4096)
	cmd := &acestepCommander{}
	model := tui.New(ring, cmd, "ACE-Step", title, 0, vol).
		WithStartupLoading("Setting up AI engine", "preparing AI engine...", 0)
	prog := tea.NewProgram(model, tea.WithAltScreen())

	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	go func() {
		select {
		case <-sigCh:
			cancel()
			prog.Quit()
		case <-streamCtx.Done():
		}
	}()

	var managerHolder atomic.Pointer[acestep.Manager]
	// Spawn the boot+playback goroutine. It sends progress messages until the
	// streamer goroutines own their own lifecycle.
	go func() {
		client, manager, err := acquireACEStepClient(streamCtx, opts, prog)
		if err != nil {
			prog.Send(tui.ACEStepStatusMsg{
				Phase:  "error",
				Detail: err.Error(),
				Err:    err,
			})
			// Give the user a moment to read the failure before quitting.
			time.Sleep(2 * time.Second)
			prog.Quit()
			return
		}
		if manager != nil {
			managerHolder.Store(manager)
		}
		// "rendering" is the final pre-ready phase; place it close to (but
		// not at) loadingCeiling so the bar visibly fills as the first track
		// generates. ACEStepReadyMsg dismisses the overlay entirely once the
		// first track is queued, so we never need to reach 1.0 here.
		prog.Send(tui.ACEStepStatusMsg{
			Phase:   "rendering",
			Title:   "Generating Music",
			Detail:  "composing first track (~10s)…",
			Percent: loadingCeiling,
		})
		prod := newRenderingProducer(client, file, opts.trackPath, opts.outputDir, prog)
		// Hook so the streamer marks "ready" as soon as the first track is
		// produced and queued.
		prod.onFirstReady = func() {
			prog.Send(tui.ACEStepReadyMsg{Detail: "AI engine ready"})
		}
		s := audio.NewStreamer(audio.StreamerConfig{
			Producer:     prod,
			QueueDepth:   opts.queueDepth,
			CrossfadeSec: opts.crossfadeSec,
			MaxTracks:    opts.maxTracks,
			Logger:       io.Discard,
			// Feed the same scope.Ring the TUI visualizer reads from, so
			// the waveform display actually moves while ACE-Step audio
			// plays. The headless path passes nil (no ring needed there).
			ScopeSink: ring,
		})
		if err := s.Start(streamCtx); err != nil {
			prog.Send(tui.ACEStepStatusMsg{
				Phase:  "error",
				Detail: "start streamer: " + err.Error(),
				Err:    err,
			})
			time.Sleep(2 * time.Second)
			prog.Quit()
			return
		}
		<-streamCtx.Done()
		s.Stop()
	}()

	if _, err := prog.Run(); err != nil {
		cancel()
		return fmt.Errorf("tui error: %w", err)
	}
	cancel()
	if m := managerHolder.Load(); m != nil {
		_ = m.Shutdown(context.Background())
	}
	return nil
}

// acestepCommander is the minimal audio.Commander needed to construct a
// tui.Model for ACE-Step. It owns no algorithm; the TUI in this mode only
// needs to render the loading overlay and a minimal now-playing view, so the
// commander's algorithm-related methods are no-ops.
type acestepCommander struct{}

// We satisfy audio.Commander with no-op implementations.
func (c *acestepCommander) SetVolume(_ int)                              {}
func (c *acestepCommander) DebugStatus() gen.DebugStatus                 { return gen.DebugStatus{} }
func (c *acestepCommander) TogglePause()                                 {}
func (c *acestepCommander) ToggleRecord() (string, error)                { return "", nil }
func (c *acestepCommander) SwapAlgorithm(_ gen.Algorithm)                {}
func (c *acestepCommander) SwapAlgorithmFade(_ gen.Algorithm, _ int)     {}

// runACEStepPlayback is kept as a thin shim for any external callers.
// Internally everything routes through runACEStep, which dispatches between
// the TUI and headless code paths.
func runACEStepPlayback(ctx context.Context, opts acestepOptions) error {
	return runACEStep(ctx, opts)
}

// loadACEStepTrack parses and validates an ACE-Step .tm file.
func loadACEStepTrack(path string) (*track.File, error) {
	file, err := track.ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if file.RenderEngine != track.RenderEngineACEStep {
		return nil, fmt.Errorf("track %s has render_engine=%q; --engine acestep requires render_engine=%q",
			path, file.RenderEngine, track.RenderEngineACEStep)
	}
	if file.Acestep == nil {
		return nil, fmt.Errorf("track %s sets render_engine=acestep but the 'acestep:' block is missing", path)
	}
	return file, nil
}

// statusSinkAdapter routes tea.Msg events produced by the legacy
// acquireACEStepClient / ensureACEStepReady code path into an
// audio.ACEStepStatusSink. Used by the SP26 Playback hot-switch factory so
// the existing message-flow code can stay unchanged.
type statusSinkAdapter struct{ s audio.ACEStepStatusSink }

func (a *statusSinkAdapter) Send(msg tea.Msg) {
	if a.s == nil {
		return
	}
	switch ev := msg.(type) {
	case tui.ACEStepInstallProgressMsg:
		a.s.OnInstallProgress(ev.Phase, ev.Title, ev.Detail, ev.Percent, ev.Err)
	case tui.ACEStepStatusMsg:
		a.s.OnStatus(ev.Phase, ev.Title, ev.Detail, ev.Percent, ev.Err)
	}
}

// buildACEStepFactory wraps the existing acquireACEStepClient code path in
// the audio.ACEStepFactory shape expected by Playback. The factory captures
// the per-run opts (port, service URL, etc.) so SwitchToACEStep doesn't have
// to repeat them.
func buildACEStepFactory(opts acestepOptions) audio.ACEStepFactory {
	return func(ctx context.Context, sink audio.ACEStepStatusSink) (*acestep.Client, *acestep.Manager, error) {
		adapter := &statusSinkAdapter{s: sink}
		return acquireACEStepClient(ctx, opts, adapter)
	}
}

// buildACEStepProducerFactory wraps the existing newRenderingProducer in the
// audio.ACEStepProducerFactory shape. SP26: per-track rendering progress flows
// through Playback's render sink (which fans out to the bus) rather than
// directly to *tea.Program, so the producer can be reused across hot-switches.
func buildACEStepProducerFactory(file *track.File, trackPath, outputDir string) audio.ACEStepProducerFactory {
	return func(client *acestep.Client, sink audio.ACEStepRenderSink) audio.AudioProducer {
		prod := newRenderingProducer(client, file, trackPath, outputDir, &renderSinkAdapter{s: sink})
		// Re-route onFirstReady through the new sink so Playback knows when
		// to dismiss the loader.
		prod.onFirstReady = func() {
			sink.OnFirstReady("AI engine ready")
		}
		return prod
	}
}

// renderSinkAdapter translates the tea.Msg events the existing producer
// emits into the audio.ACEStepRenderSink callbacks Playback exposes.
type renderSinkAdapter struct{ s audio.ACEStepRenderSink }

func (a *renderSinkAdapter) Send(msg tea.Msg) {
	if a.s == nil {
		return
	}
	if ev, ok := msg.(tui.ACEStepRenderingMsg); ok {
		a.s.OnRendering(ev.Seq, ev.Detail, ev.Done, ev.Err)
	}
}

// acquireACEStepClient connects to either a managed or externally-running
// ACE-Step daemon depending on opts. Streams install + status events to the
// sink (typically a *tea.Program; nil suppresses events).
func acquireACEStepClient(ctx context.Context, opts acestepOptions, sink messageSink) (*acestep.Client, *acestep.Manager, error) {
	if !opts.autoStart {
		client := acestep.NewClient(opts.serviceURL, opts.renderTO)
		healthCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		health, err := client.Health(healthCtx)
		cancel()
		if err != nil {
			return nil, nil, fmt.Errorf("ACE-Step service health check at %s failed: %w (did you start services/acestep/server.py? See services/acestep/README.md)",
				opts.serviceURL, err)
		}
		if !health.Loaded {
			return nil, nil, fmt.Errorf("ACE-Step service at %s reachable but not loaded (backend=%s, mock=%v, error=%q)",
				opts.serviceURL, health.Backend, health.MockMode, health.Error)
		}
		if sink != nil {
			sink.Send(tui.ACEStepStatusMsg{
				Phase:   "ready",
				Detail:  fmt.Sprintf("connected to %s", opts.serviceURL),
				Percent: 1.0,
			})
		}
		return client, nil, nil
	}
	manager, client, err := ensureACEStepReady(ctx, opts, sink)
	if err != nil {
		return nil, nil, err
	}
	return client, manager, nil
}

// ensureACEStepReady boots a Manager and, while it works, streams install +
// status events to the sink as tea.Msg values. The sink can be a *tea.Program
// in TUI mode or a stderrSink in headless mode.
func ensureACEStepReady(ctx context.Context, opts acestepOptions, sink messageSink) (*acestep.Manager, *acestep.Client, error) {
	installer := acestep.NewInstaller(opts.serviceDir, nil)
	// Single shared progress tracker so install + status phases advance the
	// same bar monotonically. Starts at 0 regardless of whether install runs
	// (SP24 fix: previously the bar jumped to ~0.65 when no install was
	// needed because each phase had a static target percent).
	progress := newLoadingProgress()
	var progressMu sync.Mutex
	if !installer.IsInstalled() {
		size := installer.EstimatedSize()
		// In TUI mode the install prompt would race with the alt-screen; the
		// non-interactive default is "yes" via the existing helper (which we
		// only call when there's a stdin and no TUI). For TUI mode we
		// auto-confirm so the loader can take over.
		if _, ok := sink.(*stderrSink); ok {
			fmt.Fprintf(os.Stderr, "termus: ACE-Step engine not installed. Setup will download:\n")
			fmt.Fprintf(os.Stderr, "  Python 3.11           ~%s\n", humanBytes(size.Python))
			fmt.Fprintf(os.Stderr, "  Dependencies          ~%s\n", humanBytes(size.Deps))
			fmt.Fprintf(os.Stderr, "  ACE-Step model        ~%s\n", humanBytes(size.Model))
			fmt.Fprintf(os.Stderr, "  Total                 ~%s\n", humanBytes(size.Total()))
			if !confirmInstall() {
				return nil, nil, errors.New("ACE-Step install declined")
			}
		} else if sink != nil {
			progressMu.Lock()
			pct := progress.observe("install:prompt")
			progressMu.Unlock()
			sink.Send(tui.ACEStepInstallProgressMsg{
				Phase:   "installing",
				Title:   "Setting up AI engine",
				Detail:  fmt.Sprintf("first run: downloading ~%s of model + deps…", humanBytes(size.Total())),
				Percent: pct,
			})
		}
	}
	evCh := make(chan acestep.InstallEvent, 32)
	installer.Events = evCh
	stCh := make(chan acestep.StatusEvent, 16)
	mgr := &acestep.Manager{
		Installer: installer,
		Port:      opts.port,
		Logger:    io.Discard,
	}
	doneEvents := make(chan struct{})
	doneStatus := make(chan struct{})
	go func() {
		defer close(doneEvents)
		for ev := range evCh {
			if sink == nil {
				continue
			}
			progressMu.Lock()
			pct := progress.observe("install:" + ev.Phase)
			progressMu.Unlock()
			sink.Send(tui.ACEStepInstallProgressMsg{
				Phase:   ev.Phase,
				Title:   installPhaseTitle(ev.Phase),
				Detail:  ev.Message,
				Percent: pct,
				Err:     ev.Err,
			})
		}
	}()
	go func() {
		defer close(doneStatus)
		for ev := range stCh {
			if sink == nil {
				continue
			}
			progressMu.Lock()
			pct := progress.observe("status:" + ev.Phase)
			progressMu.Unlock()
			sink.Send(tui.ACEStepStatusMsg{
				Phase:   ev.Phase,
				Title:   statusPhaseTitle(ev.Phase),
				Detail:  ev.Message,
				Percent: pct,
				Err:     ev.Err,
			})
		}
	}()
	client, err := mgr.EnsureReady(ctx, stCh)
	close(evCh)
	close(stCh)
	<-doneEvents
	<-doneStatus
	if err != nil {
		return nil, nil, fmt.Errorf("ACE-Step bootstrap failed: %w", err)
	}
	return mgr, client, nil
}

// installPhaseTitle is the loader title surfaced for each installer phase.
func installPhaseTitle(phase string) string {
	switch phase {
	case "install:python", "python":
		return "Setting up AI engine"
	case "install:deps", "deps", "pip":
		return "Setting up AI engine"
	case "install:model", "model":
		return "Downloading model"
	default:
		return "Setting up AI engine"
	}
}

// statusPhaseTitle is the loader title surfaced for each manager phase.
func statusPhaseTitle(phase string) string {
	switch phase {
	case "checking-install":
		return "Setting up AI engine"
	case "installing":
		return "Setting up AI engine"
	case "starting-daemon":
		return "Starting AI engine"
	case "loading-model":
		return "Starting AI engine"
	case "ready":
		return "Generating Music"
	default:
		return "Starting AI engine"
	}
}

// loadingProgress is a stateful, monotonic loader-percent tracker. It exists
// because the old static phase->percent maps assumed every cold-start phase
// would fire; when the daemon was already installed the bar jumped straight
// to ~0.65 on first paint (SP24 issue: "you start at 85%? the loading makes
// no sense.").
//
// The fix is to start at 0 and nudge the percent forward each time a *new*
// phase arrives, regardless of which phases the run hits. Each unique phase
// advances by a fixed step (loadingStep), capped at loadingCeiling until the
// final ACEStepReadyMsg lands and sets the bar to 1.0. Repeat events for the
// same phase don't double-count.
type loadingProgress struct {
	seen map[string]struct{}
	cur  float64
}

// loadingStep is how much each newly-seen phase nudges the bar forward.
// 0.15 lets a typical warm-start sequence (checking-install -> starting-daemon
// -> loading-model -> ready) progress 0.15 -> 0.30 -> 0.45 -> 0.60 before the
// first-render phase brings it the rest of the way. Cold starts hit more
// phases and converge nicely toward the ceiling.
const loadingStep = 0.15

// loadingCeiling caps the percent until ACEStepReadyMsg explicitly lands it
// at 1.0. Keeps the bar visibly short of "done" while the model loads /
// first track renders.
const loadingCeiling = 0.95

func newLoadingProgress() *loadingProgress {
	return &loadingProgress{seen: make(map[string]struct{})}
}

// observe nudges the percent forward when phase is new. Returns the current
// percent so callers can stamp it onto the outgoing tea.Msg.
func (p *loadingProgress) observe(phase string) float64 {
	if phase == "" {
		return p.cur
	}
	if _, ok := p.seen[phase]; ok {
		return p.cur
	}
	p.seen[phase] = struct{}{}
	p.cur += loadingStep
	if p.cur > loadingCeiling {
		p.cur = loadingCeiling
	}
	return p.cur
}

// confirmInstall prompts the user to confirm a ~11.5 GB download. Returns
// true if the user types y/Y/yes (or just hits enter, since default is yes).
func confirmInstall() bool {
	fmt.Fprintf(os.Stderr, "Install? [Y/n]: ")
	rdr := bufio.NewReader(os.Stdin)
	line, err := rdr.ReadString('\n')
	if err != nil {
		return false
	}
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "" || line == "y" || line == "yes"
}

// humanBytes formats a byte count in MB/GB.
func humanBytes(n int64) string {
	switch {
	case n >= 1024*1024*1024:
		return fmt.Sprintf("%.1f GB", float64(n)/(1024*1024*1024))
	case n >= 1024*1024:
		return fmt.Sprintf("%d MB", n/(1024*1024))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

// renderingProducer turns each Produce(seq) call into an HTTP /render call
// against the ACE-Step service. Seed is offset by seq so successive tracks
// are variations. It optionally streams seq-level progress to a sink so the
// TUI can show a "generating next track…" badge.
type renderingProducer struct {
	client       *acestep.Client
	baseSpec     acestep.RenderSpec
	outputDir    string
	trackName    string
	count        atomic.Int64
	sink         messageSink
	onFirstReady func()
	firstSent    atomic.Bool
}

func newRenderingProducer(client *acestep.Client, file *track.File, trackPath, outputDir string, sink messageSink) *renderingProducer {
	baseSpec, _ := acestep.CompileV3(file)
	return &renderingProducer{
		client:    client,
		baseSpec:  baseSpec,
		outputDir: outputDir,
		trackName: acestepTrackNameFromPath(trackPath),
		sink:      sink,
	}
}

func (p *renderingProducer) Produce(ctx context.Context, seq int) ([]byte, error) {
	spec := p.baseSpec
	if spec.Seed >= 0 {
		spec.Seed += int64(seq)
	}
	if p.sink != nil {
		detail := "composing first track (~10s)…"
		if seq > 0 {
			detail = fmt.Sprintf("generating track %d", seq+1)
		}
		p.sink.Send(tui.ACEStepRenderingMsg{Seq: seq, Detail: detail})
	}
	t0 := time.Now()
	data, err := p.client.Render(ctx, spec)
	if err != nil {
		if p.sink != nil {
			p.sink.Send(tui.ACEStepRenderingMsg{Seq: seq, Err: err})
		}
		return nil, err
	}

	if p.outputDir != "" {
		path := filepath.Join(p.outputDir, fmt.Sprintf("%s-%03d.wav", p.trackName, seq))
		if err := os.WriteFile(path, data, 0o644); err != nil {
			// Non-fatal: log to discard logger; sink will not see this.
			_ = err
		}
	}
	p.count.Add(1)
	if p.sink != nil {
		p.sink.Send(tui.ACEStepRenderingMsg{Seq: seq, Done: true, Detail: fmt.Sprintf("rendered seq=%d in %.1fs", seq, time.Since(t0).Seconds())})
	}
	if !p.firstSent.Swap(true) && p.onFirstReady != nil {
		p.onFirstReady()
	}
	return data, nil
}

func acestepStatusHeartbeat(ctx context.Context, s *audio.Streamer, done chan<- struct{}, w io.Writer) {
	defer close(done)
	if w == nil {
		w = io.Discard
	}
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			st := s.Status()
			fmt.Fprintf(w, "termus: status: playing=%v seq=%d queue=%d\n",
				st.Playing, st.CurrentSeq, st.QueueDepth)
		}
	}
}

func acestepTrackNameFromPath(p string) string {
	base := filepath.Base(p)
	ext := filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return base
}

// defaultACEStepServiceDir returns the absolute path to services/acestep
// relative to the running binary or current working directory.
//
// Resolution order:
//   1. $TERMUS_ACESTEP_DIR if set
//   2. <cwd>/services/acestep if it exists
//   3. <exe-dir>/services/acestep
//   4. fall back to <cwd>/services/acestep regardless
func defaultACEStepServiceDir() string {
	if v := strings.TrimSpace(os.Getenv("TERMUS_ACESTEP_DIR")); v != "" {
		return v
	}
	if wd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(wd, "services", "acestep")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "services", "acestep")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	wd, _ := os.Getwd()
	return filepath.Join(wd, "services", "acestep")
}

// runACEStepInstall is invoked by --acestep-install. It does the install
// pipeline and exits, with stderr progress reporting.
func runACEStepInstall() error {
	dir := defaultACEStepServiceDir()
	installer := acestep.NewInstaller(dir, nil)
	ch := make(chan acestep.InstallEvent, 64)
	installer.Events = ch
	done := make(chan error, 1)
	go func() {
		done <- installer.EnsureInstalled(context.Background())
	}()
	var lastPhase string
	drained := false
	for !drained {
		select {
		case ev, ok := <-ch:
			if !ok {
				drained = true
				continue
			}
			if ev.Phase != lastPhase {
				lastPhase = ev.Phase
				fmt.Fprintf(os.Stderr, "[%s] %s\n", ev.Phase, ev.Message)
			} else if ev.Message != "" {
				fmt.Fprintf(os.Stderr, "        %s\n", ev.Message)
			}
		case err := <-done:
			close(ch)
			// Drain remaining events.
			for ev := range ch {
				if ev.Phase != lastPhase {
					lastPhase = ev.Phase
					fmt.Fprintf(os.Stderr, "[%s] %s\n", ev.Phase, ev.Message)
				}
			}
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, "ACE-Step toolchain ready.")
			return nil
		}
	}
	return nil
}

// resolveEngineForTrack returns the engine to use for the given track path
// + --engine flag value. Returns "sf2", "acestep", or an error.
func resolveEngineForTrack(trackPath, engineFlag string) (string, error) {
	switch engineFlag {
	case "sf2":
		return "sf2", nil
	case "acestep":
		return "acestep", nil
	case "", "auto":
		// Inspect the .tm to decide.
		if trackPath == "" {
			return "sf2", nil
		}
		file, err := track.ParseFile(trackPath)
		if err != nil {
			return "", fmt.Errorf("inspect %s for engine resolution: %w", trackPath, err)
		}
		if file.RenderEngine == track.RenderEngineACEStep {
			return "acestep", nil
		}
		return "sf2", nil
	default:
		return "", fmt.Errorf("unknown --engine %q (want sf2, acestep, or auto)", engineFlag)
	}
}

// trackPathFromSelection returns the .tm path the user identified. Accepts
// either a direct path or a discovered ID; falls back to a relative .tm
// path on the filesystem.
func trackPathFromSelection(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", errors.New("no track specified")
	}
	if filepath.Ext(value) == ".tm" {
		if _, err := os.Stat(value); err == nil {
			return value, nil
		}
	}
	entries, err := discoverTracks()
	if err != nil {
		return "", err
	}
	entry, ok := track.Resolve(entries, value)
	if !ok {
		return "", fmt.Errorf("unknown track %q", value)
	}
	return entry.Path, nil
}

// isTerminal returns true when fd looks like an interactive terminal. We use
// this to auto-select the TUI vs the stderr-only path. A best-effort
// implementation: TERM=dumb or non-character device means headless.
func isTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	if strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}


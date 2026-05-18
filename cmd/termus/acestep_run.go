package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/mrbrutti/termus/internal/acestep"
	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/track"
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
}

// runACEStepPlayback opens the given .tm, ensures the ACE-Step daemon is
// running (auto-bootstrapping if needed), and streams forever.
func runACEStepPlayback(ctx context.Context, opts acestepOptions) error {
	file, err := track.ParseFile(opts.trackPath)
	if err != nil {
		return fmt.Errorf("parse %s: %w", opts.trackPath, err)
	}
	if file.RenderEngine != track.RenderEngineACEStep {
		return fmt.Errorf("track %s has render_engine=%q; --engine acestep requires render_engine=%q",
			opts.trackPath, file.RenderEngine, track.RenderEngineACEStep)
	}
	if file.Acestep == nil {
		return fmt.Errorf("track %s sets render_engine=acestep but the 'acestep:' block is missing", opts.trackPath)
	}

	if opts.outputDir != "" {
		if err := os.MkdirAll(opts.outputDir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", opts.outputDir, err)
		}
	}

	var client *acestep.Client
	var manager *acestep.Manager
	if opts.autoStart {
		manager, client, err = ensureACEStepReady(ctx, opts)
		if err != nil {
			return err
		}
	} else {
		client = acestep.NewClient(opts.serviceURL, opts.renderTO)
		healthCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		health, err := client.Health(healthCtx)
		cancel()
		if err != nil {
			return fmt.Errorf("ACE-Step service health check at %s failed: %w (did you start services/acestep/server.py? See services/acestep/README.md)",
				opts.serviceURL, err)
		}
		if !health.Loaded {
			return fmt.Errorf("ACE-Step service at %s reachable but not loaded (backend=%s, mock=%v, error=%q)",
				opts.serviceURL, health.Backend, health.MockMode, health.Error)
		}
		fmt.Fprintf(os.Stderr, "termus: connected to ACE-Step service at %s (backend=%s, model=%s, mock=%v, load_time=%.1fs)\n",
			opts.serviceURL, health.Backend, health.ModelName, health.MockMode, health.LoadTimeSeconds)
	}

	baseSpec, err := acestep.CompileV3(file)
	if err != nil {
		return fmt.Errorf("compile v3: %w", err)
	}
	prod := &renderingProducer{
		client:    client,
		baseSpec:  baseSpec,
		outputDir: opts.outputDir,
		trackName: acestepTrackNameFromPath(opts.trackPath),
	}

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
	go acestepStatusHeartbeat(streamCtx, s, statusDone)

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

// ensureACEStepReady boots a Manager, streams its status events to stderr,
// and returns once a Client is ready.
func ensureACEStepReady(ctx context.Context, opts acestepOptions) (*acestep.Manager, *acestep.Client, error) {
	installer := acestep.NewInstaller(opts.serviceDir, nil)
	if !installer.IsInstalled() {
		size := installer.EstimatedSize()
		fmt.Fprintf(os.Stderr, "termus: ACE-Step engine not installed. Setup will download:\n")
		fmt.Fprintf(os.Stderr, "  Python 3.11           ~%s\n", humanBytes(size.Python))
		fmt.Fprintf(os.Stderr, "  Dependencies          ~%s\n", humanBytes(size.Deps))
		fmt.Fprintf(os.Stderr, "  ACE-Step model        ~%s\n", humanBytes(size.Model))
		fmt.Fprintf(os.Stderr, "  Total                 ~%s\n", humanBytes(size.Total()))
		if !confirmInstall() {
			return nil, nil, errors.New("ACE-Step install declined")
		}
	}
	// Wire an installer that emits to stderr.
	evCh := make(chan acestep.InstallEvent, 32)
	installer.Events = evCh
	stCh := make(chan acestep.StatusEvent, 16)
	mgr := &acestep.Manager{
		Installer: installer,
		Port:      opts.port,
		Logger:    os.Stderr,
	}
	go func() {
		// Lightweight event drain: print phase changes to stderr.
		var lastPhase string
		for ev := range evCh {
			if ev.Phase != lastPhase {
				fmt.Fprintf(os.Stderr, "preparing AI engine: %s: %s\n", ev.Phase, ev.Message)
				lastPhase = ev.Phase
				continue
			}
			// Same phase, just update the line if it's short enough.
			if len(ev.Message) < 80 {
				fmt.Fprintf(os.Stderr, "  %s\n", ev.Message)
			}
		}
	}()
	go func() {
		var lastPhase string
		for ev := range stCh {
			if ev.Phase == lastPhase {
				continue
			}
			lastPhase = ev.Phase
			if ev.Err != nil {
				fmt.Fprintf(os.Stderr, "preparing AI engine: %s: ERROR: %v\n", ev.Phase, ev.Err)
				continue
			}
			fmt.Fprintf(os.Stderr, "preparing AI engine: %s: %s\n", ev.Phase, ev.Message)
		}
	}()
	client, err := mgr.EnsureReady(ctx, stCh)
	close(evCh)
	close(stCh)
	if err != nil {
		return nil, nil, fmt.Errorf("ACE-Step bootstrap failed: %w", err)
	}
	return mgr, client, nil
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
// are variations.
type renderingProducer struct {
	client    *acestep.Client
	baseSpec  acestep.RenderSpec
	outputDir string
	trackName string
	count     atomic.Int64
}

func (p *renderingProducer) Produce(ctx context.Context, seq int) ([]byte, error) {
	spec := p.baseSpec
	if spec.Seed >= 0 {
		spec.Seed += int64(seq)
	}
	t0 := time.Now()
	data, err := p.client.Render(ctx, spec)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "termus: rendered seq=%d (%d bytes, %.1fs)\n",
		seq, len(data), time.Since(t0).Seconds())

	if p.outputDir != "" {
		path := filepath.Join(p.outputDir, fmt.Sprintf("%s-%03d.wav", p.trackName, seq))
		if err := os.WriteFile(path, data, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "termus: warning: write %s: %v\n", path, err)
		}
	}
	p.count.Add(1)
	return data, nil
}

func acestepStatusHeartbeat(ctx context.Context, s *audio.Streamer, done chan<- struct{}) {
	defer close(done)
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			st := s.Status()
			fmt.Fprintf(os.Stderr, "termus: status: playing=%v seq=%d queue=%d\n",
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

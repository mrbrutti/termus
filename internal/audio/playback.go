package audio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"

	"github.com/mrbrutti/termus/internal/acestep"
	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
	"github.com/mrbrutti/termus/internal/synth"
)

// ProgramSender is the minimal surface tea.Program exposes for asynchronous
// message delivery. The Playback facade uses it to stream engine-switch
// progress (loader text, percent, ready) into the TUI without taking a hard
// dependency on bubbletea inside the audio package.
type ProgramSender interface {
	Send(msg any)
}

// nopSender is the no-op default when callers don't wire a sink. It exists so
// Playback can be exercised in unit tests without a real *tea.Program.
type nopSender struct{}

func (nopSender) Send(any) {}

// MessageBus is the catalogue of message-emit closures the Playback facade
// needs. We don't import internal/tui here (that would invert the dependency
// tree); instead, callers construct closures that wrap their own tui.*
// message types. Each field is optional: nil disables that channel.
//
// In practice, cmd/termus/main.go wires these to thin functions that build
// the matching tui.* messages (StartupLoadMsg, ACEStepStatusMsg, etc.).
type MessageBus struct {
	// StartupLoad is invoked for SF2-side loader updates ("preparing
	// soundfonts", "compiling authored arrangement", etc.). detail / percent
	// are 0/empty when not yet known.
	StartupLoad func(title, detail string, percent float64, done bool)

	// ACEInstallProgress is invoked for installer phases (downloading model,
	// installing python, etc.). Surfaced inside the same loading overlay
	// the TUI already uses for SP23 ACE-Step boot.
	ACEInstallProgress func(phase, title, detail string, percent float64, err error)

	// ACEStatus is invoked for daemon lifecycle phases (starting-daemon,
	// loading-model, ready). Once `ready` lands, the TUI flips to the
	// now-playing view; the streamer takes over from there.
	ACEStatus func(phase, title, detail string, percent float64, err error)

	// ACERendering is invoked for per-track render progress, including
	// "composing first track..." before playback begins.
	ACERendering func(seq int, detail string, done bool, err error)

	// ACEReady is invoked once the first AI track is queued and playback
	// can begin. Dismisses the loader overlay.
	ACEReady func(detail string)

	// BackendState is invoked for live-speaker startup state transitions
	// (starting / ready / hung / init failed). Mirrors the existing
	// audio.BackendState surface so callers don't need a separate channel.
	BackendState func(state BackendState)
}

func (b *MessageBus) sendStartupLoad(title, detail string, percent float64, done bool) {
	if b == nil || b.StartupLoad == nil {
		return
	}
	b.StartupLoad(title, detail, percent, done)
}

func (b *MessageBus) sendACEInstall(phase, title, detail string, percent float64, err error) {
	if b == nil || b.ACEInstallProgress == nil {
		return
	}
	b.ACEInstallProgress(phase, title, detail, percent, err)
}

func (b *MessageBus) sendACEStatus(phase, title, detail string, percent float64, err error) {
	if b == nil || b.ACEStatus == nil {
		return
	}
	b.ACEStatus(phase, title, detail, percent, err)
}

func (b *MessageBus) sendACERendering(seq int, detail string, done bool, err error) {
	if b == nil || b.ACERendering == nil {
		return
	}
	b.ACERendering(seq, detail, done, err)
}

func (b *MessageBus) sendACEReady(detail string) {
	if b == nil || b.ACEReady == nil {
		return
	}
	b.ACEReady(detail)
}

func (b *MessageBus) sendBackendState(state BackendState) {
	if b == nil || b.BackendState == nil {
		return
	}
	b.BackendState(state)
}

// SpeakerController is the narrow surface Playback needs from the global
// beep/speaker package, so tests can substitute a hermetic stub. The real
// implementation wraps speaker.Init / speaker.Play / speaker.Clear.
type SpeakerController interface {
	// Init prepares the OS speaker for playback at the given sample rate
	// and buffer size. Returns nil if already initialised at this rate.
	Init(sr beep.SampleRate, bufferSize int) error
	// Play attaches a streamer to the speaker's global mixer.
	Play(s beep.Streamer)
	// Clear removes all currently playing streamers from the mixer.
	Clear()
}

type beepSpeaker struct {
	mu          sync.Mutex
	initialised bool
	sampleRate  beep.SampleRate
}

func (b *beepSpeaker) Init(sr beep.SampleRate, bufferSize int) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.initialised && b.sampleRate == sr {
		return nil
	}
	if err := speaker.Init(sr, bufferSize); err != nil {
		return err
	}
	b.initialised = true
	b.sampleRate = sr
	return nil
}

func (b *beepSpeaker) Play(s beep.Streamer) { speaker.Play(s) }
func (b *beepSpeaker) Clear()               { speaker.Clear() }

// DefaultSpeaker returns the canonical SpeakerController backed by
// gopxl/beep/v2/speaker. A single instance is sufficient per process; the
// speaker package is itself a singleton.
func DefaultSpeaker() SpeakerController { return &beepSpeaker{} }

// ACEStepFactory builds the HTTP client and Manager pair for an ACE-Step
// session. The factory takes the install/status sink so the manager's
// bootstrap progress can be relayed to the TUI loader.
//
// Implementations live in cmd/termus/acestep_run.go (acquireACEStepClient
// adapted to this signature).
type ACEStepFactory func(ctx context.Context, sink ACEStepStatusSink) (*acestep.Client, *acestep.Manager, error)

// ACEStepStatusSink is the narrow callback surface the manager bootstrap uses
// to stream install + status progress. Implemented by Playback itself; the
// factory does not need to know about tea.Program.
type ACEStepStatusSink interface {
	OnInstallProgress(phase, title, detail string, percent float64, err error)
	OnStatus(phase, title, detail string, percent float64, err error)
}

// ACEStepProducerFactory builds the audio producer for the streamer given a
// fresh client and a sink the producer can stream rendering progress to.
// Lives in cmd/termus/acestep_run.go (newRenderingProducer).
type ACEStepProducerFactory func(client *acestep.Client, sink ACEStepRenderSink) AudioProducer

// ACEStepRenderSink relays per-track rendering progress.
type ACEStepRenderSink interface {
	OnRendering(seq int, detail string, done bool, err error)
	OnFirstReady(detail string)
}

// ACEStepSwitchOptions bundles the per-track ACE-Step parameters that aren't
// shared with the persistent session state.
type ACEStepSwitchOptions struct {
	CrossfadeSec float64
	QueueDepth   int
	MaxTracks    int
	ProducerFn   ACEStepProducerFactory
	Title        string
}

// Playback is the TUI-friendly audio facade. It owns:
//
//   - A PlaybackSession (the lifecycle logic shared with unit tests).
//   - A SpeakerController for OS-level handoff between engines.
//   - The current SF2 *Root (when SF2 is active) so Commander calls forward
//     correctly.
//   - The current ACE-Step *Streamer (via session.CurrentStreamer).
//   - The session-owned *acestep.Manager (survives engine switches).
//
// Concurrency: Switch* methods are serialised internally. Commander methods
// race-safely forward to whichever engine is active at the time of the call.
type Playback struct {
	mu sync.Mutex

	session    *PlaybackSession
	speaker    SpeakerController
	bus        *MessageBus
	sender     ProgramSender
	logger     io.Writer
	aceFactory ACEStepFactory

	// scopeRing is shared across SF2 and ACE-Step engines so the TUI
	// visualizer keeps moving across hot-switches.
	scopeRing *scope.Ring

	// SF2-side state. root is the streamer currently being fed to the
	// speaker via activeBackend.
	root          *Root
	activeBackend *LiveBackend

	// currentEngine tracks the engine the speaker is currently fed by.
	currentEngine EngineKind
	launched      bool

	// initialVol seeds new SF2 Roots when an explicit volume isn't supplied
	// to a Switch call.
	initialVol int

	// defaultSampleRate is used by the cross-engine fallback in
	// SwapAlgorithmFade so it can stand up a new SF2 backend without the
	// caller threading a beep.SampleRate through every call site. Set by
	// StartSF2 / SetDefaultSampleRate.
	defaultSampleRate beep.SampleRate

	// firstReadyCh is closed when the ACE-Step producer's first render
	// completes, signaling the loader's progress ticker to exit. Set per
	// SwitchToACEStep call; nil between switches.
	firstReadyCh chan struct{}

	// aceRecorder is the shared WAV recorder for the ACE-Step engine. It
	// lives across hot-switches so press-r works regardless of which track
	// is currently playing. SF2 uses its own Root-internal WAV tap.
	aceRecorder *Recorder
}

// NewPlayback constructs a Playback. speaker may be nil (DefaultSpeaker is
// used). sender may be nil (a no-op sender absorbs all messages). bus may be
// nil (no TUI updates are emitted).
func NewPlayback(speakerCtrl SpeakerController, sender ProgramSender, bus *MessageBus, factory ACEStepFactory, logger io.Writer, initialVol int) *Playback {
	if speakerCtrl == nil {
		speakerCtrl = DefaultSpeaker()
	}
	if sender == nil {
		sender = nopSender{}
	}
	if logger == nil {
		logger = io.Discard
	}
	return &Playback{
		session:     NewPlaybackSession(logger),
		speaker:     speakerCtrl,
		bus:         bus,
		sender:      sender,
		logger:      logger,
		aceFactory:  factory,
		initialVol:  volumeOrDefault(initialVol),
		aceRecorder: NewRecorder(synth.SampleRate),
	}
}

// AttachScopeRing wires the visualizer ring the SF2 path feeds via
// audio.Root.Stream and the ACE-Step path feeds via the streamer's scope tap.
// The same pointer must be passed to both engines so cross-engine hot-switch
// keeps the waveform display alive.
func (p *Playback) AttachScopeRing(ring *scope.Ring) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.scopeRing = ring
}

// SetMessageBus installs (or replaces) the message bus the Playback uses to
// report progress events. Callers typically build closures that translate
// audio-side events into the equivalent tui.* messages.
func (p *Playback) SetMessageBus(bus *MessageBus) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.bus = bus
}

// SetACEStepFactory installs (or replaces) the factory Playback uses to
// bootstrap an ACE-Step daemon + client on the first SwitchToACEStep call.
// Subsequent calls reuse the session's stored manager.
func (p *Playback) SetACEStepFactory(factory ACEStepFactory) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.aceFactory = factory
}

// SetDefaultSampleRate records the SF2 sample rate so the implicit
// ACE-Step -> SF2 fallback inside SwapAlgorithmFade can build a fresh
// LiveBackend without the caller threading a beep.SampleRate.
func (p *Playback) SetDefaultSampleRate(sr beep.SampleRate) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.defaultSampleRate = sr
}

// CurrentEngine returns the engine whose audio is currently feeding the
// speaker. Before the first Switch this is EngineSF2 (the zero value); use
// Launched() to discriminate.
func (p *Playback) CurrentEngine() EngineKind {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.currentEngine
}

// Launched reports whether any engine has been successfully started.
func (p *Playback) Launched() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.launched
}

// Session exposes the underlying PlaybackSession primarily for tests.
func (p *Playback) Session() *PlaybackSession { return p.session }

// Manager returns the session-owned ACE-Step manager (nil until the first
// ACE-Step switch). Caller must not Shutdown() — Playback.Stop does that.
func (p *Playback) Manager() *acestep.Manager { return p.session.Manager() }

// ActiveBackend returns the LiveBackend currently wrapping the SF2 Root, or
// nil when the active engine is ACE-Step (or the engine hasn't started yet).
// Used by the TUI's audio-control retry / render-only hooks.
func (p *Playback) ActiveBackend() *LiveBackend {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.activeBackend
}

// CurrentRoot returns the active SF2 Root, or nil when ACE-Step is active.
// Used primarily by tests; the Commander interface is the production access
// point.
func (p *Playback) CurrentRoot() *Root {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.root
}

// ----- Commander forwarding ------------------------------------------------

// SetVolume routes to the active engine. For SF2 this updates the Root; for
// ACE-Step it's stored so the next SF2 takeover starts at the right level
// (the Streamer has no global gain control yet).
func (p *Playback) SetVolume(pct int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.initialVol = volumeOrDefault(pct)
	if p.root != nil {
		p.root.SetVolume(pct)
	}
}

// TogglePause routes to the active engine. Streamer doesn't yet expose pause
// (skip is the closest analog) so ACE-Step pause is a no-op with a debug log.
func (p *Playback) TogglePause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.root != nil {
		p.root.TogglePause()
		return
	}
	fmt.Fprintln(p.logger, "playback: TogglePause: ignored — ACE-Step streamer has no pause path yet")
}

// ToggleRecord routes to whichever engine is currently driving the speaker.
// SF2 uses Root's per-stream WAV tap (filename includes the seed). ACE-Step
// uses the shared Recorder fed by the streamer's record tap (filename is
// tagged "ace" so it's distinguishable from SF2 captures). Both produce
// WAVs at synth.SampleRate stereo in the current working directory.
func (p *Playback) ToggleRecord() (string, error) {
	p.mu.Lock()
	engine := p.currentEngine
	root := p.root
	rec := p.aceRecorder
	p.mu.Unlock()
	switch engine {
	case EngineSF2:
		if root == nil {
			return "", errors.New("playback: SF2 engine not ready")
		}
		return root.ToggleRecord()
	case EngineACEStep:
		if rec == nil {
			return "", errors.New("playback: ACE-Step recorder not initialised")
		}
		active, _ := rec.Active()
		if active {
			return "", rec.ToggleStop()
		}
		return rec.ToggleStart("ace")
	default:
		return "", errors.New("playback: no active engine")
	}
}

// DebugStatus returns the SF2 Root's status when active; otherwise a zero
// snapshot. The TUI uses this for the debug overlay; it's only ever rendered
// for SF2 playback.
func (p *Playback) DebugStatus() gen.DebugStatus {
	p.mu.Lock()
	root := p.root
	p.mu.Unlock()
	if root == nil {
		return gen.DebugStatus{}
	}
	return root.DebugStatus()
}

// SwapAlgorithm hot-swaps the SF2 algorithm with the default fade. Called by
// the TUI's [n]/[p] cycle, playlist auto-advance, etc. When the active engine
// is ACE-Step the request is dropped with a log line — SF2 algorithms have no
// meaning to the streamer.
func (p *Playback) SwapAlgorithm(algo gen.Algorithm) {
	p.mu.Lock()
	root := p.root
	p.mu.Unlock()
	if root == nil {
		fmt.Fprintln(p.logger, "playback: SwapAlgorithm: ignored — current engine is not SF2")
		return
	}
	root.SwapAlgorithm(algo)
}

// SwapAlgorithmFade is like SwapAlgorithm but with a caller-specified fade
// length.
//
// SP26: when ACE-Step is the current engine, the call drops through to
// SwitchToSF2 with default settings so the TUI's existing
// TrackLoadResultMsg path "just works" for SF2 tracks picked from the
// browser while the AI engine is playing. The fadeFrames argument is
// ignored on the cross-engine path (the speaker handoff is itself a hard
// cut already).
func (p *Playback) SwapAlgorithmFade(algo gen.Algorithm, fadeFrames int) {
	p.mu.Lock()
	root := p.root
	defaultSR := p.defaultSampleRate
	p.mu.Unlock()
	if root != nil {
		root.SwapAlgorithmFade(algo, fadeFrames)
		return
	}
	// Cross-engine path: ACE-Step is currently playing; the caller wants
	// SF2. Trigger a switch and let SwitchToSF2 stand up a fresh Root.
	if defaultSR == 0 {
		defaultSR = beep.SampleRate(44100)
	}
	if err := p.SwitchToSF2(context.Background(), algo, 0, defaultSR, ""); err != nil {
		fmt.Fprintln(p.logger, "playback: implicit ACE-Step -> SF2 switch failed:", err)
	}
}

// ----- Engine switching ----------------------------------------------------

// StartSF2 is the first-time SF2 takeover used at process startup. It builds
// a new Root, hands the speaker over to it, and stores the resulting
// LiveBackend so the TUI's audio-control hooks (retry, render-only) work.
//
// Subsequent SF2 switches (algorithm cycling, track selection of another SF2
// .tm) should use the existing Commander.SwapAlgorithm* path rather than
// StartSF2 — those don't tear the speaker down.
func (p *Playback) StartSF2(ctx context.Context, algo gen.Algorithm, seed int64, sr beep.SampleRate) (*LiveBackend, error) {
	if algo == nil {
		return nil, errors.New("playback: StartSF2 requires non-nil algorithm")
	}
	p.mu.Lock()
	if p.activeBackend != nil {
		backend := p.activeBackend
		p.mu.Unlock()
		return backend, nil // already running
	}
	root := p.buildSF2RootLocked(algo, seed)
	p.mu.Unlock()

	if err := p.session.Switch(ctx, SwitchRequest{
		Engine:    EngineSF2,
		Algo:      algo,
		ScopeRing: p.scopeRing,
		Volume:    p.initialVol,
		Seed:      seed,
	}, nil); err != nil {
		return nil, fmt.Errorf("playback: session switch: %w", err)
	}

	backend := startLiveBackendWithController(root, sr, sr.N(time.Second/20), 3*time.Second, p.speaker)

	p.mu.Lock()
	p.activeBackend = backend
	p.root = root
	p.currentEngine = EngineSF2
	p.launched = true
	p.defaultSampleRate = sr
	p.mu.Unlock()

	go p.relayBackend(backend)
	return backend, nil
}

// buildSF2RootLocked constructs a new Root. The lock must already be held.
func (p *Playback) buildSF2RootLocked(algo gen.Algorithm, seed int64) *Root {
	root := NewRoot(algo, p.scopeRing)
	root.SetSeed(seed)
	root.SetVolume(p.initialVol)
	return root
}

// relayBackend pumps a LiveBackend's BackendState channel onto the message
// bus. Runs until the channel closes (LiveBackend.Close).
func (p *Playback) relayBackend(backend *LiveBackend) {
	for state := range backend.States() {
		p.currentBus().sendBackendState(state)
	}
}

// SwitchToSF2 is the hot-switch operation for picking an SF2 track from the
// browser while the SF2 engine is already active (no-op speaker teardown), or
// while the ACE-Step engine is active (full teardown + speaker handoff).
//
// algo is the freshly-built SF2 algorithm; seed feeds the Root for recording
// filenames. Emits StartupLoad messages around the speaker handoff so the TUI
// can show a loader while the algorithm warms up.
func (p *Playback) SwitchToSF2(ctx context.Context, algo gen.Algorithm, seed int64, sr beep.SampleRate, title string) error {
	if algo == nil {
		return errors.New("playback: SwitchToSF2 requires non-nil algorithm")
	}

	p.mu.Lock()
	currentEngine := p.currentEngine
	root := p.root
	p.mu.Unlock()

	if currentEngine == EngineSF2 && root != nil {
		// SF2 -> SF2: the cheapest path. Hot-swap the algorithm on the
		// existing Root — no speaker churn, no loader needed.
		root.SwapAlgorithmFade(algo, 8820)
		return nil
	}

	// Cross-engine (ACE-Step -> SF2) or recovery (no root) path: tear down
	// the previous engine via the session, then bring the speaker back up
	// with a fresh Root.
	p.currentBus().sendStartupLoad(title, "switching to procedural engine", 0.2, false)

	if err := p.session.Switch(ctx, SwitchRequest{
		Engine:    EngineSF2,
		Algo:      algo,
		ScopeRing: p.scopeRing,
		Volume:    p.initialVol,
		Seed:      seed,
	}, nil); err != nil {
		return fmt.Errorf("playback: session switch: %w", err)
	}

	// Speaker handoff: clear the global mixer (which held the now-stopped
	// streamer's audio), then re-init at the SF2 sample rate and Play the
	// new Root via a fresh LiveBackend.
	p.speaker.Clear()
	if err := p.speaker.Init(sr, sr.N(time.Second/20)); err != nil {
		return fmt.Errorf("playback: speaker.Init: %w", err)
	}

	p.mu.Lock()
	newRoot := p.buildSF2RootLocked(algo, seed)
	if p.activeBackend != nil {
		old := p.activeBackend
		go old.Close()
	}
	p.root = newRoot
	p.currentEngine = EngineSF2
	p.launched = true
	p.defaultSampleRate = sr
	p.mu.Unlock()

	backend := startLiveBackendWithController(newRoot, sr, sr.N(time.Second/20), 3*time.Second, p.speaker)
	p.mu.Lock()
	p.activeBackend = backend
	p.mu.Unlock()

	go p.relayBackend(backend)

	p.currentBus().sendStartupLoad(title, "starting audio", 1.0, true)
	return nil
}

// SwitchToACEStep tears down the current engine (SF2 or ACE-Step) and starts a
// new ACE-Step session. If the daemon is already running it reuses it; the
// first call paid the bootstrap cost. Emits ACE-Step* messages onto the bus
// throughout the boot sequence.
//
// This call returns once the streamer is *started*. The first track is
// produced asynchronously; the bus's ACEReady fires when it lands.
func (p *Playback) SwitchToACEStep(ctx context.Context, opts ACEStepSwitchOptions) error {
	p.mu.Lock()
	factory := p.aceFactory
	p.mu.Unlock()
	if factory == nil {
		return errors.New("playback: SwitchToACEStep called without ACEStepFactory")
	}
	if opts.ProducerFn == nil {
		return errors.New("playback: SwitchToACEStep requires ProducerFn")
	}
	// Bootstrap the manager (or reuse the warm one) outside the lock so
	// install progress messages can flow without contention.
	statusSink := &playbackStatusSink{p: p}
	client, manager, err := p.acquireClient(ctx, statusSink)
	if err != nil {
		return err
	}

	// Wire the manager's OnProgress to the status sink so RENDER_PROGRESS
	// lines from the daemon emit real percent + description into the TUI.
	// We map the model's 0..1 progress onto our loader's 0.5..0.95 range so
	// the bar advances from the "first track" floor (set just below) up
	// toward 95%, then ACEReady snaps it to 100%.
	if manager != nil {
		manager.OnProgress = func(modelPercent float64, detail string) {
			// 0.5 + modelPercent * 0.45 = 0.5..0.95
			mapped := 0.5 + modelPercent*0.45
			d := detail
			if d == "" {
				d = "composing first track…"
			}
			statusSink.OnStatus("generating-first-track", "Generating Music", d, mapped, nil)
		}
	}

	// Tear down the SF2 backend first (release the speaker) so the streamer
	// has a clean device.
	p.mu.Lock()
	oldBackend := p.activeBackend
	p.activeBackend = nil
	p.root = nil
	p.mu.Unlock()
	// Tear down the SF2 backend SYNCHRONOUSLY. Doing this in a goroutine
	// (the previous behavior) leaked into the next engine's setup: the
	// streamer's speakerSink would call speaker.Init while the SF2 audio
	// loop was still pushing samples, which on macOS oto/CoreAudio races
	// the device handle and silently produces no audio. Waiting here adds
	// at most a few hundred ms; the user is already on a multi-second
	// loader so this is invisible.
	if oldBackend != nil {
		oldBackend.Close()
	}
	p.speaker.Clear()

	renderSink := &playbackRenderSink{p: p}
	prod := opts.ProducerFn(client, renderSink)
	// Audio-path debug log. Streamer/sink errors are otherwise swallowed
	// (they don't reach stderr because that would smear the TUI). Tail
	// /tmp/termus-audio.log to see init/play errors during hot-switch.
	debugLog := openAudioDebugLog()
	streamCfg := StreamerConfig{
		Producer:     prod,
		QueueDepth:   opts.QueueDepth,
		CrossfadeSec: opts.CrossfadeSec,
		MaxTracks:    opts.MaxTracks,
		Logger:       debugLog,
		// Route audio through Playback's SpeakerController so there is one
		// owner of the global beep.speaker. The previous default (a fresh
		// internal speakerSink) raced Playback's controller across the
		// SF2→ACE-Step hot-switch on macOS.
		Sink: NewControllerSink(p.speaker),
	}
	if p.scopeRing != nil {
		streamCfg.ScopeSink = p.scopeRing
	}
	if p.aceRecorder != nil {
		streamCfg.RecordSink = p.aceRecorder
	}
	// The streamer's producer fires off an HTTP /render call on its first
	// loop iteration; on M-series that takes anywhere from ~15s (warm) to
	// 60-90s (cold or complex prompts). Surface a continuously-updating
	// status with elapsed time so the loader shows real progress instead
	// of sitting silently. ACEReady (from the producer's first successful
	// render) dismisses the loader entirely.
	statusSink.OnStatus(
		"generating-first-track",
		"Generating Music",
		"composing first track…",
		0.5,
		nil,
	)
	if err := p.session.Switch(ctx, SwitchRequest{
		Engine:    EngineACEStep,
		Manager:   manager,
		StreamCfg: streamCfg,
		Volume:    p.initialVol,
	}, nil); err != nil {
		return fmt.Errorf("playback: session switch: %w", err)
	}

	// Poll the daemon's GET /progress endpoint while the first render is
	// in flight. Maps the model's 0..1 onto our loader's 0.5..0.95 range
	// so the bar visibly advances during diffusion + VAE decode.
	//
	// IMPORTANT: this goroutine MUST NOT be tied to the caller's ctx,
	// which is typically a per-Cmd context from bubbletea that cancels
	// the moment SwitchToACEStep returns. Use a fresh background ctx
	// that the firstReadyCh / Shutdown path can cancel. Each Progress()
	// call gets its own short timeout so a slow daemon doesn't starve
	// the tick rate.
	firstReadyCh := make(chan struct{})
	p.mu.Lock()
	p.firstReadyCh = firstReadyCh
	p.mu.Unlock()
	go func() {
		start := time.Now()
		tick := time.NewTicker(750 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-firstReadyCh:
				return
			case <-tick.C:
				// Per-call ctx with a short deadline. Background-rooted
				// so an upstream cancel doesn't kill the poller; ours is
				// firstReadyCh.
				pctx, pcancel := context.WithTimeout(context.Background(), 2*time.Second)
				pr, err := client.Progress(pctx)
				pcancel()
				// Re-check firstReadyCh *after* the (potentially slow)
				// Progress call so we don't emit a stale 50% status that
				// arrives at the TUI after ACEStepReadyMsg and pins the
				// loader open while audio is already playing.
				select {
				case <-firstReadyCh:
					return
				default:
				}
				elapsed := time.Since(start)
				if err != nil || !pr.Active {
					// Fall back to elapsed-time tick when /progress isn't
					// reporting active (very early, very late, or older
					// server.py without the endpoint).
					detail := fmt.Sprintf("composing first track… %s elapsed", formatElapsed(elapsed))
					statusSink.OnStatus("generating-first-track", "Generating Music", detail, 0.5, nil)
					continue
				}
				mapped := 0.5 + pr.Percent*0.45
				detail := pr.Detail
				if detail == "" {
					detail = "generating…"
				}
				detail = fmt.Sprintf("%s (%s elapsed)", detail, formatElapsed(elapsed))
				statusSink.OnStatus("generating-first-track", "Generating Music", detail, mapped, nil)
			}
		}
	}()

	p.mu.Lock()
	p.currentEngine = EngineACEStep
	p.launched = true
	p.mu.Unlock()
	return nil
}

// openAudioDebugLog returns a writer for the audio-path debug log. Errors
// in the streamer/sink chain would otherwise vanish into io.Discard (we
// can't write to stderr without smearing the TUI alt-screen). Tail
// /tmp/termus-audio.log during a hot-switch to see what's actually
// happening. Falls back to io.Discard if the file can't be opened.
func openAudioDebugLog() io.Writer {
	path := "/tmp/termus-audio.log"
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return io.Discard
	}
	fmt.Fprintf(f, "\n=== playback session opened %s ===\n", time.Now().Format(time.RFC3339))
	return f
}

// formatElapsed renders an elapsed Duration as "M:SS" — the same format the
// progress ticker uses to update the AI-engine loader.
func formatElapsed(d time.Duration) string {
	secs := int(d.Seconds())
	return fmt.Sprintf("%d:%02d", secs/60, secs%60)
}

// acquireClient runs the manager bootstrap (or reuses a warm manager) outside
// the lock. Subsequent ACE-Step switches reuse the session's stored manager
// rather than re-running EnsureReady.
func (p *Playback) acquireClient(ctx context.Context, sink ACEStepStatusSink) (*acestep.Client, *acestep.Manager, error) {
	if mgr := p.session.Manager(); mgr != nil {
		sink.OnStatus("ready", "AI engine ready", "reusing existing daemon", 1.0, nil)
		return clientForManager(mgr), mgr, nil
	}
	p.mu.Lock()
	factory := p.aceFactory
	p.mu.Unlock()
	if factory == nil {
		return nil, nil, errors.New("playback: no ACEStepFactory configured")
	}
	return factory(ctx, sink)
}

// clientForManager constructs a fresh Client pointing at the manager's port.
func clientForManager(mgr *acestep.Manager) *acestep.Client {
	return acestep.NewClient(managerBaseURL(mgr), 5*time.Minute)
}

// managerBaseURL extracts the manager's HTTP base URL.
func managerBaseURL(mgr *acestep.Manager) string {
	port := mgr.Port
	if port == 0 {
		port = 7790
	}
	return fmt.Sprintf("http://localhost:%d", port)
}

// Stop tears down whichever engine is running and shuts down the session.
// Idempotent.
func (p *Playback) Stop(ctx context.Context) error {
	p.mu.Lock()
	old := p.activeBackend
	p.activeBackend = nil
	p.root = nil
	p.mu.Unlock()
	if old != nil {
		go old.Close()
	}
	p.speaker.Clear()
	return p.session.Stop(ctx)
}

// currentBus returns the bus under lock so the closures stay race-safe even
// when SetMessageBus is called concurrently with a switch.
func (p *Playback) currentBus() *MessageBus {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.bus
}

// playbackStatusSink implements ACEStepStatusSink by relaying to the message
// bus. Used by SwitchToACEStep so the factory doesn't take a hard *tea.Program.
type playbackStatusSink struct{ p *Playback }

func (s *playbackStatusSink) OnInstallProgress(phase, title, detail string, percent float64, err error) {
	s.p.currentBus().sendACEInstall(phase, title, detail, percent, err)
}

func (s *playbackStatusSink) OnStatus(phase, title, detail string, percent float64, err error) {
	s.p.currentBus().sendACEStatus(phase, title, detail, percent, err)
}

// playbackRenderSink implements ACEStepRenderSink by relaying to the message
// bus.
type playbackRenderSink struct{ p *Playback }

func (s *playbackRenderSink) OnRendering(seq int, detail string, done bool, err error) {
	s.p.currentBus().sendACERendering(seq, detail, done, err)
}

func (s *playbackRenderSink) OnFirstReady(detail string) {
	// Stop the progress ticker started in SwitchToACEStep.
	s.p.mu.Lock()
	ch := s.p.firstReadyCh
	s.p.firstReadyCh = nil
	s.p.mu.Unlock()
	if ch != nil {
		// close-once: another caller may race; recover from a double-close
		// panic (cheap and only ever triggers if there are concurrent first
		// renders, which shouldn't happen but Playback is shared mutable).
		defer func() { _ = recover() }()
		close(ch)
	}
	s.p.currentBus().sendACEReady(detail)
}

// Compile-time assertion: Playback satisfies Commander so callers can pass
// it where the SF2 Root used to go (tui.New takes audio.Commander).
var _ Commander = (*Playback)(nil)

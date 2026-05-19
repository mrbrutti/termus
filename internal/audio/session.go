package audio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/mrbrutti/termus/internal/acestep"
	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
)

// EngineKind selects which playback engine backs the session. SF2 is the
// procedural / SoundFont path (audio.Root); ACEStep is the AI streaming path
// (audio.Streamer + acestep.Manager).
type EngineKind int

const (
	// EngineSF2 is the in-process procedural / SoundFont playback path.
	EngineSF2 EngineKind = iota
	// EngineACEStep is the HTTP+queue ACE-Step playback path.
	EngineACEStep
)

// String makes EngineKind printable in logs and test failures.
func (k EngineKind) String() string {
	switch k {
	case EngineSF2:
		return "sf2"
	case EngineACEStep:
		return "acestep"
	default:
		return fmt.Sprintf("engine(%d)", int(k))
	}
}

// SwitchEvent is the streamed lifecycle event published by Switch as the new
// engine boots. Mirrors the shape of acestep.StatusEvent so callers can
// render progress in the TUI loader.
type SwitchEvent struct {
	Phase   string
	Message string
	Err     error
}

// SwitchRequest describes a target state for the session: which engine, and
// the engine-specific configuration.
//
// SF2 fields (Algo, ScopeRing) are consulted when Engine == EngineSF2.
// ACE-Step fields (Spec, Manager, Producer, ScopeSink, StreamCfg) are
// consulted when Engine == EngineACEStep. Volume is honored by both paths.
type SwitchRequest struct {
	Engine EngineKind

	// SF2 path:
	Algo      gen.Algorithm
	ScopeRing *scope.Ring // for the SF2 path; nil disables the scope tap.

	// ACE-Step path:
	Spec      acestep.RenderSpec
	Manager   *acestep.Manager
	Producer  AudioProducer  // injected by callers; tests can supply mocks
	ScopeSink ScopeSink      // mirror of the same *scope.Ring used elsewhere
	StreamCfg StreamerConfig // optional caller overrides (QueueDepth, Sink, etc.)

	// Shared:
	Volume int
	Seed   int64
}

// PlaybackSession owns the currently-playing engine. Switch() tears down the
// current engine and starts the new one; Stop() tears down whatever's
// running. The session holds a single acestep.Manager across the lifetime so
// switching SF2 -> ACE-Step -> SF2 -> ACE-Step doesn't pay daemon-bootstrap
// cost on the second AI switch.
//
// Concurrency: Switch and Stop are serialised by an internal mutex. The
// returned engine state can be inspected via CurrentEngine, CurrentRoot, and
// CurrentStreamer.
type PlaybackSession struct {
	mu sync.Mutex

	current EngineKind
	root    *Root     // active when current == EngineSF2
	stream  *Streamer // active when current == EngineACEStep

	// streamCancel cancels the per-stream context so Streamer.Stop unwinds.
	streamCancel context.CancelFunc

	// Daemon manager survives engine switches; only released by Stop().
	manager *acestep.Manager

	// Has Switch() ever been called? Used to discriminate "first launch"
	// from "hot switch" behavior in error messages.
	launched bool

	// logger receives one-line records of switch transitions for debug logs.
	logger io.Writer
}

// NewPlaybackSession constructs a fresh session in the "no engine running"
// state. The session is inert until the first Switch() call. logger may be
// nil (it defaults to io.Discard).
func NewPlaybackSession(logger io.Writer) *PlaybackSession {
	if logger == nil {
		logger = io.Discard
	}
	return &PlaybackSession{logger: logger}
}

// CurrentEngine returns the engine that is currently playing. Before the
// first Switch this is EngineSF2 (the zero value); callers should consult
// Launched() to discriminate that case.
func (s *PlaybackSession) CurrentEngine() EngineKind {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current
}

// Launched reports whether Switch has ever succeeded on this session.
func (s *PlaybackSession) Launched() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.launched
}

// CurrentRoot returns the active SF2 Root, or nil if the SF2 engine isn't the
// current one.
func (s *PlaybackSession) CurrentRoot() *Root {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current != EngineSF2 {
		return nil
	}
	return s.root
}

// CurrentStreamer returns the active ACE-Step Streamer, or nil if the
// ACE-Step engine isn't the current one.
func (s *PlaybackSession) CurrentStreamer() *Streamer {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current != EngineACEStep {
		return nil
	}
	return s.stream
}

// Manager returns the session-owned ACE-Step manager (nil until the first
// ACE-Step switch installs one). The session retains ownership; callers
// must not Shutdown() it directly — only the session does that, on Stop().
func (s *PlaybackSession) Manager() *acestep.Manager {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.manager
}

// Switch tears down the current engine and starts the new one described by
// req. Blocks until the new engine is ready to accept audio frames. Returns
// an error when:
//   - req is malformed for the chosen engine (e.g. SF2 with nil Algo)
//   - starting the new engine fails (e.g. the ACE-Step Streamer.Start returns
//     an error)
//
// events is optional; when non-nil, the channel receives one SwitchEvent for
// each meaningful phase change ("teardown-old", "starting-new", "ready").
// The channel is NOT closed by Switch; the caller owns its lifecycle.
//
// SF2 -> SF2: old Root is torn down; new Root is built. The session keeps
// the existing acestep.Manager alive across the switch (daemon stays warm).
//
// SF2 -> ACE-Step: old Root is torn down; the Streamer is started against
// req.Producer (or req.StreamCfg.Producer).
//
// ACE-Step -> SF2: Streamer is stopped (graceful); the manager remains
// alive — its daemon survives so the next switch back to ACE-Step is fast.
//
// ACE-Step -> ACE-Step: current Streamer is stopped; a new one is started
// with req.Spec / req.Producer. Daemon is reused.
func (s *PlaybackSession) Switch(ctx context.Context, req SwitchRequest, events chan<- SwitchEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := validateRequest(req); err != nil {
		return err
	}
	emit := func(phase, msg string, err error) {
		fmt.Fprintf(s.logger, "session: switch %s: %s (err=%v)\n", phase, msg, err)
		if events == nil {
			return
		}
		select {
		case events <- SwitchEvent{Phase: phase, Message: msg, Err: err}:
		default:
		}
	}

	// 1. Tear down whatever is running.
	if s.launched {
		emit("teardown-old", fmt.Sprintf("tearing down %s engine", s.current), nil)
		s.teardownLocked()
	}

	// 2. Adopt the new manager pointer when the request carries one. We
	// only ever swap to a manager pointer we don't already own — keeping
	// the existing one warm is the whole point of the session-level
	// lifecycle.
	if req.Manager != nil && s.manager == nil {
		s.manager = req.Manager
	}

	// 3. Bring up the new engine.
	emit("starting-new", fmt.Sprintf("starting %s engine", req.Engine), nil)
	switch req.Engine {
	case EngineSF2:
		s.root = NewRoot(req.Algo, req.ScopeRing)
		s.root.SetSeed(req.Seed)
		s.root.SetVolume(volumeOrDefault(req.Volume))
		s.current = EngineSF2
	case EngineACEStep:
		streamCfg := req.StreamCfg
		if streamCfg.Producer == nil {
			if req.Producer != nil {
				streamCfg.Producer = req.Producer
			} else {
				return fmt.Errorf("session: ACE-Step switch requires either Producer or StreamCfg.Producer")
			}
		}
		if streamCfg.ScopeSink == nil {
			streamCfg.ScopeSink = req.ScopeSink
		}
		streamer := NewStreamer(streamCfg)
		streamCtx, cancel := context.WithCancel(context.Background())
		if err := streamer.Start(streamCtx); err != nil {
			cancel()
			return fmt.Errorf("session: start ACE-Step streamer: %w", err)
		}
		s.stream = streamer
		s.streamCancel = cancel
		s.current = EngineACEStep
	default:
		return fmt.Errorf("session: unknown engine %v", req.Engine)
	}
	s.launched = true
	emit("ready", fmt.Sprintf("%s engine ready", req.Engine), nil)
	return nil
}

// Stop tears down whichever engine is currently running and shuts down the
// session-owned ACE-Step daemon (if one was ever started). Idempotent.
func (s *PlaybackSession) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.teardownLocked()
	if s.manager != nil {
		_ = s.manager.Shutdown(ctx)
		s.manager = nil
	}
	return nil
}

// teardownLocked stops whichever engine is current. Caller must hold s.mu.
// Does NOT touch s.manager (so the daemon survives SF2 ↔ ACE-Step swaps).
func (s *PlaybackSession) teardownLocked() {
	switch s.current {
	case EngineSF2:
		// Root has no explicit stop; the live backend drives it. The
		// caller wires the next Root via TUI SwapAlgorithm flow. We
		// just drop our pointer so it can be GC'd when the speaker
		// hands off.
		s.root = nil
	case EngineACEStep:
		if s.streamCancel != nil {
			s.streamCancel()
		}
		if s.stream != nil {
			s.stream.Stop()
		}
		s.stream = nil
		s.streamCancel = nil
	}
}

// validateRequest returns an error when req is missing fields required for
// the chosen engine.
func validateRequest(req SwitchRequest) error {
	switch req.Engine {
	case EngineSF2:
		if req.Algo == nil {
			return errors.New("session: SF2 switch requires non-nil Algo")
		}
	case EngineACEStep:
		// Producer-or-StreamCfg.Producer is checked at start time so
		// tests can pass them on the StreamerConfig directly.
		if req.Producer == nil && req.StreamCfg.Producer == nil {
			return errors.New("session: ACE-Step switch requires Producer")
		}
	default:
		return fmt.Errorf("session: unknown engine %v", req.Engine)
	}
	return nil
}

func volumeOrDefault(v int) int {
	if v <= 0 {
		return 70
	}
	if v > 100 {
		return 100
	}
	return v
}

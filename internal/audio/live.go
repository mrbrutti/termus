package audio

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

// BackendStateKind describes the live audio backend startup state.
type BackendStateKind string

const (
	BackendStateStarting        BackendStateKind = "starting"
	BackendStateReady           BackendStateKind = "ready"
	BackendStateNoDefaultDevice BackendStateKind = "no default device"
	BackendStateHung            BackendStateKind = "backend hung"
	BackendStateRenderOnly      BackendStateKind = "render-only"
	BackendStateInitFailed      BackendStateKind = "init failed"
)

// BackendState is a user-facing audio backend status update.
type BackendState struct {
	Kind   BackendStateKind
	Detail string
}

// StatusText returns a concise status line suitable for the TUI footer.
func (s BackendState) StatusText() string {
	switch s.Kind {
	case BackendStateStarting:
		return "audio: starting..."
	case BackendStateReady:
		return "audio: ready"
	case BackendStateNoDefaultDevice:
		return "audio: no default device; use --out file.wav"
	case BackendStateHung:
		return "audio: backend hung; use --out file.wav"
	case BackendStateRenderOnly:
		return "audio: render-only"
	case BackendStateInitFailed:
		if s.Detail == "" {
			return "audio: init failed"
		}
		return "audio: init failed; use --out file.wav"
	default:
		return "audio: unknown"
	}
}

// ClassifyInitError groups low-level backend errors into user-facing states.
func ClassifyInitError(err error) BackendStateKind {
	if err == nil {
		return BackendStateReady
	}
	text := strings.ToLower(err.Error())
	switch {
	case strings.Contains(text, "default-output"),
		strings.Contains(text, "default output"),
		strings.Contains(text, "default device"),
		strings.Contains(text, "no device"),
		strings.Contains(text, "no system object"):
		return BackendStateNoDefaultDevice
	default:
		return BackendStateInitFailed
	}
}

// LiveBackend reports live-speaker startup progress without blocking the UI.
type LiveBackend struct {
	states    chan BackendState
	ready     atomic.Bool
	closed    atomic.Bool
	attemptID atomic.Uint64
	closeFn   func()
	initFn    func() error
	startFn   func()
	timeout   time.Duration
	closeOnce sync.Once
	// closeMu serialises channel close vs. emit so the "send on closed
	// channel" race detector report goes away. Without it, watchAttempt
	// can race with a concurrent Close() and panic when the Stop / engine
	// switch path is exercised under -race.
	closeMu sync.Mutex
}

// StartLive starts the speaker in the background and reports state changes on
// the returned channel. The state stream begins with "starting", then usually
// transitions to "ready", "hung", or an init error classification.
func StartLive(root beep.Streamer, sr beep.SampleRate, bufferSize int, timeout time.Duration) *LiveBackend {
	return startLiveBackend(
		func() error { return speaker.Init(sr, bufferSize) },
		func() { speaker.Play(root) },
		func() {
			// Upstream beep/oto leaves the driver context alive even after
			// Close() and documents that programs usually don't need to call it.
			// Clearing the mixer is enough for quit, and avoids teardown paths
			// that can leave the next launch without working audio on macOS.
			speaker.Clear()
		},
		timeout,
	)
}

// startLiveBackendWithController is the seam used by Playback to spin up a
// LiveBackend whose audio side is driven by an injected SpeakerController.
// Tests pass a stub speaker; production callers use DefaultSpeaker() and the
// behaviour is identical to StartLive.
func startLiveBackendWithController(root beep.Streamer, sr beep.SampleRate, bufferSize int, timeout time.Duration, ctrl SpeakerController) *LiveBackend {
	if ctrl == nil {
		return StartLive(root, sr, bufferSize, timeout)
	}
	return startLiveBackend(
		func() error { return ctrl.Init(sr, bufferSize) },
		func() { ctrl.Play(root) },
		func() { ctrl.Clear() },
		timeout,
	)
}

func startLiveBackend(initFn func() error, startFn func(), closeFn func(), timeout time.Duration) *LiveBackend {
	b := &LiveBackend{
		states:  make(chan BackendState, 16),
		closeFn: closeFn,
		initFn:  initFn,
		startFn: startFn,
		timeout: timeout,
	}
	b.startAttempt()
	return b
}

// States returns the startup-state stream.
func (b *LiveBackend) States() <-chan BackendState { return b.states }

// Retry starts a fresh backend initialization attempt and emits a new
// "starting" state. Late results from previous hung attempts are ignored.
func (b *LiveBackend) Retry() {
	if b.closed.Load() {
		return
	}
	b.startAttempt()
}

// SetRenderOnly forces a user-facing render-only state without attempting
// live audio startup.
func (b *LiveBackend) SetRenderOnly() {
	if b.closed.Load() {
		return
	}
	b.ready.Store(false)
	b.emit(BackendState{Kind: BackendStateRenderOnly})
}

// Close stops live playback if it was successfully started.
func (b *LiveBackend) Close() {
	// closeMu protects against the race where a watchAttempt goroutine
	// is about to send on b.states while Close is closing the channel.
	// We hold the lock around the channel close *and* every emit so the
	// "channel closed" state is observable atomically.
	b.closed.Store(true)
	if !b.ready.Swap(false) {
		b.closeStatesLocked()
		return
	}
	if b.closeFn != nil {
		b.closeFn()
	}
	b.closeStatesLocked()
}

func (b *LiveBackend) closeStatesLocked() {
	b.closeMu.Lock()
	defer b.closeMu.Unlock()
	b.closeOnce.Do(func() { close(b.states) })
}

func (b *LiveBackend) emit(state BackendState) {
	b.closeMu.Lock()
	defer b.closeMu.Unlock()
	if b.closed.Load() {
		return
	}
	select {
	case b.states <- state:
	default:
	}
}

func (b *LiveBackend) startAttempt() {
	if b.initFn == nil || b.startFn == nil {
		return
	}
	id := b.attemptID.Add(1)
	b.ready.Store(false)
	b.emit(BackendState{Kind: BackendStateStarting})
	done := make(chan error, 1)
	go func(attempt uint64) {
		err := b.initFn()
		if b.closed.Load() || b.attemptID.Load() != attempt {
			return
		}
		if err == nil {
			b.startFn()
			b.ready.Store(true)
		}
		done <- err
	}(id)
	go b.watchAttempt(id, done, b.timeout)
}

func (b *LiveBackend) watchAttempt(id uint64, done <-chan error, timeout time.Duration) {
	timerC := (<-chan time.Time)(nil)
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		timerC = timer.C
	}

	for {
		select {
		case err := <-done:
			if b.closed.Load() || b.attemptID.Load() != id {
				return
			}
			if err != nil {
				b.emit(BackendState{
					Kind:   ClassifyInitError(err),
					Detail: err.Error(),
				})
			} else {
				b.emit(BackendState{Kind: BackendStateReady})
			}
			return
		case <-timerC:
			if b.closed.Load() || b.attemptID.Load() != id {
				return
			}
			b.emit(BackendState{Kind: BackendStateHung})
			timerC = nil
		}
	}
}

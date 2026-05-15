package audio

import (
	"strings"
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
	states  chan BackendState
	ready   atomic.Bool
	closeFn func()
}

// StartLive starts the speaker in the background and reports state changes on
// the returned channel. The state stream begins with "starting", then usually
// transitions to "ready", "hung", or an init error classification.
func StartLive(root beep.Streamer, sr beep.SampleRate, bufferSize int, timeout time.Duration) *LiveBackend {
	return startLiveBackend(
		func() error { return speaker.Init(sr, bufferSize) },
		func() { speaker.Play(root) },
		func() {
			speaker.Clear()
			speaker.Close()
		},
		timeout,
	)
}

func startLiveBackend(initFn func() error, startFn func(), closeFn func(), timeout time.Duration) *LiveBackend {
	b := &LiveBackend{
		states:  make(chan BackendState, 8),
		closeFn: closeFn,
	}
	b.emit(BackendState{Kind: BackendStateStarting})

	done := make(chan error, 1)
	go func() {
		err := initFn()
		if err == nil {
			startFn()
			b.ready.Store(true)
		}
		done <- err
	}()
	go b.watch(done, timeout)

	return b
}

// States returns the startup-state stream.
func (b *LiveBackend) States() <-chan BackendState { return b.states }

// Close shuts down the live speaker if it was successfully started.
func (b *LiveBackend) Close() {
	if !b.ready.Swap(false) {
		return
	}
	if b.closeFn != nil {
		b.closeFn()
	}
}

func (b *LiveBackend) emit(state BackendState) {
	select {
	case b.states <- state:
	default:
	}
}

func (b *LiveBackend) watch(done <-chan error, timeout time.Duration) {
	timerC := (<-chan time.Time)(nil)
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		timerC = timer.C
	}

	for {
		select {
		case err := <-done:
			if err != nil {
				b.emit(BackendState{
					Kind:   ClassifyInitError(err),
					Detail: err.Error(),
				})
			} else {
				b.emit(BackendState{Kind: BackendStateReady})
			}
			close(b.states)
			return
		case <-timerC:
			b.emit(BackendState{Kind: BackendStateHung})
			timerC = nil
		}
	}
}

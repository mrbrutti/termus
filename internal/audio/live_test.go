package audio

import (
	"errors"
	"slices"
	"sync/atomic"
	"testing"
	"time"
)

func TestClassifyInitError(t *testing.T) {
	tests := []struct {
		err  error
		want BackendStateKind
	}{
		{err: nil, want: BackendStateReady},
		{err: errors.New("AQMEIO error -66680 finding/initializing Default-Output"), want: BackendStateNoDefaultDevice},
		{err: errors.New("Could not find default device"), want: BackendStateNoDefaultDevice},
		{err: errors.New("speaker cannot be initialized more than once"), want: BackendStateInitFailed},
	}
	for _, tt := range tests {
		if got := ClassifyInitError(tt.err); got != tt.want {
			t.Fatalf("ClassifyInitError(%v) = %q, want %q", tt.err, got, tt.want)
		}
	}
}

func TestLiveBackendReportsHungThenReady(t *testing.T) {
	var started atomic.Bool
	backend := startLiveBackend(
		func() error {
			time.Sleep(20 * time.Millisecond)
			return nil
		},
		func() { started.Store(true) },
		nil,
		5*time.Millisecond,
	)

	var states []BackendStateKind
	timeout := time.After(200 * time.Millisecond)
	for {
		select {
		case state := <-backend.States():
			states = append(states, state.Kind)
			if len(states) == 3 {
				backend.Close()
				if !started.Load() {
					t.Fatal("expected startFn to run before ready state")
				}
				want := []BackendStateKind{
					BackendStateStarting,
					BackendStateHung,
					BackendStateReady,
				}
				if !slices.Equal(states, want) {
					t.Fatalf("states = %v, want %v", states, want)
				}
				return
			}
		case <-timeout:
			t.Fatal("timed out waiting for backend states")
		}
	}
}

func TestLiveBackendCloseOnlyRunsAfterReady(t *testing.T) {
	var closes atomic.Int32
	backend := startLiveBackend(
		func() error { return nil },
		func() {},
		func() { closes.Add(1) },
		0,
	)
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case state := <-backend.States():
			if state.Kind == BackendStateReady {
				goto ready
			}
		case <-timeout:
			t.Fatal("timed out waiting for ready state")
		}
	}
ready:
	backend.Close()
	backend.Close()
	if closes.Load() != 1 {
		t.Fatalf("close count = %d, want 1", closes.Load())
	}
}

func TestLiveBackendRetryStartsFreshAttempt(t *testing.T) {
	var attempts atomic.Int32
	backend := startLiveBackend(
		func() error {
			if attempts.Add(1) == 1 {
				return errors.New("no device")
			}
			return nil
		},
		func() {},
		nil,
		0,
	)
	states := []BackendStateKind{(<-backend.States()).Kind, (<-backend.States()).Kind}
	backend.Retry()
	states = append(states, (<-backend.States()).Kind, (<-backend.States()).Kind)
	backend.Close()
	want := []BackendStateKind{
		BackendStateStarting,
		BackendStateNoDefaultDevice,
		BackendStateStarting,
		BackendStateReady,
	}
	if !slices.Equal(states, want) {
		t.Fatalf("states = %v, want %v", states, want)
	}
}

func TestLiveBackendCanEnterRenderOnly(t *testing.T) {
	backend := startLiveBackend(
		func() error {
			time.Sleep(20 * time.Millisecond)
			return nil
		},
		func() {},
		nil,
		0,
	)
	<-backend.States() // initial starting
	backend.SetRenderOnly()
	if got := (<-backend.States()).Kind; got != BackendStateRenderOnly {
		t.Fatalf("render-only state = %q, want %q", got, BackendStateRenderOnly)
	}
	backend.Close()
}

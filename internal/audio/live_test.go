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
		case state, ok := <-backend.States():
			if !ok {
				if !started.Load() {
					t.Fatal("expected startFn to run before states closed")
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
			states = append(states, state.Kind)
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
	for range backend.States() {
	}
	backend.Close()
	backend.Close()
	if closes.Load() != 1 {
		t.Fatalf("close count = %d, want 1", closes.Load())
	}
}

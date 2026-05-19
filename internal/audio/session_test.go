package audio

import (
	"context"
	"testing"
	"time"

	"github.com/mrbrutti/termus/internal/acestep"
	"github.com/mrbrutti/termus/internal/scope"
)

// sessionTestAlgo is a tiny gen.Algorithm stub used to satisfy NewRoot in the
// session tests. It produces silence but is enough for the lifecycle checks.
type sessionTestAlgo struct{ name string }

func (s *sessionTestAlgo) Name() string               { return s.name }
func (s *sessionTestAlgo) Seed(int64)                 {}
func (s *sessionTestAlgo) Next(left, right []float64) {}

// TestPlaybackSession_SF2ToSF2 verifies SF2 -> SF2 switching tears down the
// old Root and stands up a new one. The session's CurrentEngine() should
// remain EngineSF2 across the switch and CurrentRoot() should change to the
// new pointer.
func TestPlaybackSession_SF2ToSF2(t *testing.T) {
	sess := NewPlaybackSession(nil)
	ring := scope.NewRing(1024)
	first := &sessionTestAlgo{name: "first"}
	second := &sessionTestAlgo{name: "second"}

	ctx := context.Background()
	if err := sess.Switch(ctx, SwitchRequest{
		Engine:    EngineSF2,
		Algo:      first,
		ScopeRing: ring,
		Volume:    70,
	}, nil); err != nil {
		t.Fatalf("first SF2 switch: %v", err)
	}
	if got := sess.CurrentEngine(); got != EngineSF2 {
		t.Fatalf("CurrentEngine = %s, want sf2", got)
	}
	firstRoot := sess.CurrentRoot()
	if firstRoot == nil {
		t.Fatalf("expected non-nil Root after first switch")
	}

	if err := sess.Switch(ctx, SwitchRequest{
		Engine:    EngineSF2,
		Algo:      second,
		ScopeRing: ring,
		Volume:    80,
	}, nil); err != nil {
		t.Fatalf("second SF2 switch: %v", err)
	}
	secondRoot := sess.CurrentRoot()
	if secondRoot == nil {
		t.Fatalf("expected non-nil Root after second switch")
	}
	if secondRoot == firstRoot {
		t.Fatalf("Root pointer did not change after SF2 -> SF2 switch")
	}
	if got := sess.CurrentEngine(); got != EngineSF2 {
		t.Fatalf("CurrentEngine after second switch = %s, want sf2", got)
	}
	if got := sess.CurrentStreamer(); got != nil {
		t.Fatalf("CurrentStreamer should be nil while SF2 is active, got %v", got)
	}
	if err := sess.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

// TestPlaybackSession_SF2ToACEStep verifies the cross-engine switch:
// teardown the SF2 Root, start an ACE-Step streamer against a mock producer.
func TestPlaybackSession_SF2ToACEStep(t *testing.T) {
	sess := NewPlaybackSession(nil)
	ring := scope.NewRing(1024)
	algo := &sessionTestAlgo{name: "sf2"}
	if err := sess.Switch(context.Background(), SwitchRequest{
		Engine:    EngineSF2,
		Algo:      algo,
		ScopeRing: ring,
	}, nil); err != nil {
		t.Fatalf("SF2 switch: %v", err)
	}
	prod := newMockProducer()
	sink := &recordingSink{}
	if err := sess.Switch(context.Background(), SwitchRequest{
		Engine: EngineACEStep,
		StreamCfg: StreamerConfig{
			Producer:     prod,
			Sink:         sink,
			CrossfadeSec: 0.05,
			MaxTracks:    2,
		},
	}, nil); err != nil {
		t.Fatalf("ACE-Step switch: %v", err)
	}
	if got := sess.CurrentEngine(); got != EngineACEStep {
		t.Fatalf("CurrentEngine = %s, want acestep", got)
	}
	if sess.CurrentRoot() != nil {
		t.Fatalf("CurrentRoot should be nil after ACE-Step takeover")
	}
	if sess.CurrentStreamer() == nil {
		t.Fatalf("CurrentStreamer should be non-nil after ACE-Step takeover")
	}

	// Wait for at least one render before tearing down. The streamer runs
	// asynchronously; without this, Stop races the producer.
	waitFor(t, 4*time.Second, func() bool {
		return len(sink.snapshot()) >= 1
	})

	if err := sess.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

// TestPlaybackSession_ACEStepToSF2 verifies switching back to SF2 cleanly
// stops the streamer but does not Shutdown() the daemon manager.
func TestPlaybackSession_ACEStepToSF2(t *testing.T) {
	sess := NewPlaybackSession(nil)
	prod := newMockProducer()
	sink := &recordingSink{}
	// Pre-install a manager so we can assert it survives.
	mgr := &acestep.Manager{}
	if err := sess.Switch(context.Background(), SwitchRequest{
		Engine:  EngineACEStep,
		Manager: mgr,
		StreamCfg: StreamerConfig{
			Producer:     prod,
			Sink:         sink,
			CrossfadeSec: 0.05,
			MaxTracks:    1,
		},
	}, nil); err != nil {
		t.Fatalf("ACE-Step switch: %v", err)
	}
	if sess.Manager() != mgr {
		t.Fatalf("session should own the manager after first ACE-Step switch")
	}

	// Switch to SF2.
	ring := scope.NewRing(1024)
	if err := sess.Switch(context.Background(), SwitchRequest{
		Engine:    EngineSF2,
		Algo:      &sessionTestAlgo{name: "back-to-sf2"},
		ScopeRing: ring,
	}, nil); err != nil {
		t.Fatalf("SF2 switch back: %v", err)
	}
	if got := sess.CurrentEngine(); got != EngineSF2 {
		t.Fatalf("CurrentEngine = %s, want sf2", got)
	}
	if sess.Manager() != mgr {
		t.Fatalf("manager pointer should survive the SF2 switch (daemon stays alive)")
	}
	if sess.CurrentStreamer() != nil {
		t.Fatalf("CurrentStreamer should be nil after switching back to SF2")
	}
	_ = sess.Stop(context.Background())
}

// TestPlaybackSession_ACEStepToACEStep starts an ACE-Step engine, then
// switches to a fresh ACE-Step engine with a new producer. The streamer
// pointer must change but the manager must stay the same.
func TestPlaybackSession_ACEStepToACEStep(t *testing.T) {
	sess := NewPlaybackSession(nil)
	prod1 := newMockProducer()
	sink1 := &recordingSink{}
	mgr := &acestep.Manager{}
	if err := sess.Switch(context.Background(), SwitchRequest{
		Engine:  EngineACEStep,
		Manager: mgr,
		StreamCfg: StreamerConfig{
			Producer:     prod1,
			Sink:         sink1,
			CrossfadeSec: 0.05,
			MaxTracks:    1,
		},
	}, nil); err != nil {
		t.Fatalf("first ACE-Step switch: %v", err)
	}
	first := sess.CurrentStreamer()
	if first == nil {
		t.Fatalf("expected first streamer to be non-nil")
	}

	prod2 := newMockProducer()
	sink2 := &recordingSink{}
	if err := sess.Switch(context.Background(), SwitchRequest{
		Engine: EngineACEStep,
		StreamCfg: StreamerConfig{
			Producer:     prod2,
			Sink:         sink2,
			CrossfadeSec: 0.05,
			MaxTracks:    1,
		},
	}, nil); err != nil {
		t.Fatalf("second ACE-Step switch: %v", err)
	}
	if sess.CurrentStreamer() == first {
		t.Fatalf("streamer pointer should change across ACE-Step switches")
	}
	if sess.Manager() != mgr {
		t.Fatalf("manager pointer must survive ACE-Step ↔ ACE-Step switch")
	}
	_ = sess.Stop(context.Background())
}

// TestPlaybackSession_StopShutsManager confirms that Stop() releases the
// manager (so the daemon goes away on process exit), even though it survives
// hot switches.
func TestPlaybackSession_StopShutsManager(t *testing.T) {
	sess := NewPlaybackSession(nil)
	mgr := &acestep.Manager{}
	if err := sess.Switch(context.Background(), SwitchRequest{
		Engine:  EngineACEStep,
		Manager: mgr,
		StreamCfg: StreamerConfig{
			Producer:     newMockProducer(),
			Sink:         &recordingSink{},
			CrossfadeSec: 0.05,
			MaxTracks:    1,
		},
	}, nil); err != nil {
		t.Fatalf("ACE-Step switch: %v", err)
	}
	if sess.Manager() == nil {
		t.Fatalf("manager should be non-nil after ACE-Step switch")
	}
	if err := sess.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if sess.Manager() != nil {
		t.Fatalf("manager should be nil after Stop()")
	}
}

// TestPlaybackSession_ValidatesRequest covers the SF2-no-algo and
// ACEStep-no-producer rejection paths so callers get a useful error rather
// than a panic.
func TestPlaybackSession_ValidatesRequest(t *testing.T) {
	sess := NewPlaybackSession(nil)
	if err := sess.Switch(context.Background(), SwitchRequest{Engine: EngineSF2}, nil); err == nil {
		t.Fatalf("expected error from SF2 switch with nil Algo")
	}
	if err := sess.Switch(context.Background(), SwitchRequest{Engine: EngineACEStep}, nil); err == nil {
		t.Fatalf("expected error from ACE-Step switch with no Producer")
	}
}

// TestPlaybackSession_EmitsEvents confirms that the optional events channel
// receives teardown-old / starting-new / ready phases. Use a buffered chan so
// the test never blocks the switch.
func TestPlaybackSession_EmitsEvents(t *testing.T) {
	sess := NewPlaybackSession(nil)
	events := make(chan SwitchEvent, 8)
	if err := sess.Switch(context.Background(), SwitchRequest{
		Engine:    EngineSF2,
		Algo:      &sessionTestAlgo{name: "a"},
		ScopeRing: scope.NewRing(256),
	}, events); err != nil {
		t.Fatalf("first switch: %v", err)
	}
	// Second switch must emit teardown-old, starting-new, ready.
	if err := sess.Switch(context.Background(), SwitchRequest{
		Engine:    EngineSF2,
		Algo:      &sessionTestAlgo{name: "b"},
		ScopeRing: scope.NewRing(256),
	}, events); err != nil {
		t.Fatalf("second switch: %v", err)
	}
	close(events)

	phases := []string{}
	for ev := range events {
		phases = append(phases, ev.Phase)
	}
	mustContain := func(p string) {
		for _, ph := range phases {
			if ph == p {
				return
			}
		}
		t.Fatalf("expected phase %q in events; got %v", p, phases)
	}
	mustContain("starting-new")
	mustContain("ready")
	mustContain("teardown-old")
	_ = sess.Stop(context.Background())
}

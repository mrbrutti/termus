package synth

import (
	"math"
	"testing"
)

// TestConvolverImpulsePass verifies that convolving with a single-sample
// impulse IR is a no-op (output equals input).
func TestConvolverImpulsePass(t *testing.T) {
	c := NewConvolver([]float64{1.0})
	for _, x := range []float64{0.7, -0.3, 0.5, 0.0, -0.1} {
		y := c.Tick(x)
		if math.Abs(y-x) > 1e-9 {
			t.Fatalf("impulse IR: input %g → output %g", x, y)
		}
	}
}

// TestConvolverDelaysSignal verifies a length-N IR of [0, 0, …, 1] delays
// the input by N-1 samples.
func TestConvolverDelaysSignal(t *testing.T) {
	const delay = 5
	ir := make([]float64, delay+1)
	ir[delay] = 1.0
	c := NewConvolver(ir)
	// Input: 1, then zeros.
	for i := 0; i < delay; i++ {
		if y := c.Tick(0); i == 0 {
			// First call after we send the 1 should be zero (delay hasn't elapsed).
			_ = y
		}
	}
	got := c.Tick(0)
	if got != 0 {
		t.Fatalf("step %d: expected 0 before sending impulse, got %g", delay, got)
	}
	// Reset and try with an impulse input at t=0, then zeros.
	c = NewConvolver(ir)
	first := c.Tick(1.0)
	if first != 0 {
		t.Fatalf("impulse, t=0: expected 0 (delay not elapsed), got %g", first)
	}
	for i := 1; i < delay; i++ {
		if v := c.Tick(0); v != 0 {
			t.Fatalf("impulse, t=%d: expected 0, got %g", i, v)
		}
	}
	v := c.Tick(0)
	if math.Abs(v-1.0) > 1e-9 {
		t.Fatalf("impulse, t=%d: expected 1.0, got %g", delay, v)
	}
}

func TestSyntheticRoomIRNonEmpty(t *testing.T) {
	ir := SyntheticRoomIR(0.05)
	if len(ir) < 8 {
		t.Fatalf("IR too short: %d samples", len(ir))
	}
	if ir[0] == 0 {
		t.Fatal("direct path missing")
	}
	// Make sure at least one of the reflections landed.
	hasReflection := false
	for i := 1; i < len(ir); i++ {
		if ir[i] != 0 {
			hasReflection = true
			break
		}
	}
	if !hasReflection {
		t.Fatal("no reflections in IR")
	}
}

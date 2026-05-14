package synth

import (
	"math"
	"testing"
)

// TestReverbProducesTail verifies that a single impulse produces a decaying
// tail of non-zero samples — the defining behavior of a reverberator.
func TestReverbProducesTail(t *testing.T) {
	r := NewReverb(1.0) // full wet so we measure the tail, not the dry impulse
	out := make([]float64, SampleRate)
	out[0] = r.Tick(1.0)
	for i := 1; i < len(out); i++ {
		out[i] = r.Tick(0)
	}
	// Energy must persist well past the comb-filter delays (~30 ms).
	// Sample N around 200 ms should still have measurable amplitude.
	probe := out[int(0.2*float64(SampleRate))]
	if math.Abs(probe) < 1e-5 {
		t.Fatalf("reverb tail at 200ms = %g, want |x| > 1e-5", probe)
	}
	// But it should also decay — by 1 second, well below the impulse height.
	tail := out[SampleRate-1]
	if math.Abs(tail) > 0.1 {
		t.Fatalf("reverb tail at 1s = %g, want |x| < 0.1 (not decaying)", tail)
	}
}

// TestReverbDryMix verifies that with wet=0 the output is exactly the input.
func TestReverbDryMix(t *testing.T) {
	r := NewReverb(0.0)
	for _, x := range []float64{1, 0.5, -0.3, 0} {
		y := r.Tick(x)
		if math.Abs(y-x) > 1e-9 {
			t.Fatalf("wet=0: Tick(%g) = %g, want %g", x, y, x)
		}
	}
}

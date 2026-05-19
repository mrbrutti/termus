package synth

import (
	"math"
	"testing"
)

func TestSampleRate(t *testing.T) {
	if SampleRate != 48000 {
		t.Fatalf("SampleRate = %d, want 48000", SampleRate)
	}
}

func TestSoftClipBounded(t *testing.T) {
	for _, x := range []float64{-10, -2, -1, 0, 1, 2, 10} {
		y := SoftClip(x)
		if y < -1 || y > 1 {
			t.Fatalf("SoftClip(%g) = %g, out of [-1, 1]", x, y)
		}
	}
}

func TestSoftClipLinearNearZero(t *testing.T) {
	// SoftClip is tanh; tanh(x) ≈ x for small x.
	for _, x := range []float64{-0.01, 0, 0.01} {
		y := SoftClip(x)
		if math.Abs(y-x) > 1e-3 {
			t.Fatalf("SoftClip(%g) = %g, want ~%g", x, y, x)
		}
	}
}

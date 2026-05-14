package synth

import "testing"

func TestDelayDelaysSignal(t *testing.T) {
	d := NewDelay(0.1, 0, 0) // 100ms, no feedback, no mix
	// Input is a single sample of 1.0 followed by zeros.
	steps := int(0.1*float64(SampleRate)) + 10
	var out [200]float64
	for i := 0; i < steps; i++ {
		var x float64
		if i == 0 {
			x = 1.0
		}
		v := d.Tick(x)
		if i < 200 {
			out[i] = v
		}
	}
	// With mix=0 the output is the dry signal only → spike at i=0.
	if out[0] < 0.99 {
		t.Fatalf("mix=0: out[0]=%g, want ~1.0", out[0])
	}
}

func TestDelayWetMix(t *testing.T) {
	d := NewDelay(0.01, 0, 1) // 10ms, no feedback, full wet
	expectedDelaySamples := int(0.01 * float64(SampleRate))
	got := make([]float64, expectedDelaySamples+5)
	for i := range got {
		var x float64
		if i == 0 {
			x = 1.0
		}
		got[i] = d.Tick(x)
	}
	// Full-wet: dry is suppressed; the spike appears at the delay tap.
	if got[0] > 0.05 {
		t.Fatalf("full wet, t=0: got %g, want near 0", got[0])
	}
	if got[expectedDelaySamples] < 0.95 {
		t.Fatalf("full wet, t=delay: got %g, want near 1", got[expectedDelaySamples])
	}
}

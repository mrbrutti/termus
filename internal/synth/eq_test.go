package synth

import (
	"math"
	"testing"
)

func TestLowShelfBoostsLow(t *testing.T) {
	// 100 Hz sine in, 200 Hz low-shelf @ +12 dB.
	o := NewOscillator(WaveSine)
	o.SetFrequency(100)
	f := NewLowShelf(200, 12, 0.707)

	var dryRMS, wetRMS float64
	const n = SampleRate
	for i := 0; i < n; i++ {
		x := o.Tick()
		y := f.Tick(x)
		if i > n/4 { // skip transient
			dryRMS += x * x
			wetRMS += y * y
		}
	}
	dryRMS = math.Sqrt(dryRMS / float64(3*n/4))
	wetRMS = math.Sqrt(wetRMS / float64(3*n/4))
	ratio := wetRMS / dryRMS
	// +12 dB → ratio ≈ 4x.
	if ratio < 3.0 || ratio > 5.0 {
		t.Fatalf("low-shelf +12dB ratio = %g, want ~4.0", ratio)
	}
}

func TestHighShelfLeavesLowAlone(t *testing.T) {
	// 100 Hz sine in, 4 kHz high-shelf @ +12 dB — should leave 100 Hz alone.
	o := NewOscillator(WaveSine)
	o.SetFrequency(100)
	f := NewHighShelf(4000, 12, 0.707)

	var dryRMS, wetRMS float64
	const n = SampleRate
	for i := 0; i < n; i++ {
		x := o.Tick()
		y := f.Tick(x)
		if i > n/4 {
			dryRMS += x * x
			wetRMS += y * y
		}
	}
	dryRMS = math.Sqrt(dryRMS / float64(3*n/4))
	wetRMS = math.Sqrt(wetRMS / float64(3*n/4))
	ratio := wetRMS / dryRMS
	if ratio < 0.9 || ratio > 1.1 {
		t.Fatalf("high-shelf @ 4kHz should leave 100Hz alone; ratio = %g", ratio)
	}
}

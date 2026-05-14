package synth

import (
	"math"
	"testing"
)

// TestCompressorReducesPeaks verifies the compressor lowers the peak amplitude
// of a loud signal while approximately preserving quieter parts.
func TestCompressorReducesPeaks(t *testing.T) {
	// 4:1 ratio, threshold at -20 dBFS, no makeup gain.
	c := NewCompressor(-20, 4.0, 5, 100, 6, 0)

	// Loud sine at -6 dBFS = amplitude 0.5.
	const amp = 0.5
	o := NewOscillator(WaveSine)
	o.SetFrequency(440)

	// Run for half a second to let envelope settle, then measure.
	for i := 0; i < SampleRate/4; i++ {
		c.Tick(amp * o.Tick())
	}
	var peak float64
	for i := 0; i < SampleRate/4; i++ {
		v := c.Tick(amp * o.Tick())
		if math.Abs(v) > peak {
			peak = math.Abs(v)
		}
	}
	// Input peak = 0.5 (-6 dB). Threshold -20 dB. Overshoot 14 dB.
	// With 4:1 ratio, output overshoot = 14/4 = 3.5 dB above threshold,
	// so output peak ≈ -16.5 dB ≈ 0.150. Allow generous tolerance.
	if peak < 0.10 || peak > 0.25 {
		t.Fatalf("compressed peak = %g, expected roughly 0.15", peak)
	}
}

// TestCompressorLeavesQuietSignalAlone verifies signals well below the
// threshold pass through with only the makeup gain applied.
func TestCompressorLeavesQuietSignalAlone(t *testing.T) {
	c := NewCompressor(-12, 4.0, 5, 100, 6, 0)
	// -40 dBFS = amplitude 0.01.
	const amp = 0.01
	o := NewOscillator(WaveSine)
	o.SetFrequency(440)
	var peak float64
	for i := 0; i < SampleRate; i++ {
		v := c.Tick(amp * o.Tick())
		if i > SampleRate/2 && math.Abs(v) > peak {
			peak = math.Abs(v)
		}
	}
	// Should be approximately unchanged: 0.01 ± 10%.
	if peak < 0.009 || peak > 0.012 {
		t.Fatalf("quiet signal peak = %g, expected ~0.01", peak)
	}
}

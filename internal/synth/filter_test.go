package synth

import (
	"math"
	"testing"
)

func rms(buf []float64) float64 {
	var s float64
	for _, x := range buf {
		s += x * x
	}
	return math.Sqrt(s / float64(len(buf)))
}

func TestLowpassPassesLow(t *testing.T) {
	o := NewOscillator(WaveSine)
	o.SetFrequency(100)
	f := NewLowpass(1000, 0.7)
	buf := make([]float64, SampleRate)
	for i := range buf {
		buf[i] = f.Tick(o.Tick())
	}
	r := rms(buf[SampleRate/2:]) // skip transient
	if r < 0.5 {
		t.Fatalf("100Hz through 1kHz LP: RMS=%g, expected near 0.707", r)
	}
}

func TestLowpassBlocksHigh(t *testing.T) {
	o := NewOscillator(WaveSine)
	o.SetFrequency(8000)
	f := NewLowpass(500, 0.7)
	buf := make([]float64, SampleRate)
	for i := range buf {
		buf[i] = f.Tick(o.Tick())
	}
	r := rms(buf[SampleRate/2:])
	if r > 0.1 {
		t.Fatalf("8kHz through 500Hz LP: RMS=%g, expected small", r)
	}
}

package synth

import (
	"math"
	"testing"
)

func TestSineZeroCrossings(t *testing.T) {
	o := NewOscillator(WaveSine)
	o.SetFrequency(440)
	const seconds = 1
	buf := make([]float64, SampleRate*seconds)
	for i := range buf {
		buf[i] = o.Tick()
	}
	crossings := 0
	for i := 1; i < len(buf); i++ {
		if (buf[i-1] <= 0 && buf[i] > 0) || (buf[i-1] >= 0 && buf[i] < 0) {
			crossings++
		}
	}
	// 440 Hz over 1 second → ~880 zero crossings (two per cycle).
	if crossings < 870 || crossings > 890 {
		t.Fatalf("440 Hz sine produced %d zero-crossings, want ~880", crossings)
	}
}

func TestSawRange(t *testing.T) {
	o := NewOscillator(WaveSaw)
	o.SetFrequency(220)
	var minv, maxv = math.Inf(1), math.Inf(-1)
	for i := 0; i < SampleRate; i++ {
		v := o.Tick()
		if v < minv {
			minv = v
		}
		if v > maxv {
			maxv = v
		}
	}
	if minv > -0.98 || maxv < 0.98 {
		t.Fatalf("saw range = [%g, %g], want ~[-1, 1]", minv, maxv)
	}
}

func TestSetFrequencyDoesNotResetPhase(t *testing.T) {
	o := NewOscillator(WaveSine)
	o.SetFrequency(440)
	for i := 0; i < 100; i++ {
		o.Tick()
	}
	before := o.Tick()
	o.SetFrequency(440) // same frequency must not jump
	after := o.Tick()
	// Two adjacent sine samples at 440Hz/44.1kHz differ by < 0.07.
	if math.Abs(after-before) > 0.1 {
		t.Fatalf("phase jumped on SetFrequency: %g → %g", before, after)
	}
}

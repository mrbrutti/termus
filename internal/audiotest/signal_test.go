package audiotest

import (
	"math"
	"testing"
)

func TestRMSOfUnitSineIsOneOverSqrtTwo(t *testing.T) {
	s := Sine(440, 1.0, 44100, 44100)
	got := RMS(s)
	want := 1.0 / math.Sqrt2
	if math.Abs(got-want) > 1e-3 {
		t.Fatalf("RMS = %g, want %g", got, want)
	}
}

func TestPeakOfSineEqualsAmp(t *testing.T) {
	s := Sine(440, 0.5, 44100, 44100)
	got := Peak(s)
	if math.Abs(got-0.5) > 1e-3 {
		t.Fatalf("Peak = %g, want 0.5", got)
	}
}

func TestToDBMatchesKnownValues(t *testing.T) {
	cases := []struct{ amp, want float64 }{
		{1.0, 0},
		{0.5, -6.02059991},
		{0.0001, -80},
	}
	for _, c := range cases {
		got := ToDB(c.amp)
		if math.Abs(got-c.want) > 0.01 {
			t.Fatalf("ToDB(%g) = %g, want %g", c.amp, got, c.want)
		}
	}
}

func TestToMonoAveragesChannels(t *testing.T) {
	in := [][2]float64{{1, 0}, {0, 1}, {0.5, 0.5}}
	got := ToMono(in)
	want := []float64{0.5, 0.5, 0.5}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			t.Fatalf("ToMono[%d] = %g, want %g", i, got[i], want[i])
		}
	}
}

func TestClickIsImpulse(t *testing.T) {
	s := Click(10, 100, 0.8)
	if math.Abs(s[10]-0.8) > 1e-9 {
		t.Fatalf("Click[10] = %g, want 0.8", s[10])
	}
	for i, v := range s {
		if i == 10 {
			continue
		}
		if math.Abs(v) > 1e-9 {
			t.Fatalf("Click[%d] = %g, want 0", i, v)
		}
	}
}

func TestModulatedSineProducesExpectedPitchExcursion(t *testing.T) {
	s := ModulatedSine(440, 15, 0.7, 44100, 4*44100)
	if Peak(s) < 0.9 || Peak(s) > 1.0 {
		t.Fatalf("ModulatedSine peak = %g, want ~1.0", Peak(s))
	}
}

// Zero depth should produce a pure 440 Hz sine: count positive-going zero
// crossings over 1 second and verify ~440 within tolerance.
func TestModulatedSineWithZeroDepthIsPureSine(t *testing.T) {
	s := ModulatedSine(440, 0, 0.7, 44100, 44100)
	crossings := countPositiveZeroCrossings(s)
	if crossings < 438 || crossings > 442 {
		t.Fatalf("zero-depth crossings = %d in 1s, want 440 ± 2", crossings)
	}
}

// Non-zero depth must produce measurably more zero-crossing-rate variance
// than zero depth. Confirms ModulatedSine actually modulates pitch, not
// just amplitude.
func TestModulatedSineWithDepthSpreadsZeroCrossings(t *testing.T) {
	const sr = 44100.0
	const seconds = 4.0
	flat := ModulatedSine(440, 0, 0.7, sr, int(seconds*sr))
	modded := ModulatedSine(440, 50, 0.7, sr, int(seconds*sr))
	flatStd := zeroCrossingRateStdDev(flat, int(sr/10))   // 100 ms windows
	moddedStd := zeroCrossingRateStdDev(modded, int(sr/10))
	if moddedStd <= flatStd*1.5 {
		t.Fatalf("expected modded stddev (%.3f) > 1.5× flat stddev (%.3f)",
			moddedStd, flatStd)
	}
}

func countPositiveZeroCrossings(s []float64) int {
	n := 0
	for i := 1; i < len(s); i++ {
		if s[i-1] < 0 && s[i] >= 0 {
			n++
		}
	}
	return n
}

func zeroCrossingRateStdDev(s []float64, windowSamples int) float64 {
	if windowSamples < 1 || len(s) < 2*windowSamples {
		return 0
	}
	var rates []float64
	for i := 0; i+windowSamples <= len(s); i += windowSamples {
		rates = append(rates, float64(countPositiveZeroCrossings(s[i:i+windowSamples])))
	}
	if len(rates) == 0 {
		return 0
	}
	var mean float64
	for _, r := range rates {
		mean += r
	}
	mean /= float64(len(rates))
	var ssq float64
	for _, r := range rates {
		d := r - mean
		ssq += d * d
	}
	return math.Sqrt(ssq / float64(len(rates)))
}

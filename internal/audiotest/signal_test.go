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

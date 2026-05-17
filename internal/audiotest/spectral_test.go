package audiotest

import (
	"math"
	"math/rand"
	"testing"
)

func TestSpectralCentroidOfPureSineIsAtSineFrequency(t *testing.T) {
	s := Sine(1000, 1.0, 44100, 16384)
	got := SpectralCentroidHz(s, 44100)
	if math.Abs(got-1000) > 50 {
		t.Fatalf("centroid = %.1f Hz, want ~1000 ± 50", got)
	}
}

func TestSpectralCentroidOfHighSineIsHigherThanLowSine(t *testing.T) {
	low := Sine(500, 1.0, 44100, 16384)
	high := Sine(4000, 1.0, 44100, 16384)
	if SpectralCentroidHz(low, 44100) >= SpectralCentroidHz(high, 44100) {
		t.Fatal("expected 4kHz sine to have higher centroid than 500Hz sine")
	}
}

func TestSpectralCentroidOfWhiteNoiseIsNearQuarterNyquist(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	n := 16384
	buf := make([]float64, n)
	for i := range buf {
		buf[i] = 2*rng.Float64() - 1
	}
	got := SpectralCentroidHz(buf, 44100)
	// White noise centroid ≈ sampleRate/4 = 11025 Hz, ± wide tolerance.
	if math.Abs(got-11025) > 2000 {
		t.Fatalf("white noise centroid = %.0f, want ~11025 ± 2000", got)
	}
}

func TestAssertSpectralCentroidHzPasses(t *testing.T) {
	s := Sine(2000, 1.0, 44100, 16384)
	AssertSpectralCentroidHz(t, s, 44100, 2000, 100)
}

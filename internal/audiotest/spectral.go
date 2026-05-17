// internal/audiotest/spectral.go
//
// Frequency-domain measurements via the FFT library already used by the
// synth package's fft_convolver.
package audiotest

import (
	"math"
	"math/cmplx"
	"testing"

	"github.com/madelynnblue/go-dsp/fft"
)

// SpectralCentroidHz returns the magnitude-weighted mean frequency of the
// buffer. A Hann window is applied before the FFT to suppress sidelobe
// leakage that would otherwise pull the centroid away from the true peak
// frequency for non-integer-bin tones. The buffer is truncated to the
// largest power-of-2 length ≤ len(buf) to satisfy the FFT. Suitable for
// stationary signals; segment first and average for evolving signals.
func SpectralCentroidHz(buf []float64, sampleRate float64) float64 {
	if len(buf) < 2 {
		return 0
	}
	n := 1
	for n*2 <= len(buf) {
		n *= 2
	}
	windowed := make([]float64, n)
	for i := 0; i < n; i++ {
		// Hann window: 0.5 * (1 - cos(2πi / (n-1)))
		w := 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(n-1)))
		windowed[i] = buf[i] * w
	}
	spec := fft.FFTReal(windowed)
	var num, den float64
	for k := 0; k < n/2; k++ {
		mag := cmplx.Abs(spec[k])
		freq := float64(k) * sampleRate / float64(n)
		num += freq * mag
		den += mag
	}
	if den == 0 {
		return 0
	}
	return num / den
}

// AssertSpectralCentroidHz fails the test if the centroid of buf differs from
// wantHz by more than tolHz.
func AssertSpectralCentroidHz(t testing.TB, buf []float64, sampleRate, wantHz, tolHz float64) {
	t.Helper()
	got := SpectralCentroidHz(buf, sampleRate)
	if math.Abs(got-wantHz) > tolHz {
		t.Errorf("centroid = %.1f Hz, want %.1f ± %.1f Hz", got, wantHz, tolHz)
	}
}

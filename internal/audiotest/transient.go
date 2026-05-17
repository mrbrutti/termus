// internal/audiotest/transient.go
//
// Energy-based transient detection. Computes short-time RMS in fixed
// windows and reports the start-of-window sample index for any window
// whose RMS exceeds the previous window's by at least thresholdDB.
package audiotest

import (
	"math"
	"testing"
)

// FindTransients returns sample indices where short-time RMS increases by
// thresholdDB relative to the previous window. windowSamples defaults to 64
// if non-positive. Each reported index is the start of the rising window.
func FindTransients(buf []float64, windowSamples int, thresholdDB float64) []int {
	if windowSamples < 1 {
		windowSamples = 64
	}
	if len(buf) < 2*windowSamples {
		return nil
	}
	count := len(buf) / windowSamples
	rms := make([]float64, count)
	for i := 0; i < count; i++ {
		rms[i] = RMS(buf[i*windowSamples : (i+1)*windowSamples])
	}
	out := []int{}
	for i := 1; i < len(rms); i++ {
		if rms[i-1] < 1e-9 {
			if rms[i] > 1e-6 {
				out = append(out, i*windowSamples)
			}
			continue
		}
		dB := 20 * math.Log10(rms[i]/rms[i-1])
		if dB >= thresholdDB {
			out = append(out, i*windowSamples)
		}
	}
	return out
}

// AssertHasTransientAt fails the test if no detected transient is within
// toleranceSamples of expectedSample.
func AssertHasTransientAt(t testing.TB, buf []float64, expectedSample, toleranceSamples int) {
	t.Helper()
	trans := FindTransients(buf, 256, 12.0)
	for _, s := range trans {
		if absInt(s-expectedSample) <= toleranceSamples {
			return
		}
	}
	t.Errorf("no transient near sample %d (±%d); found at %v",
		expectedSample, toleranceSamples, trans)
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

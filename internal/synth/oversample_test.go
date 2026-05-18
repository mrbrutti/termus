package synth_test

import (
	"math"
	"testing"

	"github.com/mrbrutti/termus/internal/audiotest"
	"github.com/mrbrutti/termus/internal/synth"
)

// TestOversamplerIsIdentityOnDC checks that feeding a constant value through
// the oversampler returns approximately that constant after filter settling.
func TestOversamplerIsIdentityOnDC(t *testing.T) {
	const factor = 16
	const dc = 0.5
	over := synth.NewOversampler(factor)

	// Warm up: let the IIR filters settle over 128 output samples.
	const warmup = 128
	for i := 0; i < warmup; i++ {
		over.Process(func() float64 { return dc })
	}

	// After settling, output should be very close to the input DC value.
	got := over.Process(func() float64 { return dc })
	const tol = 0.01
	if math.Abs(got-dc) > tol {
		t.Fatalf("DC after settling: got %g, want %g ± %g", got, dc, tol)
	}
}

// TestOversamplerSuppressesAliasingOnSawtooth verifies that a naive saw wave
// at a high frequency produces less high-band spectral energy when rendered
// through the oversampler compared to the raw naive output.
func TestOversamplerSuppressesAliasingOnSawtooth(t *testing.T) {
	const sampleRate = 44100.0
	const freqHz = 4000.0 // high frequency → lots of aliases above Nyquist
	const factor = 16
	const duration = 1.0 // second
	const nSamples = int(sampleRate * duration)

	// Naive saw phase accumulator — shared between runs via closure.
	makePhaseState := func() func() float64 {
		phase := 0.0
		inc := freqHz / (sampleRate * float64(factor))
		return func() float64 {
			v := 2*phase - 1
			phase += inc
			if phase >= 1 {
				phase -= 1
			}
			return v
		}
	}

	// --- Oversampled version ---
	over := synth.NewOversampler(factor)
	overTick := makePhaseState()
	overBuf := make([]float64, nSamples)
	for i := range overBuf {
		overBuf[i] = over.Process(overTick)
	}

	// --- Naive version (no oversampling, same fundamental) ---
	naivePhase := 0.0
	naiveInc := freqHz / sampleRate
	naiveBuf := make([]float64, nSamples)
	for i := range naiveBuf {
		naiveBuf[i] = 2*naivePhase - 1
		naivePhase += naiveInc
		if naivePhase >= 1 {
			naivePhase -= 1
		}
	}

	// Measure spectral centroid. For a 4 kHz saw, aliases fold back from
	// well above Nyquist. The oversampled version should suppress those
	// high-frequency aliases, pulling the centroid closer to the fundamental.
	overCentroid := audiotest.SpectralCentroidHz(overBuf, sampleRate)
	naiveCentroid := audiotest.SpectralCentroidHz(naiveBuf, sampleRate)

	// The oversampled centroid must be lower than the naive one, indicating
	// reduced high-frequency alias energy.
	if overCentroid >= naiveCentroid {
		t.Fatalf("oversampled centroid (%g Hz) >= naive centroid (%g Hz); expected aliasing suppression", overCentroid, naiveCentroid)
	}

	// Sanity: oversampled RMS should not blow up relative to naive.
	overRMS := audiotest.RMS(overBuf)
	naiveRMS := audiotest.RMS(naiveBuf)
	if overRMS > naiveRMS*2 {
		t.Fatalf("oversampled RMS (%g) blew up vs naive (%g)", overRMS, naiveRMS)
	}
}

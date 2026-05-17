// internal/audiotest/pitchtrack.go
//
// Zero-crossing pitch tracker for monophonic sustained inputs. Used to
// verify wow/flutter implementations produce the claimed pitch-modulation
// depth and rate.
package audiotest

import (
	"math"
	"testing"
)

// PitchTrack returns a sequence of cent-deviations from nominalHz, one
// entry per zero-crossing period. Linear interpolation between samples
// gives sub-sample crossing accuracy.
func PitchTrack(buf []float64, sampleRate, nominalHz float64) []float64 {
	crossings := make([]float64, 0, 1024)
	for i := 1; i < len(buf); i++ {
		if buf[i-1] < 0 && buf[i] >= 0 {
			denom := buf[i] - buf[i-1]
			if denom == 0 {
				continue
			}
			t := float64(i-1) - buf[i-1]/denom
			crossings = append(crossings, t)
		}
	}
	if len(crossings) < 2 {
		return nil
	}
	cents := make([]float64, 0, len(crossings)-1)
	for i := 1; i < len(crossings); i++ {
		period := crossings[i] - crossings[i-1]
		if period <= 0 {
			continue
		}
		freq := sampleRate / period
		cents = append(cents, 1200*math.Log2(freq/nominalHz))
	}
	return cents
}

// ModulationDepthCents returns the half-range of the cents trace
// (max - min) / 2.
func ModulationDepthCents(cents []float64) float64 {
	if len(cents) == 0 {
		return 0
	}
	minV, maxV := cents[0], cents[0]
	for _, c := range cents {
		if c < minV {
			minV = c
		}
		if c > maxV {
			maxV = c
		}
	}
	return (maxV - minV) / 2
}

// ModulationRateHz estimates the dominant modulation rate of the cents trace
// via autocorrelation. perSecond is the average number of cents samples per
// second (~ nominalHz for a monophonic signal).
//
// The algorithm finds the first lag where the normalised autocorrelation
// crosses from positive to negative (quarter-period of the modulation),
// then searches for the peak beyond that point (the full-period peak).
// This avoids the spurious maximum at very small lags where adjacent
// cents values are nearly identical and the ACF is monotonically
// decreasing.
func ModulationRateHz(cents []float64, perSecond float64) float64 {
	n := len(cents)
	if n < 32 {
		return 0
	}
	mean := 0.0
	for _, c := range cents {
		mean += c
	}
	mean /= float64(n)
	centered := make([]float64, n)
	for i, c := range cents {
		centered[i] = c - mean
	}
	// Compute normalisation constant (ACF at lag 0).
	var var0 float64
	for _, c := range centered {
		var0 += c * c
	}
	if var0 == 0 {
		return 0
	}

	acfAt := func(lag int) float64 {
		var s float64
		for i := 0; i+lag < n; i++ {
			s += centered[i] * centered[i+lag]
		}
		return s / var0
	}

	maxLag := n / 2

	// Step 1: find the first zero crossing (positive → negative).
	firstZero := -1
	prev := acfAt(2)
	for lag := 3; lag <= maxLag; lag++ {
		curr := acfAt(lag)
		if prev > 0 && curr <= 0 {
			firstZero = lag
			break
		}
		prev = curr
	}
	if firstZero < 0 {
		// No zero crossing found — no clear periodic modulation.
		return 0
	}

	// Step 2: find the maximum in the range [firstZero, maxLag].
	bestLag := firstZero
	bestACF := acfAt(firstZero)
	for lag := firstZero + 1; lag <= maxLag; lag++ {
		curr := acfAt(lag)
		if curr > bestACF {
			bestACF = curr
			bestLag = lag
		}
	}
	if bestACF <= 0 {
		// No positive lobe found — signal too short or no modulation.
		return 0
	}
	return perSecond / float64(bestLag)
}

// AssertPitchModulationCents fails the test if the depth or rate of pitch
// modulation in buf (measured against nominalHz) differs from the wanted
// values by more than the tolerances.
func AssertPitchModulationCents(t testing.TB, buf []float64, sampleRate, nominalHz, depthCents, rateHz, depthTol, rateTol float64) {
	t.Helper()
	cents := PitchTrack(buf, sampleRate, nominalHz)
	if len(cents) < 32 {
		t.Errorf("pitch track too short (%d entries) — input may not be monophonic at nominalHz", len(cents))
		return
	}
	duration := float64(len(buf)) / sampleRate
	perSecond := float64(len(cents)) / duration
	gotDepth := ModulationDepthCents(cents)
	if math.Abs(gotDepth-depthCents) > depthTol {
		t.Errorf("pitch modulation depth = %.2f cents, want %.2f ± %.2f", gotDepth, depthCents, depthTol)
	}
	gotRate := ModulationRateHz(cents, perSecond)
	if math.Abs(gotRate-rateHz) > rateTol {
		t.Errorf("pitch modulation rate = %.3f Hz, want %.3f ± %.3f", gotRate, rateHz, rateTol)
	}
}

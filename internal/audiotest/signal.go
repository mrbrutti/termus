// internal/audiotest/signal.go
//
// Pure-DSP utilities used by every assertion in this package. All buffers
// are mono float64 unless otherwise stated.
package audiotest

import "math"

// RMS returns the root-mean-square amplitude of a mono buffer.
func RMS(buf []float64) float64 {
	if len(buf) == 0 {
		return 0
	}
	var sum float64
	for _, x := range buf {
		sum += x * x
	}
	return math.Sqrt(sum / float64(len(buf)))
}

// Peak returns the maximum absolute sample value. Returns 0 for an empty buffer.
func Peak(buf []float64) float64 {
	if len(buf) == 0 {
		return 0
	}
	var p float64
	for _, x := range buf {
		if v := math.Abs(x); v > p {
			p = v
		}
	}
	return p
}

// ToDB converts a linear amplitude to decibels. Zero/negative returns -Inf.
func ToDB(amp float64) float64 {
	if amp <= 0 {
		return math.Inf(-1)
	}
	return 20 * math.Log10(amp)
}

// ToMono averages stereo channels into a mono buffer.
func ToMono(stereo [][2]float64) []float64 {
	out := make([]float64, len(stereo))
	for i, s := range stereo {
		out[i] = 0.5 * (s[0] + s[1])
	}
	return out
}

// Sine generates a pure sine of the given frequency and peak amplitude.
func Sine(freqHz, ampPeak, sampleRate float64, samples int) []float64 {
	out := make([]float64, samples)
	for i := range out {
		out[i] = ampPeak * math.Sin(2*math.Pi*freqHz*float64(i)/sampleRate)
	}
	return out
}

// Click returns a buffer of length n with a single impulse of value amp
// at index pos. Used for transient and IR-response tests.
func Click(pos, n int, amp float64) []float64 {
	out := make([]float64, n)
	if pos >= 0 && pos < n {
		out[pos] = amp
	}
	return out
}

// ModulatedSine generates a sine of base frequency baseHz whose instantaneous
// pitch deviates by ±depthCents at modulation rate rateHz (sine modulator).
// Used as a known-shape input to verify the pitch-modulation tracker.
func ModulatedSine(baseHz, depthCents, rateHz, sampleRate float64, samples int) []float64 {
	out := make([]float64, samples)
	phase := 0.0
	for i := range out {
		t := float64(i) / sampleRate
		cents := depthCents * math.Sin(2*math.Pi*rateHz*t)
		instFreq := baseHz * math.Pow(2, cents/1200)
		phase += 2 * math.Pi * instFreq / sampleRate
		out[i] = math.Sin(phase)
	}
	return out
}

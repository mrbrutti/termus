// Package synth provides DSP primitives used by termus generators.
package synth

import "math"

// SampleRate is the fixed audio sample rate (Hz) used throughout termus.
const SampleRate = 44100

// SoftClip applies tanh saturation, mapping any input into (-1, 1).
func SoftClip(x float64) float64 {
	return math.Tanh(x)
}

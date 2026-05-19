// Package synth provides DSP primitives used by termus generators.
package synth

import "math"

// SampleRate is the project-wide audio sample rate (Hz). 48000 so SF2 and
// ACE-Step both produce 48 kHz audio — beep v2's global speaker can only
// be Init'd once per process, so engines at different rates would collide
// on hot-switch. SF2 samples are rate-agnostic (the soundfont format
// pitch-shifts native samples at playback), so running the synth at
// 48 kHz is purely a DSP-math change.
const SampleRate = 48000

// SoftClip applies tanh saturation, mapping any input into (-1, 1).
func SoftClip(x float64) float64 {
	return math.Tanh(x)
}

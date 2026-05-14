package synth

import "math"

// Lowpass is an RBJ-style biquad lowpass filter.
type Lowpass struct {
	b0, b1, b2, a1, a2 float64
	z1, z2             float64
}

func NewLowpass(cutoffHz, q float64) *Lowpass {
	f := &Lowpass{}
	f.SetParams(cutoffHz, q)
	return f
}

// SetParams updates the filter coefficients without resetting the state.
func (f *Lowpass) SetParams(cutoffHz, q float64) {
	if cutoffHz < 10 {
		cutoffHz = 10
	}
	if cutoffHz > float64(SampleRate)/2-100 {
		cutoffHz = float64(SampleRate)/2 - 100
	}
	w0 := 2 * math.Pi * cutoffHz / float64(SampleRate)
	cosw0 := math.Cos(w0)
	alpha := math.Sin(w0) / (2 * q)

	b0 := (1 - cosw0) / 2
	b1 := 1 - cosw0
	b2 := (1 - cosw0) / 2
	a0 := 1 + alpha
	a1 := -2 * cosw0
	a2 := 1 - alpha

	f.b0 = b0 / a0
	f.b1 = b1 / a0
	f.b2 = b2 / a0
	f.a1 = a1 / a0
	f.a2 = a2 / a0
}

// Tick processes a single sample (Direct Form II Transposed).
func (f *Lowpass) Tick(x float64) float64 {
	y := f.b0*x + f.z1
	f.z1 = f.b1*x - f.a1*y + f.z2
	f.z2 = f.b2*x - f.a2*y
	return y
}

package synth

import "math"

// LowShelf is a biquad low-shelf EQ filter (RBJ cookbook). Boosts or cuts
// frequencies below the corner frequency, leaving higher frequencies
// unaffected. Useful for warming up bass or trimming muddy lows.
type LowShelf struct {
	b0, b1, b2, a1, a2 float64
	z1, z2             float64
}

// NewLowShelf builds a shelf at the given frequency with the given gain in dB
// (positive boosts, negative cuts). q ~= 0.707 gives a Butterworth-style
// response with no peaking at the corner.
func NewLowShelf(hz, dBGain, q float64) *LowShelf {
	f := &LowShelf{}
	f.SetParams(hz, dBGain, q)
	return f
}

func (f *LowShelf) SetParams(hz, dBGain, q float64) {
	if hz < 10 {
		hz = 10
	}
	if hz > float64(SampleRate)/2-100 {
		hz = float64(SampleRate)/2 - 100
	}
	A := math.Pow(10, dBGain/40)
	w0 := 2 * math.Pi * hz / float64(SampleRate)
	cosw0 := math.Cos(w0)
	sinw0 := math.Sin(w0)
	alpha := sinw0 / (2 * q)
	sqrtAalpha2 := 2 * math.Sqrt(A) * alpha

	b0 := A * ((A + 1) - (A-1)*cosw0 + sqrtAalpha2)
	b1 := 2 * A * ((A - 1) - (A+1)*cosw0)
	b2 := A * ((A + 1) - (A-1)*cosw0 - sqrtAalpha2)
	a0 := (A + 1) + (A-1)*cosw0 + sqrtAalpha2
	a1 := -2 * ((A - 1) + (A+1)*cosw0)
	a2 := (A + 1) + (A-1)*cosw0 - sqrtAalpha2

	f.b0 = b0 / a0
	f.b1 = b1 / a0
	f.b2 = b2 / a0
	f.a1 = a1 / a0
	f.a2 = a2 / a0
}

// Tick processes one sample (Direct Form II Transposed).
func (f *LowShelf) Tick(x float64) float64 {
	y := f.b0*x + f.z1
	f.z1 = f.b1*x - f.a1*y + f.z2
	f.z2 = f.b2*x - f.a2*y
	return y
}

// HighShelf is a biquad high-shelf EQ filter (RBJ cookbook). Boosts or cuts
// frequencies above the corner; leaves lower frequencies unaffected. Useful
// for "air" / sparkle on top of bright instruments, or de-essing.
type HighShelf struct {
	b0, b1, b2, a1, a2 float64
	z1, z2             float64
}

func NewHighShelf(hz, dBGain, q float64) *HighShelf {
	f := &HighShelf{}
	f.SetParams(hz, dBGain, q)
	return f
}

func (f *HighShelf) SetParams(hz, dBGain, q float64) {
	if hz < 10 {
		hz = 10
	}
	if hz > float64(SampleRate)/2-100 {
		hz = float64(SampleRate)/2 - 100
	}
	A := math.Pow(10, dBGain/40)
	w0 := 2 * math.Pi * hz / float64(SampleRate)
	cosw0 := math.Cos(w0)
	sinw0 := math.Sin(w0)
	alpha := sinw0 / (2 * q)
	sqrtAalpha2 := 2 * math.Sqrt(A) * alpha

	b0 := A * ((A + 1) + (A-1)*cosw0 + sqrtAalpha2)
	b1 := -2 * A * ((A - 1) + (A+1)*cosw0)
	b2 := A * ((A + 1) + (A-1)*cosw0 - sqrtAalpha2)
	a0 := (A + 1) - (A-1)*cosw0 + sqrtAalpha2
	a1 := 2 * ((A - 1) - (A+1)*cosw0)
	a2 := (A + 1) - (A-1)*cosw0 - sqrtAalpha2

	f.b0 = b0 / a0
	f.b1 = b1 / a0
	f.b2 = b2 / a0
	f.a1 = a1 / a0
	f.a2 = a2 / a0
}

func (f *HighShelf) Tick(x float64) float64 {
	y := f.b0*x + f.z1
	f.z1 = f.b1*x - f.a1*y + f.z2
	f.z2 = f.b2*x - f.a2*y
	return y
}

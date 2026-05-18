package synth

import "math"

// Oversampler renders a callback at N× sample rate and decimates back via
// two cascaded one-pole IIR low-pass filters (~12 dB/oct rolloff per stage,
// 24 dB/oct total). Adequate at 16× where the alias band is far above
// audible range.
//
// Usage:
//
//	over := synth.NewOversampler(16)
//	y := over.Process(func() float64 { return naiveSawTick() })
type Oversampler struct {
	factor   int
	lp1, lp2 float64 // one-pole filter states
	coeff    float64 // IIR coefficient (α)
}

// NewOversampler creates an Oversampler with the given integer upsampling
// factor. Panics if factor < 2.
func NewOversampler(factor int) *Oversampler {
	if factor < 2 {
		panic("synth: Oversampler factor must be > 1")
	}
	// Cutoff at 45% of the original sample rate (0.45 / factor of the
	// oversampled rate). One-pole IIR α = 1 - exp(-2π·fc).
	fc := 0.45 / float64(factor)
	alpha := 1 - math.Exp(-2*math.Pi*fc)
	return &Oversampler{factor: factor, coeff: alpha}
}

// Process calls sampleAtRate factor times, filters each output through two
// cascaded one-pole IIR LP filters, and returns the last filtered value.
func (o *Oversampler) Process(sampleAtRate func() float64) float64 {
	α := o.coeff
	β := 1 - α
	var last float64
	for i := 0; i < o.factor; i++ {
		x := sampleAtRate()
		// First one-pole stage.
		o.lp1 = α*x + β*o.lp1
		// Second one-pole stage.
		o.lp2 = α*o.lp1 + β*o.lp2
		last = o.lp2
	}
	return last
}

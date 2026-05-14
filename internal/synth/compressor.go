package synth

import "math"

// Compressor is a soft-knee dynamic range compressor that follows the signal
// envelope and reduces gain above a threshold. The output is mono — for
// stereo, run a linked pair where both channels share the same gain reduction
// (use NewStereoCompressor below).
//
// All thresholds and gains are in dB. The envelope follower uses simple
// exponential attack/release smoothing.
type Compressor struct {
	thresholdDB float64
	ratio       float64
	kneeDB      float64
	makeupLin   float64

	attackCoef  float64 // smoothing coefficient per sample
	releaseCoef float64

	env float64 // envelope follower state (linear, peak-following)
}

// NewCompressor builds a compressor.
//   thresholdDB: signal above this gets compressed (e.g. -12)
//   ratio:       compression ratio above the threshold (e.g. 3.0 = 3:1)
//   attackMs:    how fast the compressor reacts to peaks (e.g. 10)
//   releaseMs:   how fast it lets go after a peak (e.g. 200)
//   kneeDB:      width of the soft transition around the threshold (e.g. 6)
//   makeupDB:    fixed makeup gain applied after compression (e.g. 4)
func NewCompressor(thresholdDB, ratio, attackMs, releaseMs, kneeDB, makeupDB float64) *Compressor {
	c := &Compressor{
		thresholdDB: thresholdDB,
		ratio:       ratio,
		kneeDB:      kneeDB,
		makeupLin:   math.Pow(10, makeupDB/20),
	}
	c.attackCoef = math.Exp(-1.0 / (attackMs * 0.001 * float64(SampleRate)))
	c.releaseCoef = math.Exp(-1.0 / (releaseMs * 0.001 * float64(SampleRate)))
	return c
}

// Gain returns the gain (linear) that should be applied to a sample with the
// given absolute value. Updates the internal envelope.
func (c *Compressor) Gain(abs float64) float64 {
	// Envelope follower (peak-following with exponential attack and release).
	if abs > c.env {
		c.env = c.attackCoef*(c.env-abs) + abs
	} else {
		c.env = c.releaseCoef*(c.env-abs) + abs
	}
	envDB := 20 * math.Log10(c.env+1e-12)

	// Soft-knee compression curve in dB.
	overshoot := envDB - c.thresholdDB
	var reductionDB float64
	switch {
	case overshoot < -c.kneeDB/2:
		// Well below threshold: no compression.
		reductionDB = 0
	case overshoot < c.kneeDB/2:
		// Inside the soft knee: quadratic ramp.
		x := overshoot + c.kneeDB/2
		reductionDB = (1.0/c.ratio - 1.0) * (x * x) / (2 * c.kneeDB)
	default:
		// Above the knee: linear compression.
		reductionDB = (1.0/c.ratio - 1.0) * overshoot
	}
	return math.Pow(10, reductionDB/20) * c.makeupLin
}

// Tick processes one mono sample.
func (c *Compressor) Tick(x float64) float64 {
	abs := x
	if abs < 0 {
		abs = -abs
	}
	return x * c.Gain(abs)
}

// StereoCompressor links two channels so they receive the same gain reduction
// (driven by the louder of the two). This prevents stereo image from
// shifting when one channel is louder than the other.
type StereoCompressor struct {
	inner *Compressor
}

// NewStereoCompressor builds a stereo-linked compressor; arguments are the
// same as NewCompressor.
func NewStereoCompressor(thresholdDB, ratio, attackMs, releaseMs, kneeDB, makeupDB float64) *StereoCompressor {
	return &StereoCompressor{
		inner: NewCompressor(thresholdDB, ratio, attackMs, releaseMs, kneeDB, makeupDB),
	}
}

// Tick processes one stereo sample pair with linked gain reduction.
func (c *StereoCompressor) Tick(l, r float64) (float64, float64) {
	absL := l
	if absL < 0 {
		absL = -absL
	}
	absR := r
	if absR < 0 {
		absR = -absR
	}
	abs := absL
	if absR > abs {
		abs = absR
	}
	g := c.inner.Gain(abs)
	return l * g, r * g
}

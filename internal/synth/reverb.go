package synth

// Reverb is a Freeverb-style reverberator: 8 parallel lowpass-feedback comb
// filters summed into 4 series allpass filters. The high comb count gives a
// dense impulse response (no obvious "flutter"); the lowpass in each comb's
// feedback makes the tail darken naturally over time, like a real room.
//
// Mono in, mono out. For stereo, run two instances with slightly offset
// delay lengths so the left and right tails decorrelate.
type Reverb struct {
	combs    [8]lpfCombFilter
	allpass  [4]schroederAllpass
	wet, dry float64
}

// Freeverb's tuned delay lengths in samples (at 44.1 kHz).
var freeverbCombDelays = [8]int{1116, 1188, 1277, 1356, 1422, 1491, 1557, 1617}
var freeverbAllpassDelays = [4]int{556, 441, 341, 225}

// NewReverb builds a reverb with sensible defaults. `wet` is 0..1.
func NewReverb(wet float64) *Reverb {
	return newReverbAt(wet, 0)
}

// NewReverbRight is the same topology with delay lengths shifted by ~23
// samples (Freeverb's stereo spread) so left/right channels decorrelate.
func NewReverbRight(wet float64) *Reverb {
	return newReverbAt(wet, 23)
}

func newReverbAt(wet float64, spread int) *Reverb {
	if wet < 0 {
		wet = 0
	}
	if wet > 1 {
		wet = 1
	}
	r := &Reverb{wet: wet, dry: 1 - wet}
	const feedback = 0.84 // controls tail length
	const damping = 0.2   // lowpass coefficient in comb feedback (0=bright, 1=dark)
	for i, d := range freeverbCombDelays {
		r.combs[i] = newLPFCombFilter(d+spread, feedback, damping)
	}
	for i, d := range freeverbAllpassDelays {
		r.allpass[i] = newSchroederAllpass(d+spread, 0.5)
	}
	return r
}

// Tick processes one sample.
func (r *Reverb) Tick(x float64) float64 {
	var sum float64
	for i := range r.combs {
		sum += r.combs[i].tick(x)
	}
	sum *= 0.125
	for i := range r.allpass {
		sum = r.allpass[i].tick(sum)
	}
	return x*r.dry + sum*r.wet
}

// lpfCombFilter is a feedback comb with a one-pole lowpass in the feedback
// loop. The lowpass progressively darkens the tail, mimicking the way real
// rooms lose high frequencies first.
//
//	y[n] = buf[n-N]
//	z[n] = (1-d) * y[n] + d * z[n-1]      (lowpass smoothing)
//	buf[n] = x[n] + g * z[n]
type lpfCombFilter struct {
	buf      []float64
	w        int
	gain     float64
	damping  float64
	store    float64
}

func newLPFCombFilter(delay int, gain, damping float64) lpfCombFilter {
	if delay < 1 {
		delay = 1
	}
	return lpfCombFilter{
		buf:     make([]float64, delay),
		gain:    gain,
		damping: damping,
	}
}

func (c *lpfCombFilter) tick(x float64) float64 {
	out := c.buf[c.w]
	// One-pole lowpass on the comb's output before feeding back.
	c.store = out*(1-c.damping) + c.store*c.damping
	c.buf[c.w] = x + c.store*c.gain
	c.w++
	if c.w >= len(c.buf) {
		c.w = 0
	}
	return out
}

// schroederAllpass is a Schroeder allpass:
//   y[n] = -g*x[n] + (1-g²)*v[n-N], where v[n] = x[n] + g*v[n-N]
//
// Implemented as:
//   delayed = buf[n-N]
//   v       = x + g*delayed
//   buf[n]  = v
//   y       = delayed - g*v
type schroederAllpass struct {
	buf  []float64
	w    int
	gain float64
}

func newSchroederAllpass(delay int, gain float64) schroederAllpass {
	if delay < 1 {
		delay = 1
	}
	return schroederAllpass{buf: make([]float64, delay), gain: gain}
}

func (a *schroederAllpass) tick(x float64) float64 {
	delayed := a.buf[a.w]
	v := x + delayed*a.gain
	a.buf[a.w] = v
	a.w++
	if a.w >= len(a.buf) {
		a.w = 0
	}
	return delayed - v*a.gain
}

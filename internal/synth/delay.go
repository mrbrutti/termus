package synth

// Delay is a fractional-sample-free delay line with feedback and wet/dry mix.
type Delay struct {
	buf      []float64
	w        int     // write index
	delay    int     // delay in samples
	feedback float64 // 0..1
	mix      float64 // 0 = dry only, 1 = wet only
}

func NewDelay(seconds, feedback, mix float64) *Delay {
	n := int(seconds * float64(SampleRate))
	if n < 1 {
		n = 1
	}
	if feedback < 0 {
		feedback = 0
	}
	if feedback > 0.95 {
		feedback = 0.95
	}
	if mix < 0 {
		mix = 0
	}
	if mix > 1 {
		mix = 1
	}
	return &Delay{
		buf:      make([]float64, n),
		delay:    n,
		feedback: feedback,
		mix:      mix,
	}
}

// Tick advances by one sample. Input is `x`, output is wet/dry mix.
func (d *Delay) Tick(x float64) float64 {
	wet := d.buf[d.w]
	d.buf[d.w] = x + wet*d.feedback
	d.w = (d.w + 1) % d.delay
	return x*(1-d.mix) + wet*d.mix
}

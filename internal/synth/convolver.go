package synth

// RealtimeConvolver is the common interface satisfied by both the direct
// time-domain Convolver and the FFT-based FFTConvolver. Code that wants to
// hold "some kind of convolver" (e.g. the SF2 master bus) uses this so it
// can swap implementations based on IR length.
type RealtimeConvolver interface {
	Tick(x float64) float64
}

// Convolver is a real-time direct (time-domain) convolution filter. Use it to
// imprint the early-reflections of a real space onto an input signal by
// loading an impulse response (IR) WAV file. For best CPU behavior keep IRs
// short (≤ 6,615 samples ≈ 150 ms at 44.1 kHz); longer IRs are still correct
// but CPU cost scales linearly with IR length and may not stay real-time on
// modest hardware. For full-tail convolution (1+ seconds) an FFT-based
// partitioned implementation is the right tool — not done here.
type Convolver struct {
	ir      []float64
	history []float64 // circular delay line, length == len(ir)
	w       int       // write index into history
}

// NewConvolver builds a convolver from the given impulse response. The IR
// is copied so the caller can reuse the slice. Returns nil for an empty IR.
func NewConvolver(ir []float64) *Convolver {
	if len(ir) == 0 {
		return nil
	}
	c := &Convolver{
		ir:      make([]float64, len(ir)),
		history: make([]float64, len(ir)),
	}
	copy(c.ir, ir)
	return c
}

// Tick processes one sample. Direct convolution: output[t] is the dot
// product of the IR with the most recent N input samples, where N = len(ir).
func (c *Convolver) Tick(x float64) float64 {
	c.history[c.w] = x
	c.w++
	if c.w >= len(c.history) {
		c.w = 0
	}
	// Sum: for k in 0..N-1, ir[k] * history[(w - 1 - k) mod N].
	// Equivalent: walk history backwards from the newest sample we just wrote.
	var sum float64
	hlen := len(c.history)
	idx := c.w - 1
	if idx < 0 {
		idx += hlen
	}
	for k := 0; k < len(c.ir); k++ {
		sum += c.ir[k] * c.history[idx]
		idx--
		if idx < 0 {
			idx += hlen
		}
	}
	return sum
}

// SyntheticRoomIR returns a small synthetic impulse response that approximates
// a small-room early-reflection pattern. Useful for sanity-testing the
// convolver without a real IR file, and tasteful as a subtle "in the room"
// effect even on its own.
//
// The IR is a direct path (impulse at t=0) followed by a set of decaying
// reflections at slightly-randomized times within ~50 ms, all with amplitude
// proportional to 1/distance to mimic spherical spreading.
func SyntheticRoomIR(durationSec float64) []float64 {
	n := int(durationSec * float64(SampleRate))
	if n < 8 {
		n = 8
	}
	ir := make([]float64, n)
	ir[0] = 1.0 // direct sound
	// Five reflections at musically-spaced delays, each at decreasing amplitude.
	reflections := []struct {
		delayMs float64
		amp     float64
	}{
		{8.3, 0.50},
		{17.1, 0.36},
		{23.7, 0.28},
		{31.5, 0.20},
		{42.8, 0.14},
	}
	for _, r := range reflections {
		idx := int(r.delayMs * 0.001 * float64(SampleRate))
		if idx >= 0 && idx < n {
			ir[idx] += r.amp
		}
	}
	return ir
}

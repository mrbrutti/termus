package synth

import "math"

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

// SyntheticHallIR returns a ~1.5-second concert-hall-like impulse response:
// early reflections in the first ~50 ms, then an exponentially decaying
// noise tail with a high-frequency rolloff to simulate air absorption. Use
// it with a long convolver (FFT-based — automatic for IRs > 1024 samples).
func SyntheticHallIR(seed int64) []float64 {
	return synthSpaceIR(1.5, 60.0, 4.8, 0.08, seed)
}

// SyntheticCathedralIR returns a ~3.5-second cathedral-like impulse response.
// Long tail with denser early reflections and slower decay than the hall.
func SyntheticCathedralIR(seed int64) []float64 {
	return synthSpaceIR(3.5, 80.0, 6.5, 0.05, seed)
}

// SyntheticPlateIR returns a ~2-second plate-reverb-like IR. Plates have a
// very dense (smooth) tail with brighter coloration than a real room — we
// approximate this with a faster decay rate at high frequencies.
func SyntheticPlateIR(seed int64) []float64 {
	return synthSpaceIR(2.0, 200.0, 5.5, 0.12, seed)
}

// synthSpaceIR builds an IR that looks like real-room/hall measurements:
//   - direct impulse at t=0
//   - exponentially decaying noise tail with rate decayCoef (higher = faster)
//   - one-pole lowpass with `damping` per sample to roll off highs over time
//
// densityHz controls how many discrete "reflections" per second populate the
// tail. Higher = smoother/denser; lower = sparser/more obviously echoy.
func synthSpaceIR(durationSec, densityHz, decayCoef, damping float64, seed int64) []float64 {
	n := int(durationSec * float64(SampleRate))
	if n < 64 {
		n = 64
	}
	ir := make([]float64, n)
	ir[0] = 1.0

	// Random tail driven by deterministic seed so the same preset name
	// always produces the same IR. We don't want to import math/rand here
	// (this is the synth package — leaf module), so we use a tiny xorshift
	// instead.
	rng := uint64(seed)
	if rng == 0 {
		rng = 1
	}
	rand01 := func() float64 {
		rng ^= rng << 13
		rng ^= rng >> 7
		rng ^= rng << 17
		// Use top 53 bits for a uniform [0, 1).
		return float64(rng>>11) / float64(1<<53)
	}

	earlyEnd := int(0.04 * float64(SampleRate)) // 40 ms of early reflections
	// Early-reflection cluster.
	for t := 0; t < earlyEnd && t < n; t++ {
		// Sparse impulses with amplitude ~ random in [0, 1) * spherical falloff.
		if rand01() < densityHz/float64(SampleRate) {
			ir[t] += (2*rand01() - 1) * 0.45 * (1 - float64(t)/float64(earlyEnd))
		}
	}
	// Diffuse decaying-noise tail.
	var lp float64 // one-pole lowpass state
	for t := earlyEnd; t < n; t++ {
		if rand01() < densityHz/float64(SampleRate) {
			tail := (2*rand01() - 1) * 0.35
			env := math.Exp(-decayCoef * float64(t-earlyEnd) / float64(SampleRate))
			// One-pole lowpass to roll off high frequencies as time progresses.
			lp = lp*damping + tail*env*(1-damping)
			ir[t] += lp
		} else {
			// Decay the lowpass even when no new reflection fires so the
			// tail trails off smoothly rather than going dead silent.
			lp *= damping
			ir[t] += lp * 0.3
		}
	}
	return ir
}

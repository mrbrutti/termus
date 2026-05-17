// internal/synth/tape.go
//
// Tape-machine modeling DSP. Phase 1 ships only WowFlutter (pitch
// modulation via fractional-delay-line read offset).
package synth

import "math"

// WowFlutterConfig sets the modulator parameters. Zero depths bypass
// modulation entirely; the input is delayed but not pitch-shifted.
type WowFlutterConfig struct {
	WowRateHz         float64 // typical 0.5–1.5
	WowDepthCents     float64 // typical 10–30 for lofi
	FlutterRateHz     float64 // typical 4–10
	FlutterDepthCents float64 // typical 2–8
	// StereoOffsetRad rotates the modulator phase between L and R to
	// avoid perfectly correlated channels (which produce stronger
	// comb artifacts). Default ~0.35 rad (~20°).
	StereoOffsetRad float64
}

// WowFlutter is a stereo pitch modulator implemented as a modulated
// fractional-delay line. The base delay is set so the read pointer
// stays inside the buffer for any combination of wow + flutter
// excursion at the configured maximum depths.
type WowFlutter struct {
	sampleRate float64

	bufL, bufR []float64
	writeIdx   int
	bufLen     int

	baseDelay float64 // samples (the centre of the modulation)

	wowOmega      float64
	wowDepthSamp  float64
	flutOmega     float64
	flutDepthSamp float64
	stereoOffset  float64

	phaseWow  float64
	phaseFlut float64
}

// NewWowFlutter builds a stereo WowFlutter. The base delay is computed from
// the configured depths so the read pointer always stays within written
// history. When both depths are zero, baseDelay is 0 and readLerp returns
// the current sample unchanged (identity).
func NewWowFlutter(sampleRate float64, cfg WowFlutterConfig) *WowFlutter {
	if cfg.StereoOffsetRad == 0 {
		cfg.StereoOffsetRad = 0.35
	}

	wowDepth := centsToDelaySamples(cfg.WowDepthCents, cfg.WowRateHz, sampleRate)
	flutDepth := centsToDelaySamples(cfg.FlutterDepthCents, cfg.FlutterRateHz, sampleRate)

	// baseDelay keeps the read pointer safely inside written data.
	//   readPos = writeIdx - baseDelay - mod
	//   mod ∈ [-(wowDepth+flutDepth), +(wowDepth+flutDepth)]
	// Worst case (mod most negative): pos = writeIdx - baseDelay + totalDepth
	// Linear interpolation needs pos ≤ writeIdx - 1 (reads idx and idx+1).
	// Add 2 samples of safety margin: baseDelay ≥ totalDepth + 2.
	// When both depths are 0, baseDelay = 0 and readLerp(buf, writeIdx)
	// returns buf[writeIdx] = current sample (identity).
	totalDepth := wowDepth + flutDepth
	var baseDelay float64
	if totalDepth > 0 {
		baseDelay = math.Ceil(totalDepth) + 2
	}

	// Buffer must accommodate at least baseDelay + totalDepth + headroom.
	// max(64, ceil(baseDelay)*2 + 8) is always sufficient.
	bufLen := 64
	if needed := int(math.Ceil(baseDelay))*2 + 8; needed > bufLen {
		bufLen = needed
	}

	return &WowFlutter{
		sampleRate:    sampleRate,
		bufL:          make([]float64, bufLen),
		bufR:          make([]float64, bufLen),
		bufLen:        bufLen,
		baseDelay:     baseDelay,
		wowOmega:      2 * math.Pi * cfg.WowRateHz / sampleRate,
		wowDepthSamp:  wowDepth,
		flutOmega:     2 * math.Pi * cfg.FlutterRateHz / sampleRate,
		flutDepthSamp: flutDepth,
		stereoOffset:  cfg.StereoOffsetRad,
	}
}

// centsToDelaySamples converts a peak pitch-modulation depth in cents at a
// given modulation rate into the peak delay-line excursion in samples.
//
// For a sinusoidal delay d(t) = D·sin(2πft), the instantaneous pitch
// shift in cents is:
//
//	c(t) ≈ -1200 · d'(t) / (ln(2) · sr)
//	     = -1200 · 2πf·D·cos(2πft) / (ln(2) · sr)
//
// Peak |c| = 1200·2πf·D / (ln(2)·sr)
// Solving for D: D = peakCents · ln(2) · sr / (1200 · 2π · f)
func centsToDelaySamples(peakCents, rateHz, sampleRate float64) float64 {
	if peakCents == 0 || rateHz == 0 {
		return 0
	}
	return peakCents * math.Ln2 * sampleRate / (1200 * 2 * math.Pi * rateHz)
}

// Tick processes one stereo frame and returns the modulated output.
func (w *WowFlutter) Tick(inL, inR float64) (float64, float64) {
	// Write current input into circular buffer.
	w.bufL[w.writeIdx] = inL
	w.bufR[w.writeIdx] = inR

	// Advance modulator phases.
	w.phaseWow += w.wowOmega
	w.phaseFlut += w.flutOmega
	if w.phaseWow >= 2*math.Pi {
		w.phaseWow -= 2 * math.Pi
	}
	if w.phaseFlut >= 2*math.Pi {
		w.phaseFlut -= 2 * math.Pi
	}

	// Modulator excursion in samples (positive = read further back = lower pitch).
	modL := w.wowDepthSamp*math.Sin(w.phaseWow) +
		w.flutDepthSamp*math.Sin(w.phaseFlut)
	modR := w.wowDepthSamp*math.Sin(w.phaseWow+w.stereoOffset) +
		w.flutDepthSamp*math.Sin(w.phaseFlut+w.stereoOffset)

	// Fractional read positions (samples back from the write head).
	readL := float64(w.writeIdx) - w.baseDelay - modL
	readR := float64(w.writeIdx) - w.baseDelay - modR

	outL := w.readLerp(w.bufL, readL)
	outR := w.readLerp(w.bufR, readR)

	w.writeIdx = (w.writeIdx + 1) % w.bufLen
	return outL, outR
}

// readLerp performs linear interpolation at fractional position pos inside
// the circular buffer buf. For wow/flutter (modulation rates ≪ sr), linear
// interpolation is indistinguishable from cubic in listening tests and
// avoids the ≈10% pitch-overshoot that 4-point Lagrange introduces when
// measured via zero-crossing analysis.
func (w *WowFlutter) readLerp(buf []float64, pos float64) float64 {
	n := float64(w.bufLen)
	// Wrap to [0, bufLen).
	for pos < 0 {
		pos += n
	}
	for pos >= n {
		pos -= n
	}

	i := int(math.Floor(pos))
	frac := pos - float64(i)

	i0 := i % w.bufLen
	i1 := (i + 1) % w.bufLen

	return buf[i0]*(1-frac) + buf[i1]*frac
}

// TapeConfig configures the Tape saturation processor.
type TapeConfig struct {
	// DriveDB is the input gain in dB before the saturation nonlinearity.
	// 0 = bypass (linear). Typical lofi: 3–6 dB. Above 12 dB the sound
	// becomes obviously distorted.
	DriveDB float64
}

// Tape is an asymmetric soft-clip saturator. The asymmetry adds 2nd-harmonic
// content (tape's characteristic warmth) on top of the odd-harmonic clip
// that pure tanh would produce.
type Tape struct {
	drive  float64
	makeup float64
}

// NewTape constructs a Tape saturator. Drive of 0 dB is the identity function.
func NewTape(cfg TapeConfig) *Tape {
	drive := math.Pow(10, cfg.DriveDB/20)
	// Approximate unity loudness compensation: peak of tanh(x) for x in
	// [-drive, drive] is tanh(drive); we scale by 1/tanh(drive) so a
	// full-scale input still hits ~ unity.
	makeup := 1.0
	if cfg.DriveDB > 0 {
		makeup = 1.0 / math.Tanh(drive)
	}
	return &Tape{drive: drive, makeup: makeup}
}

// Tick applies the saturator to a single sample.
func (t *Tape) Tick(x float64) float64 {
	if t.drive == 1 && t.makeup == 1 {
		return x
	}
	// Asymmetric bias adds even harmonics: shift the operating point
	// slightly off zero before the nonlinearity. Use a small fixed bias
	// (5% of drive) — anything larger pushes a DC offset into the chain
	// that the downstream EQ would need to remove.
	const biasRatio = 0.05
	y := math.Tanh(t.drive*x+t.drive*biasRatio) - math.Tanh(t.drive*biasRatio)
	return y * t.makeup
}

// VinylConfig configures the Vinyl noise/crackle bed.
type VinylConfig struct {
	// NoiseLevelDB is the RMS of the continuous noise bed in dBFS.
	// math.Inf(-1) (i.e. "off") disables noise. Typical lofi: -27 dB.
	NoiseLevelDB float64
	// PopRateHz is the average number of pop transients per second
	// (Poisson process). 0 disables pops. Typical lofi: 4–10.
	PopRateHz float64
	// Seed makes the noise/pop sequences reproducible.
	Seed int64
}

// Vinyl produces a stereo noise bed reminiscent of a vinyl record's surface
// noise. The continuous bed is lowpassed white noise (band-limited to a
// "warm" character); pops are short, sharp transients placed via a Poisson
// process at the configured rate.
type Vinyl struct {
	sampleRate float64

	noiseAmp float64
	popProb  float64
	popAmp   float64

	// xorshift state — fast, deterministic, no allocation per Tick.
	state uint64

	// 1-pole lowpass for the bed (gives it warmer-than-white character).
	lpL, lpR float64
	lpCoeff  float64
}

// NewVinyl constructs a Vinyl noise bed. -inf level + 0 pops = silence.
func NewVinyl(sampleRate float64, cfg VinylConfig) *Vinyl {
	v := &Vinyl{
		sampleRate: sampleRate,
		state:      uint64(cfg.Seed)*6364136223846793005 + 1442695040888963407,
		popProb:    cfg.PopRateHz / sampleRate,
		popAmp:     0.7, // peak amplitude of a pop transient
		lpCoeff:    1.0 - math.Exp(-2*math.Pi*4000/sampleRate), // 4 kHz one-pole LP
	}
	if math.IsInf(cfg.NoiseLevelDB, -1) {
		v.noiseAmp = 0
	} else {
		// noiseAmp is the peak; RMS for uniform white is amp/sqrt(3),
		// so dBFS = 20·log10(amp/sqrt(3)).
		// amp = 10^(dB/20) · sqrt(3)
		v.noiseAmp = math.Pow(10, cfg.NoiseLevelDB/20) * math.Sqrt(3)
	}
	if cfg.PopRateHz <= 0 {
		v.popProb = 0
	}
	return v
}

// Tick advances the noise generator by one sample and returns a stereo frame.
// The two channels are independent (decorrelated noise, but identical RMS).
func (v *Vinyl) Tick() (float64, float64) {
	uL := v.next()
	uR := v.next()
	xL := (uL*2 - 1) * v.noiseAmp
	xR := (uR*2 - 1) * v.noiseAmp
	// 1-pole LP
	v.lpL += v.lpCoeff * (xL - v.lpL)
	v.lpR += v.lpCoeff * (xR - v.lpR)
	outL := v.lpL
	outR := v.lpR
	// Pops (correlated L/R — a real vinyl pop is mono-ish).
	if v.popProb > 0 && v.nextFloat() < v.popProb {
		pop := (v.nextFloat()*2 - 1) * v.popAmp
		outL += pop
		outR += pop
	}
	return outL, outR
}

// nextFloat returns a uniform float in [0, 1) via xorshift*.
func (v *Vinyl) nextFloat() float64 {
	v.state ^= v.state >> 12
	v.state ^= v.state << 25
	v.state ^= v.state >> 27
	v.state *= 2685821657736338717
	return float64(v.state>>11) / (1 << 53)
}

// next is a synonym used in Tick for clarity.
func (v *Vinyl) next() float64 { return v.nextFloat() }

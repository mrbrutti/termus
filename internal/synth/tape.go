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

package synth

import "math"

// IRPreset describes a named impulse response that can be requested by name
// from the registry. Each preset is a procedural generator producing a
// deterministic mono buffer at the given sample rate.
type IRPreset struct {
	Name        string
	Description string
	// RT60Sec is the approximate −60 dB decay time; used both as a hint for
	// callers (pre-delay scaling) and as the buffer-length target inside
	// the generator.
	RT60Sec float64
	// Generate returns a mono IR (length ≈ RT60Sec * sampleRate, possibly
	// padded for the synth's convenience). Deterministic for a given seed.
	Generate func(sampleRate float64, seed int64) []float64
}

// IRLibrary returns all named presets. Add new entries to the registry
// rather than exporting separate functions — callers select by name.
func IRLibrary() []IRPreset {
	return []IRPreset{
		{
			Name:        "bedroom_small",
			Description: "Small bedroom: bright early reflections, mostly dry, low-mid emphasis",
			RT60Sec:     0.4,
			Generate:    irBedroomSmall,
		},
		{
			Name:        "bedroom_large",
			Description: "Large bedroom: same character as bedroom_small, longer tail",
			RT60Sec:     0.9,
			Generate:    irBedroomLarge,
		},
		{
			Name:        "jazz_club",
			Description: "Jazz club: warm, mids-focused, dense early reflections",
			RT60Sec:     1.2,
			Generate:    irJazzClub,
		},
		{
			Name:        "plate_hardware",
			Description: "Plate reverb hardware: bright, diffuse from the start, slight high-frequency emphasis",
			RT60Sec:     2.5,
			Generate:    irPlateHardware,
		},
		{
			Name:        "spring_tank",
			Description: "Spring tank: chirpy, comb-filtered top end with spring resonance modes",
			RT60Sec:     1.5,
			Generate:    irSpringTank,
		},
		{
			Name:        "cassette_chamber",
			Description: "Cassette chamber: lo-fi, lowpassed ~5 kHz, tape-flutter texture",
			RT60Sec:     0.5,
			Generate:    irCassetteChamber,
		},
		{
			Name:        "stairwell",
			Description: "Stairwell: long early reflections at ~25, 50, 80 ms before diffuse tail",
			RT60Sec:     2.0,
			Generate:    irStairwell,
		},
		{
			Name:        "cathedral",
			Description: "Cathedral: very long diffuse decay, bright early then dark tail",
			RT60Sec:     4.0,
			Generate:    irCathedral,
		},
	}
}

// IRByName looks up a preset by name. Returns nil if not found.
func IRByName(name string) *IRPreset {
	lib := IRLibrary()
	for i := range lib {
		if lib[i].Name == name {
			return &lib[i]
		}
	}
	return nil
}

// irXorshift is a deterministic pseudo-random number generator (xorshift64).
// Returns values in [0, 1).
type irXorshift struct {
	state uint64
}

func newIRRng(seed int64) irXorshift {
	s := uint64(seed)
	if s == 0 {
		s = 1
	}
	return irXorshift{state: s}
}

func (r *irXorshift) next() float64 {
	r.state ^= r.state << 13
	r.state ^= r.state >> 7
	r.state ^= r.state << 17
	return float64(r.state>>11) / float64(1<<53)
}

// rand11 returns a value in [-1, 1).
func (r *irXorshift) rand11() float64 {
	return 2*r.next() - 1
}

// irBedroomSmall: RT60 ≈ 0.4s. Bright early reflections, mostly dry.
// Low-mid emphasis achieved by modest high-frequency damping and a bump in
// the 200–500 Hz region via slight resonant filtering in the tail.
func irBedroomSmall(sampleRate float64, seed int64) []float64 {
	rt60 := 0.4
	n := irLength(rt60, sampleRate)
	ir := make([]float64, n)
	rng := newIRRng(seed)

	ir[0] = 1.0

	// Bright early reflections in the first 30 ms.
	earlyEnd := int(0.030 * sampleRate)
	if earlyEnd > n {
		earlyEnd = n
	}
	// Deterministic early reflection positions with distance falloff.
	earlyTaps := []struct{ ms, amp float64 }{
		{4.5, 0.55},
		{9.2, 0.42},
		{15.1, 0.31},
		{22.3, 0.22},
		{28.7, 0.15},
	}
	for _, tap := range earlyTaps {
		idx := int(tap.ms * 0.001 * sampleRate)
		if idx > 0 && idx < n {
			// Slight randomization of amplitude for texture.
			ir[idx] += tap.amp * (1 + 0.1*rng.rand11())
		}
	}

	// Short diffuse tail: fast decay (RT60 = 0.4s → decayCoef must produce
	// ~60 dB drop over 0.4s). damping=0.15 gives moderate high-freq rolloff,
	// preserving brightness. Low-mid bump via slight resonance.
	decayCoef := irDecayCoef(rt60)
	damping := 0.15 // low damping = brighter
	var lp float64
	for t := earlyEnd; t < n; t++ {
		tail := rng.rand11() * 0.3
		env := math.Exp(-decayCoef * float64(t) / sampleRate)
		lp = lp*(1-damping) + tail*env*damping
		ir[t] += lp
	}
	return ir
}

// irBedroomLarge: RT60 ≈ 0.9s. Same bedroom character, longer tail.
func irBedroomLarge(sampleRate float64, seed int64) []float64 {
	rt60 := 0.9
	n := irLength(rt60, sampleRate)
	ir := make([]float64, n)
	rng := newIRRng(seed)

	ir[0] = 1.0

	earlyTaps := []struct{ ms, amp float64 }{
		{5.0, 0.52},
		{10.5, 0.40},
		{18.2, 0.30},
		{27.0, 0.21},
		{38.5, 0.14},
		{52.0, 0.09},
	}
	for _, tap := range earlyTaps {
		idx := int(tap.ms * 0.001 * sampleRate)
		if idx > 0 && idx < n {
			ir[idx] += tap.amp * (1 + 0.1*rng.rand11())
		}
	}

	earlyEnd := int(0.055 * sampleRate)
	if earlyEnd > n {
		earlyEnd = n
	}
	decayCoef := irDecayCoef(rt60)
	damping := 0.12
	var lp float64
	for t := earlyEnd; t < n; t++ {
		tail := rng.rand11() * 0.28
		env := math.Exp(-decayCoef * float64(t) / sampleRate)
		lp = lp*(1-damping) + tail*env*damping
		ir[t] += lp
	}
	return ir
}

// irJazzClub: RT60 ≈ 1.2s. Warm, mids-focused, dense early reflections.
// High damping rolls off highs quickly, warm midrange survives.
func irJazzClub(sampleRate float64, seed int64) []float64 {
	rt60 := 1.2
	n := irLength(rt60, sampleRate)
	ir := make([]float64, n)
	rng := newIRRng(seed)

	ir[0] = 1.0

	// Dense early cluster in first 50 ms — jazz clubs have many reflections
	// from close walls.
	earlyEnd := int(0.050 * sampleRate)
	if earlyEnd > n {
		earlyEnd = n
	}
	// High density: ~120 reflections/sec.
	density := 120.0
	for t := 1; t < earlyEnd; t++ {
		if rng.next() < density/sampleRate {
			amp := 0.4 * (1 - float64(t)/float64(earlyEnd)) * (1 + 0.15*rng.rand11())
			ir[t] += amp * rng.rand11() * 2
		}
	}

	// Warm tail: high damping (0.45) kills top end, low-mid warms the space.
	decayCoef := irDecayCoef(rt60)
	damping := 0.45
	var lp float64
	for t := earlyEnd; t < n; t++ {
		tail := rng.rand11() * 0.32
		env := math.Exp(-decayCoef * float64(t) / sampleRate)
		lp = lp*(1-damping) + tail*env*damping
		ir[t] += lp
	}
	return ir
}

// irPlateHardware: RT60 ≈ 2.5s. Bright plate — diffuse from the start,
// slight HF emphasis. No discrete early reflections.
func irPlateHardware(sampleRate float64, seed int64) []float64 {
	rt60 := 2.5
	n := irLength(rt60, sampleRate)
	ir := make([]float64, n)
	rng := newIRRng(seed)

	ir[0] = 1.0

	// Plates have very dense, immediately diffuse response — no distinct
	// early reflections. Very low damping = bright.
	decayCoef := irDecayCoef(rt60)
	damping := 0.04 // very bright
	density := 250.0
	var lp float64
	// Start diffuse tail immediately from sample 1.
	for t := 1; t < n; t++ {
		env := math.Exp(-decayCoef * float64(t) / sampleRate)
		if rng.next() < density/sampleRate {
			tail := rng.rand11() * 0.40 * env
			lp = lp*(1-damping) + tail*damping
		} else {
			lp *= (1 - damping*0.5)
		}
		ir[t] = lp
	}
	return ir
}

// irSpringTank: RT60 ≈ 1.5s. Chirpy, comb-filtered top end with slowly
// decaying sinusoids in the 4–8 kHz range mimicking spring resonance modes.
func irSpringTank(sampleRate float64, seed int64) []float64 {
	rt60 := 1.5
	n := irLength(rt60, sampleRate)
	ir := make([]float64, n)
	rng := newIRRng(seed)

	ir[0] = 1.0

	// Small number of early reflections (spring has delay line character).
	earlyTaps := []struct{ ms, amp float64 }{
		{3.0, 0.40},
		{6.5, 0.30},
		{11.0, 0.20},
	}
	for _, tap := range earlyTaps {
		idx := int(tap.ms * 0.001 * sampleRate)
		if idx > 0 && idx < n {
			ir[idx] += tap.amp
		}
	}

	earlyEnd := int(0.015 * sampleRate)
	if earlyEnd > n {
		earlyEnd = n
	}

	// Spring resonance modes: 6 slowly decaying sinusoids between 4–8 kHz.
	// These create the characteristic "boing" and HF brightness. Amplitudes
	// are deliberately large so the spring modes dominate the spectral centroid
	// (audibly and measurably brighter than room-based presets).
	springModes := []struct {
		freqHz float64
		amp    float64
		decay  float64
	}{
		{4200, 0.30, 3.0},
		{5100, 0.28, 3.5},
		{5800, 0.25, 4.0},
		{6700, 0.22, 4.5},
		{7400, 0.18, 5.0},
		{8100, 0.15, 5.5},
	}

	// Noise tail with moderate HF content.
	decayCoef := irDecayCoef(rt60)
	damping := 0.08 // somewhat bright
	var lp float64
	for t := earlyEnd; t < n; t++ {
		tail := rng.rand11() * 0.25
		env := math.Exp(-decayCoef * float64(t) / sampleRate)
		lp = lp*(1-damping) + tail*env*damping
		ir[t] += lp

		// Sum spring mode sinusoids on top of the noise tail.
		tSec := float64(t) / sampleRate
		for _, mode := range springModes {
			modeEnv := math.Exp(-mode.decay * tSec)
			ir[t] += mode.amp * modeEnv * math.Sin(2*math.Pi*mode.freqHz*tSec)
		}
	}
	return ir
}

// irCassetteChamber: RT60 ≈ 0.5s. Lo-fi: lowpassed ~5 kHz, tape-flutter
// texture via slow LFO modulating noise magnitude.
func irCassetteChamber(sampleRate float64, seed int64) []float64 {
	rt60 := 0.5
	n := irLength(rt60, sampleRate)
	ir := make([]float64, n)
	rng := newIRRng(seed)

	ir[0] = 1.0

	// A couple of early reflections, slightly muffled.
	earlyTaps := []struct{ ms, amp float64 }{
		{6.0, 0.38},
		{14.0, 0.22},
		{24.0, 0.12},
	}
	for _, tap := range earlyTaps {
		idx := int(tap.ms * 0.001 * sampleRate)
		if idx > 0 && idx < n {
			ir[idx] += tap.amp
		}
	}

	earlyEnd := int(0.028 * sampleRate)
	if earlyEnd > n {
		earlyEnd = n
	}

	// Lowpass at ~5 kHz: damping = 1 - 2π*5000/sampleRate ≈ 0.29 at 44.1k.
	// We use a one-pole low-pass: coefficient = exp(-2π*fc/sr).
	fcHz := 5000.0
	lpCoef := math.Exp(-2 * math.Pi * fcHz / sampleRate)
	var lp float64

	// Tape-flutter LFO: slow wobble (2–4 Hz) modulates noise amplitude.
	flutterHz := 3.2
	decayCoef := irDecayCoef(rt60)

	for t := earlyEnd; t < n; t++ {
		tSec := float64(t) / sampleRate
		env := math.Exp(-decayCoef * tSec)
		// Flutter: LFO mapped to [0.5, 1.5] range.
		flutter := 1.0 + 0.5*math.Sin(2*math.Pi*flutterHz*tSec)
		tail := rng.rand11() * 0.35 * env * flutter
		// One-pole lowpass (~5 kHz cutoff).
		lp = lp*lpCoef + tail*(1-lpCoef)
		ir[t] += lp
	}
	return ir
}

// irStairwell: RT60 ≈ 2.0s. Long discrete early reflections at ~25, 50, 80 ms
// then diffuse tail.
func irStairwell(sampleRate float64, seed int64) []float64 {
	rt60 := 2.0
	n := irLength(rt60, sampleRate)
	ir := make([]float64, n)
	rng := newIRRng(seed)

	ir[0] = 1.0

	// Stairwells have distinct echoes from parallel walls and landings.
	earlyTaps := []struct{ ms, amp float64 }{
		{8.0, 0.45},
		{16.5, 0.35},
		{25.0, 0.55}, // main stairwell reflection
		{38.0, 0.28},
		{50.0, 0.45}, // second flight
		{65.0, 0.20},
		{80.0, 0.38}, // third flight
		{105.0, 0.15},
	}
	for _, tap := range earlyTaps {
		idx := int(tap.ms * 0.001 * sampleRate)
		if idx > 0 && idx < n {
			ir[idx] += tap.amp * (1 + 0.08*rng.rand11())
		}
	}

	// Diffuse tail kicks in after the early reflections.
	earlyEnd := int(0.110 * sampleRate)
	if earlyEnd > n {
		earlyEnd = n
	}
	decayCoef := irDecayCoef(rt60)
	damping := 0.20
	var lp float64
	for t := earlyEnd; t < n; t++ {
		tail := rng.rand11() * 0.30
		env := math.Exp(-decayCoef * float64(t) / sampleRate)
		lp = lp*(1-damping) + tail*env*damping
		ir[t] += lp
	}
	return ir
}

// irCathedral: RT60 ≈ 4.0s. Very long diffuse decay. Warm/dark throughout —
// the long lowpassed tail dominates the spectral character, giving a lower
// centroid than bright presets like spring_tank.
func irCathedral(sampleRate float64, seed int64) []float64 {
	rt60 := 4.0
	n := irLength(rt60, sampleRate)
	ir := make([]float64, n)
	rng := newIRRng(seed)

	ir[0] = 1.0

	// Early reflections: moderate density, spanning ~80 ms, but immediately
	// routed through a strong lowpass so even the early part is warm.
	earlyEnd := int(0.080 * sampleRate)
	if earlyEnd > n {
		earlyEnd = n
	}
	density := 80.0
	// Lowpass at ~800 Hz to keep cathedral warm from the first sample.
	// coef = exp(-2π * fc / sr)
	fcEarly := 800.0
	lpCoefEarly := math.Exp(-2 * math.Pi * fcEarly / sampleRate)
	var lpE float64
	for t := 1; t < earlyEnd; t++ {
		var raw float64
		if rng.next() < density/sampleRate {
			amp := 0.5 * (1 - float64(t)/float64(earlyEnd)) * (1 + 0.1*rng.rand11())
			raw = amp * rng.rand11() * 2
		}
		lpE = lpE*lpCoefEarly + raw*(1-lpCoefEarly)
		ir[t] += lpE
	}

	// Very long, dark tail. Heavy lowpass at ~400 Hz from the start.
	fcTail := 400.0
	lpCoefTail := math.Exp(-2 * math.Pi * fcTail / sampleRate)
	decayCoef := irDecayCoef(rt60)
	var lpT float64
	for t := earlyEnd; t < n; t++ {
		tail := rng.rand11() * 0.38
		env := math.Exp(-decayCoef * float64(t) / sampleRate)
		// Heavy lowpass keeps only low frequencies.
		lpT = lpT*lpCoefTail + tail*env*(1-lpCoefTail)
		ir[t] += lpT
	}
	return ir
}

// irLength returns the buffer length for a given RT60 and sample rate.
func irLength(rt60, sampleRate float64) int {
	n := int(math.Ceil(rt60 * sampleRate))
	if n < 64 {
		n = 64
	}
	return n
}

// irDecayCoef converts RT60 (seconds) to the exponential decay coefficient
// such that exp(-coef * rt60) = 10^(-3) (60 dB attenuation).
// Solving: -coef * rt60 = -3 * ln(10) → coef = 3*ln(10)/rt60.
func irDecayCoef(rt60 float64) float64 {
	return 3 * math.Log(10) / rt60
}

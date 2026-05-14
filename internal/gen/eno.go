package gen

import (
	"math"
	"math/rand"

	"github.com/mrbrutti/termus/internal/synth"
)

// Compile-time assertion that *Eno implements Algorithm.
var _ Algorithm = (*Eno)(nil)

// Eno is the v2 "Music for Airports" algorithm: layered pad-bell voices on
// incommensurate loop periods over an evolving sine-additive drone, all
// soaked in a Freeverb-style stereo reverb.
type Eno struct {
	rng      *rand.Rand
	rootMidi int
	voices   []*padBellVoice // slow pad-bell bed
	leads    []*padBellVoice // faster lead voices on top of the bed
	drone    *sineDrone
	revL     *synth.Reverb
	revR     *synth.Reverb
	delayL   *synth.Delay
	delayR   *synth.Delay
	t        int64
}

// loopPeriods are pairwise-near-irrational; voices realign on the order of hours.
var loopPeriods = []float64{7.0, 11.0, 13.3, 17.7, 23.1}

// leadPeriods are shorter — these voices carry the melodic motion that sits
// on top of the slow pad bed.
var leadPeriods = []float64{4.3, 5.9}

// scaleMinor is natural minor: root, +2, +3, +5, +7, +8, +10 semitones.
var scaleMinor = []int{0, 2, 3, 5, 7, 8, 10}

// NewEno constructs an Eno generator. Caller must call Seed before Next.
func NewEno() *Eno { return &Eno{} }

func (e *Eno) Name() string { return "eno-drift" }

func (e *Eno) Seed(s int64) {
	e.rng = rand.New(rand.NewSource(s)) //nolint:gosec
	e.rootMidi = 36 + e.rng.Intn(12) // C2..B2

	// Slow pad bed.
	e.voices = make([]*padBellVoice, len(loopPeriods))
	for i, period := range loopPeriods {
		notes := make([]int, 2+e.rng.Intn(3)) // 2..4 notes per phrase
		for j := range notes {
			degree := scaleMinor[e.rng.Intn(len(scaleMinor))]
			octave := 12 * (2 + e.rng.Intn(3))
			notes[j] = e.rootMidi + degree + octave
		}
		e.voices[i] = newPadBellVoice(period, notes, e.rng.Float64(), e.rng.Float64())
	}
	// Faster lead voices on top. Quicker attack, more notes per phrase,
	// higher register — these carry the perceived melody.
	e.leads = make([]*padBellVoice, len(leadPeriods))
	for i, period := range leadPeriods {
		notes := make([]int, 4+e.rng.Intn(3)) // 4..6 notes per phrase
		for j := range notes {
			degree := scaleMinor[e.rng.Intn(len(scaleMinor))]
			octave := 12 * (3 + e.rng.Intn(2)) // higher register: +36, +48
			notes[j] = e.rootMidi + degree + octave
		}
		v := newPadBellVoice(period, notes, e.rng.Float64(), e.rng.Float64())
		v.makeLead()
		e.leads[i] = v
	}
	e.drone = newSineDrone(e.rootMidi)
	e.revL = synth.NewReverb(0.45)
	e.revR = synth.NewReverbRight(0.45)
	e.delayL = synth.NewDelay(0.290, 0.30, 0.18)
	e.delayR = synth.NewDelay(0.410, 0.30, 0.18)
	e.t = 0
}

func (e *Eno) Next(left, right []float64) {
	for i := range left {
		var l, r float64
		// Slow pad bed.
		for vi, v := range e.voices {
			s := v.tick(e.t)
			if vi%2 == 0 {
				l += s * 0.55
				r += s * 0.30
			} else {
				l += s * 0.30
				r += s * 0.55
			}
		}
		// Lead voices — quieter so they sit on top without dominating.
		for vi, v := range e.leads {
			s := v.tick(e.t)
			if vi%2 == 0 {
				l += s * 0.55
				r += s * 0.40
			} else {
				l += s * 0.40
				r += s * 0.55
			}
		}
		d := e.drone.tick(e.t)
		l += d
		r += d
		// Reverb (per-channel for stereo width).
		l = e.revL.Tick(l)
		r = e.revR.Tick(r)
		// Light cross-delay for additional motion.
		newL := e.delayL.Tick(r)
		newR := e.delayR.Tick(l)
		l, r = newL, newR
		// Master soft-clip.
		left[i] = synth.SoftClip(l * 2.0)
		right[i] = synth.SoftClip(r * 2.0)
		e.t++
	}
}

// --- padBellVoice: 4 sine partials + vibrato + envelope-modulated lowpass ---

type padBellVoice struct {
	periodSamples int64
	phaseOffset   int64
	notes         []int

	// 4 partials at f, 2.01f, 3.02f, 4.04f (slightly stretched for shimmer).
	osc     [4]*synth.Oscillator
	partAmp [4]float64

	vibrato *synth.Oscillator
	vibAmt  float64 // pitch modulation depth (in semitones)
	vibSeed float64 // randomized vibrato phase

	env *synth.Envelope
	lp  *synth.Lowpass

	curNote  int
	gateOn   bool
	baseFreq float64
	leadMode bool

	ctrlPhase int // sub-sampled filter envelope updates
}

// makeLead retunes this voice for the lead role: faster attack so individual
// notes pop out as melodic events instead of blending into the pad wash. Also
// brighter filter (the cutoff envelope opens wider) so leads cut through.
func (v *padBellVoice) makeLead() {
	v.env = synth.NewEnvelope(0.4, 0.6, 0.45, 2.5)
	v.lp = synth.NewLowpass(800, 0.7)
	v.leadMode = true
}

func newPadBellVoice(periodSec float64, notes []int, phase01, vibSeed float64) *padBellVoice {
	v := &padBellVoice{
		periodSamples: int64(periodSec * float64(synth.SampleRate)),
		notes:         notes,
		env:           synth.NewEnvelope(1.6, 0.8, 0.55, 4.0),
		lp:            synth.NewLowpass(400, 0.7),
		vibrato:       synth.NewOscillator(synth.WaveSine),
		vibAmt:        0.06, // ~0.35% pitch wobble — felt-piano warmth
		vibSeed:       vibSeed,
		curNote:       -1,
		partAmp:       [4]float64{1.00, 0.45, 0.20, 0.10},
	}
	for i := range v.osc {
		v.osc[i] = synth.NewOscillator(synth.WaveSine)
	}
	v.vibrato.SetFrequency(0.6 + 0.5*vibSeed) // 0.6..1.1 Hz
	v.phaseOffset = int64(phase01 * float64(v.periodSamples))
	return v
}

// partialRatios are slightly stretched harmonics — close to integer multiples
// but not exact, which gives a bell-like shimmer instead of an organ-like pure
// harmonic stack.
var partialRatios = [4]float64{1.000, 2.005, 3.012, 4.022}

func (v *padBellVoice) tick(t int64) float64 {
	p := (t + v.phaseOffset) % v.periodSamples
	slot := int(p) * len(v.notes) / int(v.periodSamples)
	noteOnSample := int64(slot) * v.periodSamples / int64(len(v.notes))
	pos := p - noteOnSample

	if slot != v.curNote {
		v.curNote = slot
		v.baseFreq = midiToHz(v.notes[slot])
		for i := range v.osc {
			v.osc[i].SetFrequency(v.baseFreq * partialRatios[i])
		}
		v.env.Gate(true)
		v.gateOn = true
	}
	slotLen := v.periodSamples / int64(len(v.notes))
	if v.gateOn && pos > slotLen*65/100 {
		v.env.Gate(false)
		v.gateOn = false
	}

	envVal := v.env.Tick()
	// Vibrato: each partial's pitch wobbles around base.
	vib := v.vibrato.Tick()
	pitchFactor := math.Exp2(vib * v.vibAmt / 12.0)
	for i := range v.osc {
		v.osc[i].SetFrequency(v.baseFreq * partialRatios[i] * pitchFactor)
	}

	// Sum partials.
	var s float64
	for i := range v.osc {
		s += v.partAmp[i] * v.osc[i].Tick()
	}
	s *= envVal

	// Filter envelope: cutoff opens with envelope.
	v.ctrlPhase++
	if v.ctrlPhase >= 32 {
		v.ctrlPhase = 0
		var cutoff float64
		if v.leadMode {
			cutoff = 800 + 3200*envVal // brighter for leads
		} else {
			cutoff = 350 + 2600*envVal
		}
		v.lp.SetParams(cutoff, 0.7)
	}
	s = v.lp.Tick(s)
	if v.leadMode {
		return s * 0.16
	}
	return s * 0.22
}

// --- sineDrone: 5-partial additive bed with very slow movement ---

type sineDrone struct {
	osc  [5]*synth.Oscillator
	amp  [5]float64
	vib  *synth.Oscillator
	lp   *synth.Lowpass
	lfo  *synth.Oscillator
	base float64

	ctrlPhase int
}

func newSineDrone(rootMidi int) *sineDrone {
	d := &sineDrone{
		base: midiToHz(rootMidi),
		amp:  [5]float64{1.0, 0.6, 0.32, 0.16, 0.08},
	}
	for i := range d.osc {
		d.osc[i] = synth.NewOscillator(synth.WaveSine)
	}
	d.osc[0].SetFrequency(d.base * 0.5) // sub-octave
	d.osc[1].SetFrequency(d.base * 1.002) // root, slight detune up
	d.osc[2].SetFrequency(d.base * 2.0 * 0.998) // octave, detune down
	d.osc[3].SetFrequency(d.base * 3.0)
	d.osc[4].SetFrequency(d.base * 4.0 * 1.001)
	d.vib = synth.NewOscillator(synth.WaveSine)
	d.vib.SetFrequency(0.07) // very slow swell
	d.lp = synth.NewLowpass(450, 0.6)
	d.lfo = synth.NewOscillator(synth.WaveSine)
	d.lfo.SetFrequency(0.03) // 33s cycle on the filter
	return d
}

func (d *sineDrone) tick(_ int64) float64 {
	var s float64
	for i := range d.osc {
		s += d.amp[i] * d.osc[i].Tick()
	}
	// Slow amplitude swell on the whole drone.
	swell := 0.7 + 0.3*d.vib.Tick()
	s *= swell

	// Update filter cutoff every 64 samples (~690 Hz control rate).
	d.ctrlPhase++
	if d.ctrlPhase >= 64 {
		d.ctrlPhase = 0
		cutoff := 500 + 350*d.lfo.Tick()
		d.lp.SetParams(cutoff, 0.6)
	}
	s = d.lp.Tick(s)
	return s * 0.06
}

// --- helpers ---

func midiToHz(midi int) float64 {
	return 440.0 * math.Exp2((float64(midi)-69.0)/12.0)
}

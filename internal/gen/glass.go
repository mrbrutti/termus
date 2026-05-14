package gen

import (
	"math"
	"math/rand"

	"github.com/mrbrutti/termus/internal/synth"
)

// Compile-time assertion that *Glass implements Algorithm.
var _ Algorithm = (*Glass)(nil)

// Glass is the "Aphex Twin SAW II" style: bright FM bell voices on short
// incommensurate loop periods, with light reverb. Crystalline and a bit
// dissonant — better for late-night focus than warm pad music.
type Glass struct {
	rng      *rand.Rand
	rootMidi int
	voices   []*fmBellVoice
	revL     *synth.Reverb
	revR     *synth.Reverb
	t        int64
}

// glassLoopPeriods are shorter than eno's so notes happen more frequently.
var glassLoopPeriods = []float64{3.7, 5.1, 6.7, 8.3, 10.1, 13.7}

// scalePentatonicMinor: root, +3, +5, +7, +10 (minor pentatonic — sounds good
// with sparser melodic material than the full minor scale).
var scalePentatonicMinor = []int{0, 3, 5, 7, 10}

// NewGlass constructs a Glass generator. Caller must call Seed before Next.
func NewGlass() *Glass { return &Glass{} }

func (g *Glass) Name() string { return "glass-fm" }

func (g *Glass) Seed(s int64) {
	g.rng = rand.New(rand.NewSource(s)) //nolint:gosec
	g.rootMidi = 48 + g.rng.Intn(7) // C3..F#3 — brighter starting register
	g.voices = make([]*fmBellVoice, len(glassLoopPeriods))
	for i, period := range glassLoopPeriods {
		notes := make([]int, 1+g.rng.Intn(2)) // 1..2 notes per phrase — sparse
		for j := range notes {
			degree := scalePentatonicMinor[g.rng.Intn(len(scalePentatonicMinor))]
			octave := 12 * (1 + g.rng.Intn(3)) // +12, +24, +36 from a high root
			notes[j] = g.rootMidi + degree + octave
		}
		g.voices[i] = newFMBellVoice(period, notes, g.rng.Float64())
	}
	g.revL = synth.NewReverb(0.30)
	g.revR = synth.NewReverbRight(0.30)
	g.t = 0
}

func (g *Glass) Next(left, right []float64) {
	for i := range left {
		var l, r float64
		// Pan voices across stereo field using their index.
		for vi, v := range g.voices {
			s := v.tick(g.t)
			pan := float64(vi) / float64(len(g.voices)-1) // 0..1
			l += s * (1 - pan*0.6)
			r += s * (0.4 + pan*0.6)
		}
		l = g.revL.Tick(l)
		r = g.revR.Tick(r)
		left[i] = synth.SoftClip(l * 2.0)
		right[i] = synth.SoftClip(r * 2.0)
		g.t++
	}
}

// --- fmBellVoice: 2-operator FM (carrier + modulator) for bell-like tones ---
//
// FM synthesis: the modulator's output (a sine at m*f) modulates the
// instantaneous phase of the carrier (a sine at f). With modulation index I,
// the carrier output is sin(2π·f·t + I·sin(2π·m·f·t)). This produces a series
// of harmonic partials whose amplitudes follow Bessel functions of I — a very
// efficient way to generate bell timbres.

type fmBellVoice struct {
	periodSamples int64
	phaseOffset   int64
	notes         []int

	carrierPhase float64
	modPhase     float64
	carrierInc   float64
	modInc       float64

	modIndex float64 // FM depth (radians)

	env *synth.Envelope

	curNote int
	gateOn  bool
}

func newFMBellVoice(periodSec float64, notes []int, phase01 float64) *fmBellVoice {
	v := &fmBellVoice{
		periodSamples: int64(periodSec * float64(synth.SampleRate)),
		notes:         notes,
		// Bell envelopes: fast attack, slow decay/release, low sustain.
		env:      synth.NewEnvelope(0.005, 1.2, 0.15, 2.8),
		curNote:  -1,
		modIndex: 4.0, // moderate modulation depth = bell-ish, not gong
	}
	v.phaseOffset = int64(phase01 * float64(v.periodSamples))
	return v
}

func (v *fmBellVoice) tick(t int64) float64 {
	p := (t + v.phaseOffset) % v.periodSamples
	slot := int(p) * len(v.notes) / int(v.periodSamples)
	noteOnSample := int64(slot) * v.periodSamples / int64(len(v.notes))
	pos := p - noteOnSample

	if slot != v.curNote {
		v.curNote = slot
		f := midiToHz(v.notes[slot])
		v.carrierInc = f / float64(synth.SampleRate)
		// Modulator at 1.4× carrier — produces inharmonic bell partials.
		v.modInc = (f * 1.4) / float64(synth.SampleRate)
		v.env.Gate(true)
		v.gateOn = true
	}
	slotLen := v.periodSamples / int64(len(v.notes))
	if v.gateOn && pos > slotLen*30/100 {
		v.env.Gate(false)
		v.gateOn = false
	}

	// FM synthesis: modulator drives carrier phase.
	modOut := math.Sin(v.modPhase * 2 * math.Pi)
	envVal := v.env.Tick()
	// Modulation depth tracks the envelope so the timbre brightens on attack
	// and mellows on release — a key part of FM bell realism.
	carrierVal := math.Sin(v.carrierPhase*2*math.Pi + v.modIndex*envVal*modOut)

	v.carrierPhase += v.carrierInc
	if v.carrierPhase >= 1 {
		v.carrierPhase -= 1
	}
	v.modPhase += v.modInc
	if v.modPhase >= 1 {
		v.modPhase -= 1
	}

	return carrierVal * envVal * 0.20
}

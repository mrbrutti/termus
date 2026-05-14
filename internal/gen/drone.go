package gen

import (
	"math/rand"

	"github.com/mrbrutti/termus/internal/synth"
)

// Compile-time assertion that *Drone implements Algorithm.
var _ Algorithm = (*Drone)(nil)

// Drone is the "Stars of the Lid" style: a few voices each holding a single
// note for ~40–80 seconds, swelling in and out over very slow envelopes, very
// large reverb. Almost no melodic motion — pure evolving harmonic texture.
type Drone struct {
	rng      *rand.Rand
	rootMidi int
	voices   []*droneVoice
	revL     *synth.Reverb
	revR     *synth.Reverb
	t        int64
}

// droneLoopPeriods are very long — each voice holds one note for tens of
// seconds before retriggering with a different note from the scale.
var droneLoopPeriods = []float64{47.0, 61.0, 73.0, 89.0}

// NewDrone constructs a Drone generator. Caller must call Seed before Next.
func NewDrone() *Drone { return &Drone{} }

func (d *Drone) Name() string { return "drone-bed" }

func (d *Drone) Seed(s int64) {
	d.rng = rand.New(rand.NewSource(s)) //nolint:gosec
	d.rootMidi = 24 + d.rng.Intn(7) // C1..F#1 — really low for the bed
	d.voices = make([]*droneVoice, len(droneLoopPeriods))
	for i, period := range droneLoopPeriods {
		// Each voice cycles through 2..4 notes from the minor scale over
		// its (long) period. Notes are placed in the 2nd–3rd octave above
		// the root so the texture is rich but not muddy.
		notes := make([]int, 2+d.rng.Intn(3))
		for j := range notes {
			degree := scaleMinor[d.rng.Intn(len(scaleMinor))]
			octave := 12 * (1 + d.rng.Intn(3)) // +12, +24, +36
			notes[j] = d.rootMidi + degree + octave
		}
		d.voices[i] = newDroneVoice(period, notes, d.rng.Float64())
	}
	d.revL = synth.NewReverb(0.65)
	d.revR = synth.NewReverbRight(0.65)
	d.t = 0
}

func (d *Drone) Next(left, right []float64) {
	for i := range left {
		var l, r float64
		for vi, v := range d.voices {
			s := v.tick(d.t)
			// Mild stereo spread.
			if vi%2 == 0 {
				l += s * 0.6
				r += s * 0.4
			} else {
				l += s * 0.4
				r += s * 0.6
			}
		}
		l = d.revL.Tick(l)
		r = d.revR.Tick(r)
		left[i] = synth.SoftClip(l * 2.2)
		right[i] = synth.SoftClip(r * 2.2)
		d.t++
	}
}

// --- droneVoice: 5 partial sines, very slow ADSR, slow individual vibratos ---

type droneVoice struct {
	periodSamples int64
	phaseOffset   int64
	notes         []int

	osc [5]*synth.Oscillator
	amp [5]float64
	vib [5]*synth.Oscillator // per-partial vibrato

	env *synth.Envelope
	lp  *synth.Lowpass

	curNote  int
	gateOn   bool
	baseFreq float64

	ctrlPhase int
}

func newDroneVoice(periodSec float64, notes []int, phase01 float64) *droneVoice {
	v := &droneVoice{
		periodSamples: int64(periodSec * float64(synth.SampleRate)),
		notes:         notes,
		env:           synth.NewEnvelope(8.0, 2.0, 0.75, 10.0), // 8s attack, 10s release
		lp:            synth.NewLowpass(900, 0.6),
		curNote:       -1,
		amp:           [5]float64{1.0, 0.55, 0.28, 0.14, 0.07},
	}
	for i := range v.osc {
		v.osc[i] = synth.NewOscillator(synth.WaveSine)
		v.vib[i] = synth.NewOscillator(synth.WaveSine)
		// Each partial gets a slightly different slow vibrato rate.
		v.vib[i].SetFrequency(0.08 + 0.04*float64(i))
	}
	v.phaseOffset = int64(phase01 * float64(v.periodSamples))
	return v
}

// dronePartialRatios are exact integer harmonics for a clean drone bed.
var dronePartialRatios = [5]float64{1.0, 2.0, 3.0, 4.0, 5.0}

func (v *droneVoice) tick(t int64) float64 {
	p := (t + v.phaseOffset) % v.periodSamples
	slot := int(p) * len(v.notes) / int(v.periodSamples)
	noteOnSample := int64(slot) * v.periodSamples / int64(len(v.notes))
	pos := p - noteOnSample

	if slot != v.curNote {
		v.curNote = slot
		v.baseFreq = midiToHz(v.notes[slot])
		for i := range v.osc {
			v.osc[i].SetFrequency(v.baseFreq * dronePartialRatios[i])
		}
		v.env.Gate(true)
		v.gateOn = true
	}
	slotLen := v.periodSamples / int64(len(v.notes))
	// Release at 70% of slot so attack+release overlap with the next note.
	if v.gateOn && pos > slotLen*70/100 {
		v.env.Gate(false)
		v.gateOn = false
	}

	envVal := v.env.Tick()

	// Each partial has its own slow pitch wobble (a few cents) for life.
	v.ctrlPhase++
	if v.ctrlPhase >= 64 {
		v.ctrlPhase = 0
		for i := range v.osc {
			cents := 0.04 * v.vib[i].Tick() // ~0.04 semitones
			factor := 1.0 + cents*0.0578     // small-angle approx of 2^(cents/12)
			v.osc[i].SetFrequency(v.baseFreq * dronePartialRatios[i] * factor)
		}
	}

	var s float64
	for i := range v.osc {
		s += v.amp[i] * v.osc[i].Tick()
	}
	s *= envVal
	s = v.lp.Tick(s)
	return s * 0.18
}

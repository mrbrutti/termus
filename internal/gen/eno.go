package gen

import (
	"math"
	"math/rand"

	"github.com/mrbrutti/termus/internal/synth"
)

// Compile-time assertion that *Eno implements Algorithm.
var _ Algorithm = (*Eno)(nil)

// Eno is the v1 algorithm: multiple short melodic phrases on incommensurate
// loop periods over a slow filter-swept drone pad, with stereo cross-delay.
type Eno struct {
	rng      *rand.Rand
	rootMidi int // root note as MIDI number
	voices   []*enoVoice
	drone    *enoDrone
	delayL   *synth.Delay
	delayR   *synth.Delay
	t        int64 // sample index since Seed
}

// loopPeriods are pairwise-near-irrational; voices realign on the order of hours.
var loopPeriods = []float64{7.0, 11.0, 13.3, 17.7, 23.1}

// scaleMinor is natural minor: root, +2, +3, +5, +7, +8, +10 semitones.
var scaleMinor = []int{0, 2, 3, 5, 7, 8, 10}

// NewEno constructs an Eno generator. Caller must call Seed before Next.
func NewEno() *Eno {
	return &Eno{}
}

func (e *Eno) Name() string { return "eno-drift" }

func (e *Eno) Seed(s int64) {
	e.rng = rand.New(rand.NewSource(s)) //nolint:gosec
	// Pick a minor root in {C2..B2} = MIDI 36..47.
	e.rootMidi = 36 + e.rng.Intn(12)

	// Build N voices, one per loop period.
	e.voices = make([]*enoVoice, len(loopPeriods))
	for i, period := range loopPeriods {
		notes := make([]int, 1+e.rng.Intn(3)) // 1..3 notes per phrase
		for j := range notes {
			// Scale degree from the minor scale, then push up some octaves.
			degree := scaleMinor[e.rng.Intn(len(scaleMinor))]
			octave := 12 * (2 + e.rng.Intn(3)) // +24, +36, or +48 semitones
			notes[j] = e.rootMidi + degree + octave
		}
		e.voices[i] = newEnoVoice(period, notes, e.rng.Float64())
	}
	e.drone = newEnoDrone(e.rootMidi)
	e.delayL = synth.NewDelay(0.300, 0.25, 0.30)
	e.delayR = synth.NewDelay(0.420, 0.25, 0.30)
	e.t = 0
}

func (e *Eno) Next(left, right []float64) {
	for i := range left {
		var l, r float64
		// Sum voices, panned alternately for stereo spread.
		for vi, v := range e.voices {
			s := v.tick(e.t)
			if vi%2 == 0 {
				l += s * 0.7
				r += s * 0.3
			} else {
				l += s * 0.3
				r += s * 0.7
			}
		}
		// Drone pad in mono, low gain.
		d := e.drone.tick()
		l += d
		r += d
		// Stereo cross-delay (ping-pong): each channel's signal echoes through
		// the opposite-side delay line. This widens the stereo image and is the
		// canonical ambient-music space effect.
		newL := e.delayL.Tick(r)
		newR := e.delayR.Tick(l)
		l, r = newL, newR
		// Master gain into soft-clip. The pre-clip gain is set high enough
		// that average output sits comfortably around -12..-18 dBFS — quiet
		// enough for deep work, loud enough to actually be audible without
		// cranking system volume. tanh smoothly limits peaks below clipping.
		left[i] = synth.SoftClip(l * 2.5)
		right[i] = synth.SoftClip(r * 2.5)
		e.t++
	}
}

// --- voice ---

type enoVoice struct {
	periodSamples int64
	phaseOffset   int64
	notes         []int // MIDI notes
	osc1          *synth.Oscillator
	osc2          *synth.Oscillator
	osc3          *synth.Oscillator
	env           *synth.Envelope
	lp            *synth.Lowpass
	curNote       int
	gateOn        bool
}

func newEnoVoice(periodSec float64, notes []int, phase01 float64) *enoVoice {
	v := &enoVoice{
		periodSamples: int64(periodSec * float64(synth.SampleRate)),
		notes:         notes,
		osc1:          synth.NewOscillator(synth.WaveSine),
		osc2:          synth.NewOscillator(synth.WaveSaw),
		osc3:          synth.NewOscillator(synth.WaveSaw),
		env:           synth.NewEnvelope(1.8, 0.5, 0.6, 3.5),
		lp:            synth.NewLowpass(2000, 0.7),
		curNote:       -1,
	}
	v.phaseOffset = int64(phase01 * float64(v.periodSamples))
	return v
}

func (v *enoVoice) tick(t int64) float64 {
	// Phase within the loop period.
	p := (t + v.phaseOffset) % v.periodSamples
	// Which note slot are we in? Spread evenly across the period.
	slot := int(p) * len(v.notes) / int(v.periodSamples)
	noteOnSample := int64(slot) * v.periodSamples / int64(len(v.notes))
	pos := p - noteOnSample

	if slot != v.curNote {
		// New note: gate on, set frequencies.
		v.curNote = slot
		freq := midiToHz(v.notes[slot])
		v.osc1.SetFrequency(freq)
		v.osc2.SetFrequency(freq * 1.005)
		v.osc3.SetFrequency(freq * 0.995)
		v.env.Gate(true)
		v.gateOn = true
	}
	// Release after 60% of the slot's lifetime.
	slotLen := v.periodSamples / int64(len(v.notes))
	if v.gateOn && pos > slotLen*60/100 {
		v.env.Gate(false)
		v.gateOn = false
	}

	s := 0.7*v.osc1.Tick() + 0.15*v.osc2.Tick() + 0.15*v.osc3.Tick()
	s *= v.env.Tick()
	s = v.lp.Tick(s)
	return s * 0.25 // per-voice gain
}

// --- drone ---

type enoDrone struct {
	o1, o2 *synth.Oscillator
	lp     *synth.Lowpass
	lfo    *synth.Oscillator
}

func newEnoDrone(rootMidi int) *enoDrone {
	d := &enoDrone{
		o1:  synth.NewOscillator(synth.WaveSaw),
		o2:  synth.NewOscillator(synth.WaveSaw),
		lp:  synth.NewLowpass(500, 0.5),
		lfo: synth.NewOscillator(synth.WaveSine),
	}
	d.o1.SetFrequency(midiToHz(rootMidi) * 1.0025)    // root, slight detune up
	d.o2.SetFrequency(midiToHz(rootMidi+12) * 0.9975) // octave, slight detune down
	d.lfo.SetFrequency(0.05)
	return d
}

func (d *enoDrone) tick() float64 {
	// LFO maps [-1, 1] → cutoff [200, 800] Hz.
	lfo := d.lfo.Tick()
	cutoff := 500 + 300*lfo
	d.lp.SetParams(cutoff, 0.6)
	s := 0.5*d.o1.Tick() + 0.5*d.o2.Tick()
	s = d.lp.Tick(s)
	return s * 0.125 // ~-18 dB
}

// --- helpers ---

func midiToHz(midi int) float64 {
	return 440.0 * math.Exp2((float64(midi)-69.0)/12.0)
}

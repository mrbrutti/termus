package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

var _ Algorithm = (*SF2Markov)(nil)

// SF2Markov is the markov-melody algorithm rendered through a SoundFont.
// The transition-matrix-driven melodies sound especially "composed" on
// piano, where small intervallic motion is most audible. Strings layer
// underneath at lower velocity for warmth.
type SF2Markov struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
	rng  *rand.Rand

	samplesElapsed int64
	nextSwapAt     int64
}

func NewSF2Markov(sf *meltysynth.SoundFont) *SF2Markov {
	return &SF2Markov{sf: sf}
}

func (a *SF2Markov) Name() string { return "markov-sf2" }

func (a *SF2Markov) Seed(seedVal int64) {
	rng := rand.New(rand.NewSource(seedVal)) //nolint:gosec
	a.rng = rng
	rootMidi := 36 + rng.Intn(12)
	a.samplesElapsed = 0
	a.scheduleNextSwap()

	core, err := newSF2Core(a.sf, 3.2, seedVal)
	if err != nil {
		a.core = nil
		return
	}
	core.setProgram(0, 0)  // Acoustic Grand Piano (center)
	core.setProgram(1, 49) // Slow Strings         (left)
	core.setProgram(2, 71) // Clarinet              (right — solo voice)
	core.setProgram(3, 32) // Acoustic Bass         (low foundation, center)
	core.setPan(0, 64)
	core.setPan(1, 40)
	core.setPan(2, 90)
	core.setPan(3, 64)

	core.setReverbSend(0, 72)  // piano: medium
	core.setReverbSend(1, 100) // strings: deeply wet
	core.setReverbSend(2, 88)  // clarinet solo voice: wet for "concert hall" feel
	core.setReverbSend(3, 36)  // bass: dry
	core.setChorusSend(1, 40)

	for i, period := range markovLoopPeriods {
		count := 10 + rng.Intn(5)
		notes := markovWalk(rng, rootMidi, count)
		ch := int32(0)
		var vel int32 = 86
		if i == len(markovLoopPeriods)-1 {
			ch = 1
			vel = 52
		}
		root := rootMidi
		mutate := func(_ int, prev int) int {
			loc := findClosestScalePitch(prev, root, scaleMinor)
			nextDeg := nextMarkovDegree(rng, loc.degreeIdx)
			return root + scaleMinor[nextDeg] + loc.octaveOffset*12
		}
		core.addTrack(SF2Track{
			Channel: ch, Velocity: vel, Notes: notes,
			PeriodSec: period, Phase01: rng.Float64(),
			MutationRate: 0.12, MutateOne: mutate,
			VelocityJitter: 8, TimingJitterSec: 0.015,
		})
	}

	// Clarinet solo voice: sparse Markov phrases in a higher register.
	clarNotes := markovWalk(rng, rootMidi+12, 6)
	clarMutate := func(_ int, prev int) int {
		loc := findClosestScalePitch(prev, rootMidi+12, scaleMinor)
		nextDeg := nextMarkovDegree(rng, loc.degreeIdx)
		return (rootMidi + 12) + scaleMinor[nextDeg] + loc.octaveOffset*12
	}
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 70, Notes: clarNotes,
		PeriodSec: 23.0, Phase01: rng.Float64(),
		MutationRate: 0.18, MutateOne: clarMutate,
		VelocityJitter: 10, TimingJitterSec: 0.018,
	})

	// Soft brush snare on the GM drum channel — sparse hits to give the
	// "composed" Markov feel a hint of rhythmic structure without imposing
	// a beat. GM 39 is "Hand Clap"; GM 26 is "Snare 1"; we use GM 38 (snare)
	// at low velocity for a "brushed" feel via the standard kit.
	const drumCh = 9
	core.setProgram(drumCh, 0)
	core.setPan(drumCh, 64)
	core.setReverbSend(drumCh, 90) // wet for that "cathedral percussion" feel
	// 4 hits per 14-second cycle — roughly half-note feel at a slow tempo.
	brushNotes := []int{38, 38, 38, 38}
	core.addTrack(SF2Track{
		Channel: drumCh, Velocity: 36, Notes: brushNotes,
		PeriodSec: 14.0, Phase01: 0.25, // start on the offbeat
		VelocityJitter: 12, TimingJitterSec: 0.025,
	})

	// Acoustic bass: long, slow Markov walk one octave below the root.
	bassNotes := markovWalk(rng, rootMidi-12, 5)
	bassMutate := func(_ int, prev int) int {
		loc := findClosestScalePitch(prev, rootMidi-12, scaleMinor)
		nextDeg := nextMarkovDegree(rng, loc.degreeIdx)
		return (rootMidi - 12) + scaleMinor[nextDeg] + loc.octaveOffset*12
	}
	core.addTrack(SF2Track{
		Channel: 3, Velocity: 82, Notes: bassNotes,
		PeriodSec: 18.5, Phase01: rng.Float64(),
		MutationRate: 0.08, MutateOne: bassMutate,
		VelocityJitter: 6, TimingJitterSec: 0.012,
	})

	a.core = core
}

// markovWalk produces a Markov-chain sequence of MIDI notes using the same
// minorTransitions matrix as gen.Markov. Extracted here as a shared helper.
func markovWalk(rng *rand.Rand, rootMidi, count int) []int {
	degree := 0
	if rng.Float64() < 0.3 {
		degree = 4
	}
	octave := 12 * (2 + rng.Intn(3))
	notes := make([]int, count)
	notes[0] = rootMidi + scaleMinor[degree] + octave
	for i := 1; i < count; i++ {
		degree = nextMarkovDegree(rng, degree)
		switch r := rng.Float64(); {
		case r < 0.10:
			octave += 12
		case r < 0.18:
			octave -= 12
		}
		if octave < 12 {
			octave = 12
		}
		if octave > 60 {
			octave = 60
		}
		notes[i] = rootMidi + scaleMinor[degree] + octave
	}
	return notes
}

// SetReverbIR installs a convolution reverb on the master bus.
func (a *SF2Markov) SetReverbIR(ir []float64, wet float64) {
	if a.core != nil {
		a.core.setConvolutionIR(ir, wet)
	}
}

func (a *SF2Markov) Next(left, right []float64) {
	if a.core == nil {
		for i := range left {
			left[i] = 0
			right[i] = 0
		}
		return
	}
	a.core.renderInto(left, right)
	a.samplesElapsed += int64(len(left))
	if a.samplesElapsed >= a.nextSwapAt {
		a.swapOneInstrument()
		a.scheduleNextSwap()
	}
}

var markovChannelAlternatives = map[int32][]int32{
	0: {0, 1, 4, 5},        // Acoustic Grand (default), Bright, EP1, EP2
	1: {49, 48, 50, 51},    // Slow Strings (default), Strings 1, Synth Strings 1, 2
	2: {71, 68, 69, 64},    // Clarinet (default), Oboe, English Horn, Soprano Sax
	3: {32, 33, 87, 38},    // Acoustic Bass (default), Electric Bass, Lead Bass, Synth Bass 1
}

func (a *SF2Markov) scheduleNextSwap() {
	secs := 200.0 + 180.0*a.rng.Float64()
	a.nextSwapAt = a.samplesElapsed + int64(secs*44100)
}

func (a *SF2Markov) swapOneInstrument() {
	channels := []int32{0, 1, 2, 3}
	ch := channels[a.rng.Intn(len(channels))]
	a.core.programSwap(ch, markovChannelAlternatives[ch], a.rng)
}

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
}

func NewSF2Markov(sf *meltysynth.SoundFont) *SF2Markov {
	return &SF2Markov{sf: sf}
}

func (a *SF2Markov) Name() string { return "markov-sf2" }

func (a *SF2Markov) Seed(seedVal int64) {
	rng := rand.New(rand.NewSource(seedVal)) //nolint:gosec
	rootMidi := 36 + rng.Intn(12)

	core, err := newSF2Core(a.sf, 3.2)
	if err != nil {
		a.core = nil
		return
	}
	core.setProgram(0, 0)  // Acoustic Grand Piano
	core.setProgram(1, 49) // String Ensemble 2 (slow)

	// Same Markov-chain melody as gen.Markov.
	for i, period := range markovLoopPeriods {
		count := 10 + rng.Intn(5)
		notes := markovWalk(rng, rootMidi, count)
		// Piano voices on top, strings doubling lower voices.
		ch := int32(0)
		var vel int32 = 86
		if i == len(markovLoopPeriods)-1 {
			ch = 1
			vel = 52
		}
		core.addTrack(SF2Track{
			Channel: ch, Velocity: vel, Notes: notes,
			PeriodSec: period, Phase01: rng.Float64(),
		})
	}
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

func (a *SF2Markov) Next(left, right []float64) {
	if a.core == nil {
		for i := range left {
			left[i] = 0
			right[i] = 0
		}
		return
	}
	a.core.renderInto(left, right)
}

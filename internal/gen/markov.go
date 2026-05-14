package gen

import (
	"math/rand"

	"github.com/mrbrutti/termus/internal/synth"
)

// Compile-time assertion that *Markov implements Algorithm.
var _ Algorithm = (*Markov)(nil)

// Markov generates melodies by sampling a hand-crafted transition matrix over
// the 7 degrees of the natural minor scale. Unlike Pentatonic's random walk,
// Markov biases its next-note choices toward functionally musical moves —
// 5→1, 7→1, 4→3, 6→5 — so the line has a sense of "wanting to resolve" that
// random walks lack. The result is more "composed" feeling than the other
// algorithms.
type Markov struct {
	rng      *rand.Rand
	rootMidi int
	voices   []*padBellVoice
	revL     *synth.Reverb
	revR     *synth.Reverb
	t        int64
}

// markovLoopPeriods: each voice plays a fairly long phrase (10–14 notes) on
// these periods, letting the transition matrix really wander.
// Slowed ~25% from 7.5/11.0/16.5 so Markov-generated melodies don't crowd
// the listener and the transition weights' "wanting to resolve" character
// has more time to breathe between resolutions.
var markovLoopPeriods = []float64{9.4, 13.8, 20.6}

// minorTransitions[i][j] = weight of moving from minor-scale degree i to j.
// Degrees: 0=root, 1=2, 2=b3, 3=4, 4=5, 5=b6, 6=b7.
//
// Biases (musical-theory motivated):
//   - 5th (idx 4) likes to resolve down to root or to b6, less often to b7
//   - 7th (idx 6, b7) likes to resolve up to root
//   - 4th (idx 3) likes to fall to b3 (suspension resolution)
//   - 2nd (idx 1) likes to resolve down to root
//   - From root (idx 0), wider distribution (any move plausible)
//   - Stepwise motion is generally preferred to leaps
var minorTransitions = [7][7]int{
	//        0   1   2   3   4   5   6   ← to
	/* 0 */ {2, 5, 6, 4, 6, 3, 4},
	/* 1 */ {7, 1, 5, 3, 2, 1, 1},
	/* 2 */ {3, 6, 1, 6, 4, 2, 2},
	/* 3 */ {2, 3, 8, 1, 5, 2, 1},
	/* 4 */ {8, 2, 3, 5, 2, 5, 3},
	/* 5 */ {3, 1, 2, 2, 7, 1, 4},
	/* 6 */ {9, 2, 2, 2, 3, 3, 1},
}

// NewMarkov constructs the algorithm. Caller must call Seed before Next.
func NewMarkov() *Markov { return &Markov{} }

func (m *Markov) Name() string { return "markov-melody" }

func (m *Markov) Seed(s int64) {
	m.rng = rand.New(rand.NewSource(s)) //nolint:gosec
	m.rootMidi = 36 + m.rng.Intn(12)

	m.voices = make([]*padBellVoice, len(markovLoopPeriods))
	for i, period := range markovLoopPeriods {
		count := 10 + m.rng.Intn(5) // 10..14 notes per phrase
		notes := m.markovNotes(count)
		m.voices[i] = newPadBellVoice(period, notes, m.rng.Float64(), m.rng.Float64())
	}
	m.revL = synth.NewReverb(0.55)
	m.revR = synth.NewReverbRight(0.55)
	m.t = 0
}

// markovNotes walks the transition matrix to produce count notes, then maps
// scale-degree indices to MIDI numbers in a chosen octave (with occasional
// octave shifts to keep the line from getting flat).
func (m *Markov) markovNotes(count int) []int {
	// Always start on root or 5 for a stable opening.
	degree := 0
	if m.rng.Float64() < 0.3 {
		degree = 4
	}
	octave := 12 * (2 + m.rng.Intn(3))

	notes := make([]int, count)
	notes[0] = m.rootMidi + scaleMinor[degree] + octave

	for i := 1; i < count; i++ {
		degree = nextMarkovDegree(m.rng, degree)
		// Occasional octave jump for variety, but bias toward staying.
		switch r := m.rng.Float64(); {
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
		notes[i] = m.rootMidi + scaleMinor[degree] + octave
	}
	return notes
}

// nextMarkovDegree samples the next scale degree from the row of the
// transition matrix indexed by the current degree.
func nextMarkovDegree(rng *rand.Rand, cur int) int {
	row := minorTransitions[cur]
	totalW := 0
	for _, w := range row {
		totalW += w
	}
	pick := rng.Intn(totalW)
	acc := 0
	for j, w := range row {
		acc += w
		if pick < acc {
			return j
		}
	}
	return cur
}

func (m *Markov) Next(left, right []float64) {
	for i := range left {
		var l, r float64
		for vi, v := range m.voices {
			s := v.tick(m.t)
			if vi%2 == 0 {
				l += s * 0.65
				r += s * 0.35
			} else {
				l += s * 0.35
				r += s * 0.65
			}
		}
		l = m.revL.Tick(l)
		r = m.revR.Tick(r)
		left[i] = synth.SoftClip(l * 2.1)
		right[i] = synth.SoftClip(r * 2.1)
		m.t++
	}
}

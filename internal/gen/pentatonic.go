package gen

import (
	"math/rand"

	"github.com/mrbrutti/termus/internal/synth"
)

// Compile-time assertion that *Pentatonic implements Algorithm.
var _ Algorithm = (*Pentatonic)(nil)

// Pentatonic is a random walk through the minor pentatonic scale played by
// soft pad-bell voices. Because the pentatonic scale has no minor seconds
// between adjacent notes, any walk through it sounds consonant — perfect for
// generative material that's forgiving to tune.
type Pentatonic struct {
	rng      *rand.Rand
	rootMidi int
	voices   []*padBellVoice
	revL     *synth.Reverb
	revR     *synth.Reverb
	t        int64
}

// pentaLoopPeriods: moderate-length cycles. Each voice plays a longer phrase
// (more notes per period) than eno so the walk has time to actually wander.
// Slowed ~25% from 6.0/8.5/11.5/15.0 for slower walk pace.
var pentaLoopPeriods = []float64{7.5, 10.6, 14.4, 18.8}

// NewPentatonic constructs the algorithm. Caller must call Seed before Next.
func NewPentatonic() *Pentatonic { return &Pentatonic{} }

func (p *Pentatonic) Name() string { return "pentatonic-walk" }

func (p *Pentatonic) Seed(s int64) {
	p.rng = rand.New(rand.NewSource(s)) //nolint:gosec
	p.rootMidi = 36 + p.rng.Intn(12)

	p.voices = make([]*padBellVoice, len(pentaLoopPeriods))
	for i, period := range pentaLoopPeriods {
		// Generate notes by random walk over the pentatonic-minor scale.
		// Each voice gets 6..10 notes per phrase — enough to feel like a
		// melody, not just two notes alternating.
		count := 6 + p.rng.Intn(5)
		notes := p.walkNotes(count)
		p.voices[i] = newPadBellVoice(period, notes, p.rng.Float64(), p.rng.Float64())
	}
	p.revL = synth.NewReverb(0.50)
	p.revR = synth.NewReverbRight(0.50)
	p.t = 0
}

// walkNotes produces a count-length walk through the pentatonic-minor scale,
// expressed as MIDI numbers anchored to p.rootMidi.
//
// The walk uses these transition weights from the current scale-index k:
//
//	stay (k → k):                weight 1
//	step  (k → k±1):             weight 5
//	skip  (k → k±2):             weight 2
//	leap  (k → k±3 or k±4):      weight 1
//
// This biases the walk toward stepwise motion (musical) while still allowing
// the occasional bigger interval to keep the line interesting.
func (p *Pentatonic) walkNotes(count int) []int {
	scale := scalePentatonicMinor // {0, 3, 5, 7, 10}
	// Start at a random scale index, in a random octave.
	idx := p.rng.Intn(len(scale))
	octave := 12 * (2 + p.rng.Intn(3)) // +24..+48
	notes := make([]int, count)
	notes[0] = p.rootMidi + scale[idx] + octave

	for i := 1; i < count; i++ {
		idx = walkStep(p.rng, idx, len(scale))
		// Allow the walk to occasionally cross octaves.
		if p.rng.Float64() < 0.18 {
			if p.rng.Float64() < 0.5 {
				octave += 12
			} else {
				octave -= 12
			}
			if octave < 12 {
				octave = 12
			}
			if octave > 60 {
				octave = 60
			}
		}
		notes[i] = p.rootMidi + scale[idx] + octave
	}
	return notes
}

// walkStep picks the next scale index given the current one, using weighted
// probabilities favoring stepwise motion.
func walkStep(rng *rand.Rand, cur, scaleLen int) int {
	// Candidate offsets and their weights.
	type cand struct {
		off, w int
	}
	cands := []cand{
		{0, 1},
		{-1, 5}, {1, 5},
		{-2, 2}, {2, 2},
		{-3, 1}, {3, 1},
		{-4, 1}, {4, 1},
	}
	totalW := 0
	for _, c := range cands {
		totalW += c.w
	}
	pick := rng.Intn(totalW)
	acc := 0
	for _, c := range cands {
		acc += c.w
		if pick < acc {
			next := cur + c.off
			if next < 0 {
				next += scaleLen
			}
			if next >= scaleLen {
				next -= scaleLen
			}
			return next
		}
	}
	return cur
}

func (p *Pentatonic) Next(left, right []float64) {
	for i := range left {
		var l, r float64
		for vi, v := range p.voices {
			s := v.tick(p.t)
			pan := float64(vi) / float64(len(p.voices)-1)
			l += s * (0.75 - pan*0.45)
			r += s * (0.30 + pan*0.45)
		}
		l = p.revL.Tick(l)
		r = p.revR.Tick(r)
		left[i] = synth.SoftClip(l * 2.1)
		right[i] = synth.SoftClip(r * 2.1)
		p.t++
	}
}

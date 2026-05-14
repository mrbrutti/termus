package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

var _ Algorithm = (*SF2Pentatonic)(nil)

// SF2Pentatonic is the pentatonic-walk algorithm rendered through a
// SoundFont. The random walk through pentatonic minor sounds especially
// good on piano + orchestral harp — both are immediately associated with
// "music box" / "lullaby" textures.
type SF2Pentatonic struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
}

func NewSF2Pentatonic(sf *meltysynth.SoundFont) *SF2Pentatonic {
	return &SF2Pentatonic{sf: sf}
}

func (a *SF2Pentatonic) Name() string { return "pentatonic-sf2" }

func (a *SF2Pentatonic) Seed(seedVal int64) {
	rng := rand.New(rand.NewSource(seedVal)) //nolint:gosec
	rootMidi := 36 + rng.Intn(12)

	core, err := newSF2Core(a.sf, 3.2, seedVal)
	if err != nil {
		a.core = nil
		return
	}
	core.setProgram(0, 0)  // Acoustic Grand Piano
	core.setProgram(1, 46) // Orchestral Harp

	// Same walk-based note generation as gen.Pentatonic. Mutation uses a
	// short random walk from the previous note's pentatonic-scale degree —
	// so mutated notes stay smoothly connected to the existing melody.
	for i, period := range pentaLoopPeriods {
		count := 6 + rng.Intn(5)
		notes := pentatonicWalk(rng, rootMidi, count)
		ch := int32(i % 2)
		var vel int32 = 80
		if ch == 1 {
			vel = 68
		}
		// Capture rootMidi + scale ref for the mutation closure.
		root := rootMidi
		mutate := func(_ int, prev int) int {
			// Find closest pentatonic-minor pitch to prev, walk by ±1..2 steps.
			closest := findClosestScalePitch(prev, root, scalePentatonicMinor)
			step := walkStep(rng, closest.degreeIdx, len(scalePentatonicMinor))
			return root + scalePentatonicMinor[step] + closest.octaveOffset*12
		}
		core.addTrack(SF2Track{
			Channel: ch, Velocity: vel, Notes: notes,
			PeriodSec: period, Phase01: rng.Float64(),
			MutationRate: 0.15, MutateOne: mutate,
		})
	}
	a.core = core
}

// scalePitchLoc describes where a MIDI note sits in a scale anchored at
// rootMidi: which degree (0..len(scale)-1) and how many octaves above the
// root. Used by mutation to find a "nearby" scale note to walk from.
type scalePitchLoc struct {
	degreeIdx    int
	octaveOffset int
}

// findClosestScalePitch returns the closest scale-anchored pitch to the given
// MIDI note. Useful for mutation closures that want to walk from "wherever
// we currently are" rather than re-rolling from scratch.
func findClosestScalePitch(midi, rootMidi int, scale []int) scalePitchLoc {
	rel := midi - rootMidi
	// Decompose into (octave, semitone within octave) where semitone is in 0..11.
	octave := rel / 12
	semi := rel % 12
	if semi < 0 {
		semi += 12
		octave--
	}
	bestIdx := 0
	bestDist := 1 << 30
	for i, s := range scale {
		d := semi - s
		if d < 0 {
			d = -d
		}
		if d < bestDist {
			bestDist = d
			bestIdx = i
		}
	}
	return scalePitchLoc{degreeIdx: bestIdx, octaveOffset: octave}
}

// pentatonicWalk produces a random walk through the pentatonic-minor scale,
// matching gen.Pentatonic's logic. Extracted here as a shared helper.
func pentatonicWalk(rng *rand.Rand, rootMidi, count int) []int {
	scale := scalePentatonicMinor
	idx := rng.Intn(len(scale))
	octave := 12 * (2 + rng.Intn(3))
	notes := make([]int, count)
	notes[0] = rootMidi + scale[idx] + octave
	for i := 1; i < count; i++ {
		idx = walkStep(rng, idx, len(scale))
		if rng.Float64() < 0.18 {
			if rng.Float64() < 0.5 {
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
		notes[i] = rootMidi + scale[idx] + octave
	}
	return notes
}

// SetReverbIR installs a convolution reverb on the master bus.
func (a *SF2Pentatonic) SetReverbIR(ir []float64, wet float64) {
	if a.core != nil {
		a.core.setConvolutionIR(ir, wet)
	}
}

func (a *SF2Pentatonic) Next(left, right []float64) {
	if a.core == nil {
		for i := range left {
			left[i] = 0
			right[i] = 0
		}
		return
	}
	a.core.renderInto(left, right)
}

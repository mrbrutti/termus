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

	core, err := newSF2Core(a.sf, 3.2)
	if err != nil {
		a.core = nil
		return
	}
	core.setProgram(0, 0)  // Acoustic Grand Piano
	core.setProgram(1, 46) // Orchestral Harp

	// Same walk-based note generation as gen.Pentatonic.
	for i, period := range pentaLoopPeriods {
		count := 6 + rng.Intn(5)
		notes := pentatonicWalk(rng, rootMidi, count)
		// Alternate between piano and harp per voice for instrumental variety.
		ch := int32(i % 2)
		var vel int32 = 80
		if ch == 1 {
			vel = 68
		}
		core.addTrack(SF2Track{
			Channel: ch, Velocity: vel, Notes: notes,
			PeriodSec: period, Phase01: rng.Float64(),
		})
	}
	a.core = core
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

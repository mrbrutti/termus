package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

var _ Algorithm = (*SF2Drone)(nil)

// SF2Drone is the drone-bed algorithm rendered through a SoundFont. Long
// sustained notes on string ensemble and slow strings, with a high-register
// flute shimmer voice — close in feel to the Eno-meets-Stars-of-the-Lid
// orchestrations.
type SF2Drone struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
}

func NewSF2Drone(sf *meltysynth.SoundFont) *SF2Drone {
	return &SF2Drone{sf: sf}
}

func (a *SF2Drone) Name() string { return "drone-sf2" }

func (a *SF2Drone) Seed(seedVal int64) {
	rng := rand.New(rand.NewSource(seedVal)) //nolint:gosec
	rootMidi := 24 + rng.Intn(7) // C1..F#1

	core, err := newSF2Core(a.sf, 3.0, seedVal)
	if err != nil {
		a.core = nil
		return
	}
	core.setProgram(0, 48) // String Ensemble 1 (left)
	core.setProgram(1, 49) // String Ensemble 2 slow (right)
	core.setProgram(2, 73) // Flute (shimmer, high right)
	core.setProgram(3, 53) // Choir Voice "Oohs" (warmth, center)
	core.setProgram(4, 32) // Acoustic Bass (foundation, center)
	core.setPan(0, 38)
	core.setPan(1, 90)
	core.setPan(2, 100)
	core.setPan(3, 64)
	core.setPan(4, 64)

	// Bed voices on long periods. Mutation is gentle here — drone wants
	// to feel stable; abrupt note changes would betray the aesthetic.
	bedMutate := func(_ int, _ int) int {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		octave := 12 * (1 + rng.Intn(3))
		return rootMidi + degree + octave
	}
	for _, period := range droneLoopPeriods {
		notes := make([]int, 3+rng.Intn(3))
		for j := range notes {
			notes[j] = bedMutate(0, 0)
		}
		phase := rng.Float64()
		core.addTrack(SF2Track{
			Channel: 0, Velocity: 64, Notes: notes,
			PeriodSec: period, Phase01: phase,
			MutationRate: 0.05, MutateOne: bedMutate,
			VelocityJitter: 4, TimingJitterSec: 0.020,
		})
		core.addTrack(SF2Track{
			Channel: 1, Velocity: 48, Notes: notes,
			PeriodSec: period, Phase01: phase,
			VelocityJitter: 4, TimingJitterSec: 0.020,
		})
	}

	// Shimmer voice on a 19s period — flute in the high register.
	shimmerMutate := func(_ int, _ int) int {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		octave := 12 * (3 + rng.Intn(2))
		return rootMidi + degree + octave
	}
	shimmerNotes := make([]int, 4+rng.Intn(3))
	for j := range shimmerNotes {
		shimmerNotes[j] = shimmerMutate(0, 0)
	}
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 70, Notes: shimmerNotes,
		PeriodSec: shimmerPeriod, Phase01: rng.Float64(),
		MutationRate: 0.15, MutateOne: shimmerMutate,
		VelocityJitter: 10, TimingJitterSec: 0.025,
	})

	// Choir voice "oohs" in mid-register for human warmth. Slow cycle so
	// the same chord-tone-ish notes hold for a long time.
	choirMutate := func(_ int, _ int) int {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		return rootMidi + degree + 12 // one octave above root
	}
	choirNotes := make([]int, 3)
	for j := range choirNotes {
		choirNotes[j] = choirMutate(0, 0)
	}
	core.addTrack(SF2Track{
		Channel: 3, Velocity: 54, Notes: choirNotes,
		PeriodSec: 37.0, Phase01: rng.Float64(),
		MutationRate: 0.08, MutateOne: choirMutate,
		VelocityJitter: 4, TimingJitterSec: 0.020,
	})

	// Acoustic bass: very slow walk through scale tones at the root register.
	bassMutate := func(_ int, _ int) int {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		return rootMidi + degree // base octave (rootMidi is already low C1-F#1)
	}
	bassNotes := make([]int, 3)
	for j := range bassNotes {
		bassNotes[j] = bassMutate(0, 0)
	}
	core.addTrack(SF2Track{
		Channel: 4, Velocity: 76, Notes: bassNotes,
		PeriodSec: 41.0, Phase01: rng.Float64(),
		MutationRate: 0.06, MutateOne: bassMutate,
		VelocityJitter: 4, TimingJitterSec: 0.020,
	})

	a.core = core
}

// SetReverbIR installs a convolution reverb on the master bus.
func (a *SF2Drone) SetReverbIR(ir []float64, wet float64) {
	if a.core != nil {
		a.core.setConvolutionIR(ir, wet)
	}
}

func (a *SF2Drone) Next(left, right []float64) {
	if a.core == nil {
		for i := range left {
			left[i] = 0
			right[i] = 0
		}
		return
	}
	a.core.renderInto(left, right)
}

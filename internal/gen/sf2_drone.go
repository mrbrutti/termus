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

	core, err := newSF2Core(a.sf, 3.0)
	if err != nil {
		a.core = nil
		return
	}
	// Channel layout: 0 = String Ensemble, 1 = Slow Strings, 2 = Flute.
	core.setProgram(0, 48) // String Ensemble 1
	core.setProgram(1, 49) // String Ensemble 2 (slow)
	core.setProgram(2, 73) // Flute

	// Bed voices on long periods.
	for _, period := range droneLoopPeriods {
		notes := make([]int, 3+rng.Intn(3))
		for j := range notes {
			degree := scaleMinor[rng.Intn(len(scaleMinor))]
			octave := 12 * (1 + rng.Intn(3))
			notes[j] = rootMidi + degree + octave
		}
		phase := rng.Float64()
		// Layer both string programs at lower velocity for a soft texture.
		core.addTrack(SF2Track{
			Channel: 0, Velocity: 64, Notes: notes,
			PeriodSec: period, Phase01: phase,
		})
		core.addTrack(SF2Track{
			Channel: 1, Velocity: 48, Notes: notes,
			PeriodSec: period, Phase01: phase,
		})
	}

	// Shimmer voice on a 19s period — flute in the high register.
	shimmerNotes := make([]int, 4+rng.Intn(3))
	for j := range shimmerNotes {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		octave := 12 * (3 + rng.Intn(2))
		shimmerNotes[j] = rootMidi + degree + octave
	}
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 70, Notes: shimmerNotes,
		PeriodSec: shimmerPeriod, Phase01: rng.Float64(),
	})
	a.core = core
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

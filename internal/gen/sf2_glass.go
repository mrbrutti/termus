package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

var _ Algorithm = (*SF2Glass)(nil)

// SF2Glass is the glass-fm algorithm rendered through a SoundFont. The FM
// bells are replaced by Tubular Bells (GM #14), which sound like real bells
// played by a real player — sweeter and warmer than synthesis. Plus a Glass
// Harmonica (GM #98) layer for higher partials.
type SF2Glass struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
}

func NewSF2Glass(sf *meltysynth.SoundFont) *SF2Glass {
	return &SF2Glass{sf: sf}
}

func (a *SF2Glass) Name() string { return "glass-sf2" }

func (a *SF2Glass) Seed(seedVal int64) {
	rng := rand.New(rand.NewSource(seedVal)) //nolint:gosec
	rootMidi := 48 + rng.Intn(7) // C3..F#3

	core, err := newSF2Core(a.sf, 3.2)
	if err != nil {
		a.core = nil
		return
	}
	core.setProgram(0, 14) // Tubular Bells
	core.setProgram(1, 98) // Crystal (FX 3) — bell-like ringing

	for _, period := range glassLoopPeriods {
		notes := make([]int, 1+rng.Intn(2))
		for j := range notes {
			degree := scalePentatonicMinor[rng.Intn(len(scalePentatonicMinor))]
			octave := 12 * (1 + rng.Intn(3))
			notes[j] = rootMidi + degree + octave
		}
		phase := rng.Float64()
		core.addTrack(SF2Track{
			Channel: 0, Velocity: 84, Notes: notes,
			PeriodSec: period, Phase01: phase,
		})
		// Crystal layer at lower velocity, slight phase offset for shimmer.
		core.addTrack(SF2Track{
			Channel: 1, Velocity: 52, Notes: notes,
			PeriodSec: period, Phase01: phase + 0.03,
		})
	}
	a.core = core
}

func (a *SF2Glass) Next(left, right []float64) {
	if a.core == nil {
		for i := range left {
			left[i] = 0
			right[i] = 0
		}
		return
	}
	a.core.renderInto(left, right)
}

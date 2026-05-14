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

	core, err := newSF2Core(a.sf, 3.2, seedVal)
	if err != nil {
		a.core = nil
		return
	}
	core.setProgram(0, 14) // Tubular Bells           (left)
	core.setProgram(1, 98) // Crystal (FX 3)          (right)
	core.setProgram(2, 92) // Pad 5 (Bowed Glass)     (low pad, center)
	core.setProgram(3, 12) // Marimba                 (sub-octave reinforcement)
	core.setPan(0, 40)
	core.setPan(1, 88)
	core.setPan(2, 64)
	core.setPan(3, 64)

	pentMutate := func(_ int, _ int) int {
		degree := scalePentatonicMinor[rng.Intn(len(scalePentatonicMinor))]
		octave := 12 * (1 + rng.Intn(3))
		return rootMidi + degree + octave
	}
	for _, period := range glassLoopPeriods {
		notes := make([]int, 1+rng.Intn(2))
		for j := range notes {
			notes[j] = pentMutate(0, 0)
		}
		phase := rng.Float64()
		core.addTrack(SF2Track{
			Channel: 0, Velocity: 84, Notes: notes,
			PeriodSec: period, Phase01: phase,
			MutationRate: 0.20, MutateOne: pentMutate,
			VelocityJitter: 10, TimingJitterSec: 0.022,
		})
		core.addTrack(SF2Track{
			Channel: 1, Velocity: 52, Notes: notes,
			PeriodSec: period, Phase01: phase + 0.03,
			VelocityJitter: 6, TimingJitterSec: 0.022,
		})
	}

	// Bowed glass pad in low register — sustained tones for body underneath
	// the bell shimmer. Long period, low velocity so it's atmospheric, not
	// the focus.
	padMutate := func(_ int, _ int) int {
		degree := scalePentatonicMinor[rng.Intn(len(scalePentatonicMinor))]
		return rootMidi + degree // base octave
	}
	padNotes := make([]int, 4)
	for j := range padNotes {
		padNotes[j] = padMutate(0, 0)
	}
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 50, Notes: padNotes,
		PeriodSec: 33.0, Phase01: 0,
		MutationRate: 0.15, MutateOne: padMutate,
		VelocityJitter: 4, TimingJitterSec: 0.025,
	})

	// Marimba reinforcement: occasional low strikes echoing the bell tones.
	marimNotes := make([]int, 4)
	for j := range marimNotes {
		marimNotes[j] = padMutate(0, 0) // same scale, low register
	}
	core.addTrack(SF2Track{
		Channel: 3, Velocity: 56, Notes: marimNotes,
		PeriodSec: 19.0, Phase01: rng.Float64(),
		MutationRate: 0.15, MutateOne: padMutate,
		VelocityJitter: 12, TimingJitterSec: 0.030,
	})

	a.core = core
}

// SetReverbIR installs a convolution reverb on the master bus.
func (a *SF2Glass) SetReverbIR(ir []float64, wet float64) {
	if a.core != nil {
		a.core.setConvolutionIR(ir, wet)
	}
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

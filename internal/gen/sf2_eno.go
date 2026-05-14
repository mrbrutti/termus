package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

// Compile-time assertion.
var _ Algorithm = (*SF2Eno)(nil)

// SF2Eno is the eno-drift algorithm rendered through a SoundFont. Same
// scheduling structure as Eno (5 pad voices on incommensurate loops + 2 lead
// voices on shorter loops) but voices are mapped to sampled instruments —
// strings + warm pad for the bed, acoustic grand piano for the leads.
type SF2Eno struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
}

// NewSF2Eno builds the algorithm. Caller must call Seed before Next.
func NewSF2Eno(sf *meltysynth.SoundFont) *SF2Eno {
	return &SF2Eno{sf: sf}
}

func (a *SF2Eno) Name() string { return "eno-sf2" }

func (a *SF2Eno) Seed(seedVal int64) {
	rng := rand.New(rand.NewSource(seedVal)) //nolint:gosec
	rootMidi := 36 + rng.Intn(12)

	core, err := newSF2Core(a.sf, 3.5)
	if err != nil {
		a.core = nil
		return
	}
	// Channel layout: 0 = strings, 1 = warm pad (both for the bed); 2 = piano (leads).
	core.setProgram(0, 48) // String Ensemble 1
	core.setProgram(1, 89) // Warm Pad
	core.setProgram(2, 0)  // Acoustic Grand Piano

	// Slow pad bed — same logic as gen.Eno.Seed but with two tracks per
	// musical voice so strings and warm pad layer together.
	for _, period := range loopPeriods {
		notes := make([]int, 2+rng.Intn(3))
		for j := range notes {
			degree := scaleMinor[rng.Intn(len(scaleMinor))]
			octave := 12 * (2 + rng.Intn(3))
			notes[j] = rootMidi + degree + octave
		}
		phase := rng.Float64()
		core.addTrack(SF2Track{
			Channel: 0, Velocity: 70, Notes: notes,
			PeriodSec: period, Phase01: phase,
		})
		core.addTrack(SF2Track{
			Channel: 1, Velocity: 56, Notes: notes,
			PeriodSec: period, Phase01: phase,
		})
	}

	// Lead voices — shorter periods, more notes, higher register, piano.
	for _, period := range leadPeriods {
		notes := make([]int, 4+rng.Intn(3))
		for j := range notes {
			degree := scaleMinor[rng.Intn(len(scaleMinor))]
			octave := 12 * (3 + rng.Intn(2))
			notes[j] = rootMidi + degree + octave
		}
		core.addTrack(SF2Track{
			Channel: 2, Velocity: 92, Notes: notes,
			PeriodSec: period, Phase01: rng.Float64(),
		})
	}
	a.core = core
}

// SetReverbIR installs a convolution reverb on the master bus.
func (a *SF2Eno) SetReverbIR(ir []float64, wet float64) {
	if a.core != nil {
		a.core.setConvolutionIR(ir, wet)
	}
}

func (a *SF2Eno) Next(left, right []float64) {
	if a.core == nil {
		for i := range left {
			left[i] = 0
			right[i] = 0
		}
		return
	}
	a.core.renderInto(left, right)
}

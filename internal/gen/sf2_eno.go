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

	core, err := newSF2Core(a.sf, 3.5, seedVal)
	if err != nil {
		a.core = nil
		return
	}
	core.setProgram(0, 48) // String Ensemble 1     (left)
	core.setProgram(1, 89) // Warm Pad              (right)
	core.setProgram(2, 0)  // Acoustic Grand Piano  (center, leads)
	core.setProgram(3, 8)  // Celesta               (high sparkle, right)
	core.setProgram(4, 60) // French Horn           (low foundation, center)
	core.setPan(0, 38)
	core.setPan(1, 90)
	core.setPan(2, 64)
	core.setPan(3, 100)
	core.setPan(4, 64)

	// Slow pad bed — same logic as gen.Eno.Seed but with two tracks per
	// musical voice so strings and warm pad layer together.
	padMutate := func(_ int, _ int) int {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		octave := 12 * (2 + rng.Intn(3))
		return rootMidi + degree + octave
	}
	for _, period := range loopPeriods {
		notes := make([]int, 2+rng.Intn(3))
		for j := range notes {
			notes[j] = padMutate(0, 0)
		}
		phase := rng.Float64()
		core.addTrack(SF2Track{
			Channel: 0, Velocity: 70, Notes: notes,
			PeriodSec: period, Phase01: phase,
			MutationRate: 0.08, MutateOne: padMutate,
			VelocityJitter: 6,
		})
		core.addTrack(SF2Track{
			Channel: 1, Velocity: 56, Notes: notes,
			PeriodSec: period, Phase01: phase,
			VelocityJitter: 4,
		})
	}

	// Lead voices — shorter periods, more notes, higher register, piano.
	leadMutate := func(_ int, _ int) int {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		octave := 12 * (3 + rng.Intn(2))
		return rootMidi + degree + octave
	}
	for _, period := range leadPeriods {
		notes := make([]int, 4+rng.Intn(3))
		for j := range notes {
			notes[j] = leadMutate(0, 0)
		}
		core.addTrack(SF2Track{
			Channel: 2, Velocity: 92, Notes: notes,
			PeriodSec: period, Phase01: rng.Float64(),
			MutationRate: 0.18, MutateOne: leadMutate,
			VelocityJitter: 12,
		})
	}

	// Celesta: sparse high-register sparkle, very slow trigger rate.
	celesta := func(_ int, _ int) int {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		return rootMidi + degree + 48 // +4 octaves
	}
	celestaNotes := make([]int, 4)
	for j := range celestaNotes {
		celestaNotes[j] = celesta(0, 0)
	}
	core.addTrack(SF2Track{
		Channel: 3, Velocity: 56, Notes: celestaNotes,
		PeriodSec: 27.0, Phase01: rng.Float64(),
		MutationRate: 0.20, MutateOne: celesta,
		VelocityJitter: 12,
	})

	// French horn: warm low sustained voice on chord roots. Cycles slowly
	// (chord change every ~25s) — anchors the harmony without intruding on
	// the pad bed.
	horn := func(_ int, _ int) int {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		return rootMidi + degree // low register
	}
	hornNotes := make([]int, 3)
	for j := range hornNotes {
		hornNotes[j] = horn(0, 0)
	}
	core.addTrack(SF2Track{
		Channel: 4, Velocity: 50, Notes: hornNotes,
		PeriodSec: 25.0, Phase01: rng.Float64(),
		MutationRate: 0.10, MutateOne: horn,
		VelocityJitter: 6,
	})

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

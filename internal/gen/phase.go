package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

var _ Algorithm = (*Phase)(nil)
var _ SF2Reverberator = (*Phase)(nil)

// Phase is a Reich-style phase-shift algorithm. Two marimba voices play
// the same 6-note melodic figure at slightly different tempos: voice A at
// some base period, voice B at ~0.8% slower. Over time the rhythmic
// alignment between them drifts continuously, producing ever-changing
// interlocking patterns — the algorithmic technique Steve Reich pioneered
// in "Piano Phase" (1967) and "Music for 18 Musicians" (1976).
//
// Underneath, a sustained Choir Aahs pad cycles through a slow 4-chord
// minor-key progression (~30 s per chord), with an acoustic bass on the
// chord roots an octave lower. The melodic figure is drawn from the
// pentatonic minor scale, so it stays consonant against any chord in the
// progression.
//
// Best results with --ir set to a real cathedral or hall IR — the
// long-tail reverb dramatically enhances the interlocking-pattern effect
// because rhythmic offsets between the two voices create dense reflections.
type Phase struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
}

// NewPhase constructs the algorithm. Caller must call Seed before Next.
func NewPhase(sf *meltysynth.SoundFont) *Phase { return &Phase{sf: sf} }

func (a *Phase) Name() string { return "phase-shift" }

func (a *Phase) Seed(seedVal int64) {
	rng := rand.New(rand.NewSource(seedVal)) //nolint:gosec
	rootMidi := 45 + rng.Intn(7) // A2..D#3 — a register that puts the marimba mid-range

	core, err := newSF2Core(a.sf, 3.0, seedVal)
	if err != nil {
		a.core = nil
		return
	}
	// Channel layout:
	//   0 — Marimba (GM #12)    voice A (slightly faster figure)
	//   1 — Marimba (GM #12)    voice B (slightly slower figure)
	//   2 — Choir Aahs (#52)    sustained chord pad
	//   3 — Acoustic Bass (#32) chord roots, octave lower
	core.setProgram(0, 12)
	core.setProgram(1, 12)
	core.setProgram(2, 52)
	core.setProgram(3, 32)

	// Build the melodic figure: 6 notes from the pentatonic-minor scale,
	// distributed across two octaves above the root for a "music box" feel.
	figure := make([]int, 6)
	for i := range figure {
		deg := scalePentatonicMinor[rng.Intn(len(scalePentatonicMinor))]
		// Slight upward bias as the figure unfolds — gives a melodic arc
		// even though the order within is random.
		octaveBias := 24
		if i >= 4 {
			octaveBias = 36
		}
		figure[i] = rootMidi + deg + octaveBias
	}

	// Tempo: ~3.0 to 3.6 seconds per 6-note cycle.
	basePeriod := 3.0 + 0.6*rng.Float64()
	// Voice B is exactly 0.8% slower. With 6 notes per cycle, the phase
	// offset accumulates one full note position every ~125 cycles, so
	// after ~7 minutes the two voices have shifted by an entire beat.
	driftRatio := 1.008

	// Mutation: the figure slowly mutates over time so the patterns evolve
	// instead of repeating verbatim. Both voices share the same figure
	// slice (Voice A and Voice B literally play the same notes) so mutating
	// once changes the figure for both — perfect, since the whole point of
	// the algorithm is that both voices play the SAME thing.
	figMutate := func(_ int, _ int) int {
		deg := scalePentatonicMinor[rng.Intn(len(scalePentatonicMinor))]
		oct := 24
		if rng.Float64() < 0.35 {
			oct = 36
		}
		return rootMidi + deg + oct
	}
	core.addTrack(SF2Track{
		Channel: 0, Velocity: 92, Notes: figure,
		PeriodSec: basePeriod, Phase01: 0,
		MutationRate: 0.10, MutateOne: figMutate,
	})
	core.addTrack(SF2Track{
		Channel: 1, Velocity: 86, Notes: figure,
		PeriodSec: basePeriod * driftRatio, Phase01: 0,
		// Voice B inherits any mutations on the shared figure slice; no
		// own MutateOne needed.
	})

	// Chord progression: i-VI-III-VII (Andalusian feel), one chord per
	// ~30 seconds. The pad track cycles through 4 chord roots; the same
	// list raised down an octave is the bass.
	progDegs := [][]int{
		{0, 5, 2, 6}, // i-VI-III-VII
		{0, 3, 5, 6}, // i-iv-VI-VII
		{0, 6, 5, 3}, // i-VII-VI-iv
		{0, 2, 3, 6}, // i-III-iv-VII
	}[rng.Intn(4)]
	chordRoots := make([]int, len(progDegs))
	bassRoots := make([]int, len(progDegs))
	for i, deg := range progDegs {
		chordRoots[i] = rootMidi + scaleMinor[deg] // pad in same register as root
		bassRoots[i] = rootMidi + scaleMinor[deg] - 12
		// Pull the bass into a reasonable register (no subsonic).
		if bassRoots[i] < 24 {
			bassRoots[i] += 12
		}
	}
	const chordCycleSec = 120.0 // 4 chords × 30 s each
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 60, Notes: chordRoots,
		PeriodSec: chordCycleSec, Phase01: 0,
	})
	core.addTrack(SF2Track{
		Channel: 3, Velocity: 80, Notes: bassRoots,
		PeriodSec: chordCycleSec, Phase01: 0,
	})

	a.core = core
}

// SetReverbIR installs a convolution reverb on the master bus. Highly
// recommended for this algorithm — the phase-shift effect lives or dies
// on the rhythmic interaction of the two voices, and a long reverb tail
// dramatically thickens the resulting interlocking patterns.
func (a *Phase) SetReverbIR(ir []float64, wet float64) {
	if a.core != nil {
		a.core.setConvolutionIR(ir, wet)
	}
}

func (a *Phase) Next(left, right []float64) {
	if a.core == nil {
		for i := range left {
			left[i] = 0
			right[i] = 0
		}
		return
	}
	a.core.renderInto(left, right)
}

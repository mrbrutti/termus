package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"

	"github.com/mrbrutti/termus/internal/synth"
)

var _ Algorithm = (*Phase)(nil)
var _ SF2Reverberator = (*Phase)(nil)

// Phase is a Reich-style phase-shift algorithm tuned for ambient listening.
// Two vibraphone voices play the same 4-note pentatonic-minor figure at
// slightly different tempos; the long sustain of the vibraphone fills the
// space between attacks so the music feels like a continuous shimmer rather
// than a sequence of struck notes. A choir-aahs pad and acoustic bass move
// underneath at a much slower chord cycle (~60 s per chord).
//
// For hours-long listening, the algorithm performs two kinds of mutation:
//   - per-slot: at each note trigger, with ~10% probability one of the
//     figure slots is re-rolled (handled by sf2Core).
//   - macro key-drift: every 4–7 minutes a transposition of ±1..±2 semitones
//     is rolled in. The drift takes effect gradually as notes mutate over
//     the next minute, so the key change feels like Eno-style "obvious but
//     gradual" modulation rather than a jump cut.
//
// The algorithm auto-installs the SyntheticHallIR reverb on the master bus
// unless --ir is provided. Phase-shift especially benefits from a long
// reverb tail since the dense interlocking patterns blur into a wash.
type Phase struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
	rng  *rand.Rand

	rootMidi  int // base tonic
	keyOffset int // current macro-transposition (semitones)

	samplesElapsed int64 // since Seed
	nextDriftAt    int64 // absolute sample index for next macro key drift
	nextSwapAt     int64 // absolute sample index for next instrument swap
}

func NewPhase(sf *meltysynth.SoundFont) *Phase { return &Phase{sf: sf} }

func (a *Phase) Name() string { return "phase-shift" }

// currentRoot returns the effective root MIDI accounting for any macro
// key-drift that has happened. Mutator closures read this through the
// *Phase pointer so they always pick notes in the current key, not the
// initial one.
func (a *Phase) currentRoot() int { return a.rootMidi + a.keyOffset }

func (a *Phase) Seed(seedVal int64) {
	a.rng = rand.New(rand.NewSource(seedVal)) //nolint:gosec
	a.rootMidi = 45 + a.rng.Intn(7)            // A2..D#3
	a.keyOffset = 0
	a.samplesElapsed = 0
	a.scheduleNextDrift()
	a.scheduleNextSwap()

	core, err := newSF2Core(a.sf, 3.0, seedVal)
	if err != nil {
		a.core = nil
		return
	}
	// Channel layout:
	//   0 — Vibraphone (#11)    voice A — long sustain, ambient hover
	//   1 — Vibraphone (#11)    voice B — slightly slower tempo
	//   2 — Choir Aahs (#52)    sustained pad
	//   3 — Acoustic Bass (#32) chord roots, soft
	core.setProgram(0, 11)
	core.setProgram(1, 11)
	core.setProgram(2, 52)
	core.setProgram(3, 32)

	// Filter LFO on the choir pad — sustained, benefits from breathing.
	// Vibraphones get no LFO (would mask the phase-shift rhythmic effect).
	core.addFilterLFO(2, 1.0/18.0, 60, 32)

	// Per-channel base cutoffs. Vibraphones bright so attack transients
	// stay defined (essential for hearing the rhythmic interference);
	// choir and bass darker for atmosphere.
	core.setChannelCutoff(0, 92) // vibe A
	core.setChannelCutoff(1, 92) // vibe B
	core.setChannelCutoff(2, 50) // choir pad
	core.setChannelCutoff(3, 60) // bass

	// Phase wants HEAVY reverb on the vibraphones — the long tail is what
	// blurs the two voices' interlocking patterns into a continuous wash.
	core.setReverbSend(0, 115)
	core.setReverbSend(1, 115)
	core.setReverbSend(2, 100) // choir pad
	core.setReverbSend(3, 50)  // bass dry
	core.setChorusSend(2, 48)

	// 4-note figure, drawn from pentatonic-minor. Fewer notes than the v1
	// 6-note figure → each note rings longer, more ambient.
	figure := make([]int, 4)
	for i := range figure {
		figure[i] = a.pickFigureNote()
	}
	figMutate := func(_ int, _ int) int { return a.pickFigureNote() }

	// Cycle ~7.5–10 s — slightly slower than before to give each note even
	// more room to ring. Phase is already the most ambient algorithm; just
	// a gentle adjustment to align with the broader "slower across the
	// board" pass.
	basePeriod := 7.5 + 2.5*a.rng.Float64()
	driftRatio := 1.004 // even gentler tempo offset — patterns drift more slowly

	// Vibraphone voices get small velocity jitter for "real player" feel but
	// NO timing jitter — the entire phase-shift technique depends on each
	// voice's tempo being precisely defined. Randomizing timing here would
	// erase the rhythmic-interference effect the algorithm exists to produce.
	core.addTrack(SF2Track{
		Channel: 0, Velocity: 76, Notes: figure,
		PeriodSec: basePeriod, Phase01: 0,
		MutationRate: 0.10, MutateOne: figMutate,
		VelocityJitter: 8,
	})
	core.addTrack(SF2Track{
		Channel: 1, Velocity: 70, Notes: figure,
		PeriodSec: basePeriod * driftRatio, Phase01: 0,
		VelocityJitter: 8,
	})

	// Chord progression: 4 chords, ~60 s each (was 30 s) — the harmonic bed
	// barely moves so the listener's attention rests on the interlocking
	// vibraphone patterns. Pad and bass share the same chord roots.
	progDegs := [][]int{
		{0, 5, 2, 6},
		{0, 3, 5, 6},
		{0, 6, 5, 3},
		{0, 2, 3, 6},
	}[a.rng.Intn(4)]

	// Recompute chord notes from currentRoot() on every cycle by mutating
	// the slots. The chord cycle itself is 240 s, so this isn't a hot path.
	chordRoots := make([]int, len(progDegs))
	bassRoots := make([]int, len(progDegs))
	for i, deg := range progDegs {
		chordRoots[i] = a.currentRoot() + scaleMinor[deg]
		bassRoots[i] = a.currentRoot() + scaleMinor[deg] - 12
		if bassRoots[i] < 24 {
			bassRoots[i] += 12
		}
	}
	// Mutators for pad/bass: re-roll uses the CURRENT key, so over time the
	// progression itself drifts along with macro key shifts. Stays in the
	// same scale degrees, just shifted.
	padMutate := func(slot int, _ int) int {
		return a.currentRoot() + scaleMinor[progDegs[slot%len(progDegs)]]
	}
	bassMutate := func(slot int, _ int) int {
		k := a.currentRoot() + scaleMinor[progDegs[slot%len(progDegs)]] - 12
		if k < 24 {
			k += 12
		}
		return k
	}
	const chordCycleSec = 240.0
	// Pad and bass have very slow change rate (one chord per minute), so
	// timing jitter is meaningless — kept on velocity only for breathing feel.
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 50, Notes: chordRoots,
		PeriodSec: chordCycleSec, Phase01: 0,
		MutationRate: 1.0, MutateOne: padMutate,
		VelocityJitter: 4,
	})
	core.addTrack(SF2Track{
		Channel: 3, Velocity: 70, Notes: bassRoots,
		PeriodSec: chordCycleSec, Phase01: 0,
		MutationRate: 1.0, MutateOne: bassMutate,
		VelocityJitter: 6,
	})

	// Auto-install a long synthetic hall reverb — phase-shift sounds dramatic
	// in a big space and lifeless dry. Overridden if the user passes --ir.
	core.setConvolutionIR(synth.SyntheticHallIR(seedVal), 0.55)

	a.core = core
}

// pickFigureNote returns a random pentatonic-minor note in the figure's
// register, using the current (possibly key-drifted) root.
func (a *Phase) pickFigureNote() int {
	deg := scalePentatonicMinor[a.rng.Intn(len(scalePentatonicMinor))]
	oct := 24
	if a.rng.Float64() < 0.35 {
		oct = 36
	}
	return a.currentRoot() + deg + oct
}

func (a *Phase) scheduleNextDrift() {
	// 4–7 minutes between key drifts.
	secs := 240.0 + 180.0*a.rng.Float64()
	a.nextDriftAt = a.samplesElapsed + int64(secs*float64(synth.SampleRate))
}

// shiftKey rolls a ±1..±2 semitone macro transposition. Drifts are
// clamped to ±5 semitones from the original key to keep the long-term
// listening centered around the seed's chosen tonic.
func (a *Phase) shiftKey() {
	shift := a.rng.Intn(5) - 2 // -2..+2
	if shift == 0 {
		shift = 1
	}
	a.keyOffset += shift
	if a.keyOffset > 5 {
		a.keyOffset = 5 - a.rng.Intn(3)
	}
	if a.keyOffset < -5 {
		a.keyOffset = -5 + a.rng.Intn(3)
	}
}

// SetReverbIR installs a convolution reverb on the master bus. Phase auto-
// installs a hall by default; --ir overrides.
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
	a.samplesElapsed += int64(len(left))
	if a.samplesElapsed >= a.nextDriftAt {
		a.shiftKey()
		a.scheduleNextDrift()
	}
	if a.samplesElapsed >= a.nextSwapAt {
		a.swapOneInstrument()
		a.scheduleNextSwap()
	}
}

// phaseChannelAlternatives — vibraphone is the genre-defining choice; swaps
// happen between mallet-percussion variants that preserve the rhythmic
// shimmer character.
var phaseChannelAlternatives = map[int32][]int32{
	0: {11, 9, 12, 13, 14}, // Vibraphone (default), Glockenspiel, Marimba, Xylophone, Tubular Bells
	1: {11, 9, 12, 13, 14}, // (voice B must match voice A's instrument — see below)
	2: {52, 53, 91, 89},    // Choir Aahs (default), Choir Oohs, Choir Pad, Warm Pad
	3: {32, 33, 38, 60},    // Acoustic Bass (default), Electric Bass, Synth Bass, French Horn
}

func (a *Phase) scheduleNextSwap() {
	secs := 300.0 + 240.0*a.rng.Float64() // 5–9 min — phase wants slow change
	a.nextSwapAt = a.samplesElapsed + int64(secs*44100)
}

// swapOneInstrument: phase is special — channels 0 and 1 MUST share the same
// program so the phase-shift effect works (two voices of the SAME instrument
// drifting against each other). When swapping the marimba family, swap
// both 0 and 1 to the same program.
func (a *Phase) swapOneInstrument() {
	// Roll: 60% swap mallet pair, 40% swap pad/bass.
	if a.rng.Float64() < 0.60 {
		alts := phaseChannelAlternatives[0]
		newProg := alts[a.rng.Intn(len(alts))]
		a.core.setProgram(0, newProg)
		a.core.setProgram(1, newProg)
		return
	}
	channels := []int32{2, 3}
	ch := channels[a.rng.Intn(len(channels))]
	a.core.programSwap(ch, phaseChannelAlternatives[ch], a.rng)
}

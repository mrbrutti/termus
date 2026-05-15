package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"

	"github.com/mrbrutti/termus/internal/synth"
)

var _ Algorithm = (*Phase)(nil)
var _ SF2Reverberator = (*Phase)(nil)

// Phase is a Steve-Reich-style phase-shift algorithm. Two mallet voices
// (Marimba + Vibraphone) play the same 8-note pentatonic pattern at slightly
// different tempos; over minutes they drift in and out of unison, creating
// shifting polyrhythms — the canonical Music-for-18-Musicians effect.
//
// Underneath:
//   - A slow harmonic bed of choir + FM-EP pad on long held chord tones.
//   - A sub-bass pedal that drones the chord root for the entire chord cycle.
//   - A high "crotales" track that triggers very occasionally for a glint of
//     metallic upper-partial on top of the interlocking mallets.
//
// The harmonic field shifts every ~60 s across 4 chord centers, so the same
// phasing pattern is reheard against different tonal contexts.
//
// Preferred SF: fm-dx (DX-style metallic mallets + sustained FM textures —
// closer to Reich's vibe/marimba ensemble + Yamaha electric-piano cluster).
type Phase struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
	rng  *rand.Rand

	rootMidi  int
	keyOffset int

	chordRoots      []int // 4 chord-root offsets from key center
	currentChordIdx int

	samplesElapsed int64
	nextChordAt    int64
	nextDriftAt    int64
}

func NewPhase(sf *meltysynth.SoundFont) *Phase { return &Phase{sf: sf} }

func (a *Phase) Name() string { return "phase" }

func (a *Phase) currentRoot() int { return a.rootMidi + a.keyOffset }

func (a *Phase) Seed(seedVal int64) {
	a.rng = rand.New(rand.NewSource(seedVal)) //nolint:gosec
	a.rootMidi = 48 + a.rng.Intn(7) // C3..F#3
	a.keyOffset = 0
	a.samplesElapsed = 0
	a.scheduleNextDrift()

	// 4-chord harmonic bed cycling slowly — chord roots are scale degrees
	// from pentatonic minor (so all melodic notes always fit).
	a.chordRoots = []int{
		0,
		scalePentatonicMinor[2],
		scalePentatonicMinor[3],
		scalePentatonicMinor[1],
	}
	a.currentChordIdx = 0
	a.scheduleNextChord()

	core, err := newSF2Core(a.sf, 2.8, seedVal)
	if err != nil {
		a.core = nil
		return
	}

	// Channel layout:
	//   0 — Vibraphone     (program 11)  phase voice A
	//   1 — Marimba        (program 12)  phase voice B (slightly faster)
	//   2 — Choir Aahs     (program 52)  pad bed
	//   3 — Electric Piano (program 4)   FM-EP harmonic backing
	//   4 — Synth Bass 1   (program 38)  sub-bass pedal
	//   5 — Glockenspiel   (program 9)   high crotales accent
	core.setProgram(0, 11)
	core.setProgram(1, 12)
	core.setProgram(2, 52)
	core.setProgram(3, 4)
	core.setProgram(4, 38)
	core.setProgram(5, 9)
	core.setPan(0, 44) // vibes left
	core.setPan(1, 84) // marimba right
	core.setPan(2, 64)
	core.setPan(3, 56)
	core.setPan(4, 64)
	core.setPan(5, 100)

	// Mallets bright (attack transients carry the phase effect); pad darker.
	core.setChannelCutoff(0, 100)
	core.setChannelCutoff(1, 100)
	core.setChannelCutoff(2, 64)
	core.setChannelCutoff(3, 72)
	core.setChannelCutoff(4, 50)
	core.setChannelCutoff(5, 120)

	// Slow pad LFO for breathing.
	core.addFilterLFO(2, 1.0/18.0, 64, 24)
	core.addFilterLFO(3, 1.0/23.0, 76, 18)

	// Heavy reverb on mallets — the long tail is what blurs the two voices'
	// interlocking patterns into one continuous shimmer.
	core.setReverbSend(0, 115)
	core.setReverbSend(1, 115)
	core.setReverbSend(2, 96)
	core.setReverbSend(3, 80)
	core.setReverbSend(4, 30)
	core.setReverbSend(5, 120)
	core.setChorusSend(2, 32)

	// Pre-install hall reverb if the user hasn't installed one — phase loves
	// space.
	if core.convL == nil {
		ir := synth.SyntheticHallIR(seedVal)
		core.setConvolutionIR(ir, 0.45)
	}

	// --- The phase pattern: 8 pentatonic notes (descending then ascending).
	// Same pattern on both vibe and marimba.
	figure := a.makePhaseFigure()

	// Cycle period: 6.5–8.5 s for the 8-note pattern → ~1 note per second-ish.
	basePeriod := 6.5 + 2.0*a.rng.Float64()
	// Drift ratio: voice B's period is 0.7% shorter than voice A's, so each
	// pass voice B "gains" ~0.05 s. After many minutes they fully wrap around.
	const driftRatio = 0.993

	// --- Vibraphone voice A: fixed tempo. NO timing jitter — phase technique
	// depends on each voice's tempo being precisely defined.
	core.addTrack(SF2Track{
		Channel: 0, Velocity: 78, Notes: append([]int{}, figure...),
		PeriodSec: basePeriod, Phase01: 0,
		VelocityJitter: 6,
	})
	// --- Marimba voice B: slightly faster.
	core.addTrack(SF2Track{
		Channel: 1, Velocity: 72, Notes: append([]int{}, figure...),
		PeriodSec: basePeriod * driftRatio, Phase01: 0,
		VelocityJitter: 6,
	})

	// --- Choir aahs pad: 2 chord-tone voices (thinned from 3 — the mallet
	// patterns are the focus, the pad just provides a tonal ground).
	for voice := 0; voice < 2; voice++ {
		v := voice
		core.addTrack(SF2Track{
			Channel: 2, Velocity: 38, Notes: []int{a.padTone(v)},
			PeriodSec: 19.3 + 9*float64(v), Phase01: a.rng.Float64(),
			MutationRate: 0.40,
			MutateOne:    func(_ int, _ int) int { return a.padTone(v) },
			VelocityJitter: 4, TimingJitterSec: 0.05,
		})
	}

	// --- FM-EP harmonic backing: 1 voice in upper register (thinned from 2).
	core.addTrack(SF2Track{
		Channel: 3, Velocity: 36, Notes: []int{a.padTone(1) + 12},
		PeriodSec: 27.3, Phase01: a.rng.Float64(),
		MutationRate: 0.30,
		MutateOne:    func(_ int, _ int) int { return a.padTone(1) + 12 },
		VelocityJitter: 4, TimingJitterSec: 0.06,
	})

	// --- Sub-bass pedal: chord root, very slow retrigger.
	core.addTrack(SF2Track{
		Channel: 4, Velocity: 60, Notes: []int{a.bassRoot()},
		PeriodSec: 41.7, Phase01: 0,
		MutationRate: 0.50,
		MutateOne:    func(_ int, _ int) int { return a.bassRoot() },
		VelocityJitter: 4, TimingJitterSec: 0.03,
	})

	// --- Glockenspiel crotales: very rare high accents — 1 hit every ~45 s.
	core.addTrack(SF2Track{
		Channel: 5, Velocity: 40, Notes: []int{a.crotalesNote()},
		PeriodSec: 47.3, Phase01: a.rng.Float64(),
		MutationRate: 0.50,
		MutateOne:    func(_ int, _ int) int { return a.crotalesNote() },
		VelocityJitter: 14, TimingJitterSec: 0.10,
	})

	a.core = core
}

// makePhaseFigure builds the literal Steve Reich "Piano Phase" (1967)
// pattern: 12 sixteenth notes using 5 distinct pitches (E4, F#4, B4, C#5,
// D5). The pattern is asymmetric so each phase offset sounds perceptually
// distinct from the others.
//
// Original Piano Phase: E4 F#4 B4 C#5 D5 F#4 E4 F#4 B4 C#5 D5 F#4
// We transpose into the current key for variety while preserving the
// scale-degree intervals.
func (a *Phase) makePhaseFigure() []int {
	// Scale-degree pattern from Piano Phase (in F# dorian-ish):
	//   E4 = root - 2  → degree -1
	//   F#4 = root     → degree 0
	//   B4 = root + 5  → degree 3
	//   C#5 = root + 7 → degree 4
	//   D5 = root + 8  → degree 5
	// 12-note sequence:
	pattern := []int{-1, 0, 3, 4, 5, 0, -1, 0, 3, 4, 5, 0}
	// Use dorian/aeolian-ish scale so the b3 lands correctly.
	scale := []int{0, 2, 3, 5, 7, 8, 10}
	root := a.currentRoot() + 24 // around C5
	out := make([]int, len(pattern))
	for i, deg := range pattern {
		oct := 0
		for deg < 0 {
			deg += len(scale)
			oct--
		}
		for deg >= len(scale) {
			deg -= len(scale)
			oct++
		}
		key := root + scale[deg] + 12*oct
		for key < 72 {
			key += 12
		}
		for key > 88 {
			key -= 12
		}
		out[i] = key
	}
	return out
}

// padTone returns one chord-tone in the mid register for the pad bed.
func (a *Phase) padTone(voice int) int {
	if len(a.chordRoots) == 0 {
		return 60
	}
	cr := a.chordRoots[a.currentChordIdx]
	// Add octave bump per voice so 3 voices spread across the register.
	key := a.currentRoot() + cr + scalePentatonicMinor[voice%len(scalePentatonicMinor)] + 12 + 12*voice
	for key < 60 {
		key += 12
	}
	for key > 84 {
		key -= 12
	}
	return key
}

// bassRoot returns the chord root in the bass register.
func (a *Phase) bassRoot() int {
	if len(a.chordRoots) == 0 {
		return 36
	}
	cr := a.chordRoots[a.currentChordIdx]
	key := a.currentRoot() + cr - 24
	for key > 48 {
		key -= 12
	}
	for key < 24 {
		key += 12
	}
	return key
}

// crotalesNote returns a very-high chord-tone for the glockenspiel accents
// (C6–C7).
func (a *Phase) crotalesNote() int {
	if len(a.chordRoots) == 0 {
		return 84
	}
	cr := a.chordRoots[a.currentChordIdx]
	deg := scalePentatonicMinor[a.rng.Intn(len(scalePentatonicMinor))]
	key := a.currentRoot() + cr + deg + 36
	for key < 84 {
		key += 12
	}
	for key > 96 {
		key -= 12
	}
	return key
}

func (a *Phase) scheduleNextChord() {
	// 50-80 s per chord — slow enough to feel static, fast enough to feel
	// motion over a few minutes.
	secs := 50.0 + 30.0*a.rng.Float64()
	a.nextChordAt = a.samplesElapsed + int64(secs*44100)
}

func (a *Phase) scheduleNextDrift() {
	mins := 4.0 + 3.0*a.rng.Float64()
	a.nextDriftAt = a.samplesElapsed + int64(mins*60*44100)
}

func (a *Phase) advance() {
	if a.samplesElapsed >= a.nextChordAt {
		a.currentChordIdx = (a.currentChordIdx + 1) % len(a.chordRoots)
		a.scheduleNextChord()
	}
	if a.samplesElapsed >= a.nextDriftAt {
		drift := []int{-2, -1, 1, 2}[a.rng.Intn(4)]
		a.keyOffset += drift
		if a.keyOffset > 5 {
			a.keyOffset -= 12
		}
		if a.keyOffset < -5 {
			a.keyOffset += 12
		}
		a.scheduleNextDrift()
	}
}

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
	a.advance()
	a.core.renderInto(left, right)
	a.samplesElapsed += int64(len(left))
}

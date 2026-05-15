package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

var _ Algorithm = (*SF2Markov)(nil)
var _ SF2Reverberator = (*SF2Markov)(nil)

// SF2Markov is the classical algorithm — chamber-music feel with real
// cadential phrase structure. Two 8-bar phrases (antecedent + consequent)
// per chorus, in a tonal idiom: I-IV-V-I, ii-V-I, deceptive cadences.
//
//   - Cello walks an arpeggiated bass line: root on beat 1, then chord-tones
//     in alternation with passing tones — Baroque-style "walking" bass.
//   - Harpsichord continuo plays the chord on every beat (4 hits per bar),
//     evocative of Bach's keyboard parts. Volume is modest so it sits
//     underneath the melody.
//   - Violin carries the melody — 2 notes per bar with bebop-free phrasing,
//     emphasizing strong-beat chord tones with passing-tone ornaments.
//   - String Ensemble pad sustains the full chord for body.
//   - Optional Oboe counter-melody comes in/out every ~60 s for
//     antecedent/consequent feel.
//
// Tempo: 80–120 BPM allegro moderato. 4/4. Preferred SF: timbres-of-heaven
// for orchestral richness.
type SF2Markov struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
	rng  *rand.Rand

	rootMidi  int
	keyOffset int

	progression []classicalChord

	samplesElapsed int64
	nextSectionAt  int64
	oboeOn         *bool

	violinPhrase []int
}

// classicalChord is one bar of harmony.
type classicalChord struct {
	rootSemi int
	tones    []int // root, 3rd, 5th — and 7th for some
	label    string
	majMin   int // 0=major, 1=minor — used by violin melody for scale choice
}

// classicalProgressions: 8-bar cadential progressions in classical voice
// leading. Each ends with a strong V-I or V-I-IV-I cadence.
var classicalProgressions = [][]classicalChord{
	// Major: I-IV-V-I / I-V-vi-V (Beethoven-like)
	{
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I", majMin: 0},
		{rootSemi: 5, tones: []int{0, 4, 7}, label: "IV", majMin: 0},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V", majMin: 0},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I", majMin: 0},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I", majMin: 0},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V", majMin: 0},
		{rootSemi: 9, tones: []int{0, 3, 7}, label: "vi", majMin: 1},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V", majMin: 0},
	},
	// Minor: i-iv-V-i / i-VI-iv-V (Bach-like minor)
	{
		{rootSemi: 0, tones: []int{0, 3, 7}, label: "i", majMin: 1},
		{rootSemi: 5, tones: []int{0, 3, 7}, label: "iv", majMin: 1},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V", majMin: 0},
		{rootSemi: 0, tones: []int{0, 3, 7}, label: "i", majMin: 1},
		{rootSemi: 0, tones: []int{0, 3, 7}, label: "i", majMin: 1},
		{rootSemi: 8, tones: []int{0, 4, 7}, label: "VI", majMin: 0},
		{rootSemi: 5, tones: []int{0, 3, 7}, label: "iv", majMin: 1},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V", majMin: 0},
	},
	// Major: I-vi-IV-V / I-V-IV-I (pop-classical hybrid)
	{
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I", majMin: 0},
		{rootSemi: 9, tones: []int{0, 3, 7}, label: "vi", majMin: 1},
		{rootSemi: 5, tones: []int{0, 4, 7}, label: "IV", majMin: 0},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V", majMin: 0},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I", majMin: 0},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V", majMin: 0},
		{rootSemi: 5, tones: []int{0, 4, 7}, label: "IV", majMin: 0},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I", majMin: 0},
	},
}

func NewSF2Markov(sf *meltysynth.SoundFont) *SF2Markov { return &SF2Markov{sf: sf} }

func (a *SF2Markov) Name() string { return "classical" }

func (a *SF2Markov) currentRoot() int { return a.rootMidi + a.keyOffset }

func (a *SF2Markov) Seed(seedVal int64) {
	a.rng = rand.New(rand.NewSource(seedVal)) //nolint:gosec
	a.rootMidi = 48 + a.rng.Intn(7)
	a.keyOffset = 0
	a.progression = classicalProgressions[a.rng.Intn(len(classicalProgressions))]
	a.samplesElapsed = 0

	oboeStart := true
	a.oboeOn = &oboeStart
	a.scheduleNextSection()

	// Master gain reduced (2.6 → 2.2) — classical was the loudest of the
	// genres (RMS 0.23). Dial it down so it sits with the others in a
	// playlist.
	core, err := newSF2Core(a.sf, 2.2, seedVal)
	if err != nil {
		a.core = nil
		return
	}

	// Channel layout:
	//   0 — Violin             (program 40)  melody
	//   1 — Cello              (program 42)  walking bass
	//   2 — Harpsichord        (program 6)   continuo
	//   3 — String Ensemble 1  (program 48)  sustained pad
	//   4 — Oboe               (program 68)  counter-melody (toggleable)
	core.setProgram(0, 40)
	core.setProgram(1, 42)
	core.setProgram(2, 6)
	core.setProgram(3, 48)
	core.setProgram(4, 68)
	core.setPan(0, 84) // violin right (1st-chair position)
	core.setPan(1, 44) // cello left
	core.setPan(2, 64) // harpsichord center
	core.setPan(3, 64) // strings center
	core.setPan(4, 76) // oboe slightly right

	// Bright, natural voicings — classical instruments don't get the
	// muffled/dark treatment that lo-fi gets.
	core.setChannelCutoff(0, 110)
	core.setChannelCutoff(1, 90)
	core.setChannelCutoff(2, 100)
	core.setChannelCutoff(3, 84)
	core.setChannelCutoff(4, 100)

	// Reverb sends: chamber-hall sized. Violin + oboe most wet (the solo
	// voices); cello and harpsichord moderate; strings sit IN the reverb.
	core.setReverbSend(0, 84)
	core.setReverbSend(1, 56)
	core.setReverbSend(2, 50)
	core.setReverbSend(3, 100)
	core.setReverbSend(4, 88)

	// Pick a progression and tempo (80–120 BPM allegro moderato).
	bpm := 80.0 + 40.0*a.rng.Float64()
	beatSec := 60.0 / bpm
	barSec := beatSec * 4
	numBars := len(a.progression)
	cycleSec := barSec * float64(numBars)

	// --- Cello walking bass: 4 notes per bar, arpeggiated chord tones with
	// passing tones — classical "Alberti bass" / "walking bass" hybrid.
	cellaNotes := make([]int, 4*numBars)
	for i := range cellaNotes {
		cellaNotes[i] = a.celloPattern(i)
	}
	core.addTrack(SF2Track{
		Channel: 1, Velocity: 70, Notes: cellaNotes,
		PeriodSec: cycleSec, Phase01: 0,
		MutationRate: 0.10,
		MutateOne:    func(slot int, _ int) int { return a.celloPattern(slot) },
		VelocityJitter: 8, TimingJitterSec: 0.008,
	})

	// --- Harpsichord continuo: 2 stabs per bar (beats 1 and 3 — half-time
	// continuo). The previous "stab on every beat" was too dense — real
	// Baroque continuo breathes between accents.
	for voiceIdx := 0; voiceIdx < 3; voiceIdx++ {
		voice := voiceIdx
		stabs := make([]int, 2*numBars)
		for i := range stabs {
			// slot i maps to beat 1 or 3 of bar i/2
			stabs[i] = a.harpsichordTone(i*2, voice) // i*2 because we're firing on every other beat
		}
		core.addTrack(SF2Track{
			Channel: 2, Velocity: 48, Notes: stabs,
			PeriodSec: cycleSec, Phase01: 0,
			MutationRate: 0.05,
			MutateOne: func(slot int, _ int) int {
				return a.harpsichordTone(slot*2, voice)
			},
			VelocityJitter: 8, TimingJitterSec: 0.006,
		})
	}

	// --- Violin melody: a coherent 8-note phrase spread across the 8-bar
	// form (2 notes/bar). Pre-computed so the melody has a recognizable
	// arc — antecedent half rises, consequent half resolves — instead of
	// each note being chosen independently.
	a.violinPhrase = a.makeViolinPhrase(2 * numBars)
	core.addTrack(SF2Track{
		Channel: 0, Velocity: 82, Notes: a.violinPhrase,
		PeriodSec: cycleSec, Phase01: 0,
		MutationRate: 0.08, // very light mutation — keep the phrase shape
		MutateOne: func(slot int, _ int) int {
			return a.violinPhrase[slot%len(a.violinPhrase)]
		},
		VelocityJitter: 12, TimingJitterSec: 0.020,
	})

	// --- String pad: holds the chord for the entire bar. 3 voices for the
	// triad, on incommensurate periods so subtle re-attacks happen at
	// different times across the chord (Baroque-style staggered breathing).
	for voiceIdx := 0; voiceIdx < 3; voiceIdx++ {
		voice := voiceIdx
		notes := make([]int, numBars)
		for i := range notes {
			notes[i] = a.padTone(i, voice)
		}
		core.addTrack(SF2Track{
			Channel: 3, Velocity: 38, Notes: notes,
			PeriodSec: cycleSec, Phase01: float64(voice) * 0.07,
			MutationRate: 0.15,
			MutateOne:    func(slot int, _ int) int { return a.padTone(slot, voice) },
			VelocityJitter: 4, TimingJitterSec: 0.020,
		})
	}

	// --- Oboe counter-melody: 1 note per bar in the alto register, on beat 3.
	// Toggleable section dynamics (antecedent vs consequent feel).
	oboeNotes := make([]int, numBars)
	for i := range oboeNotes {
		oboeNotes[i] = a.oboeNote(i)
	}
	core.addTrack(SF2Track{
		Channel: 4, Velocity: 56, Notes: oboeNotes,
		PeriodSec: cycleSec,
		Phase01:   0.5 / float64(numBars), // beat 3 of bar 0
		MutationRate: 0.30,
		MutateOne:    func(slot int, _ int) int { return a.oboeNote(slot) },
		VelocityJitter: 10, TimingJitterSec: 0.025,
		Enabled: a.oboeOn,
	})

	a.core = core
}

// celloPattern returns one beat of the walking-bass cello part. Per bar:
//
//	beat 1: chord root (low)
//	beat 2: chord 5th
//	beat 3: chord root (octave up)
//	beat 4: chord 3rd (or chromatic passing tone to next chord's root)
func (a *SF2Markov) celloPattern(slot int) int {
	totalBeats := 4 * len(a.progression)
	slot = ((slot % totalBeats) + totalBeats) % totalBeats
	bar := slot / 4
	beat := slot % 4
	chord := a.progression[bar]
	root := a.currentRoot() + chord.rootSemi - 12
	switch beat {
	case 0:
		return root
	case 1:
		return root + chord.tones[2] // 5th
	case 2:
		return root + 12 // root octave up
	case 3:
		// Passing tone to next chord's root.
		nextBar := (bar + 1) % len(a.progression)
		nextRoot := a.currentRoot() + a.progression[nextBar].rootSemi - 12
		if nextRoot > root {
			return nextRoot - 2 // step from below by whole step
		}
		return nextRoot + 2 // step from above by whole step
	}
	return root
}

// harpsichordTone returns one voice (0..2 = root/3rd/5th) of the continuo
// stab on the current beat. All voices land on the same beat for a chord
// stab effect.
func (a *SF2Markov) harpsichordTone(slot, voice int) int {
	totalBeats := 4 * len(a.progression)
	slot = ((slot % totalBeats) + totalBeats) % totalBeats
	bar := slot / 4
	chord := a.progression[bar]
	if voice >= len(chord.tones) {
		voice = len(chord.tones) - 1
	}
	key := a.currentRoot() + chord.rootSemi + chord.tones[voice]
	for key < 60 {
		key += 12
	}
	for key > 76 {
		key -= 12
	}
	return key
}

// makeViolinPhrase builds the violin melody for the entire 8-bar form
// (numSlots = 2 * numBars). Uses a peak-and-fall contour on the major scale
// — every classical period closes with a downward resolution to the tonic.
// Pentatonic-major gives consonance against every chord in the progression.
func (a *SF2Markov) makeViolinPhrase(numSlots int) []int {
	// Pick contour: bias toward "peak-and-fall" or "wave" — both have the
	// classical-feeling antecedent/consequent shape.
	contour := melodicPhrases[2+a.rng.Intn(2)] // {2,3} = peak-fall, wave
	if len(contour) != numSlots {
		extended := make([]int, numSlots)
		for i := range extended {
			extended[i] = contour[i%len(contour)]
		}
		contour = extended
	}
	// Violin register: G4–E6 (67–88). Start at +24 semitones above currentRoot.
	root := a.currentRoot() + 24
	notes := applyPhraseToScale(contour, majorPentatonic, root, 2, 0)
	for i, k := range notes {
		for k < 67 {
			k += 12
		}
		for k > 88 {
			k -= 12
		}
		notes[i] = k
	}
	return notes
}

// violinNote returns one of the 2 melody notes for the i-th half-bar slot.
// Picks chord tones with stepwise resolution for "composed" voice leading.
func (a *SF2Markov) violinNote(slot int) int {
	totalSlots := 2 * len(a.progression)
	slot = ((slot % totalSlots) + totalSlots) % totalSlots
	bar := slot / 2
	chord := a.progression[bar]
	// First half: 5th of chord; second half: 3rd. Gives downward voice leading.
	var idx int
	if slot%2 == 0 {
		idx = 2 // 5th
	} else {
		idx = 1 // 3rd
	}
	if idx >= len(chord.tones) {
		idx = len(chord.tones) - 1
	}
	key := a.currentRoot() + chord.rootSemi + chord.tones[idx]
	// Violin sweet spot: G4–E6 (67..88).
	for key < 67 {
		key += 12
	}
	for key > 88 {
		key -= 12
	}
	// 25% chance to embellish with a scale neighbor (passing tone).
	if a.rng.Float64() < 0.25 {
		var scale []int
		if chord.majMin == 0 {
			scale = []int{0, 2, 4, 5, 7, 9, 11}
		} else {
			scale = []int{0, 2, 3, 5, 7, 8, 11} // harmonic minor
		}
		deg := scale[a.rng.Intn(len(scale))]
		key = a.currentRoot() + deg + 12*((key-a.currentRoot())/12)
		for key < 67 {
			key += 12
		}
		for key > 88 {
			key -= 12
		}
	}
	return key
}

// padTone returns one chord-tone (0..2) for the pad-bed track in the mid
// register.
func (a *SF2Markov) padTone(slot, voice int) int {
	bar := ((slot % len(a.progression)) + len(a.progression)) % len(a.progression)
	chord := a.progression[bar]
	if voice >= len(chord.tones) {
		voice = len(chord.tones) - 1
	}
	key := a.currentRoot() + chord.rootSemi + chord.tones[voice]
	for key < 55 {
		key += 12
	}
	for key > 76 {
		key -= 12
	}
	return key
}

// oboeNote returns one note of the oboe counter-melody — the 3rd of the
// current chord, in the oboe register (around C5–C6).
func (a *SF2Markov) oboeNote(slot int) int {
	bar := ((slot % len(a.progression)) + len(a.progression)) % len(a.progression)
	chord := a.progression[bar]
	if len(chord.tones) < 2 {
		return 72
	}
	key := a.currentRoot() + chord.rootSemi + chord.tones[1] + 12
	for key < 72 {
		key += 12
	}
	for key > 84 {
		key -= 12
	}
	return key
}

func (a *SF2Markov) scheduleNextSection() {
	secs := 50.0 + 60.0*a.rng.Float64()
	a.nextSectionAt = a.samplesElapsed + int64(secs*44100)
}

func (a *SF2Markov) advance() {
	if a.samplesElapsed >= a.nextSectionAt {
		*a.oboeOn = !*a.oboeOn
		a.scheduleNextSection()
	}
}

func (a *SF2Markov) SetReverbIR(ir []float64, wet float64) {
	if a.core != nil {
		a.core.setConvolutionIR(ir, wet)
	}
}

func (a *SF2Markov) Next(left, right []float64) {
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

package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

var _ Algorithm = (*SF2Pentatonic)(nil)
var _ SF2Reverberator = (*SF2Pentatonic)(nil)

// SF2Pentatonic is the lullaby algorithm — a slow music-box waltz in 3/4. The
// previous version was a random pentatonic walk; this rewrite gives it real
// lullaby structure:
//
//   - 3/4 waltz time at 56–72 BPM (the cradle tempo of Brahms' Wiegenlied).
//   - Harp bass on beat 1, chord accompaniment on beats 2 and 3 (the
//     classic "oom-pah-pah" waltz figure).
//   - Music-box melody on the strong beats with sparse upper-register
//     ornaments from celesta and glockenspiel.
//   - Simple cadential chord progressions in major key (I-vi-IV-V,
//     I-IV-ii-V, I-V-vi-IV) — the kind of changes the listener anticipates
//     in a real lullaby.
//   - Soft choir-aahs pad underneath for warmth.
//
// Preferred SF: fairy-tale (music box, celesta, glockenspiel, harp all in one
// bank).
type SF2Pentatonic struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
	rng  *rand.Rand

	rootMidi  int
	keyOffset int

	progression []lullabyChord

	samplesElapsed int64
	nextSectionAt  int64
	glockOn        *bool

	melodyPhrase []int
}

// lullabyChord is one bar of harmony in 3/4. tones are semitone offsets from
// the chord root for the accompanying voicing (3-note triad). rootSemi is the
// chord root's semitone offset from the key center.
type lullabyChord struct {
	rootSemi int
	tones    []int // 3 voicing tones (root, 3rd, 5th typically)
	label    string
}

// lullabyProgressions: 8-bar cadential progressions in major key.
var lullabyProgressions = [][]lullabyChord{
	// Brahms' "Wiegenlied" style: I-V-I-V / I-V-IV-I (simplified to 8 bars).
	{
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V"},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V"},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 5, tones: []int{0, 4, 7}, label: "IV"},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V"},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
	},
	// I-vi-IV-V doo-wop / Hush Little Baby:
	{
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 9, tones: []int{0, 3, 7}, label: "vi"},
		{rootSemi: 5, tones: []int{0, 4, 7}, label: "IV"},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V"},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 9, tones: []int{0, 3, 7}, label: "vi"},
		{rootSemi: 5, tones: []int{0, 4, 7}, label: "IV"},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V"},
	},
	// I-IV-V-I / vi-ii-V-I (gentle plagal-dominant alternation):
	{
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 5, tones: []int{0, 4, 7}, label: "IV"},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V"},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 9, tones: []int{0, 3, 7}, label: "vi"},
		{rootSemi: 2, tones: []int{0, 3, 7}, label: "ii"},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V"},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
	},
}

func NewSF2Pentatonic(sf *meltysynth.SoundFont) *SF2Pentatonic {
	return &SF2Pentatonic{sf: sf}
}

func (a *SF2Pentatonic) Name() string { return "lullaby" }

func (a *SF2Pentatonic) currentRoot() int { return a.rootMidi + a.keyOffset }

func (a *SF2Pentatonic) Seed(seedVal int64) {
	a.rng = rand.New(rand.NewSource(seedVal)) //nolint:gosec
	// C4 / D4 / E4 / G4 — sweet upper-mid range for music box.
	a.rootMidi = 48 + a.rng.Intn(8) // C3..G3
	a.keyOffset = 0
	a.progression = lullabyProgressions[a.rng.Intn(len(lullabyProgressions))]
	a.samplesElapsed = 0

	glockStart := false // glockenspiel starts off — appears later for variety
	a.glockOn = &glockStart
	a.scheduleNextSection()

	// Master gain raised (2.4 → 3.0) — lullaby was significantly quieter
	// than the others, made it disappear in a mixed playlist.
	core, err := newSF2Core(a.sf, 3.0, seedVal)
	if err != nil {
		a.core = nil
		return
	}

	// Channel layout:
	//   0 — Orchestral Harp (program 46)  bass on beat 1
	//   1 — Music Box       (program 10)  main melody
	//   2 — Celesta         (program 8)   accompaniment on beats 2-3
	//   3 — Glockenspiel    (program 9)   ornament (toggleable)
	//   4 — Choir Aahs      (program 52)  soft pad bed
	core.setProgram(0, 46)
	core.setProgram(1, 10)
	core.setProgram(2, 8)
	core.setProgram(3, 9)
	core.setProgram(4, 52)
	core.setPan(0, 64)
	core.setPan(1, 56)
	core.setPan(2, 80)
	core.setPan(3, 100)
	core.setPan(4, 64)

	// All bright (music box / celesta / glock are pristine); pad darkened.
	core.setChannelCutoff(0, 96)
	core.setChannelCutoff(1, 120)
	core.setChannelCutoff(2, 120)
	core.setChannelCutoff(3, 120)
	core.setChannelCutoff(4, 64)

	// Gentle pad LFO.
	core.addFilterLFO(4, 1.0/15.0, 64, 18)

	// Lots of reverb on the bell-family, moderate on pad/harp.
	core.setReverbSend(0, 70)
	core.setReverbSend(1, 100)
	core.setReverbSend(2, 110)
	core.setReverbSend(3, 110)
	core.setReverbSend(4, 84)
	core.setChorusSend(1, 24)
	core.setChorusSend(2, 24)

	// Tempo: 48–62 BPM. True cradle-rocking tempo — slower than the previous
	// 56-72 range, which was sitting closer to "slow song" than "lullaby".
	bpm := 48.0 + 14.0*a.rng.Float64()
	beatSec := 60.0 / bpm
	barSec := beatSec * 3 // 3/4 time
	numBars := len(a.progression)
	cycleSec := barSec * float64(numBars)

	// --- Harp bass on beat 1: 1 hit per bar, root of the chord.
	bassNotes := make([]int, numBars)
	for i := range bassNotes {
		bassNotes[i] = a.bassRoot(i)
	}
	core.addTrack(SF2Track{
		Channel: 0, Velocity: 70, Notes: bassNotes,
		PeriodSec: cycleSec, Phase01: 0,
		MutationRate: 0.10,
		MutateOne:    func(slot int, _ int) int { return a.bassRoot(slot) },
		VelocityJitter: 8, TimingJitterSec: 0.012,
	})

	// --- Celesta accompaniment on beats 2 and 3: 2 hits per bar, 3rd and 5th
	// of the chord (oom-pah-pah upper voices). Use 2 separate tracks for the
	// two beats so each lands on a uniform grid.
	// Beat 2 starts at bar fraction 1/3, beat 3 at 2/3.
	// 1 hit per bar, evenly spaced at beat positions, with phase offset.
	for ti, beatNum := range []int{1, 2} { // beat 2 and beat 3
		beat := beatNum
		voice := ti
		notes := make([]int, numBars)
		for i := range notes {
			notes[i] = a.compTone(i, voice+1) // 3rd then 5th
		}
		core.addTrack(SF2Track{
			Channel: 2, Velocity: 54, Notes: notes,
			PeriodSec: cycleSec,
			Phase01:   float64(beat) / (3 * float64(numBars)),
			MutationRate: 0.15,
			MutateOne:    func(slot int, _ int) int { return a.compTone(slot, voice+1) },
			VelocityJitter: 10, TimingJitterSec: 0.020,
		})
	}

	// --- Music box melody: pre-computed coherent 8-note phrase in
	// pentatonic major. Cycles through the 8 bars deterministically — the
	// same melody every loop is what makes a lullaby memorable. NO mutation
	// (a music box plays the same tune; that's its charm).
	a.melodyPhrase = a.makeMelodyPhrase(numBars)
	core.addTrack(SF2Track{
		Channel: 1, Velocity: 76, Notes: a.melodyPhrase,
		PeriodSec: cycleSec, Phase01: 0,
		VelocityJitter: 12, TimingJitterSec: 0.018,
	})

	// --- Glockenspiel ornament: very high, sparse — 1 hit per 4 bars on the
	// 3rd of the chord.
	glockNotes := make([]int, numBars/4)
	if len(glockNotes) < 1 {
		glockNotes = make([]int, 1)
	}
	for i := range glockNotes {
		glockNotes[i] = a.compTone(i*4, 1) + 24
	}
	core.addTrack(SF2Track{
		Channel: 3, Velocity: 48, Notes: glockNotes,
		PeriodSec: cycleSec,
		Phase01:   0.125 / float64(numBars), // slight offset from beat 1
		MutationRate: 0.35,
		MutateOne: func(slot int, _ int) int {
			return a.compTone(slot*4, 1) + 24
		},
		VelocityJitter: 12, TimingJitterSec: 0.030,
		Enabled: a.glockOn,
	})

	// --- Choir aahs pad: very sustained, slow retrigger. 2 voices.
	for ti, period := range []float64{13.3, 19.7} {
		voice := ti
		core.addTrack(SF2Track{
			Channel: 4, Velocity: 36, Notes: []int{a.padNote(voice)},
			PeriodSec: period, Phase01: a.rng.Float64(),
			MutationRate: 0.20,
			MutateOne:    func(_ int, _ int) int { return a.padNote(voice) },
			VelocityJitter: 4, TimingJitterSec: 0.040,
		})
	}

	a.core = core
}

// bassRoot returns the harp-bass root for the i-th bar. Lower octave to
// support the melody above.
func (a *SF2Pentatonic) bassRoot(slot int) int {
	bar := ((slot % len(a.progression)) + len(a.progression)) % len(a.progression)
	chord := a.progression[bar]
	key := a.currentRoot() + chord.rootSemi - 12
	for key > 48 {
		key -= 12
	}
	for key < 32 {
		key += 12
	}
	return key
}

// compTone returns the i-th bar's chord-tone at the requested interval
// position (1=3rd, 2=5th, 0=root). Placed in the mid register (C4–B4).
func (a *SF2Pentatonic) compTone(slot, idx int) int {
	bar := ((slot % len(a.progression)) + len(a.progression)) % len(a.progression)
	chord := a.progression[bar]
	if idx >= len(chord.tones) {
		idx = len(chord.tones) - 1
	}
	key := a.currentRoot() + chord.rootSemi + chord.tones[idx]
	for key < 60 {
		key += 12
	}
	for key > 76 {
		key -= 12
	}
	return key
}

// makeMelodyPhrase builds the music-box's coherent melodic contour for the
// entire 8-bar form. Pentatonic major guarantees consonance with every chord
// in the progression. Picks one of the stylized contours so the listener
// hears an antecedent → consequent → resolution shape rather than random
// notes.
func (a *SF2Pentatonic) makeMelodyPhrase(numBars int) []int {
	contour := pickMelodicPhrase(a.rng)
	// Resize to match the bar count by repeating or trimming the contour.
	if len(contour) != numBars {
		extended := make([]int, numBars)
		for i := range extended {
			extended[i] = contour[i%len(contour)]
		}
		contour = extended
	}
	// Music box sweet spot is C5–C6 (60..72). Start in mid register.
	root := a.currentRoot() + 24
	notes := applyPhraseToScale(contour, majorPentatonic, root, 2, 0)
	for i, k := range notes {
		for k < 72 {
			k += 12
		}
		for k > 88 {
			k -= 12
		}
		notes[i] = k
	}
	return notes
}

// melodyNote returns the music-box melody note for the i-th bar. Mostly
// chord tones (so it always resolves), occasionally a scale tone for color.
func (a *SF2Pentatonic) melodyNote(slot int) int {
	bar := ((slot % len(a.progression)) + len(a.progression)) % len(a.progression)
	chord := a.progression[bar]
	var key int
	if a.rng.Float64() < 0.75 {
		// Chord tone, raised into the music-box register (C5–C6).
		idx := a.rng.Intn(len(chord.tones))
		key = a.currentRoot() + chord.rootSemi + chord.tones[idx] + 12
	} else {
		// Major-scale tone for stepwise motion.
		major := []int{0, 2, 4, 5, 7, 9, 11}
		deg := major[a.rng.Intn(len(major))]
		key = a.currentRoot() + deg + 12
	}
	for key < 72 {
		key += 12
	}
	for key > 88 {
		key -= 12
	}
	return key
}

// padNote returns a soft chord-tone for the choir-aahs bed. Cycles through
// the first 3 chords of the progression.
func (a *SF2Pentatonic) padNote(voice int) int {
	if len(a.progression) == 0 {
		return 60
	}
	chord := a.progression[a.rng.Intn(len(a.progression))]
	idx := voice % len(chord.tones)
	key := a.currentRoot() + chord.rootSemi + chord.tones[idx] + 12
	for key < 60 {
		key += 12
	}
	for key > 80 {
		key -= 12
	}
	return key
}

func (a *SF2Pentatonic) scheduleNextSection() {
	secs := 60.0 + 60.0*a.rng.Float64()
	a.nextSectionAt = a.samplesElapsed + int64(secs*44100)
}

func (a *SF2Pentatonic) advance() {
	if a.samplesElapsed >= a.nextSectionAt {
		*a.glockOn = !*a.glockOn
		a.scheduleNextSection()
	}
}

func (a *SF2Pentatonic) SetReverbIR(ir []float64, wet float64) {
	if a.core != nil {
		a.core.setConvolutionIR(ir, wet)
	}
}

func (a *SF2Pentatonic) Next(left, right []float64) {
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

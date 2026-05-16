package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"

	"github.com/mrbrutti/termus/internal/synth"
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
	barSamples  int64

	samplesElapsed int64
	nextSectionAt  int64
	glockOn        *bool
	section        FormSection

	melodyPhrase []int
	glockMotifs  MotifMemory
	profile      ControlProfile
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

func (a *SF2Pentatonic) ApplyControlProfile(profile ControlProfile) {
	a.profile = profileOrDefault(profile)
}

func (a *SF2Pentatonic) currentRoot() int { return a.rootMidi + a.keyOffset }

func (a *SF2Pentatonic) Seed(seedVal int64) {
	a.rng = rand.New(rand.NewSource(seedVal)) //nolint:gosec
	// C4 / D4 / E4 / G4 — sweet upper-mid range for music box.
	a.rootMidi = 48 + a.rng.Intn(8) // C3..G3
	a.keyOffset = 0
	a.progression = lullabyProgressions[a.rng.Intn(len(lullabyProgressions))]
	a.samplesElapsed = 0
	a.glockMotifs = a.makeGlockMotifs()

	glockStart := false // glockenspiel starts off — appears later for variety
	a.glockOn = &glockStart
	a.scheduleNextSection()
	a.syncSection()

	// Master gain raised aggressively (2.4 → 3.8) — like bells, lullaby's
	// sparse 3/4-waltz content has long pauses between hits, so the
	// effective RMS is much lower than a denser genre even with comparable
	// peak levels.
	core, err := newSF2Core(a.sf, 3.8, seedVal)
	if err != nil {
		a.core = nil
		return
	}
	applyMaxSF2Palette(core, "lullaby")

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

	// Tempo: 72–84 BPM. Research surfaced that Brahms Wiegenlied is
	// performed at 72-84 — slower feels like "slow song", not lullaby. The
	// rocking feel comes from the dotted-quarter + eighth + quarter rhythm,
	// not from absolute tempo.
	bpm := 72.0 + 12.0*a.rng.Float64()
	beatSec := 60.0 / bpm
	barSec := beatSec * 3 // 3/4 time
	a.barSamples = int64(barSec * float64(synth.SampleRate))
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
		MutationRate:   0.10,
		MutateOne:      func(slot int, _ int) int { return a.bassRoot(slot) },
		Gate:           0.82,
		Legato:         true,
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
			PeriodSec:      cycleSec,
			Phase01:        float64(beat) / (3 * float64(numBars)),
			MutationRate:   0.15,
			MutateOne:      func(slot int, _ int) int { return a.compTone(slot, voice+1) },
			Gate:           0.46,
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
		Gate:   0.88,
		Legato: true,
		ResolveExpression: func(slot int, key int) SF2ExpressionCurve {
			return SF2ExpressionCurve{Start: 88, Peak: 108, End: 92, PeakAt01: 0.40}
		},
		ResolveModWheel:    func(slot int, key int) SF2ExpressionCurve { return gentleVibratoCurve(0, 10, 4) },
		ResolveBrightness:  func(slot int, key int) SF2ExpressionCurve { return brightnessBloomCurve(110, 124, 112) },
		ResolveDetuneCents: slotDetunePattern(-1, 2, -2, 1),
		VelocityJitter:     12, TimingJitterSec: 0.018,
	})

	// --- Glockenspiel ornament: high answering figures that only answer the
	// music-box phrase in later bars and cadences.
	glockNotes := make([]int, numBars)
	core.addTrack(SF2Track{
		Channel: 3, Velocity: 48, Notes: glockNotes,
		PeriodSec: cycleSec,
		Phase01:   0.125 / float64(numBars), // slight offset from beat 1
		ResolveNote: func(slot int, _ int) int {
			return a.glockNoteAt(slot)
		},
		ResolveBrightness:  func(slot int, key int) SF2ExpressionCurve { return brightnessBloomCurve(118, 127, 120) },
		ResolveDetuneCents: slotDetunePattern(0, 2, -1, 3),
		Gate:               0.42,
		VelocityJitter:     12, TimingJitterSec: 0.030,
		Enabled: a.glockOn,
	})

	// --- Choir aahs pad: very sustained, slow retrigger. 2 voices.
	for ti, period := range []float64{13.3, 19.7} {
		voice := ti
		core.addTrack(SF2Track{
			Channel: 4, Velocity: 36, Notes: []int{a.padNote(voice)},
			PeriodSec: period, Phase01: a.rng.Float64(),
			MutationRate:       0.20,
			MutateOne:          func(_ int, _ int) int { return a.padNote(voice) },
			ResolveNote:        func(_ int, _ int) int { return a.padNote(voice) },
			ResolveModWheel:    func(slot int, key int) SF2ExpressionCurve { return gentleVibratoCurve(0, 16, 8) },
			ResolveBrightness:  func(slot int, key int) SF2ExpressionCurve { return brightnessBloomCurve(64, 72, 66) },
			ResolveDetuneCents: slotDetunePattern(-2, 1, -1, 2),
			Gate:               0.98,
			Legato:             true,
			VelocityJitter:     4, TimingJitterSec: 0.040,
		})
	}

	a.core = core
	a.applyArrangement()
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

// makeMelodyPhrase builds the music-box's melody. Specifically uses the
// Brahms Wiegenlied Op. 49 No. 4 contour: scale degrees
//
//	3 3 | 5 5 | 3 3 | 5 5 | 4 5 | 6 5 | 4 3 | 2 2
//
// (2 notes per bar across the 8-bar period). The iconic shape is
// repetition on 3rd-and-5th, peak on 6th, stepwise descent 5-4-3-2.
// Uses the major scale (NOT pentatonic) since Brahms is fully diatonic.
func (a *SF2Pentatonic) makeMelodyPhrase(numBars int) []int {
	// Brahms scale-degree contour, converted to 0-based index into the major
	// scale. Degrees 3,5,4,6,2 → indices 2,4,3,5,1.
	contour := []int{
		2, 2, // bar 1: degree 3
		4, 4, // bar 2: degree 5 (the "rising 3rd" jump from previous A)
		2, 2, // bar 3: degree 3
		4, 4, // bar 4: degree 5
		3, 4, // bar 5: degree 4 → 5 (consequent begins)
		5, 4, // bar 6: degree 6 (peak) → 5
		3, 2, // bar 7: stepwise descent
		1, 1, // bar 8: degree 2 (held — close)
	}
	if len(contour) != 2*numBars {
		// Fall back if progression isn't 8 bars: extend by cycling.
		extended := make([]int, 2*numBars)
		for i := range extended {
			extended[i] = contour[i%len(contour)]
		}
		contour = extended
	}
	majorScale := []int{0, 2, 4, 5, 7, 9, 11}
	root := a.currentRoot() + 24
	notes := make([]int, len(contour))
	for i := range contour {
		target := clampMidiToRange(scaleNoteAt(contour, i, majorScale, root, 0, 0), 72, 88)
		chord := a.progression[(i/2)%len(a.progression)]
		chordRoot := a.currentRoot() + chord.rootSemi + 24
		if i%2 == 0 {
			// Strong melodic positions land on chord tones so the lullaby
			// actually outlines the harmony instead of floating over it.
			notes[i] = nearestRelativeNote(target, chordRoot, chord.tones, 72, 88)
			continue
		}
		notes[i] = nearestRelativeNote(target, root, majorScale, 72, 88)
		if i > 0 && absInt(notes[i]-notes[i-1]) > 5 {
			// Weak positions prefer stepwise rocking motion.
			up := nearestRelativeNote(notes[i-1]+2, root, majorScale, 72, 88)
			down := nearestRelativeNote(notes[i-1]-2, root, majorScale, 72, 88)
			if absInt(up-target) <= absInt(down-target) {
				notes[i] = up
			} else {
				notes[i] = down
			}
		}
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

func (a *SF2Pentatonic) currentBar() int {
	if len(a.progression) == 0 || a.barSamples <= 0 {
		return 0
	}
	return int((a.samplesElapsed / a.barSamples) % int64(len(a.progression)))
}

func (a *SF2Pentatonic) glockNoteAt(slot int) int {
	phrase := a.glockMotifs.PhraseFor(a.section.Kind)
	if len(phrase) == 0 {
		return -1
	}
	code := phrase[((slot%len(phrase))+len(phrase))%len(phrase)]
	if code < 0 {
		return -1
	}
	if a.section.Kind != FormCadence && slot%len(phrase) < len(phrase)/2 {
		return -1
	}
	idx := 1 + code%2
	return clampMidiToRange(a.compTone(slot, idx)+24, 84, 100)
}

// padNote returns a soft chord-tone for the choir-aahs bed. It follows the
// current bar's harmony rather than pulling a random chord from the form.
func (a *SF2Pentatonic) padNote(voice int) int {
	if len(a.progression) == 0 {
		return 60
	}
	chord := a.progression[a.currentBar()]
	idx := voice % len(chord.tones)
	key := a.currentRoot() + chord.rootSemi + chord.tones[idx] + 12
	return clampMidiToRange(key, 60, 80)
}

func (a *SF2Pentatonic) scheduleNextSection() {
	secs := 60.0 + 60.0*a.rng.Float64()
	step := a.barSamples * 4
	if step <= 0 {
		step = int64(4 * synth.SampleRate)
	}
	a.nextSectionAt = scheduleQuantizedAfter(a.samplesElapsed, secs, step)
}

func (a *SF2Pentatonic) advance() {
	if a.samplesElapsed >= a.nextSectionAt {
		*a.glockOn = !*a.glockOn
		a.scheduleNextSection()
	}
	a.syncSection()
}

func (a *SF2Pentatonic) DebugStatus() DebugStatus {
	chord := ""
	if len(a.progression) > 0 {
		chord = a.progression[a.currentBar()].label
	}
	return DebugStatus{
		Chord:   chord,
		Section: string(a.section.Kind),
		Bar:     a.currentBar() + 1,
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
	prev := a.samplesElapsed
	a.advance()
	a.core.renderInto(left, right)
	a.samplesElapsed += int64(len(left))
	if crossedQuantizedBoundary(prev, a.samplesElapsed, a.barSamples) {
		a.syncSection()
	}
}

func (a *SF2Pentatonic) makeGlockMotifs() MotifMemory {
	base := []int{-1, 0, -1, 1, -1, 0, -1, 1}
	return MotifMemory{
		A:       base,
		Aprime:  []int{-1, 1, -1, 0, -1, 1, -1, 0},
		B:       []int{-1, 0, 1, -1, -1, 1, 0, -1},
		Cadence: []int{-1, 1, -1, 1, -1, 1, 0, 1},
		Outro:   []int{-1, -1, -1, 0},
	}
}

func (a *SF2Pentatonic) syncSection() {
	a.section = waltzTextureSection(a.currentBar(), len(a.progression), a.glockOn != nil && *a.glockOn)
	if a.samplesElapsed == 0 {
		a.section = sectionTemplate(FormIntro)
	}
	a.applyArrangement()
}

func (a *SF2Pentatonic) applyArrangement() {
	if a.core == nil {
		return
	}
	profile := profileOrDefault(a.profile)
	bass := SectionSceneFor(a.section, RoleBass)
	lead := SectionSceneFor(a.section, RoleLead)
	comp := SectionSceneFor(a.section, RoleComp)
	texture := SectionSceneFor(a.section, RoleTexture)
	reverbDelta := ReverbDelta(profile)
	brightDelta := BrightnessDelta(profile)
	densityDelta := int32(ProfileCentered(profile.Density) * 8)
	droneDelta := DroneDepthDelta(profile)
	a.core.setReverbSend(0, SectionCC(70, bass.ReverbDelta+reverbDelta/3))
	a.core.setReverbSend(1, SectionCC(100, lead.ReverbDelta+reverbDelta))
	a.core.setReverbSend(2, SectionCC(110, comp.ReverbDelta+reverbDelta))
	a.core.setReverbSend(3, SectionCC(110, lead.ReverbDelta+reverbDelta))
	a.core.setReverbSend(4, SectionCC(84, texture.ReverbDelta+reverbDelta))
	a.core.setChannelCutoff(0, SectionCC(96, bass.BrightnessDelta+brightDelta/2))
	a.core.setChannelCutoff(1, SectionCC(120, lead.BrightnessDelta+brightDelta))
	a.core.setChannelCutoff(2, SectionCC(120, comp.BrightnessDelta+brightDelta))
	a.core.setChannelCutoff(3, SectionCC(120, lead.BrightnessDelta+brightDelta))
	a.core.setChannelCutoff(4, SectionCC(64, texture.BrightnessDelta+brightDelta))
	a.core.setChannelExpression(0, SectionCC(100, bass.ExpressionDelta+droneDelta))
	a.core.setChannelExpression(1, SectionCC(108, lead.ExpressionDelta+densityDelta))
	a.core.setChannelExpression(2, SectionCC(102, comp.ExpressionDelta+densityDelta/2))
	a.core.setChannelExpression(3, SectionCC(104, lead.ExpressionDelta+densityDelta))
	a.core.setChannelExpression(4, SectionCC(98, texture.ExpressionDelta+densityDelta/2))
}

func (a *SF2Pentatonic) SectionGain() float64 {
	return SectionMixProfileFor(a.section).Gain
}

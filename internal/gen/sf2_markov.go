package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

var _ Algorithm = (*SF2Markov)(nil)
var _ SF2Reverberator = (*SF2Markov)(nil)

// SF2Markov is the classical algorithm — Classical period (Haydn/Mozart)
// chamber-orchestra texture. Major fixes after research surfaced that the
// previous version sounded Baroque rather than Classical:
//
//   - REMOVED the harpsichord. Harpsichord continuo was obsolete by ~1780;
//     Mozart's mature symphonies do not use it. Its presence was the main
//     reason listeners said "this doesn't sound classical."
//   - Added Alberti bass — broken-chord eighth-note figuration (low-high-
//     mid-high) in the inner voice. This is the single most recognizable
//     Classical-period texture (Mozart K.545 opens with it). Replaces the
//     harpsichord stab pattern.
//   - 8-bar period structure with antecedent (ends on V, half cadence) +
//     consequent (ends on I, authentic cadence). Both halves use diatonic
//     I-IV-V-vi chords only — no modal flavoring (chromaticism is rare in
//     Classical-period writing, and what's there is for expression, not
//     color).
//   - Violin melody pre-computed with a Classical contour: triadic opening
//     (1-3-5), stepwise development, downward resolution to tonic.
//   - Oboe doubles the violin at unison in tuttis (Classical orchestration
//     convention) — toggleable.
//
// Preferred SF: timbres-of-heaven (best orchestral strings/brass for the
// Classical chamber texture).
type SF2Markov struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
	rng  *rand.Rand

	rootMidi  int
	keyOffset int

	progression []classicalChord
	barSamples  int64
	form        EpisodePlan
	section     FormSection

	samplesElapsed int64
	oboeOn         *bool

	violinPhrase []int
	profile      ControlProfile
}

// classicalChord is one bar of harmony — diatonic only.
type classicalChord struct {
	rootSemi int
	tones    []int // root, 3rd, 5th
	label    string
}

// classicalProgressions: 8-bar periods with antecedent (4 bars ending on V)
// + consequent (4 bars ending on I). All diatonic in major key. Drawn from
// the Mozart / Haydn idiom.
var classicalProgressions = [][]classicalChord{
	// Mozart K.545 opening period (transposed to scale degrees):
	// Antecedent: I  V7  I  V    Consequent: I  V7  I  I
	{
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 7, tones: []int{0, 4, 7, 10}, label: "V7"},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V"}, // half cadence
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 7, tones: []int{0, 4, 7, 10}, label: "V7"},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"}, // authentic cadence
	},
	// Standard period: I IV I V / I IV V I
	{
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 5, tones: []int{0, 4, 7}, label: "IV"},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V"},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 5, tones: []int{0, 4, 7}, label: "IV"},
		{rootSemi: 7, tones: []int{0, 4, 7, 10}, label: "V7"},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
	},
	// Haydn "Surprise" style: I I V V / I IV V I
	{
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V"},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V"},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
		{rootSemi: 5, tones: []int{0, 4, 7}, label: "IV"},
		{rootSemi: 7, tones: []int{0, 4, 7, 10}, label: "V7"},
		{rootSemi: 0, tones: []int{0, 4, 7}, label: "I"},
	},
}

func NewSF2Markov(sf *meltysynth.SoundFont) *SF2Markov { return &SF2Markov{sf: sf} }

func (a *SF2Markov) Name() string { return "classical" }

func (a *SF2Markov) ApplyControlProfile(profile ControlProfile) {
	a.profile = profileOrDefault(profile)
}

func (a *SF2Markov) currentRoot() int { return a.rootMidi + a.keyOffset }

func (a *SF2Markov) Seed(seedVal int64) {
	a.rng = rand.New(rand.NewSource(seedVal)) //nolint:gosec
	// Classical period favors C, G, D, F, Bb, Eb major (few accidentals
	// because of natural-horn limitations). Pick from these specifically.
	hornKeys := []int{48, 50, 53, 55, 58, 60} // C3, D3, F3, G3, Bb3, C4
	a.rootMidi = hornKeys[a.rng.Intn(len(hornKeys))]
	a.keyOffset = 0
	a.progression = classicalProgressions[a.rng.Intn(len(classicalProgressions))]
	a.samplesElapsed = 0

	oboeStart := false
	a.oboeOn = &oboeStart

	core, err := newSF2Core(a.sf, 2.4, seedVal)
	if err != nil {
		a.core = nil
		return
	}

	// Channel layout (Classical-period chamber orchestra — NO HARPSICHORD):
	//   0 — Violin             (program 40)  1st violins, melody
	//   1 — Cello              (program 42)  bass line + basses (octave doubled)
	//   2 — String Ensemble 1  (program 48)  2nd violins / violas, Alberti bass
	//   3 — String Ensemble 2  (program 49)  sustained pad (full ensemble bed)
	//   4 — Oboe               (program 68)  doubles violin in tuttis
	core.setProgram(0, 40)
	core.setProgram(1, 42)
	core.setProgram(2, 48)
	core.setProgram(3, 49)
	core.setProgram(4, 68)
	core.setPan(0, 84) // 1st violins right (concert seating)
	core.setPan(1, 44) // cellos left
	core.setPan(2, 56) // 2nd violins center-left
	core.setPan(3, 64) // ensemble center
	core.setPan(4, 76) // oboe right of center

	// All voices reasonably bright — Classical music is meant to read
	// clearly, not be obscured.
	core.setChannelCutoff(0, 110)
	core.setChannelCutoff(1, 90)
	core.setChannelCutoff(2, 88)
	core.setChannelCutoff(3, 80)
	core.setChannelCutoff(4, 100)

	// Concert-hall reverb. Solo voices (violin, oboe) most wet.
	core.setReverbSend(0, 84)
	core.setReverbSend(1, 56)
	core.setReverbSend(2, 64)
	core.setReverbSend(3, 100)
	core.setReverbSend(4, 88)

	// Tempo: allegro moderato 96–132 BPM (Mozart-symphony range).
	bpm := 96.0 + 36.0*a.rng.Float64()
	beatSec := 60.0 / bpm
	barSec := beatSec * 4
	a.barSamples = secondsToSamples(barSec)
	a.form = NewEpisodePlan(a.rng, a.barSamples, "classical")
	a.section = a.form.SectionAt(0)
	numBars := len(a.progression)
	cycleSec := barSec * float64(numBars)

	// --- Cello bass line: 2 hits per bar (beats 1 and 3), root and 5th —
	// the standard Classical bass figuration. The bass moves diatonically
	// step-wise toward chord changes; doubled at octave below by basses.
	cellaNotes := make([]int, 2*numBars)
	for i := range cellaNotes {
		cellaNotes[i] = a.celloLine(i)
	}
	core.addTrack(SF2Track{
		Channel: 1, Velocity: 76, Notes: cellaNotes,
		PeriodSec: cycleSec, Phase01: 0,
		MutationRate:   0.05, // very low — bass lines stay stable in Classical
		MutateOne:      func(slot int, _ int) int { return a.celloLine(slot) },
		Gate:           0.86,
		Legato:         true,
		VelocityJitter: 6, TimingJitterSec: 0.008,
	})

	// --- Alberti bass: the heart of Classical-period texture.
	// 8 eighth notes per bar in low-high-mid-high pattern across the chord.
	// Played by 2nd violins/violas (channel 2 = String Ensemble 1).
	// For C major chord (root C, 3rd E, 5th G): C-G-E-G C-G-E-G per bar.
	albertiNotes := make([]int, 8*numBars)
	for i := range albertiNotes {
		albertiNotes[i] = a.albertiNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 56, Notes: albertiNotes,
		PeriodSec: cycleSec, Phase01: 0,
		MutationRate:   0.02, // Alberti is fixed pattern — almost no mutation
		MutateOne:      func(slot int, _ int) int { return a.albertiNoteAt(slot) },
		Gate:           0.52,
		VelocityJitter: 4, TimingJitterSec: 0.006,
	})

	// --- Violin melody: pre-computed 16-note phrase across the 8-bar
	// period (2 notes/bar). Uses a Classical-period contour:
	//   antecedent (bars 1-4): triadic-opening 1-3-5-3 then walks toward 2
	//                          (half cadence)
	//   consequent (bars 5-8): mirrors antecedent then descends 3-2-1
	//                          (authentic cadence resolving to tonic)
	a.violinPhrase = a.makeClassicalMelody(2 * numBars)
	core.addTrack(SF2Track{
		Channel: 0, Velocity: 88, Notes: a.violinPhrase,
		PeriodSec: cycleSec, Phase01: 0,
		MutationRate: 0.04, // melody is composed — almost no per-note mutation
		MutateOne: func(slot int, _ int) int {
			return a.violinPhrase[slot%len(a.violinPhrase)]
		},
		Gate:   0.94,
		Legato: true,
		ResolveExpression: func(slot int, key int) SF2ExpressionCurve {
			return SF2ExpressionCurve{Start: 84, Peak: 108, End: 92, PeakAt01: 0.38}
		},
		VelocityJitter: 10, TimingJitterSec: 0.014,
	})

	// --- Oboe doubles violin at unison (Classical tutti orchestration).
	// Toggleable — when off, just violin sings; when on, the tutti sound.
	core.addTrack(SF2Track{
		Channel: 4, Velocity: 52, Notes: a.violinPhrase,
		PeriodSec: cycleSec, Phase01: 0,
		MutationRate: 0.04,
		MutateOne: func(slot int, _ int) int {
			return a.violinPhrase[slot%len(a.violinPhrase)]
		},
		Gate:           0.92,
		Legato:         true,
		VelocityJitter: 8, TimingJitterSec: 0.018,
		Enabled: a.oboeOn,
	})

	// --- String pad: sustained chord, one note per bar per voice. 2 voices
	// for the upper triad (3rd + 5th); cello already covers root + octave.
	// This is the "string halo" supporting the texture.
	for voiceIdx := 1; voiceIdx <= 2; voiceIdx++ {
		voice := voiceIdx
		notes := a.buildPadLine(voice, numBars)
		core.addTrack(SF2Track{
			Channel: 3, Velocity: 36, Notes: notes,
			PeriodSec: cycleSec, Phase01: float64(voice) * 0.04,
			MutationRate:   0.06,
			MutateOne:      func(slot int, _ int) int { return a.padTone(slot, voice) },
			Gate:           0.98,
			Legato:         true,
			VelocityJitter: 4, TimingJitterSec: 0.015,
		})
	}

	a.core = core
	a.applyArrangement()
}

// celloLine returns the cello bass note for the i-th half-bar slot.
// Pattern: root on beat 1, 5th on beat 3 — standard Classical bass.
// Octave below the chord root.
func (a *SF2Markov) celloLine(slot int) int {
	totalSlots := 2 * len(a.progression)
	slot = ((slot % totalSlots) + totalSlots) % totalSlots
	bar := slot / 2
	half := slot % 2
	chord := a.progression[bar]
	tone := chord.tones[0] // root
	if half == 1 {
		tone = chord.tones[2] // 5th
	}
	key := a.currentRoot() + chord.rootSemi + tone - 12
	for key > 50 {
		key -= 12
	}
	for key < 28 {
		key += 12
	}
	return key
}

// albertiNoteAt returns one eighth-note of the Alberti bass figuration.
// Per bar: 8 eighth notes in the pattern low-high-mid-high (× 2). For a
// chord whose tones in 4-note order are [root, 3rd, 5th], the pattern is:
// root - 5th - 3rd - 5th - root - 5th - 3rd - 5th (Mozart K.545 figure).
// Placed in the tenor register (C3–C4).
func (a *SF2Markov) albertiNoteAt(slot int) int {
	totalSlots := 8 * len(a.progression)
	slot = ((slot % totalSlots) + totalSlots) % totalSlots
	bar := slot / 8
	eighth := slot % 8
	chord := a.progression[bar]
	// Pattern within a 4-eighth cycle: low(root), high(5th), mid(3rd), high(5th).
	// Cycle repeats twice per bar.
	patternIdx := eighth % 4
	var toneIdx int
	switch patternIdx {
	case 0:
		toneIdx = 0 // root (low)
	case 1:
		toneIdx = 2 // 5th (high)
	case 2:
		toneIdx = 1 // 3rd (mid)
	case 3:
		toneIdx = 2 // 5th (high)
	}
	key := a.currentRoot() + chord.rootSemi + chord.tones[toneIdx]
	// Place in tenor register (C3–C4, MIDI 48–60).
	for key < 48 {
		key += 12
	}
	for key > 64 {
		key -= 12
	}
	return key
}

// makeClassicalMelody builds the violin melody for the entire 8-bar period
// (numSlots = 2 * numBars). Uses Classical-period contour:
// triadic-opening + stepwise development + downward resolution to tonic.
// Stays in pure major scale (no modal flavoring).
func (a *SF2Markov) makeClassicalMelody(numSlots int) []int {
	// Build the melody as a 2-bar motive, a 2-bar answer, a varied restatement,
	// then a cadential tail. This reads more like sentence/period writing than
	// a single fixed 8-bar contour.
	motives := [][]int{
		{0, 2, 4, 2},
		{0, 2, 4, 1},
		{0, 1, 2, 4},
	}
	motive := append([]int(nil), motives[a.rng.Intn(len(motives))]...)
	answer := []int{motive[1], motive[2], motive[3], 1}
	variant := []int{motive[0], motive[1], motive[2], 3}
	cadence := []int{2, 1, 0, 0}
	contour := append(append(append(append([]int{}, motive...), answer...), variant...), cadence...)
	if len(contour) != numSlots {
		extended := make([]int, numSlots)
		for i := range extended {
			extended[i] = contour[i%len(contour)]
		}
		contour = extended
	}
	root := a.currentRoot() + 24
	majorScale := []int{0, 2, 4, 5, 7, 9, 11}
	notes := make([]int, len(contour))
	for i := range contour {
		target := clampMidiToRange(scaleNoteAt(contour, i, majorScale, root, 0, 0), 67, 88)
		chord := a.progression[(i/2)%len(a.progression)]
		chordRoot := a.currentRoot() + chord.rootSemi + 24
		if i%2 == 0 {
			notes[i] = nearestRelativeNote(target, chordRoot, chord.tones, 67, 88)
			continue
		}
		notes[i] = nearestRelativeNote(target, root, majorScale, 67, 88)
		if i > 0 && absInt(notes[i]-notes[i-1]) > 7 {
			up := nearestRelativeNote(notes[i-1]+2, root, majorScale, 67, 88)
			down := nearestRelativeNote(notes[i-1]-2, root, majorScale, 67, 88)
			if absInt(up-target) <= absInt(down-target) {
				notes[i] = up
			} else {
				notes[i] = down
			}
		}
	}
	return notes
}

func (a *SF2Markov) buildPadLine(voice, numBars int) []int {
	notes := make([]int, numBars)
	prev := 0
	for i := 0; i < numBars; i++ {
		chord := a.progression[i%len(a.progression)]
		target := a.currentRoot() + chord.rootSemi
		best := voiceLeadNearest(prev, target, []int{chord.tones[voice]}, 55, 76)
		notes[i] = best
		prev = best
	}
	return notes
}

// padTone returns one chord-tone (1 = 3rd, 2 = 5th) for the string-pad
// halo voice.
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
	prev := a.samplesElapsed
	a.core.renderInto(left, right)
	a.samplesElapsed += int64(len(left))
	if a.form.SectionBoundaryCrossed(prev, a.samplesElapsed) {
		a.applyArrangement()
	}
}

func (a *SF2Markov) applyArrangement() {
	a.section = a.form.SectionAt(a.samplesElapsed)
	profile := profileOrDefault(a.profile)
	if a.oboeOn != nil {
		*a.oboeOn = profile.Density > 0 && (a.section.TextureLevel > 1 || a.section.Kind == FormCadence) &&
			a.section.Kind != FormOutro
	}
	if a.core == nil {
		return
	}
	lead := SectionSceneFor(a.section, RoleLead)
	bass := SectionSceneFor(a.section, RoleBass)
	comp := SectionSceneFor(a.section, RoleComp)
	texture := SectionSceneFor(a.section, RoleTexture)
	reverbDelta := ReverbDelta(profile)
	brightDelta := BrightnessDelta(profile)
	densityDelta := int32(ProfileCentered(profile.Density) * 8)
	droneDelta := DroneDepthDelta(profile)
	a.core.setReverbSend(0, SectionCC(84, lead.ReverbDelta+reverbDelta))
	a.core.setReverbSend(1, SectionCC(56, bass.ReverbDelta+reverbDelta/3))
	a.core.setReverbSend(2, SectionCC(64, comp.ReverbDelta+reverbDelta/2))
	a.core.setReverbSend(3, SectionCC(100, texture.ReverbDelta+reverbDelta))
	a.core.setReverbSend(4, SectionCC(88, lead.ReverbDelta+reverbDelta))
	a.core.setChannelCutoff(0, SectionCC(110, lead.BrightnessDelta+brightDelta))
	a.core.setChannelCutoff(1, SectionCC(90, bass.BrightnessDelta+brightDelta/2))
	a.core.setChannelCutoff(2, SectionCC(88, comp.BrightnessDelta+brightDelta))
	a.core.setChannelCutoff(3, SectionCC(80, texture.BrightnessDelta+brightDelta))
	a.core.setChannelCutoff(4, SectionCC(100, lead.BrightnessDelta+brightDelta))
	a.core.setChannelExpression(0, SectionCC(112, lead.ExpressionDelta+densityDelta))
	a.core.setChannelExpression(1, SectionCC(104, bass.ExpressionDelta+droneDelta/2))
	a.core.setChannelExpression(2, SectionCC(100, comp.ExpressionDelta+densityDelta/2))
	a.core.setChannelExpression(3, SectionCC(102, texture.ExpressionDelta+densityDelta/2))
	a.core.setChannelExpression(4, SectionCC(106, lead.ExpressionDelta+densityDelta))
}

func (a *SF2Markov) SectionGain() float64 {
	return SectionMixProfileFor(a.section).Gain
}

func (a *SF2Markov) DebugStatus() DebugStatus {
	chord := ""
	if len(a.progression) > 0 {
		bar := sampleBarIndex(a.samplesElapsed, a.barSamples) % len(a.progression)
		chord = a.progression[bar].label
	}
	return DebugStatus{
		Chord:   chord,
		Section: string(a.section.Kind),
		Bar:     a.form.BarAt(a.samplesElapsed),
	}
}

func (a *SF2Markov) ListeningMarkers() []ListeningMarker {
	return a.form.ListeningMarkers(4)
}

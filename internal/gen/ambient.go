package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

var _ Algorithm = (*Ambient)(nil)
var _ SF2Reverberator = (*Ambient)(nil)

// Ambient is a "Music for Airports" style algorithm. The old version layered
// 5 voices on incommensurate loops but didn't have the canonical bell-voice
// motif or any harmonic skeleton — pads were just modal scale notes picked
// at random. This rewrite gives it:
//
//   - A 4-chord modal "slow harmony" cycle, one chord every 45–75 seconds,
//     so the texture changes color but never stops feeling still.
//   - Two sustained pad layers (strings + warm pad) playing chord tones in
//     long-period loops with slow filter-cutoff LFOs for breathing.
//   - A choir "aahs" layer on the upper chord tones — the signature Eno
//     "voice in the pad" sound.
//   - A sparse tubular-bell motif: pentatonic chord tones triggered every
//     8–14 seconds on long incommensurate loops, drenched in reverb.
//   - A celesta sparkle layer (optional via section toggle).
//   - A sub-bass pedal that follows the chord root.
//
// There is no rhythm. Reverb sends are very high on every voice except the
// bass, which stays present.
//
// Preferred SF: arachno (D-50/M1/MU/Fairlight blend — perfect for retro pads).
type Ambient struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
	rng  *rand.Rand

	rootMidi  int
	keyOffset int

	// Chord cycle (modal — usually i / III / VI / iv in minor, or I / IV / vi
	// / V in major). Slot index per chord rotates very slowly.
	chords          []ambientChord
	currentChordIdx int

	samplesElapsed  int64
	nextChordAt     int64
	nextSectionAt   int64
	nextEvolutionAt int64
	section         FormSection

	bellsOn   *bool
	celestaOn *bool

	// Contours are fixed per seed, but their concrete notes are resolved
	// against the current chord at fire time so the harmony actually moves.
	bellContour          []int
	bellStartDegree      int
	bellRegisterShift    int
	bellMotifs           MotifMemory
	celestaContour       []int
	celestaStartDegree   int
	celestaRegisterShift int
	celestaMotifs        MotifMemory
	profile              ControlProfile
}

// ambientChord is one harmonic center: a root offset from the key center,
// plus the chord tones (semitone offsets from THE chord root).
type ambientChord struct {
	rootSemi int
	tones    []int // 3 or 4 chord tones from chord root
	label    string
}

// Modal ambient cycles: each is 4 chords, drifting slowly through related
// modes. Chord changes happen every 45–75 s so the listener feels harmonic
// motion without "song" structure.
var ambientCycles = [][]ambientChord{
	// Aeolian (natural minor) drift: i — III — VI — iv
	{
		{rootSemi: 0, tones: []int{0, 3, 7}, label: "i"},
		{rootSemi: 3, tones: []int{0, 4, 7}, label: "III"},
		{rootSemi: 8, tones: []int{0, 4, 7}, label: "VI"},
		{rootSemi: 5, tones: []int{0, 3, 7}, label: "iv"},
	},
	// Ionian (major) drift: I — V — vi — IV (50s changes slowed to ambient)
	{
		{rootSemi: 0, tones: []int{0, 4, 7, 11}, label: "Imaj7"},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V"},
		{rootSemi: 9, tones: []int{0, 3, 7}, label: "vi"},
		{rootSemi: 5, tones: []int{0, 4, 7, 11}, label: "IVmaj7"},
	},
	// Dorian: i — IV — bIII — v (modal, no leading tone)
	{
		{rootSemi: 0, tones: []int{0, 3, 7}, label: "i_dor"},
		{rootSemi: 5, tones: []int{0, 4, 7}, label: "IV_dor"},
		{rootSemi: 3, tones: []int{0, 4, 7}, label: "bIII"},
		{rootSemi: 7, tones: []int{0, 3, 7}, label: "v"},
	},
	// Mixolydian: I — bVII — IV — I (very Eno-2/1)
	{
		{rootSemi: 0, tones: []int{0, 4, 7, 10}, label: "I7"},
		{rootSemi: 10, tones: []int{0, 4, 7}, label: "bVII"},
		{rootSemi: 5, tones: []int{0, 4, 7}, label: "IV"},
		{rootSemi: 0, tones: []int{0, 4, 7, 10}, label: "I7"},
	},
}

// NewAmbient constructs the algorithm. Caller must call Seed before Next.
func NewAmbient(sf *meltysynth.SoundFont) *Ambient { return &Ambient{sf: sf} }

func (a *Ambient) Name() string { return "ambient" }

func (a *Ambient) currentRoot() int { return a.rootMidi + a.keyOffset }

func (a *Ambient) ApplyControlProfile(profile ControlProfile) { a.profile = profileOrDefault(profile) }

func (a *Ambient) phraseScale() float64 { return PhraseScale(profileOrDefault(a.profile)) }

func (a *Ambient) Seed(seedVal int64) {
	a.rng = rand.New(rand.NewSource(seedVal)) //nolint:gosec
	// Root in the bass register so chord tones can stack upward through the
	// pad/choir/bell registers.
	a.rootMidi = 36 + a.rng.Intn(7) // C2..F#2
	a.keyOffset = 0
	a.samplesElapsed = 0
	a.currentChordIdx = 0
	a.chords = ambientCycles[a.rng.Intn(len(ambientCycles))]
	a.bellContour = variedContour(a.rng, 6, 8)
	a.bellStartDegree = a.rng.Intn(3)
	a.bellRegisterShift = variedRegisterShift(a.rng)
	a.celestaContour = variedContour(a.rng, 4, 6)
	a.celestaStartDegree = 1 + a.rng.Intn(2)
	a.celestaRegisterShift = variedRegisterShift(a.rng)
	a.bellMotifs = a.makeBellMotifs()
	a.celestaMotifs = a.makeCelestaMotifs()
	a.scheduleNextChord()
	phraseScale := a.phraseScale()

	bellsStart := true
	celestaStart := true
	a.bellsOn = &bellsStart
	a.celestaOn = &celestaStart
	a.scheduleNextSection()
	a.scheduleNextEvolution()
	a.syncSection()

	core, err := newSF2Core(a.sf, 3.2, seedVal)
	if err != nil {
		a.core = nil
		return
	}

	// Channel layout — all sustained, sparse trigger rate. Mapping:
	//   0 — String Ensemble 1   (program 48)  pad bed
	//   1 — Warm Pad            (program 89)  pad bed (parallel)
	//   2 — Choir Aahs          (program 52)  vocal pad
	//   3 — Tubular Bells       (program 14)  bell motif (sparse)
	//   4 — Celesta             (program 8)   high sparkle (sparse)
	//   5 — Synth Bass 1        (program 38)  sub-bass pedal
	core.setProgram(0, 48)
	core.setProgram(1, 89)
	core.setProgram(2, 52)
	core.setProgram(3, 14)
	core.setProgram(4, 8)
	core.setProgram(5, 38)
	core.setPan(0, 40)
	core.setPan(1, 88)
	core.setPan(2, 64)
	core.setPan(3, 92)
	core.setPan(4, 36)
	core.setPan(5, 64)

	// Brightness: pads darkened (the "warm wash"), choir mid, bells & celesta
	// full bright for the halo effect.
	core.setChannelCutoff(0, 72)
	core.setChannelCutoff(1, 64)
	core.setChannelCutoff(2, 80)
	core.setChannelCutoff(3, 110)
	core.setChannelCutoff(4, 110)
	core.setChannelCutoff(5, 56)

	// Filter LFOs on the pads — gives the texture a slow "breathing" quality.
	// Different rates so the two pads pulse out of sync.
	core.addFilterLFO(0, 1.0/14.0, 70, 22)
	core.addFilterLFO(1, 1.0/19.0, 60, 26)
	core.addFilterLFO(2, 1.0/11.0, 78, 18) // choir slightly faster

	// Reverb sends: everyone wet except the bass.
	core.setReverbSend(0, 100)
	core.setReverbSend(1, 96)
	core.setReverbSend(2, 110)
	core.setReverbSend(3, 120) // bells in halo
	core.setReverbSend(4, 120) // celesta in halo
	core.setReverbSend(5, 30)
	core.setChorusSend(1, 48)
	core.setChorusSend(2, 32)

	// --- String pad: 2-voice spread. Periods chosen from Eno's documented
	// Music for Airports loop lengths (19.8s, 25.7s) — coprime in seconds
	// so they never realign within the listener's attention window.
	for ti, period := range []float64{19.8 * phraseScale, 25.7 * phraseScale} {
		voice := ti
		core.addTrack(SF2Track{
			Channel: 0, Velocity: 56, Notes: []int{a.padNote(voice, 0)},
			PeriodSec: period, Phase01: a.rng.Float64(),
			MutationRate:       0.30,
			MutateOne:          func(_ int, _ int) int { return a.padNote(voice, 0) },
			ResolveNote:        func(_ int, _ int) int { return a.padNote(voice, 0) },
			ResolveModWheel:    func(slot int, key int) SF2ExpressionCurve { return gentleVibratoCurve(0, 16, 8) },
			ResolveBrightness:  func(slot int, key int) SF2ExpressionCurve { return brightnessBloomCurve(70, 80, 72) },
			ResolveDetuneCents: slotDetunePattern(-4, 3, -2, 5),
			Gate:               0.98,
			Legato:             true,
			VelocityJitter:     8, TimingJitterSec: 0.05,
		})
	}
	// Warm pad layered in parallel — also documented Eno periods.
	for ti, period := range []float64{23.6 * phraseScale, 31.0 * phraseScale} {
		voice := ti
		core.addTrack(SF2Track{
			Channel: 1, Velocity: 50, Notes: []int{a.padNote(voice, 0)},
			PeriodSec: period, Phase01: a.rng.Float64(),
			MutationRate:       0.30,
			MutateOne:          func(_ int, _ int) int { return a.padNote(voice, 0) },
			ResolveNote:        func(_ int, _ int) int { return a.padNote(voice, 0) },
			ResolveModWheel:    func(slot int, key int) SF2ExpressionCurve { return gentleVibratoCurve(0, 14, 7) },
			ResolveBrightness:  func(slot int, key int) SF2ExpressionCurve { return brightnessBloomCurve(62, 74, 64) },
			ResolveDetuneCents: slotDetunePattern(2, -3, 4, -2),
			Gate:               0.98,
			Legato:             true,
			VelocityJitter:     6, TimingJitterSec: 0.05,
		})
	}

	// --- Choir aahs: single upper-register voice on a 29.2s period (an
	// Eno-documented loop length).
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 48, Notes: []int{a.choirNote(0)},
		PeriodSec: 29.2 * phraseScale, Phase01: a.rng.Float64(),
		MutationRate:       0.35,
		MutateOne:          func(_ int, _ int) int { return a.choirNote(0) },
		ResolveNote:        func(_ int, _ int) int { return a.choirNote(0) },
		ResolveModWheel:    func(slot int, key int) SF2ExpressionCurve { return gentleVibratoCurve(0, 18, 10) },
		ResolveBrightness:  func(slot int, key int) SF2ExpressionCurve { return brightnessBloomCurve(78, 90, 80) },
		ResolveDetuneCents: slotDetunePattern(-2, 2, -1, 3),
		Gate:               0.98,
		Legato:             true,
		VelocityJitter:     8, TimingJitterSec: 0.06,
	})

	// --- Tubular bell motif: three canon-like loop voices sharing one contour
	// but offset against each other. The contour stays recognizable while the
	// actual notes are resolved against the current chord at fire time.
	bellPeriods := []float64{25.7 * phraseScale, 31.0 * phraseScale, 36.1 * phraseScale}
	for ti, period := range bellPeriods {
		voice := ti
		bellSlots := make([]int, len(a.bellContour))
		core.addTrack(SF2Track{
			Channel: 3, Velocity: 64, Notes: bellSlots,
			PeriodSec: period, Phase01: a.rng.Float64(),
			ResolveNote:        func(slot int, _ int) int { return a.bellNoteAt(voice, slot) },
			ResolveBrightness:  func(slot int, key int) SF2ExpressionCurve { return brightnessBloomCurve(108, 124, 110) },
			ResolveDetuneCents: slotDetunePattern(0, 2, -1, 3),
			Gate:               0.72,
			VelocityJitter:     14, TimingJitterSec: 0.30, // tape-loop slippage feel
			Enabled: a.bellsOn,
		})
	}

	// --- Celesta sparkle: a shorter answering phrase high above the bells.
	celestaSlots := make([]int, len(a.celestaContour))
	core.addTrack(SF2Track{
		Channel: 4, Velocity: 44, Notes: celestaSlots,
		PeriodSec: 53.7 * phraseScale, Phase01: a.rng.Float64(),
		ResolveNote:        func(slot int, _ int) int { return a.celestaNote(slot) },
		ResolveBrightness:  func(slot int, key int) SF2ExpressionCurve { return brightnessBloomCurve(104, 122, 108) },
		ResolveDetuneCents: slotDetunePattern(1, -2, 0, 2),
		Gate:               0.58,
		VelocityJitter:     12, TimingJitterSec: 0.10,
		Enabled: a.celestaOn,
	})

	// --- Sub-bass pedal: very slow rate, holds the chord root in the very
	// bottom of the register so the texture has a foundation.
	core.addTrack(SF2Track{
		Channel: 5, Velocity: 60, Notes: []int{a.bassRoot()},
		PeriodSec: 41.7 * phraseScale, Phase01: 0,
		MutationRate:   0.50,
		MutateOne:      func(_ int, _ int) int { return a.bassRoot() },
		ResolveNote:    func(_ int, _ int) int { return a.bassRoot() },
		Gate:           0.96,
		Legato:         true,
		VelocityJitter: 6, TimingJitterSec: 0.03,
	})

	a.core = core
	a.applyArrangement()
}

// padNote returns one MIDI key for the pad/strings, picking chord tone
// `voice` (0..2) from the current chord, in the pad register (around C4–B5).
func (a *Ambient) padNote(voice, _ int) int {
	if len(a.chords) == 0 {
		return 60
	}
	c := a.chords[a.currentChordIdx]
	idx := voice % len(c.tones)
	key := a.currentRoot() + c.rootSemi + c.tones[idx]
	// Spread voices across octaves so a 3-tone chord covers the pad register.
	octaveBump := 24 + 12*voice
	key += octaveBump
	for key < 60 {
		key += 12
	}
	for key > 84 {
		key -= 12
	}
	return key
}

// choirNote returns an upper chord-tone key in the soprano range (C5–C6).
func (a *Ambient) choirNote(voice int) int {
	if len(a.chords) == 0 {
		return 72
	}
	c := a.chords[a.currentChordIdx]
	// Prefer 3rd and 5th over root for the choir (more "vocal").
	pick := []int{1, 2, 1, 2, 0}
	idx := pick[voice%len(pick)]
	if idx >= len(c.tones) {
		idx = len(c.tones) - 1
	}
	key := a.currentRoot() + c.rootSemi + c.tones[idx] + 36
	for key < 72 {
		key += 12
	}
	for key > 84 {
		key -= 12
	}
	return key
}

// makeBellPhrase builds the bell-motif's coherent 8-note melodic contour
// using the current chord's tones, transposed into the bell register
// (C5–C7). Picks a stylized contour (peak-and-fall by default for the Eno
// 2/1 vibe) and uses chord-tone scale degrees so every note resolves.
func (a *Ambient) bellNoteAt(voice, slot int) int {
	if len(a.chords) == 0 {
		return 72
	}
	c := a.chords[a.currentChordIdx]
	phrase := a.bellMotifs.PhraseFor(a.section.Kind)
	key := scaleNoteAt(phrase, slot+voice*2, c.tones, a.currentRoot()+c.rootSemi+36+a.bellRegisterShift, a.bellStartDegree, 0)
	return clampMidiToRange(key, 72, 96)
}

// celestaNote returns one slot of the answering celesta phrase (C6–C7).
func (a *Ambient) celestaNote(slot int) int {
	if len(a.chords) == 0 {
		return 84
	}
	phrase := a.celestaMotifs.PhraseFor(a.section.Kind)
	if len(phrase) == 0 {
		return -1
	}
	if a.bellsOn != nil && *a.bellsOn && a.section.Kind != FormB && slot%len(phrase) < len(phrase)/2 {
		return -1
	}
	c := a.chords[a.currentChordIdx]
	key := scaleNoteAt(phrase, slot, c.tones, a.currentRoot()+c.rootSemi+48+a.celestaRegisterShift, a.celestaStartDegree, 0)
	return clampMidiToRange(key, 84, 96)
}

// bassRoot returns the chord's root in the bass register (around C2).
func (a *Ambient) bassRoot() int {
	if len(a.chords) == 0 {
		return 36
	}
	c := a.chords[a.currentChordIdx]
	key := a.currentRoot() + c.rootSemi
	for key > 48 {
		key -= 12
	}
	for key < 24 {
		key += 12
	}
	return key
}

func (a *Ambient) scheduleNextChord() {
	// 45-75 seconds per chord — very slow harmonic motion.
	secs := (45.0 + 30.0*a.rng.Float64()) * a.phraseScale()
	a.nextChordAt = a.samplesElapsed + int64(secs*44100)
}

func (a *Ambient) scheduleNextSection() {
	// 2–4 min between section toggles (which ornaments are on).
	secs := (120.0 + 120.0*a.rng.Float64()) * a.phraseScale()
	a.nextSectionAt = a.samplesElapsed + int64(secs*44100)
}

func (a *Ambient) scheduleNextEvolution() {
	secs := (240.0 + 180.0*a.rng.Float64()) * a.phraseScale()
	a.nextEvolutionAt = a.samplesElapsed + int64(secs*44100)
}

func (a *Ambient) evolveTexture() {
	a.chords = ambientCycles[a.rng.Intn(len(ambientCycles))]
	if len(a.chords) > 0 {
		a.currentChordIdx %= len(a.chords)
	}
	a.bellContour = variedContour(a.rng, 5, 8)
	a.celestaContour = variedContour(a.rng, 4, 6)
	a.bellStartDegree = a.rng.Intn(3)
	a.celestaStartDegree = a.rng.Intn(3)
	a.bellRegisterShift = variedRegisterShift(a.rng)
	a.celestaRegisterShift = variedRegisterShift(a.rng)
	if len(a.bellMotifs.A) > 0 && a.rng.Float64() < 0.55 {
		a.bellMotifs = transformNumericMotifMemory(a.rng, a.bellMotifs)
	} else {
		a.bellMotifs = a.makeBellMotifs()
	}
	if len(a.celestaMotifs.A) > 0 && a.rng.Float64() < 0.55 {
		a.celestaMotifs = transformNumericMotifMemory(a.rng, a.celestaMotifs)
	} else {
		a.celestaMotifs = a.makeCelestaMotifs()
	}
	if a.bellsOn != nil && a.celestaOn != nil {
		if a.rng.Float64() < 0.5 {
			*a.bellsOn = true
			*a.celestaOn = false
		} else {
			*a.bellsOn = false
			*a.celestaOn = true
		}
	}
	a.scheduleNextEvolution()
}

func (a *Ambient) advance() {
	chordAdvanced := false
	if a.samplesElapsed >= a.nextChordAt {
		a.currentChordIdx = (a.currentChordIdx + 1) % len(a.chords)
		a.scheduleNextChord()
		chordAdvanced = true
	}
	if chordAdvanced && a.samplesElapsed >= a.nextEvolutionAt {
		a.evolveTexture()
	}
	if chordAdvanced && a.samplesElapsed >= a.nextSectionAt {
		// Toggle one of the ornament layers.
		if a.rng.Float64() < 0.5 {
			*a.bellsOn = !*a.bellsOn
		} else {
			*a.celestaOn = !*a.celestaOn
		}
		a.scheduleNextSection()
	}
	if chordAdvanced {
		a.syncSection()
	}
}

func (a *Ambient) DebugStatus() DebugStatus {
	chord := ""
	if len(a.chords) > 0 {
		chord = a.chords[a.currentChordIdx].label
	}
	return DebugStatus{
		Chord:   chord,
		Section: string(a.section.Kind),
		Bar:     a.currentChordIdx + 1,
	}
}

// SetReverbIR installs a convolution reverb on the master bus.
func (a *Ambient) SetReverbIR(ir []float64, wet float64) {
	if a.core != nil {
		a.core.setConvolutionIR(ir, wet)
	}
}

func (a *Ambient) Next(left, right []float64) {
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

func (a *Ambient) makeBellMotifs() MotifMemory {
	base := trimOrRepeatPhrase(a.bellContour, 8, 0)
	return MotifMemory{
		A:       base,
		Aprime:  rotatePhrase(sequencePhrase(base, map[int]int{6: 4, 4: 2, -2: 0}), 1),
		B:       reversePhrase(base),
		Cadence: []int{0, 2, 0, -2},
		Outro:   []int{0, -2, 0, -2},
	}
}

func (a *Ambient) makeCelestaMotifs() MotifMemory {
	base := trimOrRepeatPhrase(a.celestaContour, 6, 0)
	answer := stitchPhrase([]int{-2, 0}, rotatePhrase(base, 1))
	return MotifMemory{
		A:       base,
		Aprime:  sequencePhrase(base, map[int]int{4: 2, 2: 0}),
		B:       trimOrRepeatPhrase(answer, len(base), 0),
		Cadence: []int{0, 2, 0, 2},
		Outro:   []int{0, -2},
	}
}

func (a *Ambient) syncSection() {
	cadence := len(a.chords) > 0 && a.currentChordIdx == len(a.chords)-1
	bells := a.bellsOn != nil && *a.bellsOn
	celesta := a.celestaOn != nil && *a.celestaOn
	a.section = textureSectionForLayers(bells, celesta, cadence)
	profile := profileOrDefault(a.profile)
	if profile.Density <= 1 && a.celestaOn != nil {
		*a.celestaOn = false
	}
	if profile.Density == 0 && a.bellsOn != nil && a.section.Kind != FormCadence {
		*a.bellsOn = false
	}
	if a.samplesElapsed == 0 {
		a.section = sectionTemplate(FormIntro)
	}
	a.applyArrangement()
}

func (a *Ambient) applyArrangement() {
	if a.core == nil {
		return
	}
	profile := profileOrDefault(a.profile)
	texture := SectionSceneFor(a.section, RoleTexture)
	lead := SectionSceneFor(a.section, RoleLead)
	bass := SectionSceneFor(a.section, RoleBass)
	reverbDelta := ReverbDelta(profile)
	brightDelta := BrightnessDelta(profile)
	densityDelta := int32(ProfileCentered(profile.Density) * 8)
	droneDelta := DroneDepthDelta(profile)
	a.core.setReverbSend(0, SectionCC(100, texture.ReverbDelta+reverbDelta))
	a.core.setReverbSend(1, SectionCC(96, texture.ReverbDelta+reverbDelta))
	a.core.setReverbSend(2, SectionCC(110, texture.ReverbDelta+reverbDelta))
	a.core.setReverbSend(3, SectionCC(120, lead.ReverbDelta+reverbDelta))
	a.core.setReverbSend(4, SectionCC(120, lead.ReverbDelta+reverbDelta))
	a.core.setReverbSend(5, SectionCC(30, bass.ReverbDelta+reverbDelta/3))
	a.core.setChannelCutoff(0, SectionCC(72, texture.BrightnessDelta+brightDelta))
	a.core.setChannelCutoff(1, SectionCC(64, texture.BrightnessDelta+brightDelta))
	a.core.setChannelCutoff(2, SectionCC(80, texture.BrightnessDelta+brightDelta))
	a.core.setChannelCutoff(3, SectionCC(110, lead.BrightnessDelta+brightDelta))
	a.core.setChannelCutoff(4, SectionCC(110, lead.BrightnessDelta+brightDelta))
	a.core.setChannelCutoff(5, SectionCC(56, bass.BrightnessDelta+brightDelta/2))
	a.core.setChannelExpression(0, SectionCC(98, texture.ExpressionDelta+densityDelta/2))
	a.core.setChannelExpression(1, SectionCC(96, texture.ExpressionDelta+densityDelta/2))
	a.core.setChannelExpression(2, SectionCC(100, texture.ExpressionDelta+densityDelta/2))
	a.core.setChannelExpression(3, SectionCC(108, lead.ExpressionDelta+densityDelta))
	a.core.setChannelExpression(4, SectionCC(104, lead.ExpressionDelta+densityDelta))
	a.core.setChannelExpression(5, SectionCC(100, bass.ExpressionDelta+droneDelta))
}

func (a *Ambient) SectionGain() float64 {
	return SectionMixProfileFor(a.section).Gain
}

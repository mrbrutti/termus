package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

var _ Algorithm = (*Jazz)(nil)
var _ SF2Reverberator = (*Jazz)(nil)

// Jazz is a proper small-group swing algorithm. The previous "jazz" was
// actually slow modal ambient; this one has the things that make a listener
// recognize it as jazz on the first bar:
//
//   - Medium swing tempo (130–150 BPM) with triplet-feel 8ths
//   - 4/4 with a walking acoustic-bass line — root on beat 1, mostly chord
//     tones in between, chromatic-approach to next chord's root on beat 4
//   - Ride cymbal: quarters on every beat plus the swung "& of 2" and
//     "& of 4" — the classic ding-da-ding-ding-da-ding pattern, split here
//     across two tracks so each subset stays on its uniform grid
//   - Hi-hat chick on 2 and 4
//   - Brushed snare comping (occasional weak hits on 4)
//   - Piano comping in Charleston style — beat 1 stab + "& of 2" stab,
//     using shell voicings (root–3rd–7th) so the harmony reads instantly
//   - Sparse alto-sax melody on top, mostly chord tones with bebop scale
//     passing tones, played with timing jitter so it doesn't feel
//     programmed
//
// Form: an 8-bar progression made of two 4-bar ii-V-I-vi cycles. Per chord
// changes happen one per bar.
//
// SoundFont: prefers tyros4 (excellent jazz brass + walking bass + brushed
// kit). Falls back to whatever SF is loaded.
type Jazz struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
	rng  *rand.Rand

	rootMidi  int
	keyOffset int

	progression []jazzChord

	samplesElapsed int64
	nextDriftAt    int64

	saxOn         *bool
	nextSectionAt int64
}

// jazzChord is one bar of harmony. tones are MIDI semitone offsets from
// (rootMidi + keyOffset): root, 3rd, 5th, 7th in that order. rootSemi is the
// semitone offset of the chord's root from the key center — used by the
// walking-bass generator to plan chromatic approach to the next chord.
type jazzChord struct {
	rootSemi int
	tones    []int
	label    string
	// majMin: 0=major-quality (maj7 / dom7), 1=minor-quality (m7 / m7b5).
	// Used by the melody generator to pick the right bebop scale.
	majMin int
}

// jazzProgressions: 4-bar ii-V-I-vi cycles in semitone offsets from the
// tonic. Two cycles played back to back makes one 8-bar form. Hand-picked to
// cover the three most-played small-group changes.
var jazzProgressions = [][]jazzChord{
	// In C: |Dm7|G7|Cmaj7|Am7|
	{
		{rootSemi: 2, tones: []int{2, 5, 9, 12}, label: "Dm7", majMin: 1},
		{rootSemi: 7, tones: []int{7, 11, 14, 17}, label: "G7", majMin: 0},
		{rootSemi: 0, tones: []int{0, 4, 7, 11}, label: "Cmaj7", majMin: 0},
		{rootSemi: 9, tones: []int{9, 12, 16, 19}, label: "Am7", majMin: 1},
	},
	// "Autumn Leaves" feel: |Am7|D7|Gmaj7|Cmaj7|
	{
		{rootSemi: 9, tones: []int{9, 12, 16, 19}, label: "Am7", majMin: 1},
		{rootSemi: 2, tones: []int{2, 6, 9, 12}, label: "D7", majMin: 0},
		{rootSemi: 7, tones: []int{7, 11, 14, 18}, label: "Gmaj7", majMin: 0},
		{rootSemi: 0, tones: []int{0, 4, 7, 11}, label: "Cmaj7", majMin: 0},
	},
	// Minor blues: |Cm7|Fm7|Gm7|Cm7|
	{
		{rootSemi: 0, tones: []int{0, 3, 7, 10}, label: "Cm7", majMin: 1},
		{rootSemi: 5, tones: []int{5, 8, 12, 15}, label: "Fm7", majMin: 1},
		{rootSemi: 7, tones: []int{7, 10, 14, 17}, label: "Gm7", majMin: 1},
		{rootSemi: 0, tones: []int{0, 3, 7, 10}, label: "Cm7", majMin: 1},
	},
}

// Jazz drum-kit GM keys on channel 9.
const (
	jazzKickKey      = 36 // C2 — Bass Drum (light, used 2&4 in "feathered" jazz)
	jazzSnareBrushed = 38 // D2 — Snare (will be hit with brush velocity in jazz kit)
	jazzHiHatChick   = 44 // G#2 — Pedal Hi-Hat (the "chick" on 2 & 4)
	jazzRideBell     = 53 // F3 — Ride Bell (slightly brighter on quarters)
	jazzRideCymbal   = 51 // D#3 — Ride Cymbal 1
)

// NewJazz constructs a Jazz algorithm bound to the given SoundFont. Seed must
// be called before Next.
func NewJazz(sf *meltysynth.SoundFont) *Jazz { return &Jazz{sf: sf} }

func (a *Jazz) Name() string { return "jazz" }

func (a *Jazz) currentRoot() int { return a.rootMidi + a.keyOffset }

func (a *Jazz) Seed(seedVal int64) {
	a.rng = rand.New(rand.NewSource(seedVal)) //nolint:gosec
	// Bb / F / Eb / C are common horn keys — pick from this group.
	hornKeys := []int{46, 48, 51, 53} // Bb2, C3, Eb3, F3
	a.rootMidi = hornKeys[a.rng.Intn(len(hornKeys))]
	a.keyOffset = 0
	a.samplesElapsed = 0
	a.scheduleNextDrift()

	// Section flag — sax solo drops in/out every 30–60 seconds for verse/solo
	// feel. Start with sax on so the listener hears it within the first chorus.
	saxStart := true
	a.saxOn = &saxStart
	a.scheduleNextSection()

	core, err := newSF2Core(a.sf, 2.6, seedVal)
	if err != nil {
		a.core = nil
		return
	}

	// Channel layout:
	//   0 — Acoustic Grand Piano (program 0)        comping
	//   1 — Acoustic Bass        (program 32)       walking bass
	//   2 — Alto Sax             (program 65)       solo melody
	//   9 — Jazz Drum Kit        (bank 128, prog 32) ride/hihat/snare/kick
	core.setProgram(0, 0)
	core.setProgram(1, 32)
	core.setProgram(2, 65)
	core.setPan(0, 56) // piano slightly left
	core.setPan(1, 64) // bass center
	core.setPan(2, 72) // sax slightly right (classic stage placement)

	// Jazz drum kit on the standard drum channel.
	core.syn.ProcessMidiMessage(drumChannel, ccBankSelect, drumBankMSB, 0)
	const drumKitJazz = 32
	core.setProgram(drumChannel, drumKitJazz)
	core.setPan(drumChannel, 64)

	// Brighter character than lofi — jazz instruments are EQ'd to read
	// clearly; let the natural SF tone through.
	core.setChannelCutoff(0, 96) // piano — fairly bright (no muffled-tape feel)
	core.setChannelCutoff(1, 88) // bass — woody, mid-forward
	core.setChannelCutoff(2, 110) // sax — bright + present

	// Reverb sends — small-club reverb on everyone except the bass.
	core.setReverbSend(0, 48)
	core.setReverbSend(1, 18) // bass stays dry-ish to keep its definition
	core.setReverbSend(2, 86) // sax gets the most space (solo)
	core.setReverbSend(drumChannel, 42)
	// Light chorus on piano only — gives it a slight Bill-Evans shimmer.
	core.setChorusSend(0, 24)

	// Pick a progression.
	base := jazzProgressions[a.rng.Intn(len(jazzProgressions))]
	// Two repeats = 8-bar form.
	a.progression = make([]jazzChord, 0, 2*len(base))
	a.progression = append(a.progression, base...)
	a.progression = append(a.progression, base...)

	// Tempo: 120–148 BPM medium swing.
	bpm := 120.0 + 28.0*a.rng.Float64()
	beatSec := 60.0 / bpm
	barSec := beatSec * 4
	numBars := len(a.progression)
	cycleSec := barSec * float64(numBars)

	// Swing amount for triplet 8ths: 0.165 puts the off-beat 8th at ~0.665
	// of the beat (long-short feel). 0.0 = straight 8ths, 0.16 = jazz swing,
	// 0.18+ = heavily shuffled.
	const swingAmt = 0.165

	// --- Walking bass: 4 quarter notes per bar, hits every beat.
	bassNotes := make([]int, 4*numBars)
	for i := range bassNotes {
		bassNotes[i] = a.walkingBassAt(i)
	}
	core.addTrack(SF2Track{
		Channel: 1, Velocity: 92, Notes: bassNotes,
		PeriodSec: cycleSec, Phase01: 0,
		MutationRate: 0.35, // the line shifts a bit each cycle
		MutateOne:    func(slot int, _ int) int { return a.walkingBassAt(slot) },
		VelocityJitter: 8, TimingJitterSec: 0.006, // upright bassists are tight
	})

	// --- Piano comp on beat 1 — shell voicing, one hit per bar.
	pianoDownNotes := make([]int, numBars)
	for i := range pianoDownNotes {
		pianoDownNotes[i] = a.compShell(i, 3) // 3rd in upper voice for the down-beat
	}
	core.addTrack(SF2Track{
		Channel: 0, Velocity: 78, Notes: pianoDownNotes,
		PeriodSec: cycleSec, Phase01: 0,
		MutationRate: 0.20,
		MutateOne:    func(slot int, _ int) int { return a.compShell(slot, 3) },
		VelocityJitter: 10, TimingJitterSec: 0.014,
	})
	// The 7th of the shell voicing on a separate track so the chord rings
	// as a 2-note dyad — small group jazz comping rarely uses full triads.
	pianoDown7Notes := make([]int, numBars)
	for i := range pianoDown7Notes {
		pianoDown7Notes[i] = a.compShell(i, 7) // 7th in lower voice
	}
	core.addTrack(SF2Track{
		Channel: 0, Velocity: 76, Notes: pianoDown7Notes,
		PeriodSec: cycleSec, Phase01: 0,
		MutationRate: 0.20,
		MutateOne:    func(slot int, _ int) int { return a.compShell(slot, 7) },
		VelocityJitter: 10, TimingJitterSec: 0.014,
	})

	// --- Piano comp on "& of 2" — Charleston accent. Slightly softer hit;
	// gives the rhythm forward motion.
	// Position fraction within bar = 1.667/4 = 0.417 (swung 8th of beat 2).
	pianoAccentNotes := make([]int, numBars)
	for i := range pianoAccentNotes {
		pianoAccentNotes[i] = a.compShell(i, 5) // 5th voicing
	}
	core.addTrack(SF2Track{
		Channel: 0, Velocity: 64, Notes: pianoAccentNotes,
		PeriodSec: cycleSec,
		Phase01:   0.417 / float64(numBars),
		MutationRate: 0.20,
		MutateOne:    func(slot int, _ int) int { return a.compShell(slot, 5) },
		VelocityJitter: 12, TimingJitterSec: 0.020,
		FireProbability: 0.85, // not every bar — comp leaves space
	})

	// --- Ride cymbal: quarter notes (4 hits per bar). Bell on beat 1, plain
	// ride on 2/3/4 — gives the bell its emphasis.
	rideQuarterNotes := make([]int, 4*numBars)
	for i := range rideQuarterNotes {
		if i%4 == 0 {
			rideQuarterNotes[i] = jazzRideBell
		} else {
			rideQuarterNotes[i] = jazzRideCymbal
		}
	}
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 78, Notes: rideQuarterNotes,
		PeriodSec: cycleSec, Phase01: 0,
		VelocityJitter: 10, TimingJitterSec: 0.004,
	})
	// --- Ride: swung "& of 2" and "& of 4" — completes the jazz ride pattern.
	// 2 hits per bar at swung-8th positions 1.667/4 and 3.667/4.
	rideSwungNotes := make([]int, 2*numBars)
	for i := range rideSwungNotes {
		rideSwungNotes[i] = jazzRideCymbal
	}
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 62, Notes: rideSwungNotes,
		PeriodSec: cycleSec,
		Phase01:   0.417 / float64(numBars), // start at "& of 2" of bar 0
		VelocityJitter: 10, TimingJitterSec: 0.006,
		FireProbability: 0.92,
	})

	// --- Hi-hat chick on 2 and 4: 2 hits per bar at beats 1 and 3 + half-bar.
	// Actually beats 2 and 4 → bar fractions 1/4 and 3/4. 2 evenly-spaced
	// slots with phase 0.25/numBars.
	hatNotes := make([]int, 2*numBars)
	for i := range hatNotes {
		hatNotes[i] = jazzHiHatChick
	}
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 66, Notes: hatNotes,
		PeriodSec: cycleSec,
		Phase01:   0.25 / float64(numBars), // beats 2 & 4
		VelocityJitter: 8, TimingJitterSec: 0.004,
	})

	// --- Brushed snare backing: occasional weak hit on beat 4. Sparse —
	// fires only ~40% of bars.
	snareNotes := make([]int, numBars)
	for i := range snareNotes {
		snareNotes[i] = jazzSnareBrushed
	}
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 52, Notes: snareNotes,
		PeriodSec: cycleSec,
		Phase01:   0.75 / float64(numBars), // beat 4
		VelocityJitter: 12, TimingJitterSec: 0.010,
		FireProbability: 0.40, // sparse fills only
	})

	// --- Feathered kick on beats 1 and 3 — barely audible in modern jazz,
	// just enough to anchor the time. Velocity very low.
	kickNotes := make([]int, 2*numBars)
	for i := range kickNotes {
		kickNotes[i] = jazzKickKey
	}
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 38, Notes: kickNotes,
		PeriodSec: cycleSec, Phase01: 0,
		VelocityJitter: 6, TimingJitterSec: 0.004,
		FireProbability: 0.60, // sometimes drops out, like a real bassist
	})

	// --- Sax solo: 4 melodic notes spread across the 8-bar form, in the
	// alto's sweet register. With mutation each chorus shifts.
	saxNotes := make([]int, 4)
	for i := range saxNotes {
		saxNotes[i] = a.saxNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 70, Notes: saxNotes,
		PeriodSec: cycleSec,
		Phase01:   0.5 / float64(numBars), // enter on beat 3 of bar 0
		MutationRate: 0.5,
		MutateOne:    func(slot int, _ int) int { return a.saxNoteAt(slot) },
		VelocityJitter: 16, TimingJitterSec: 0.040, // sax is the most expressive — loose timing
		Enabled: a.saxOn,
		SwingAmount: swingAmt,
	})

	a.core = core
}

// walkingBassAt returns the MIDI key for the i-th beat of the cycle. Standard
// walking pattern: root-3-5-(approach), where the approach is a chromatic
// step (above or below) leading into the next chord's root.
func (a *Jazz) walkingBassAt(slot int) int {
	if len(a.progression) == 0 {
		return a.currentRoot()
	}
	totalBeats := 4 * len(a.progression)
	slot = ((slot % totalBeats) + totalBeats) % totalBeats
	bar := slot / 4
	beat := slot % 4
	chord := a.progression[bar]
	root := a.currentRoot() + chord.rootSemi
	// Drop the bass an octave below the chord root for upright-bass register.
	root -= 12

	switch beat {
	case 0:
		return root // root on beat 1
	case 1:
		// 3rd or 5th of the chord — pick deterministically from rng.
		offsets := []int{chord.tones[1] - chord.rootSemi, chord.tones[2] - chord.rootSemi}
		return root + offsets[a.rng.Intn(len(offsets))]
	case 2:
		// 5th or 7th.
		offsets := []int{chord.tones[2] - chord.rootSemi, chord.tones[3] - chord.rootSemi}
		return root + offsets[a.rng.Intn(len(offsets))]
	case 3:
		// Approach tone — chromatic semitone above or below the NEXT chord's root.
		nextBar := (bar + 1) % len(a.progression)
		nextRoot := a.currentRoot() + a.progression[nextBar].rootSemi - 12
		// 60% approach from below, 40% from above (below is slightly more common).
		if a.rng.Float64() < 0.60 {
			return nextRoot - 1
		}
		return nextRoot + 1
	}
	return root
}

// compShell returns the MIDI key for one voice of the piano shell voicing on
// the current bar. interval is 3 (third), 5 (fifth), or 7 (seventh) — chosen
// by the caller to differentiate per-voice tracks. Voicings are placed in the
// piano "middle voice" register (around middle C / C4 = 60).
func (a *Jazz) compShell(slot, interval int) int {
	if len(a.progression) == 0 {
		return 60
	}
	bar := ((slot % len(a.progression)) + len(a.progression)) % len(a.progression)
	chord := a.progression[bar]
	// Pick the chord tone by interval-number; chord.tones is [root,3rd,5th,7th].
	var idx int
	switch interval {
	case 3:
		idx = 1
	case 5:
		idx = 2
	case 7:
		idx = 3
	default:
		idx = 1
	}
	if idx >= len(chord.tones) {
		idx = len(chord.tones) - 1
	}
	// Comping voicings sit ~around C4–E4. Add octave bump if too low.
	key := a.currentRoot() + chord.tones[idx]
	for key < 60 {
		key += 12
	}
	for key > 76 {
		key -= 12
	}
	return key
}

// saxNoteAt returns one note for the sax-melody track. Picks a chord-tone or
// a bebop-scale neighbor in the alto's strong middle register (E4–C6).
func (a *Jazz) saxNoteAt(slot int) int {
	if len(a.progression) == 0 {
		return 72
	}
	// Map melodic slot to a chord — distribute 4 melody notes across the form.
	chordIdx := (slot * len(a.progression)) / 4
	if chordIdx >= len(a.progression) {
		chordIdx = len(a.progression) - 1
	}
	chord := a.progression[chordIdx]
	// 75% chord tone, 25% bebop-scale neighbor.
	root := a.currentRoot()
	var key int
	if a.rng.Float64() < 0.75 {
		// Chord tone, raised into the alto's sweet spot (E4–C6 = 64–84).
		toneIdx := a.rng.Intn(len(chord.tones))
		key = root + chord.tones[toneIdx]
	} else {
		// Bebop scale: major scale + b6 passing tone for major chords;
		// dorian for minor chords. Pick a random degree.
		var scale []int
		if chord.majMin == 0 {
			scale = []int{0, 2, 4, 5, 7, 8, 9, 11} // major + b6 passing
		} else {
			scale = []int{0, 2, 3, 5, 7, 9, 10} // dorian
		}
		deg := scale[a.rng.Intn(len(scale))]
		key = root + chord.rootSemi + deg
	}
	// Raise into alto register (sweet range E4–E6 → 64–88).
	for key < 64 {
		key += 12
	}
	for key > 86 {
		key -= 12
	}
	return key
}

// scheduleNextDrift picks when the next macro key-drift will fire.
// 4–8 minutes between drifts, deterministic from rng.
func (a *Jazz) scheduleNextDrift() {
	mins := 4.0 + 4.0*a.rng.Float64()
	a.nextDriftAt = a.samplesElapsed + int64(mins*60*44100)
}

// scheduleNextSection picks when the sax flips on/off next.
// 25–55 seconds — solos in small-group jazz are short, choruses get
// passed between players.
func (a *Jazz) scheduleNextSection() {
	secs := 25.0 + 30.0*a.rng.Float64()
	a.nextSectionAt = a.samplesElapsed + int64(secs*44100)
}

// applyMacroMutations is called once per render block. Currently handles
// section toggles and key drift.
func (a *Jazz) applyMacroMutations() {
	if a.samplesElapsed >= a.nextSectionAt {
		*a.saxOn = !*a.saxOn
		a.scheduleNextSection()
	}
	if a.samplesElapsed >= a.nextDriftAt {
		// ±1 or ±2 semitones to the key, occasionally.
		drift := []int{-2, -1, 1, 2}[a.rng.Intn(4)]
		a.keyOffset += drift
		// Keep key within a comfortable horn range.
		if a.keyOffset > 5 {
			a.keyOffset -= 12
		}
		if a.keyOffset < -5 {
			a.keyOffset += 12
		}
		a.scheduleNextDrift()
	}
}

// SetReverbIR installs a convolution reverb on the master bus (delegates to
// the shared engine).
func (a *Jazz) SetReverbIR(ir []float64, wet float64) {
	if a.core != nil {
		a.core.setConvolutionIR(ir, wet)
	}
}

func (a *Jazz) Next(left, right []float64) {
	if a.core == nil {
		for i := range left {
			left[i] = 0
			right[i] = 0
		}
		return
	}
	a.applyMacroMutations()
	a.core.renderInto(left, right)
	a.samplesElapsed += int64(len(left))
}

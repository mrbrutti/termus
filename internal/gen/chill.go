package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"

	"github.com/mrbrutti/termus/internal/synth"
)

var _ Algorithm = (*Chill)(nil)
var _ SF2Reverberator = (*Chill)(nil)

// Chill is a lofi-style algorithm with a real drum beat at its core — the
// element that makes lofi feel like lofi rather than ambient jazz. Layout:
//
//   ch 0 — Electric Piano 2 (Rhodes, chorused)  chord stabs (1 chord/bar)
//   ch 1 — Acoustic Bass                        root note on each downbeat
//   ch 2 — Vibraphone                           sparse melody (1 note/chord)
//   ch 9 — GM percussion                        kick (1 & 3), snare (2 & 4),
//                                                hi-hat (every 8th)
//
// Tempo: ~75 BPM, 4 beats per chord × 4 chords = 12.8 s per loop.
//
// The chord progression is one of five hand-picked turnarounds, mixing
// major-key (ii-V-I-VI, I-vi-IV-V) and minor-key (i-iv-VII-III, i-VI-III-VII)
// jazz/lofi progressions. The EP plays chord stabs (the Rhodes envelope
// decays naturally between hits, giving the classic lofi "wet stab" feel)
// using four chord-tone tracks summed into one channel.
//
// Tape character: a master-bus low-pass at 6.5 kHz "muffles" the high end
// (the canonical lofi sound), and a low-level white-noise hiss layer adds
// the "playing through a cassette" feel.
//
// For hours-long listening:
//   - per-track mutation: melody and (occasionally) bass re-roll within
//     the current chord's tones
//   - macro key-drift: every 4–7 minutes the key transposes ±1..2
//     semitones; chord-tone tracks have MutationRate 1.0 so they fully
//     re-roll in the new key on each cycle
type Chill struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
	rng  *rand.Rand

	rootMidi  int // base key tonic
	keyOffset int

	// Active progression — referenced by all mutator closures.
	progression []chillChord

	samplesElapsed int64
	nextDriftAt    int64
	nextSwapAt     int64
}

// chillChord is one chord in the loop, expressed as semitone offsets from
// the major-key tonic (rootMidi+keyOffset). For minor-key progressions the
// tonic is still treated as "key center" — the chord-tone offsets define
// the actual chord quality.
type chillChord struct {
	tones []int  // 4-note voicing: root, 3rd, 5th, 7th of the chord
	label string // human label, for debug/logging
}

// chillProgressions: each is a 4-chord turnaround. The first three are
// major-key jazz turnarounds; the last two are minor-key (very common in
// modern lofi).
var chillProgressions = [][]chillChord{
	// Major: ii-V-I-VI (classic jazz)
	{
		{tones: []int{2, 5, 9, 12}, label: "ii7"},
		{tones: []int{7, 11, 14, 17}, label: "V7"},
		{tones: []int{0, 4, 7, 11}, label: "Imaj7"},
		{tones: []int{9, 12, 16, 19}, label: "vi7"},
	},
	// Major: I-vi-IV-V (50s changes, lofi'd)
	{
		{tones: []int{0, 4, 7, 11}, label: "Imaj7"},
		{tones: []int{9, 12, 16, 19}, label: "vi7"},
		{tones: []int{5, 9, 12, 16}, label: "IVmaj7"},
		{tones: []int{7, 11, 14, 17}, label: "V7"},
	},
	// Major: Imaj7-IVmaj7-iii7-vi7 (wistful)
	{
		{tones: []int{0, 4, 7, 11}, label: "Imaj7"},
		{tones: []int{5, 9, 12, 16}, label: "IVmaj7"},
		{tones: []int{4, 7, 11, 14}, label: "iii7"},
		{tones: []int{9, 12, 16, 19}, label: "vi7"},
	},
	// Minor: i7-iv7-VII7-IIImaj7 (classic minor blues turnaround)
	{
		{tones: []int{0, 3, 7, 10}, label: "i7"},
		{tones: []int{5, 8, 12, 15}, label: "iv7"},
		{tones: []int{10, 14, 17, 20}, label: "VII7"},
		{tones: []int{3, 7, 10, 14}, label: "IIImaj7"},
	},
	// Minor: i7-VI-VII-i7 (Andalusian-leaning lofi)
	{
		{tones: []int{0, 3, 7, 10}, label: "i7"},
		{tones: []int{8, 12, 15, 19}, label: "VImaj7"},
		{tones: []int{10, 14, 17, 21}, label: "VIImaj7"},
		{tones: []int{0, 3, 7, 10}, label: "i7"},
	},
}

// GM standard drum keys on channel 9 (channel 10 in 1-indexed MIDI).
const (
	drumKick      = 36 // C2  — Bass Drum 1
	drumSnare     = 38 // D2  — Acoustic Snare
	drumHiHatC    = 42 // F#2 — Closed Hi-Hat
	drumChannel   = 9
	drumBankMSB   = 128 // bank 128 = drum kit in standard MIDI
	ccBankSelect  = 0xB0
	ccBankNumber  = 0x00
	progStandardKit = 0
)

// NewChill constructs the algorithm. Caller must call Seed before Next.
func NewChill(sf *meltysynth.SoundFont) *Chill { return &Chill{sf: sf} }

func (a *Chill) Name() string { return "chill" }

func (a *Chill) currentRoot() int { return a.rootMidi + a.keyOffset }

func (a *Chill) Seed(seedVal int64) {
	a.rng = rand.New(rand.NewSource(seedVal)) //nolint:gosec
	a.rootMidi = 48 + a.rng.Intn(7) // C3..F#3
	a.keyOffset = 0
	a.samplesElapsed = 0
	a.scheduleNextDrift()
	a.scheduleNextSwap()

	core, err := newSF2Core(a.sf, 2.8, seedVal)
	if err != nil {
		a.core = nil
		return
	}

	// Melodic channels.
	core.setProgram(0, 5)  // Electric Piano 2 (chorused Rhodes)  center
	core.setProgram(1, 32) // Acoustic Bass                       center
	core.setProgram(2, 11) // Vibraphone                          right
	core.setProgram(3, 64) // Soprano Sax                         left (solo)
	core.setProgram(4, 24) // Nylon Guitar                        right (comp)
	core.setPan(0, 64)
	core.setPan(1, 64)
	core.setPan(2, 88)
	core.setPan(3, 40)
	core.setPan(4, 90)

	// Channel 9 = standard MIDI drum channel. Select bank 128 (drum kit) and
	// program 0 (standard kit). Most SF2 files including TimGM6mb honor
	// channel 9 as percussion automatically, but explicitly setting the
	// bank+program is the robust path.
	core.syn.ProcessMidiMessage(drumChannel, ccBankSelect, drumBankMSB, 0)
	core.setProgram(drumChannel, progStandardKit)
	core.setPan(drumChannel, 64)

	// Filter LFO on the Rhodes — classic lofi "wow" effect, like a slowly-
	// detuning tape head. Slow rate, modest depth so the brightness gently
	// rocks back and forth.
	core.addFilterLFO(0, 1.0/8.0, 60, 22)

	// Pick a progression.
	a.progression = chillProgressions[a.rng.Intn(len(chillProgressions))]

	// Tempo: ~75 BPM. 4 beats per chord × 4 chords = 16 beats per loop.
	const bpm = 75.0
	beatSec := 60.0 / bpm
	barSec := beatSec * 4
	cycleSec := barSec * float64(len(a.progression))
	numBars := len(a.progression)

	// --- EP chord stabs: two stabs per bar (beats 1 and 3), same chord both
	// times. Four tracks (one per chord tone), all on channel 0, each with
	// 2*numBars slots. Slot k plays chord (k/2). The Rhodes envelope decays
	// across each half-bar, giving the lofi "stab → tail → stab → tail" feel.
	for toneIdx := 0; toneIdx < 4; toneIdx++ {
		ti := toneIdx
		notes := make([]int, 2*numBars)
		for s := range notes {
			notes[s] = a.epChordToneAt(s, ti)
		}
		mutate := func(slot int, _ int) int { return a.epChordToneAt(slot, ti) }
		core.addTrack(SF2Track{
			Channel: 0, Velocity: 72, Notes: notes,
			PeriodSec: cycleSec, Phase01: 0,
			MutationRate: 1.0, MutateOne: mutate,
			VelocityJitter: 8, TimingJitterSec: 0.008, // EP stab — lazy but not sloppy
		})
	}

	// --- Walking bass: root on beat 1, 5th on beat 3 (half-note feel).
	bassNotes := make([]int, 2*numBars)
	for i := range bassNotes {
		bassNotes[i] = a.bassNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: 1, Velocity: 88, Notes: bassNotes,
		PeriodSec: cycleSec, Phase01: 0,
		MutationRate: 1.0,
		MutateOne:    func(slot int, _ int) int { return a.bassNoteAt(slot) },
		VelocityJitter: 6, TimingJitterSec: 0.005, // bass — tight
	})

	// --- Vibraphone melody: one note per chord, sparse and high-register.
	vibeNotes := make([]int, numBars)
	for i := range vibeNotes {
		vibeNotes[i] = a.vibeNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 68, Notes: vibeNotes,
		PeriodSec: cycleSec, Phase01: 0,
		MutationRate: 0.35,
		MutateOne:    func(slot int, _ int) int { return a.vibeNoteAt(slot) },
		VelocityJitter: 12, TimingJitterSec: 0.020, // vibe — laid back
	})

	// --- Nylon Guitar: comping with extended chord notes on beat 2-and (the
	// "and" of beat 2) of each bar. One hit per bar at offset 1.5 beats.
	guitarNotes := make([]int, numBars)
	for i := range guitarNotes {
		guitarNotes[i] = a.guitarNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: 4, Velocity: 50, Notes: guitarNotes,
		PeriodSec: cycleSec,
		Phase01:   1.5 / float64(4*numBars), // 1.5 beats into the first bar
		MutationRate: 1.0,
		MutateOne:    func(slot int, _ int) int { return a.guitarNoteAt(slot) },
		VelocityJitter: 10, TimingJitterSec: 0.025, // nylon comping — humans don't quantize
	})

	// --- Soprano Sax: very sparse solo. Only 2 notes per loop (one in the
	// middle, one near the end), high register, jazzy color tones. With
	// mutation it'll wander to different chord tones across cycles.
	saxNotes := make([]int, 2)
	for i := range saxNotes {
		saxNotes[i] = a.saxNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: 3, Velocity: 64, Notes: saxNotes,
		PeriodSec: cycleSec,
		Phase01:   0.4, // come in 40% through the cycle
		MutationRate: 0.4,
		MutateOne:    func(slot int, _ int) int { return a.saxNoteAt(slot) },
		VelocityJitter: 14, TimingJitterSec: 0.035, // sax solo — most expressive, most loose
	})

	// --- Drum beat: kick on 1 & 3, snare on 2 & 4, hi-hat every 8th note.
	// All on channel 9. Each drum hit is just a NoteOn of the appropriate
	// percussion key. NoteOff has no effect on GM drum kits — they're
	// one-shots — but the engine fires it anyway and it's harmless.
	kickNotes := make([]int, 2*numBars)
	for i := range kickNotes {
		kickNotes[i] = drumKick
	}
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 92, Notes: kickNotes,
		PeriodSec: cycleSec, Phase01: 0,
		VelocityJitter: 8, TimingJitterSec: 0.003, // kick — anchors the groove, must be tight
	})
	snareNotes := make([]int, 2*numBars)
	for i := range snareNotes {
		snareNotes[i] = drumSnare
	}
	// Snare offset by 1 beat = beat 2 of each bar (since the 2-per-bar slot
	// pattern lands on beats 1 & 3 by default; shifting by half a slot lands
	// it on beats 2 & 4).
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 82, Notes: snareNotes,
		PeriodSec: cycleSec, Phase01: 0.5 / float64(2*numBars),
		VelocityJitter: 6, TimingJitterSec: 0.004, // snare — tight, slight behind-the-beat
	})
	hihatNotes := make([]int, 8*numBars) // 8 hits per bar
	for i := range hihatNotes {
		hihatNotes[i] = drumHiHatC
	}
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 55, Notes: hihatNotes,
		PeriodSec: cycleSec, Phase01: 0,
		VelocityJitter:  14,    // hi-hat benefits most from "swing" velocity feel
		TimingJitterSec: 0.006, // and a bit of swing timing
	})

	// --- Tape character ---
	// Light master-bus low-pass at 6.5 kHz: rolls off the brightest harmonics
	// the way a cassette would. q=0.5 keeps the rolloff smooth (no resonant
	// peak).
	core.setMasterLowpass(6500, 0.5)
	// Subtle white-noise hiss at ~-50 dBFS. Just barely audible behind quiet
	// passages, masked by everything louder.
	core.setTapeHiss(0.003)

	// Soft small-room reverb by default.
	core.setConvolutionIR(synth.SyntheticRoomIR(0.12), 0.35)

	a.core = core
}

// epChordToneAt returns the MIDI note for one tone of the chord that should
// be played in the given EP slot. EP has 2 stabs per bar, so slot/2 indexes
// the progression.
func (a *Chill) epChordToneAt(slot, toneIdx int) int {
	chordIdx := (slot / 2) % len(a.progression)
	c := a.progression[chordIdx]
	return a.currentRoot() + c.tones[toneIdx] + 24
}

// bassNoteAt returns the bass note for half-note-feel beat `slot`. Pattern
// per chord: chord root (beat 1) → chord fifth (beat 3) → next chord →
// root → fifth → etc. Always in the low register one octave below the
// chord root.
func (a *Chill) bassNoteAt(slot int) int {
	chordIdx := (slot / 2) % len(a.progression)
	half := slot % 2 // 0 = beat 1, 1 = beat 3
	c := a.progression[chordIdx]
	tone := c.tones[0] // root
	if half == 1 {
		tone = c.tones[2] // 5th
	}
	return a.currentRoot() + tone - 12
}

// guitarNoteAt returns a single nylon-guitar comp note per bar. Plays a
// chord-tone in the +12-semitone register (between bass and EP) at the "and"
// of beat 2 — classic jazz/bossa comping placement.
func (a *Chill) guitarNoteAt(slot int) int {
	chordIdx := slot % len(a.progression)
	c := a.progression[chordIdx]
	// 60% root, 30% 5th, 10% 9th (chord extension)
	switch r := a.rng.Float64(); {
	case r < 0.10:
		return a.currentRoot() + c.tones[0] + 14 // 9th of chord root
	case r < 0.40:
		return a.currentRoot() + c.tones[2] + 12 // 5th
	default:
		return a.currentRoot() + c.tones[0] + 12 // root
	}
}

// saxNoteAt returns one soprano-sax note. The sax track has only 2 slots
// per cycle so the sax phrases are very sparse — soloistic, not constant
// melody. Plays jazzy intervals: chord tones, 9th, 11th, or 13th.
func (a *Chill) saxNoteAt(slot int) int {
	// Pick a chord from the current cycle proportional to slot position.
	chordIdx := (slot * len(a.progression) / 2) % len(a.progression)
	c := a.progression[chordIdx]
	chordRoot := a.currentRoot() + c.tones[0]
	switch r := a.rng.Float64(); {
	case r < 0.20:
		return chordRoot + 14 // 9th
	case r < 0.35:
		return chordRoot + 17 // 11th
	case r < 0.50:
		return chordRoot + 21 // 13th
	default:
		// Chord tone in high register
		return a.currentRoot() + c.tones[a.rng.Intn(len(c.tones))] + 36
	}
}

// vibeNoteAt returns one melody note per chord. 65% chord tone, 35% color
// tone (9th or 13th) for jazzy character.
func (a *Chill) vibeNoteAt(slot int) int {
	chordIdx := slot % len(a.progression)
	c := a.progression[chordIdx]
	chordRoot := a.currentRoot() + c.tones[0]
	switch r := a.rng.Float64(); {
	case r < 0.20:
		return chordRoot + 14 // 9th
	case r < 0.35:
		return chordRoot + 21 // 13th
	default:
		return a.currentRoot() + c.tones[a.rng.Intn(len(c.tones))] + 36
	}
}

func (a *Chill) scheduleNextDrift() {
	secs := 240.0 + 180.0*a.rng.Float64()
	a.nextDriftAt = a.samplesElapsed + int64(secs*float64(synth.SampleRate))
}

func (a *Chill) shiftKey() {
	shift := a.rng.Intn(5) - 2
	if shift == 0 {
		shift = 1
	}
	a.keyOffset += shift
	if a.keyOffset > 4 {
		a.keyOffset = 4 - a.rng.Intn(3)
	}
	if a.keyOffset < -4 {
		a.keyOffset = -4 + a.rng.Intn(3)
	}
}

// SetReverbIR installs a convolution reverb on the master bus. Chill auto-
// installs a small room by default; --ir overrides.
func (a *Chill) SetReverbIR(ir []float64, wet float64) {
	if a.core != nil {
		a.core.setConvolutionIR(ir, wet)
	}
}

func (a *Chill) Next(left, right []float64) {
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

// chillChannelAlternatives — staying inside the lofi soundscape. Drums (ch 9)
// are deliberately excluded; swapping the kit mid-track would feel jarring.
var chillChannelAlternatives = map[int32][]int32{
	0: {5, 4, 88, 89}, // EP2 (default), EP1, New Age Pad, Warm Pad
	1: {32, 33, 36, 38}, // Acoustic Bass (default), Electric Bass Finger, Slap Bass, Synth Bass 1
	2: {11, 9, 13},    // Vibraphone (default), Glockenspiel, Xylophone
	3: {64, 65, 66, 67}, // Soprano Sax (default), Alto Sax, Tenor Sax, Baritone Sax
	4: {24, 25, 26, 27}, // Nylon Guitar (default), Steel String, Jazz Guitar, Electric Clean
}

func (a *Chill) scheduleNextSwap() {
	secs := 180.0 + 120.0*a.rng.Float64() // 3–5 min — chill wants gentle variety
	a.nextSwapAt = a.samplesElapsed + int64(secs*44100)
}

func (a *Chill) swapOneInstrument() {
	channels := []int32{0, 1, 2, 3, 4}
	ch := channels[a.rng.Intn(len(channels))]
	a.core.programSwap(ch, chillChannelAlternatives[ch], a.rng)
}

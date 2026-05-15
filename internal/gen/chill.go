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
//	ch 0 — Electric Piano 2 (Rhodes, chorused)  chord stabs (1 chord/bar)
//	ch 1 — Acoustic Bass                        root note on each downbeat
//	ch 2 — Vibraphone                           sparse melody (1 note/chord)
//	ch 9 — GM percussion                        kick (1 & 3), snare (2 & 4),
//	                                             hi-hat (every 8th)
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

	barSamples int64
	form       FormPlan
	section    FormSection

	samplesElapsed int64
	nextDriftAt    int64

	// Section state: sax solo and nylon guitar comp drop in/out every 90–180s
	// to give the track verse/chorus/bridge dynamics over a long listen.
	saxOn    *bool
	guitarOn *bool
	vibeOn   *bool

	vibePlan     []int
	guitarPlan   []int
	saxPlan      []int
	vibeMotifs   MotifMemory
	guitarMotifs MotifMemory
	saxMotifs    MotifMemory
}

// chillChord is one chord in the loop, expressed as semitone offsets from
// the major-key tonic (rootMidi+keyOffset). For minor-key progressions the
// tonic is still treated as "key center" — the chord-tone offsets define
// the actual chord quality.
type chillChord struct {
	tones []int  // 4-note voicing: root, 3rd, 5th, 7th of the chord
	label string // human label, for debug/logging
}

func chillMaj7(rootSemi int, label string) chillChord {
	return chillChord{
		tones: []int{rootSemi, rootSemi + 4, rootSemi + 7, rootSemi + 11},
		label: label,
	}
}

func chillMin7(rootSemi int, label string) chillChord {
	return chillChord{
		tones: []int{rootSemi, rootSemi + 3, rootSemi + 7, rootSemi + 10},
		label: label,
	}
}

func chillDom7(rootSemi int, label string) chillChord {
	return chillChord{
		tones: []int{rootSemi, rootSemi + 4, rootSemi + 7, rootSemi + 10},
		label: label,
	}
}

// chillChordOption extends chillChord with a list of valid next-chord
// indices in the same palette. Pattern lifted from meel-hd/lofi-engine's
// state-machine chord grammar — each chord knows what can follow it,
// producing musically-coherent progressions of arbitrary length without
// us hand-curating every possibility. Beats hand-curated 4-chord loops
// for variety; loses a little of their tightness.
type chillChordOption struct {
	tones    []int
	label    string
	nextIdxs []int
}

// Major-key chord grammar. Indices match diatonic scale degrees so the
// nextIdxs read like Roman-numeral progression rules:
//
//	I    → ii iii IV V vi vii (root chord, can go anywhere)
//	ii   → iii V vii          (predominant)
//	iii  → IV vi              (median, weak)
//	IV   → ii V               (subdominant → predominant or dominant)
//	V    → I iii vi           (dominant → tonic or deceptive)
//	vi   → ii IV              (relative-minor area)
//	vii  → I iii              (leading-tone)
var chillMajorChordGrammar = []chillChordOption{
	{tones: []int{0, 4, 7, 11}, label: "Imaj7", nextIdxs: []int{1, 2, 3, 4, 5, 6}},
	{tones: []int{2, 5, 9, 12}, label: "ii7", nextIdxs: []int{2, 4, 6}},
	{tones: []int{4, 7, 11, 14}, label: "iii7", nextIdxs: []int{3, 5}},
	{tones: []int{5, 9, 12, 16}, label: "IVmaj7", nextIdxs: []int{1, 4}},
	{tones: []int{7, 11, 14, 17}, label: "V7", nextIdxs: []int{0, 2, 5}},
	{tones: []int{9, 12, 16, 19}, label: "vi7", nextIdxs: []int{1, 3}},
	{tones: []int{11, 14, 17, 21}, label: "vii_m7b5", nextIdxs: []int{0, 2}},
}

// Minor-key chord grammar:
//
//	i   → iv V VI VII III     (tonic, goes most places)
//	iv  → i V VII             (subdominant)
//	V   → i VI                (cadential dominant; borrowed from harmonic minor)
//	VI  → iv VII              (relative-major area)
//	VII → i III               (subtonic resolves to tonic or relative major)
//	III → iv VI               (relative major)
var chillMinorChordGrammar = []chillChordOption{
	{tones: []int{0, 3, 7, 10}, label: "i7", nextIdxs: []int{1, 2, 3, 4, 5}},
	{tones: []int{5, 8, 12, 15}, label: "iv7", nextIdxs: []int{0, 2, 4}},
	{tones: []int{7, 11, 14, 17}, label: "V7", nextIdxs: []int{0, 3}},
	{tones: []int{8, 12, 15, 19}, label: "VImaj7", nextIdxs: []int{1, 4}},
	{tones: []int{10, 14, 17, 20}, label: "VII7", nextIdxs: []int{0, 5}},
	{tones: []int{3, 7, 10, 14}, label: "IIImaj7", nextIdxs: []int{1, 3}},
}

// markovWalkChords generates `length` chords by walking the chord grammar.
// Starts on a stable degree (tonic 70% of the time, V 20%, vi 10%) and
// follows each chord's nextIdxs list to pick the next one.
func markovWalkChords(rng *rand.Rand, grammar []chillChordOption, length int) []chillChord {
	out := make([]chillChord, length)
	var idx int
	switch r := rng.Float64(); {
	case r < 0.70:
		idx = 0 // tonic
	case r < 0.90:
		// "V" / dominant chord — for both grammars, that's index 4 (V7) or
		// for the minor table also index 4 (VII7). Both are legitimate
		// "non-tonic" openings.
		idx = 4
	default:
		// "relative minor / VI" — index 5 in both tables
		idx = 5
	}
	if idx >= len(grammar) {
		idx = 0
	}
	out[0] = chillChord{tones: grammar[idx].tones, label: grammar[idx].label}
	for i := 1; i < length; i++ {
		nexts := grammar[idx].nextIdxs
		if len(nexts) == 0 {
			idx = 0
		} else {
			idx = nexts[rng.Intn(len(nexts))]
		}
		out[i] = chillChord{tones: grammar[idx].tones, label: grammar[idx].label}
	}
	return out
}

// chillProgressions: legacy hand-curated 4-chord turnarounds. Kept as a
// fallback for the 25% of seeds that go this route — tight Imaj7-VI-IV-V
// loops still have their charm vs. the 8-chord Markov walks below.
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
	drumKick        = 36 // C2  — Bass Drum 1
	drumSnare       = 38 // D2  — Acoustic Snare
	drumHiHatC      = 42 // F#2 — Closed Hi-Hat
	drumHiHatOpen   = 46 // A#2 — Open Hi-Hat
	drumCrash       = 49 // C#3 — Crash Cymbal 1
	drumChannel     = 9
	drumBankMSB     = 128 // bank 128 = drum kit in standard MIDI
	ccBankSelect    = 0xB0
	ccBankNumber    = 0x00
	progStandardKit = 0
)

const (
	chillPlanRest = iota
	chillPlanRoot
	chillPlanThird
	chillPlanFifth
	chillPlanSeventh
	chillPlanNinth
	chillPlanEleventh
	chillPlanThirteenth
	chillPlanPickupAbove
	chillPlanPickupBelow
	chillPlanSuspendFourth
	chillPlanResolveThird
)

// NewChill constructs the algorithm. Caller must call Seed before Next.
func NewChill(sf *meltysynth.SoundFont) *Chill { return &Chill{sf: sf} }

func (a *Chill) Name() string { return "chill" }

func (a *Chill) currentRoot() int { return a.rootMidi + a.keyOffset }

func (a *Chill) Seed(seedVal int64) {
	a.rng = rand.New(rand.NewSource(seedVal)) //nolint:gosec
	a.rootMidi = 48 + a.rng.Intn(7)           // C3..F#3
	a.keyOffset = 0
	a.samplesElapsed = 0
	// Section state. Start in "intro" — sax off, guitar on (just chords +
	// rhythm section). First section flip usually brings the sax in.
	saxStart := false
	guitarStart := true
	vibeStart := true
	a.saxOn = &saxStart
	a.guitarOn = &guitarStart
	a.vibeOn = &vibeStart

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

	// Channel 9 = standard MIDI drum channel. Bank 128 selects the drum
	// bank; the PROGRAM within that bank picks WHICH kit. GeneralUser-GS
	// has 13 different drum kits at GM standard slots. For chill, the
	// JAZZ KIT (program 32) is the right starting point — softer cymbals,
	// brushed snare, warm kick — exactly the "lofi study beat" sound.
	core.syn.ProcessMidiMessage(drumChannel, ccBankSelect, drumBankMSB, 0)
	const drumKitJazz = 32
	core.setProgram(drumChannel, drumKitJazz)
	core.setPan(drumChannel, 64)

	// Per-channel base cutoffs — the lofi-engine trick. Set each
	// melodic instrument's CC 74 to a low static value so SF2 voices with
	// filter mappings render dramatically darker (lofi-engine uses a
	// 1 kHz hardware lowpass on its piano channel; CC 74 ≈ 32 is the
	// MIDI equivalent in most SoundFonts including TimGM6mb).
	core.setChannelCutoff(0, 32) // Rhodes EP — very darkened
	core.setChannelCutoff(2, 56) // vibraphone — slight darkening only
	core.setChannelCutoff(3, 70) // sax solo — left bright so it cuts through
	core.setChannelCutoff(4, 42) // nylon guitar — moderately dark for comping
	// Bass and drums left at full brightness — bass IS the low end, drums
	// need transient definition.

	// Filter LFO on the Rhodes — classic lofi "wow" effect. Now centered
	// LOW (32 instead of 60) so the LFO modulates around the new darkened
	// base cutoff rather than re-brightening past it.
	core.addFilterLFO(0, 1.0/8.0, 32, 16)

	// Chill master EQ override: the engine default boosts highs by +3 dB at
	// 7.5 kHz, which fights against the tape lowpass and the "dark" lofi
	// aesthetic. For chill specifically, CUT highs slightly so the master
	// chain genuinely darkens the top end rather than re-adding what the
	// LP took away.
	core.setMasterEQ(180, 1.5, 7500, -4.0)
	// Also drop the master LP cutoff a bit (was 6500 Hz) — pulls the top
	// end down further for that fully-muffled tape feel.
	core.setMasterLowpass(5500, 0.55)

	// Lofi reverb is generally short and close (already configured via
	// SyntheticRoomIR), but per-channel sends shape the mix character.
	// Sax solo gets the most reverb for "soloistic space"; drums stay dry
	// to keep the beat punchy; bass dry to keep the low end tight.
	core.setReverbSend(0, 56)           // Rhodes: light room verb
	core.setReverbSend(1, 24)           // bass: dry
	core.setReverbSend(2, 80)           // vibraphone: wet, halo
	core.setReverbSend(3, 96)           // sax: most wet — soloistic space
	core.setReverbSend(4, 50)           // nylon guitar: moderate
	core.setReverbSend(drumChannel, 30) // drums: mostly dry, just a touch of room
	core.setChorusSend(0, 56)           // Rhodes loves chorus
	core.setChorusSend(2, 32)
	core.setChorusSend(4, 24)

	// Sidechain ducking — the kick triggers a -4 dB duck on the master bus
	// that recovers over 250 ms. This is the squelchy "pump" of modern lofi
	// where the bass and pad get briefly pulled down each time the kick
	// hits, making the kick feel huge without it being loud.
	core.configureSidechain(-4, 12, 240)

	// Tape saturation — gentler than before (0.20 vs 0.28) so it doesn't
	// generate as many harsh upper harmonics that the listener perceives
	// as "always-present sharpness."
	core.setTapeSaturation(0.20)

	// Vinyl crackle — much sparser than v1 (was 15 pops/sec at 0.045 amp;
	// now 6 pops/sec at 0.022 amp with longer pop duration). Real dusty
	// vinyl pops a few times per second, not constantly. The reduction
	// removes the "always there" hash that the prior amplitude+rate were
	// producing.
	core.setVinylCrackle(6, 0.022, 1.5)

	// Pick a progression. 75% of seeds get a Markov-walked 8-chord progression
	// (per chord grammar rules above); 25% get a hand-curated 4-chord
	// turnaround. The two modes have different feels — Markov walks tend
	// to wander more "compositionally" and are less predictable; the
	// hand-curated turnarounds are tight loops that feel like classic
	// lofi study-beat backings.
	if a.rng.Float64() < 0.75 {
		// 60% major-key, 40% minor for the Markov walks.
		grammar := chillMajorChordGrammar
		if a.rng.Float64() < 0.40 {
			grammar = chillMinorChordGrammar
		}
		a.progression = markovWalkChords(a.rng, grammar, 8)
	} else {
		a.progression = chillProgressions[a.rng.Intn(len(chillProgressions))]
	}
	a.progression = a.reharmonizeProgression(a.progression)
	numBars := len(a.progression)
	a.vibeMotifs = a.makeVibeMotifs()
	a.guitarMotifs = a.makeGuitarMotifs()
	a.saxMotifs = a.makeSaxMotifs()
	a.vibePlan = trimOrRepeatPhrase(a.vibeMotifs.A, numBars, chillPlanThird)
	a.guitarPlan = trimOrRepeatPhrase(a.guitarMotifs.A, numBars, chillPlanNinth)
	a.saxPlan = trimOrRepeatPhrase(a.saxMotifs.A, numBars, chillPlanRest)

	// Tempo: 65 BPM ± 4 (61–69 BPM range, seed-driven). Per research, lofi
	// sits at 65–95 BPM and the sweet spot for "doesn't tire the listener
	// over hours" is the lower half of that range. We were at 75 — drop to
	// 65 nominal for a noticeably slower, deeper feel.
	bpm := 61.0 + 8.0*a.rng.Float64()
	beatSec := 60.0 / bpm
	barSec := beatSec * 4
	a.barSamples = secondsToSamples(barSec)
	a.form = NewFormPlan(a.rng, a.barSamples, "lofi")
	a.section = a.form.SectionAt(0)
	a.scheduleNextDrift()
	cycleSec := barSec * float64(len(a.progression))
	a.applyArrangement()

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
			Gate:                   0.52,
			ResolveTimingOffsetSec: cyclicTimingOffset(0, 12),
			ResolveVelocity: func(slot int, key int, base int32) int32 {
				if slot%2 == 0 {
					return base + 8
				}
				return base - 6
			},
			VelocityJitter: 8, TimingJitterSec: 0.008, // EP stab — lazy but not sloppy
		})
	}

	// --- Walking bass: 4 quarter notes per bar — root, 3rd, 5th, chromatic
	// approach to the next chord's root. Jazz-influenced lofi bass — more
	// melodic than half-note root-fifth, but still tight enough to read as
	// rhythm-section. Same idea as the jazz algorithm's walking bass.
	bassNotes := make([]int, 4*numBars)
	for i := range bassNotes {
		bassNotes[i] = a.bassNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: 1, Velocity: 88, Notes: bassNotes,
		PeriodSec: cycleSec, Phase01: 0,
		MutationRate: 0.4,
		MutateOne:    func(slot int, _ int) int { return a.bassNoteAt(slot) },
		Gate:         0.82,
		Legato:       true,
		TieRepeats:   true,
		OverlapSec:   0.012,
		ResolveTimingOffsetSec: cyclicTimingOffset(
			4, 7, 5, 0,
		),
		ResolveVelocity: func(slot int, key int, base int32) int32 {
			if slot%4 == 0 {
				return base + 5
			}
			return base - 4
		},
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
		ResolveNote:            func(slot int, _ int) int { return a.vibeNoteAt(slot) },
		Gate:                   0.68,
		ResolveTimingOffsetSec: cyclicTimingOffset(16, 11, 14, 9),
		ResolveVelocity: func(slot int, key int, base int32) int32 {
			if a.section.TextureLevel > 1 {
				return base + 6
			}
			return base - 2
		},
		VelocityJitter: 12, TimingJitterSec: 0.020, // vibe — laid back
		Enabled: a.vibeOn,
	})

	// --- Nylon Guitar: comping with extended chord notes on beat 2-and (the
	// "and" of beat 2) of each bar. One hit per bar at offset 1.5 beats.
	guitarNotes := make([]int, numBars)
	for i := range guitarNotes {
		guitarNotes[i] = a.guitarNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: 4, Velocity: 50, Notes: guitarNotes,
		PeriodSec:              cycleSec,
		Phase01:                1.5 / float64(4*numBars), // 1.5 beats into the first bar
		ResolveNote:            func(slot int, _ int) int { return a.guitarNoteAt(slot) },
		Gate:                   0.44,
		ResolveTimingOffsetSec: cyclicTimingOffset(18),
		ResolveVelocity: func(slot int, key int, base int32) int32 {
			if a.section.Kind == FormCadence {
				return base + 8
			}
			return base - 3
		},
		VelocityJitter: 10, TimingJitterSec: 0.025, // nylon comping — humans don't quantize
		Enabled: a.guitarOn,
	})

	// --- Soprano Sax: a recurring 8-bar phrase with explicit rests.
	saxNotes := make([]int, len(a.saxPlan))
	for i := range saxNotes {
		saxNotes[i] = a.saxNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: 3, Velocity: 64, Notes: saxNotes,
		PeriodSec:              cycleSec,
		Phase01:                0.5 / float64(numBars), // enter on beat 3 of bar 1
		ResolveNote:            func(slot int, _ int) int { return a.saxNoteAt(slot) },
		Gate:                   0.94,
		Legato:                 true,
		TieRepeats:             true,
		OverlapSec:             0.026,
		ResolveTimingOffsetSec: chillLeadTiming(a.saxPlanCodeAt),
		ResolveVelocity: func(slot int, key int, base int32) int32 {
			switch a.section.Kind {
			case FormB, FormCadence:
				return base + 10
			case FormIntro, FormBreakdown:
				return base - 6
			default:
				return base + 2
			}
		},
		ResolveExpression: func(slot int, key int) SF2ExpressionCurve {
			curve := SF2ExpressionCurve{Start: 82, Peak: 106, End: 90, PeakAt01: 0.32}
			if a.section.Kind == FormCadence {
				curve = SF2ExpressionCurve{Start: 88, Peak: 114, End: 96, PeakAt01: 0.42}
			}
			return curve
		},
		VelocityJitter: 14, TimingJitterSec: 0.035, // sax solo — most expressive, most loose
		Enabled: a.saxOn,
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
		Gate:                   0.08,
		ResolveTimingOffsetSec: cyclicTimingOffset(-4, 0),
		VelocityJitter:         8, TimingJitterSec: 0.003, // kick — anchors the groove, must be tight
		FireProbability: 0.90, // occasional skip so the groove varies subtly
		OnFire:          core.triggerDuck,
	})
	snareNotes := make([]int, 2*numBars)
	for i := range snareNotes {
		snareNotes[i] = drumSnare
	}
	// Snare on beats 2 & 4, with the canonical "Dilla swing" 30 ms late
	// offset — research-documented signature of J Dilla's Donuts drumming
	// (snare pushed 25-35 ms behind the grid creates the "drunk" lofi feel).
	// 0.030s as a fraction of cycleSec adds to the beat-2&4 phase offset.
	const dillaSnareLagSec = 0.030
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 82, Notes: snareNotes,
		PeriodSec:              cycleSec,
		Phase01:                0.5/float64(2*numBars) + dillaSnareLagSec/cycleSec,
		Gate:                   0.10,
		ResolveTimingOffsetSec: cyclicTimingOffset(4, 6),
		VelocityJitter:         6, TimingJitterSec: 0.004,
		FireProbability: 0.88, // snare almost always lands, with rare skips
	})
	hihatNotes := make([]int, 8*numBars) // 8 hits per bar
	for i := range hihatNotes {
		hihatNotes[i] = drumHiHatC
	}
	// Dilla-style hi-hat: 55:45 long-short ratio (SwingAmount 0.05 — barely
	// swung) instead of the previous 0.13 medium-shuffle. Research found
	// Dilla's 16ths sit at ~55:45, much closer to straight than to triplet
	// swing.
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 38, Notes: hihatNotes,
		PeriodSec: cycleSec, Phase01: 0,
		Gate:                   0.06,
		ResolveTimingOffsetSec: cyclicTimingOffset(0, 12, -2, 11, -1, 13, -3, 8),
		VelocityJitter:         10,
		TimingJitterSec:        0.006,
		SwingAmount:            0.05,
		FireProbability:        0.78,
	})
	ghostNotes := make([]int, numBars)
	for i := range ghostNotes {
		ghostNotes[i] = a.ghostSnareNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 34, Notes: ghostNotes,
		PeriodSec:      cycleSec,
		Phase01:        0.875 / float64(numBars),
		ResolveNote:    func(slot int, _ int) int { return a.ghostSnareNoteAt(slot) },
		Gate:           0.08,
		VelocityJitter: 8, TimingJitterSec: 0.005,
	})
	crashNotes := make([]int, numBars)
	for i := range crashNotes {
		crashNotes[i] = a.crashNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 68, Notes: crashNotes,
		PeriodSec:      cycleSec,
		Phase01:        0,
		ResolveNote:    func(slot int, _ int) int { return a.crashNoteAt(slot) },
		Gate:           0.12,
		VelocityJitter: 10, TimingJitterSec: 0.004,
	})
	openHatNotes := make([]int, numBars)
	for i := range openHatNotes {
		openHatNotes[i] = a.openHatNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 42, Notes: openHatNotes,
		PeriodSec:      cycleSec,
		Phase01:        0.875 / float64(numBars),
		ResolveNote:    func(slot int, _ int) int { return a.openHatNoteAt(slot) },
		Gate:           0.10,
		VelocityJitter: 8, TimingJitterSec: 0.005,
	})

	// Tape hiss — subtle white-noise floor at ~-50 dBFS.
	core.setTapeHiss(0.003)
	// (Master LP and EQ are set above, near the FilterLFO config.)

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

// bassNoteAt returns the bass note for 4-per-bar walking-bass beat `slot`.
// Pattern per chord: root → 3rd → 5th → chromatic approach to next chord's
// root. Always in the low register one octave below the chord root.
func (a *Chill) bassNoteAt(slot int) int {
	totalBeats := 4 * len(a.progression)
	slot = ((slot % totalBeats) + totalBeats) % totalBeats
	chordIdx := slot / 4
	beat := slot % 4
	c := a.progression[chordIdx]
	root := a.currentRoot() + c.tones[0] - 12
	switch beat {
	case 0:
		return root
	case 1:
		return a.currentRoot() + c.tones[1] - 12 // 3rd
	case 2:
		return a.currentRoot() + c.tones[2] - 12 // 5th
	case 3:
		// Chromatic approach: ±1 semitone leading into the next chord's root.
		nextIdx := (chordIdx + 1) % len(a.progression)
		nextRoot := a.currentRoot() + a.progression[nextIdx].tones[0] - 12
		if a.rng.Float64() < 0.6 {
			return nextRoot - 1
		}
		return nextRoot + 1
	}
	return root
}

// guitarNoteAt returns a single nylon-guitar comp note per bar. Plays a
// chord-tone in the +12-semitone register (between bass and EP) at the "and"
// of beat 2 — classic jazz/bossa comping placement.
func (a *Chill) guitarNoteAt(slot int) int {
	chordIdx := slot % len(a.progression)
	return a.resolvePlanNote(slot, a.progression[chordIdx], a.guitarPlanCodeAt(slot), 12+a.section.RegisterLift, 52, 80)
}

// saxNoteAt resolves one slot of the precomputed phrase. Negative slots are
// explicit rests, which gives the solo space instead of forcing a note every
// bar.
func (a *Chill) saxNoteAt(slot int) int {
	chordIdx := slot % len(a.progression)
	return a.resolvePlanNote(slot, a.progression[chordIdx], a.saxPlanCodeAt(slot), 24+a.section.RegisterLift, 67, 92)
}

// vibeNoteAt resolves the upper-voice motif that answers the Rhodes stabs.
func (a *Chill) vibeNoteAt(slot int) int {
	chordIdx := slot % len(a.progression)
	return a.resolvePlanNote(slot, a.progression[chordIdx], a.vibePlanCodeAt(slot), 24+a.section.RegisterLift/2, 72, 94)
}

func (a *Chill) makeVibePlan(numBars int) []int {
	return trimOrRepeatPhrase(a.vibeMotifs.A, numBars, chillPlanThird)
}

func (a *Chill) vibePlanCodeAt(slot int) int {
	phrase := a.vibeMotifs.PhraseFor(a.section.Kind)
	if len(phrase) == 0 {
		phrase = a.vibePlan
	}
	if len(phrase) == 0 {
		return chillPlanThird
	}
	slot = ((slot % len(phrase)) + len(phrase)) % len(phrase)
	return phrase[slot]
}

func (a *Chill) makeVibeMotifs() MotifMemory {
	plans := [][]int{
		{chillPlanThird, chillPlanNinth},
		{chillPlanSeventh, chillPlanThirteenth},
		{chillPlanNinth, chillPlanFifth},
	}
	aCell := plans[a.rng.Intn(len(plans))]
	answerCell := plans[a.rng.Intn(len(plans))]
	aPhrase := stitchPhrase(aCell, answerCell)
	aPrime := sequencePhrase(aPhrase, map[int]int{
		chillPlanThird:      chillPlanSeventh,
		chillPlanNinth:      chillPlanEleventh,
		chillPlanFifth:      chillPlanThirteenth,
		chillPlanThirteenth: chillPlanNinth,
	})
	bPhrase := stitchPhrase(plans[a.rng.Intn(len(plans))], plans[a.rng.Intn(len(plans))])
	cadence := stitchPhrase(aPhrase[:2], []int{chillPlanEleventh, chillPlanResolveThird})
	return MotifMemory{A: aPhrase, Aprime: aPrime, B: bPhrase, Cadence: cadence, Outro: []int{chillPlanNinth, chillPlanRoot}}
}

func (a *Chill) makeGuitarPlan(numBars int) []int {
	return trimOrRepeatPhrase(a.guitarMotifs.A, numBars, chillPlanNinth)
}

func (a *Chill) guitarPlanCodeAt(slot int) int {
	phrase := a.guitarMotifs.PhraseFor(a.section.Kind)
	if len(phrase) == 0 {
		phrase = a.guitarPlan
	}
	if len(phrase) == 0 {
		return chillPlanNinth
	}
	slot = ((slot % len(phrase)) + len(phrase)) % len(phrase)
	return phrase[slot]
}

func (a *Chill) makeGuitarMotifs() MotifMemory {
	plans := [][]int{
		{chillPlanNinth, chillPlanSuspendFourth},
		{chillPlanFifth, chillPlanPickupAbove},
		{chillPlanRoot, chillPlanResolveThird},
	}
	aCell := plans[a.rng.Intn(len(plans))]
	answerCell := plans[a.rng.Intn(len(plans))]
	aPhrase := stitchPhrase(aCell, answerCell)
	aPrime := sequencePhrase(aPhrase, map[int]int{
		chillPlanPickupAbove:   chillPlanPickupBelow,
		chillPlanSuspendFourth: chillPlanResolveThird,
		chillPlanRoot:          chillPlanNinth,
	})
	bPhrase := stitchPhrase(plans[a.rng.Intn(len(plans))], []int{chillPlanNinth, chillPlanPickupAbove})
	cadence := stitchPhrase(aPhrase[:2], []int{chillPlanResolveThird, chillPlanRoot})
	return MotifMemory{A: aPhrase, Aprime: aPrime, B: bPhrase, Cadence: cadence, Outro: []int{chillPlanSuspendFourth, chillPlanRoot}}
}

func (a *Chill) makeSaxPlan(numBars int) []int {
	return trimOrRepeatPhrase(a.saxMotifs.A, numBars, chillPlanRest)
}

func (a *Chill) saxPlanCodeAt(slot int) int {
	phrase := a.saxMotifs.PhraseFor(a.section.Kind)
	if len(phrase) == 0 {
		phrase = a.saxPlan
	}
	if len(phrase) == 0 {
		return chillPlanRest
	}
	slot = ((slot % len(phrase)) + len(phrase)) % len(phrase)
	return phrase[slot]
}

func (a *Chill) makeSaxMotifs() MotifMemory {
	callTemplates := [][]int{
		{chillPlanNinth, chillPlanRest, chillPlanPickupBelow, chillPlanResolveThird},
		{chillPlanRest, chillPlanThirteenth, chillPlanSuspendFourth, chillPlanResolveThird},
		{chillPlanThird, chillPlanPickupAbove, chillPlanRest, chillPlanSeventh},
	}
	call := callTemplates[a.rng.Intn(len(callTemplates))]
	answer := []int{call[0], chillPlanEleventh, chillPlanPickupBelow, chillPlanResolveThird}
	aPhrase := stitchPhrase(call, answer)
	aPrime := sequencePhrase(aPhrase, map[int]int{
		chillPlanPickupBelow: chillPlanPickupAbove,
		chillPlanEleventh:    chillPlanThirteenth,
		chillPlanSeventh:     chillPlanNinth,
	})
	bPhrase := stitchPhrase(
		[]int{chillPlanRest, chillPlanThirteenth, chillPlanPickupAbove, chillPlanResolveThird},
		[]int{chillPlanThird, chillPlanRest, chillPlanPickupBelow, chillPlanSeventh},
	)
	cadence := stitchPhrase(aPhrase[:4], []int{chillPlanEleventh, chillPlanPickupBelow, chillPlanResolveThird, chillPlanRoot})
	return MotifMemory{A: aPhrase, Aprime: aPrime, B: bPhrase, Cadence: cadence, Outro: []int{chillPlanRest, chillPlanResolveThird, chillPlanRoot}}
}

func (a *Chill) reharmonizeProgression(base []chillChord) []chillChord {
	out := append([]chillChord(nil), base...)
	if len(out) == 0 {
		return out
	}
	for i := range out {
		next := out[(i+1)%len(out)]
		nextRoot := wrapPitchClass(next.tones[0])
		switch {
		case nextRoot == 0 && a.rng.Float64() < 0.24:
			out[i] = chillDom7(10, "bVII7")
		case chordRootSemi(out[i]) == 5 && chillHasMajorThird(out[i]) && a.rng.Float64() < 0.28:
			out[i] = chillMin7(5, "iv7")
		case nextRoot != chordRootSemi(out[i]) && a.rng.Float64() < 0.20:
			out[i] = chillSecondaryDominant(next)
		}
	}
	if len(out) > 0 && a.rng.Float64() < 0.30 {
		out[len(out)-1] = chillMaj7(8, "bVImaj7")
	}
	return out
}

func chillSecondaryDominant(target chillChord) chillChord {
	root := wrapPitchClass(chordRootSemi(target) + 7)
	return chillDom7(root, pitchClassLabel(root)+"7")
}

func chordRootSemi(chord chillChord) int {
	if len(chord.tones) == 0 {
		return 0
	}
	return wrapPitchClass(chord.tones[0])
}

func chillHasMajorThird(chord chillChord) bool {
	if len(chord.tones) < 2 {
		return false
	}
	return wrapPitchClass(chord.tones[1]-chord.tones[0]) == 4
}

func (a *Chill) resolvePlanNote(slot int, chord chillChord, code, octaveBump, low, high int) int {
	chordRoot := a.currentRoot() + chord.tones[0]
	next := a.progression[(slot+1)%len(a.progression)]
	nextRoot := a.currentRoot() + next.tones[0]
	if code == chillPlanRest {
		return -1
	}
	var key int
	switch code {
	case chillPlanRoot:
		key = chordRoot
	case chillPlanThird:
		key = a.currentRoot() + chord.tones[1]
	case chillPlanFifth:
		key = a.currentRoot() + chord.tones[2]
	case chillPlanSeventh:
		key = a.currentRoot() + chord.tones[3]
	case chillPlanNinth:
		key = chordRoot + 14
	case chillPlanEleventh:
		key = chordRoot + 17
	case chillPlanThirteenth:
		key = chordRoot + 21
	case chillPlanPickupAbove:
		key = nearestRelativeNote(chordRoot+2, nextRoot, next.tones, low, high)
		return key
	case chillPlanPickupBelow:
		key = nearestRelativeNote(chordRoot-2, nextRoot, next.tones, low, high)
		return key
	case chillPlanSuspendFourth:
		key = chordRoot + 5
	case chillPlanResolveThird:
		key = a.currentRoot() + chord.tones[1]
	default:
		key = chordRoot
	}
	return clampMidiToRange(key+octaveBump, low, high)
}

func (a *Chill) scheduleNextDrift() {
	secs := 240.0 + 180.0*a.rng.Float64()
	step := a.barSamples * 4
	if step <= 0 {
		step = int64(4 * synth.SampleRate)
	}
	a.nextDriftAt = scheduleQuantizedAfter(a.samplesElapsed, secs, step)
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
	a.applyArrangement()
	a.core.renderInto(left, right)
	prev := a.samplesElapsed
	a.samplesElapsed += int64(len(left))
	if a.samplesElapsed >= a.nextDriftAt {
		a.shiftKey()
		a.scheduleNextDrift()
	}
	if a.form.SectionBoundaryCrossed(prev, a.samplesElapsed) {
		a.applyArrangement()
	}
}

func (a *Chill) currentBar() int {
	if a.barSamples <= 0 || len(a.progression) == 0 {
		return 0
	}
	return sampleBarIndex(a.samplesElapsed, a.barSamples) % len(a.progression)
}

func (a *Chill) applyArrangement() {
	a.section = a.form.SectionAt(a.samplesElapsed)
	mix := SectionMixProfileFor(a.section)
	if a.saxOn != nil {
		*a.saxOn = a.section.LeadLevel > 0 && a.section.Kind != FormOutro
	}
	if a.guitarOn != nil {
		*a.guitarOn = a.section.TextureLevel > 0 && a.section.Kind != FormBreakdown
	}
	if a.vibeOn != nil {
		*a.vibeOn = a.section.TextureLevel > 0
	}
	if a.core == nil {
		return
	}
	core := a.core
	core.setReverbSend(3, SectionCC(96, mix.ReverbDelta))
	core.setChannelCutoff(0, SectionCC(32, mix.BrightnessDelta))
	core.setChannelCutoff(2, SectionCC(56, mix.BrightnessDelta/2))
	core.setChannelCutoff(4, SectionCC(42, mix.BrightnessDelta/2))
	core.setChannelExpression(0, SectionCC(104, mix.ExpressionDelta))
	core.setChannelExpression(2, SectionCC(100, mix.ExpressionDelta/2))
	core.setChannelExpression(3, SectionCC(108, mix.ExpressionDelta))
	core.setChannelExpression(4, SectionCC(102, mix.ExpressionDelta/2))
}

func (a *Chill) SectionGain() float64 {
	return SectionMixProfileFor(a.section).Gain
}

func (a *Chill) DebugStatus() DebugStatus {
	bar := 0
	chord := ""
	if len(a.progression) > 0 {
		bar = a.currentBar()
		chord = a.progression[bar].label
	}
	return DebugStatus{
		Chord:   chord,
		Section: string(a.section.Kind),
		Bar:     a.form.BarAt(a.samplesElapsed),
	}
}

func (a *Chill) ghostSnareNoteAt(slot int) int {
	bar := slot % maxInt(1, len(a.progression))
	if (bar+1)%4 == 0 || bar == len(a.progression)-1 {
		return drumSnare
	}
	return -1
}

func (a *Chill) crashNoteAt(slot int) int {
	bar := slot % maxInt(1, len(a.progression))
	if bar == 0 || a.section.Kind == FormCadence {
		return drumCrash
	}
	if (bar+1)%8 == 0 {
		return drumCrash
	}
	return -1
}

func (a *Chill) openHatNoteAt(slot int) int {
	bar := slot % maxInt(1, len(a.progression))
	if (bar+1)%4 == 0 {
		return drumHiHatOpen
	}
	return -1
}

// chillChannelAlternatives — staying inside the lofi soundscape. Now
// including the drum channel (9) so the kit itself rotates: Jazz Kit
// (default) → Brush Kit (40) → Standard Kit (0) → Room Kit (8). Each
// kit gives the same drum pattern a noticeably different feel — Jazz
// is warmest, Brush is softest, Standard is more "produced," Room is
// roomier. Going from one to another every few minutes is the closest
// our generator gets to "a different drummer walked in."
var chillChannelAlternatives = map[int32][]int32{
	0: {5, 4, 88, 89},   // EP2 (default), EP1, New Age Pad, Warm Pad
	1: {32, 33, 36, 38}, // Acoustic Bass, Electric Bass Finger, Slap Bass, Synth Bass 1
	2: {11, 9, 13},      // Vibraphone, Glockenspiel, Xylophone
	3: {64, 65, 66, 67}, // Soprano Sax, Alto Sax, Tenor Sax, Baritone Sax
	4: {24, 25, 26, 27}, // Nylon Guitar, Steel String, Jazz Guitar, Electric Clean
	9: {32, 40, 0, 8},   // Jazz Kit (default), Brush Kit, Standard Kit, Room Kit
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (a *Chill) ListeningMarkers() []ListeningMarker {
	return a.form.ListeningMarkers(2)
}

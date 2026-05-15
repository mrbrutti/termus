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

	barSamples int64
	form       FormPlan
	section    FormSection

	samplesElapsed int64
	nextDriftAt    int64

	saxOn *bool

	bassPlan    []int
	compLines   map[int][]int
	accentAnd2  []bool
	accentBeat4 []bool
	accentAnd4  []bool
	saxPlan     []int
	saxMotifs   MotifMemory
}

// jazzChord is one bar of harmony. tones are MIDI semitone offsets from
// (rootMidi + keyOffset): root, 3rd, 5th, 7th in that order. rootSemi is the
// semitone offset of the chord's root from the key center — used by the
// walking-bass generator to plan chromatic approach to the next chord.
type jazzChord struct {
	rootSemi int
	tones    []int
	label    string
	quality  string
}

func jazzMaj7(rootSemi int, label string) jazzChord {
	return jazzChord{
		rootSemi: rootSemi,
		tones:    []int{rootSemi, rootSemi + 4, rootSemi + 7, rootSemi + 11},
		label:    label,
		quality:  "maj7",
	}
}

func jazzMin7(rootSemi int, label string) jazzChord {
	return jazzChord{
		rootSemi: rootSemi,
		tones:    []int{rootSemi, rootSemi + 3, rootSemi + 7, rootSemi + 10},
		label:    label,
		quality:  "m7",
	}
}

func jazzDom7(rootSemi int, label string) jazzChord {
	return jazzChord{
		rootSemi: rootSemi,
		tones:    []int{rootSemi, rootSemi + 4, rootSemi + 7, rootSemi + 10},
		label:    label,
		quality:  "7",
	}
}

// jazzProgressions: 4-bar ii-V-I-vi cycles in semitone offsets from the
// tonic. Two cycles played back to back makes one 8-bar form. Hand-picked to
// cover the three most-played small-group changes.
var jazzProgressions = [][]jazzChord{
	// In C: |Dm7|G7|Cmaj7|Am7|
	{
		jazzMin7(2, "Dm7"),
		jazzDom7(7, "G7"),
		jazzMaj7(0, "Cmaj7"),
		jazzMin7(9, "Am7"),
	},
	// "Autumn Leaves" feel: |Am7|D7|Gmaj7|Cmaj7|
	{
		jazzMin7(9, "Am7"),
		jazzDom7(2, "D7"),
		jazzMaj7(7, "Gmaj7"),
		jazzMaj7(0, "Cmaj7"),
	},
	// Minor blues: |Cm7|Fm7|Gm7|Cm7|
	{
		jazzMin7(0, "Cm7"),
		jazzMin7(5, "Fm7"),
		jazzMin7(7, "Gm7"),
		jazzMin7(0, "Cm7"),
	},
	// Secondary-dominant cycle.
	{
		jazzMin7(2, "Dm7"),
		jazzDom7(7, "G7"),
		jazzMaj7(0, "Cmaj7"),
		jazzDom7(9, "A7"),
		jazzMin7(2, "Dm7"),
		jazzDom7(7, "G7"),
		jazzMaj7(0, "Cmaj7"),
		jazzDom7(9, "A7"),
	},
	// Borrowed iv + turnaround.
	{
		jazzMin7(2, "Dm7"),
		jazzDom7(7, "G7"),
		jazzMaj7(0, "Cmaj7"),
		jazzMin7(5, "Fm7"),
		jazzMin7(4, "Em7"),
		jazzDom7(9, "A7"),
		jazzMin7(2, "Dm7"),
		jazzDom7(7, "G7"),
	},
	// Tritone-sub color.
	{
		jazzMin7(2, "Dm7"),
		jazzDom7(1, "Db7"),
		jazzMaj7(0, "Cmaj7"),
		jazzDom7(9, "A7"),
		jazzMin7(2, "Dm7"),
		jazzDom7(7, "G7"),
		jazzMaj7(0, "Cmaj7"),
		jazzDom7(8, "Ab7"),
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

const (
	jazzPlanRest = iota
	jazzPlanRoot
	jazzPlanThird
	jazzPlanFifth
	jazzPlanSeventh
	jazzPlanNinth
	jazzPlanApproachAbove
	jazzPlanApproachBelow
	jazzPlanSuspendFourth
	jazzPlanResolveThird
	jazzPlanAnticipateNextRoot
)

const (
	jazzVoicingA = iota
	jazzVoicingB
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

	// Form starts with the lead muted so the intro feels like a band count-in
	// rather than a soloist appearing on beat one.
	saxStart := false
	a.saxOn = &saxStart

	// Master gain reduced (2.6 → 1.9) — jazz drum transients were pinning
	// the soft-clipper at peak. 1.9 gives ~3 dB headroom so the compressor
	// can handle dynamics without the limiter doing work.
	core, err := newSF2Core(a.sf, 1.9, seedVal)
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
	core.processMIDI(drumChannel, ccBankSelect, drumBankMSB, 0)
	const drumKitJazz = 32
	core.setProgram(drumChannel, drumKitJazz)
	core.setPan(drumChannel, 64)

	// Brighter character than lofi — jazz instruments are EQ'd to read
	// clearly; let the natural SF tone through.
	core.setChannelCutoff(0, 96)  // piano — fairly bright (no muffled-tape feel)
	core.setChannelCutoff(1, 88)  // bass — woody, mid-forward
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
	base = a.reharmonizeProgression(base)
	if len(base) < 8 {
		a.progression = make([]jazzChord, 0, 2*len(base))
		a.progression = append(a.progression, base...)
		a.progression = append(a.progression, base...)
	} else {
		a.progression = append([]jazzChord(nil), base...)
	}

	// Tempo: 120–148 BPM medium swing.
	bpm := 120.0 + 28.0*a.rng.Float64()
	beatSec := 60.0 / bpm
	barSec := beatSec * 4
	a.barSamples = secondsToSamples(barSec)
	a.form = NewFormPlan(a.rng, a.barSamples, "jazz")
	a.section = a.form.SectionAt(0)
	a.scheduleNextDrift()
	numBars := len(a.progression)
	cycleSec := barSec * float64(numBars)
	a.bassPlan = a.makeBassPlan(4 * numBars)
	a.compLines = map[int][]int{
		3: a.buildCompLine(3, numBars),
		5: a.buildCompLine(5, numBars),
		7: a.buildCompLine(7, numBars),
		9: a.buildCompLine(9, numBars),
	}
	a.accentAnd2, a.accentBeat4, a.accentAnd4 = a.makeCompAccentPlans(numBars)
	a.saxMotifs = a.makeSaxMotifs()
	a.saxPlan = trimOrRepeatPhrase(a.saxMotifs.A, 2*numBars, jazzPlanRest)
	a.applyArrangement()

	// --- Walking bass: 4 quarter notes per bar, hits every beat.
	bassNotes := make([]int, 4*numBars)
	for i := range bassNotes {
		bassNotes[i] = a.walkingBassAt(i)
	}
	core.addTrack(SF2Track{
		Channel: 1, Velocity: 92, Notes: bassNotes,
		PeriodSec: cycleSec, Phase01: 0,
		ResolveNote: func(slot int, _ int) int { return a.walkingBassAt(slot) },
		Gate:        0.84,
		Legato:      true,
		TieRepeats:  true,
		OverlapSec:  0.010,
		ResolveTimingOffsetSec: cyclicTimingOffset(
			0, 3, 1, -6,
		),
		ResolveVelocity: func(slot int, key int, base int32) int32 {
			if slot%4 == 0 {
				return base + 4
			}
			return base - 3
		},
		VelocityJitter: 8, TimingJitterSec: 0.006, // upright bassists are tight
	})

	// --- Piano comp on beat 1 — rootless A voicing (3rd + 5th + 7th + 9th).
	// The 4-note rootless voicing is the Bill-Evans / Red-Garland / Wynton-
	// Kelly foundation. Each interval gets its own track so all four sound
	// simultaneously on the down-beat, producing a proper jazz chord stab
	// rather than a triadic block.
	for _, interval := range []int{3, 5, 7, 9} {
		intv := interval
		notes := make([]int, numBars)
		for i := range notes {
			notes[i] = a.compRootless(i, intv)
		}
		core.addTrack(SF2Track{
			Channel: 0, Velocity: 74, Notes: notes,
			PeriodSec: cycleSec, Phase01: 0,
			ResolveNote:            func(slot int, _ int) int { return a.compRootless(slot, intv) },
			Gate:                   0.48,
			ResolveTimingOffsetSec: cyclicTimingOffset(0),
			ResolveVelocity: func(slot int, key int, base int32) int32 {
				if a.section.TextureLevel > 1 {
					return base + 5
				}
				return base - 2
			},
			VelocityJitter: 10, TimingJitterSec: 0.014,
		})
	}

	// --- Comping accents selected from a small rhythmic-cell library. Each
	// position has its own fixed phase; per-bar bool plans determine whether it
	// sounds or rests.
	pianoAccentAnd2 := make([]int, numBars)
	pianoAccentBeat4 := make([]int, numBars)
	pianoAccentAnd4 := make([]int, numBars)
	for i := 0; i < numBars; i++ {
		pianoAccentAnd2[i] = a.compAccentAnd2At(i)
		pianoAccentBeat4[i] = a.compAccentBeat4At(i)
		pianoAccentAnd4[i] = a.compAccentAnd4At(i)
	}
	core.addTrack(SF2Track{
		Channel: 0, Velocity: 60, Notes: pianoAccentAnd2,
		PeriodSec:              cycleSec,
		Phase01:                0.417 / float64(numBars),
		ResolveNote:            func(slot int, _ int) int { return a.compAccentAnd2At(slot) },
		Gate:                   0.34,
		ResolveTimingOffsetSec: cyclicTimingOffset(11),
		VelocityJitter:         12, TimingJitterSec: 0.020,
	})
	core.addTrack(SF2Track{
		Channel: 0, Velocity: 54, Notes: pianoAccentBeat4,
		PeriodSec:              cycleSec,
		Phase01:                0.75 / float64(numBars),
		ResolveNote:            func(slot int, _ int) int { return a.compAccentBeat4At(slot) },
		Gate:                   0.32,
		ResolveTimingOffsetSec: cyclicTimingOffset(7),
		VelocityJitter:         10, TimingJitterSec: 0.018,
	})
	core.addTrack(SF2Track{
		Channel: 0, Velocity: 58, Notes: pianoAccentAnd4,
		PeriodSec:              cycleSec,
		Phase01:                0.917 / float64(numBars),
		ResolveNote:            func(slot int, _ int) int { return a.compAccentAnd4At(slot) },
		Gate:                   0.34,
		ResolveTimingOffsetSec: cyclicTimingOffset(10),
		VelocityJitter:         12, TimingJitterSec: 0.020,
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
		Gate:                   0.08,
		ResolveTimingOffsetSec: cyclicTimingOffset(-2, -4, -1, -3),
		VelocityJitter:         10, TimingJitterSec: 0.004,
	})
	// --- Ride: swung "& of 2" and "& of 4" — completes the jazz ride pattern.
	// 2 hits per bar at swung-8th positions 1.667/4 and 3.667/4.
	rideSwungNotes := make([]int, 2*numBars)
	for i := range rideSwungNotes {
		rideSwungNotes[i] = jazzRideCymbal
	}
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 62, Notes: rideSwungNotes,
		PeriodSec:              cycleSec,
		Phase01:                0.417 / float64(numBars), // start at "& of 2" of bar 0
		Gate:                   0.08,
		ResolveTimingOffsetSec: cyclicTimingOffset(-6, -4),
		VelocityJitter:         10, TimingJitterSec: 0.006,
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
		PeriodSec:              cycleSec,
		Phase01:                0.25 / float64(numBars), // beats 2 & 4
		Gate:                   0.08,
		ResolveTimingOffsetSec: cyclicTimingOffset(8, 10),
		VelocityJitter:         8, TimingJitterSec: 0.004,
	})

	// --- Brushed snare backing: occasional weak hit on beat 4. Sparse —
	// fires only ~40% of bars.
	snareNotes := make([]int, numBars)
	for i := range snareNotes {
		snareNotes[i] = a.snareCompNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 52, Notes: snareNotes,
		PeriodSec:              cycleSec,
		Phase01:                0.75 / float64(numBars), // beat 4
		ResolveNote:            func(slot int, _ int) int { return a.snareCompNoteAt(slot) },
		Gate:                   0.08,
		ResolveTimingOffsetSec: cyclicTimingOffset(18),
		VelocityJitter:         12, TimingJitterSec: 0.010,
		FireProbability: 0.82,
	})

	// --- Feathered kick on beats 1 and 3 — barely audible in modern jazz,
	// just enough to anchor the time. Velocity very low.
	kickNotes := make([]int, 2*numBars)
	for i := range kickNotes {
		kickNotes[i] = a.kickNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 38, Notes: kickNotes,
		PeriodSec: cycleSec, Phase01: 0,
		ResolveNote:            func(slot int, _ int) int { return a.kickNoteAt(slot) },
		Gate:                   0.08,
		ResolveTimingOffsetSec: cyclicTimingOffset(0, 3),
		VelocityJitter:         6, TimingJitterSec: 0.004,
		FireProbability: 0.90,
	})

	// --- Sax solo: 2-slot-per-bar phrase with explicit rests, so the line can
	// breathe and answer itself across 1-2 bar spans.
	saxNotes := make([]int, len(a.saxPlan))
	for i := range saxNotes {
		saxNotes[i] = a.saxNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 70, Notes: saxNotes,
		PeriodSec:              cycleSec,
		Phase01:                0,
		ResolveNote:            func(slot int, _ int) int { return a.saxNoteAt(slot) },
		Gate:                   0.96,
		Legato:                 true,
		TieRepeats:             true,
		OverlapSec:             0.022,
		ResolveTimingOffsetSec: jazzSaxTiming(a.saxPlanCodeAt),
		ResolveVelocity: func(slot int, key int, base int32) int32 {
			switch a.section.Kind {
			case FormB, FormCadence:
				return base + 8
			case FormIntro, FormOutro:
				return base - 8
			default:
				return base + 2
			}
		},
		ResolveExpression: func(slot int, key int) SF2ExpressionCurve {
			curve := SF2ExpressionCurve{Start: 84, Peak: 108, End: 92, PeakAt01: 0.34}
			if a.section.Kind == FormCadence {
				curve = SF2ExpressionCurve{Start: 90, Peak: 116, End: 96, PeakAt01: 0.42}
			}
			return curve
		},
		ResolveModWheel: func(slot int, key int) SF2ExpressionCurve {
			if a.section.Kind == FormCadence {
				return gentleVibratoCurve(0, 26, 12)
			}
			return gentleVibratoCurve(0, 20, 10)
		},
		ResolveBrightness: func(slot int, key int) SF2ExpressionCurve {
			if a.section.Kind == FormCadence {
				return brightnessBloomCurve(102, 120, 106)
			}
			return brightnessBloomCurve(96, 112, 100)
		},
		ResolveDetuneCents: slotDetunePattern(-3, 2, -1, 4),
		VelocityJitter:     16, TimingJitterSec: 0.040, // sax is the most expressive — loose timing
		Enabled: a.saxOn,
	})
	crashNotes := make([]int, numBars)
	for i := range crashNotes {
		crashNotes[i] = a.crashNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 74, Notes: crashNotes,
		PeriodSec:      cycleSec,
		Phase01:        0,
		ResolveNote:    func(slot int, _ int) int { return a.crashNoteAt(slot) },
		Gate:           0.12,
		VelocityJitter: 10, TimingJitterSec: 0.004,
	})
	ghostNotes := make([]int, numBars)
	for i := range ghostNotes {
		ghostNotes[i] = a.ghostSnareNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 30, Notes: ghostNotes,
		PeriodSec:      cycleSec,
		Phase01:        0.875 / float64(numBars),
		ResolveNote:    func(slot int, _ int) int { return a.ghostSnareNoteAt(slot) },
		Gate:           0.08,
		VelocityJitter: 8, TimingJitterSec: 0.006,
	})
	fillNotes := make([]int, 4*numBars)
	for i := range fillNotes {
		fillNotes[i] = a.brushFillNoteAt(i)
	}
	core.addTrack(SF2Track{
		Channel: drumChannel, Velocity: 42, Notes: fillNotes,
		PeriodSec:      cycleSec,
		Phase01:        0,
		ResolveNote:    func(slot int, _ int) int { return a.brushFillNoteAt(slot) },
		Gate:           0.06,
		VelocityJitter: 8, TimingJitterSec: 0.005,
		FireProbability: 0.88,
	})

	a.core = core
}

// walkingBassAt returns the MIDI key for the i-th beat of the cycle. Standard
// walking pattern now follows a precomputed plan so the line has consistent
// guide-tone motion across bars instead of rolling note choices on demand.
func (a *Jazz) walkingBassAt(slot int) int {
	if len(a.progression) == 0 {
		return a.currentRoot()
	}
	totalBeats := 4 * len(a.progression)
	slot = ((slot % totalBeats) + totalBeats) % totalBeats
	bar := slot / 4
	chord := a.progression[bar]
	nextBar := (bar + 1) % len(a.progression)
	switch a.bassPlan[slot%len(a.bassPlan)] {
	case jazzPlanRoot:
		return clampMidiToRange(a.currentRoot()+chord.tones[0]-12, 36, 55)
	case jazzPlanThird:
		return clampMidiToRange(a.currentRoot()+chord.tones[1]-12, 36, 55)
	case jazzPlanFifth:
		return clampMidiToRange(a.currentRoot()+chord.tones[2]-12, 36, 55)
	case jazzPlanSeventh:
		return clampMidiToRange(a.currentRoot()+chord.tones[3]-12, 36, 55)
	case jazzPlanApproachAbove:
		return clampMidiToRange(a.currentRoot()+a.progression[nextBar].tones[0]-11, 36, 55)
	case jazzPlanApproachBelow:
		return clampMidiToRange(a.currentRoot()+a.progression[nextBar].tones[0]-13, 36, 55)
	default:
		return clampMidiToRange(a.currentRoot()+chord.tones[0]-12, 36, 55)
	}
}

// compRootless returns one voice of the rootless 4-note jazz voicing
// (3-5-7-9) on the current bar. interval = 3, 5, 7, or 9 picks which chord
// tone. The 9th is computed as the chord's root + 14 semitones. Bars switch
// between A and B voicings so the comp breathes instead of repeating one
// piano grip forever.
func (a *Jazz) compRootless(slot, interval int) int {
	if line, ok := a.compLines[interval]; ok && len(line) > 0 {
		return line[((slot%len(line))+len(line))%len(line)]
	}
	if len(a.progression) == 0 {
		return 60
	}
	bar := ((slot % len(a.progression)) + len(a.progression)) % len(a.progression)
	chord := a.progression[bar]
	if len(chord.tones) < 4 {
		return 60
	}
	var rel int
	switch interval {
	case 3:
		rel = chord.tones[1]
	case 5:
		rel = chord.tones[2]
	case 7:
		rel = chord.tones[3]
	case 9:
		// 9th = chord root + 14 semitones (= one octave + one whole step).
		rel = chord.tones[0] + 14
	default:
		rel = chord.tones[1]
	}
	key := a.currentRoot() + rel
	return clampMidiToRange(key, 58, 76)
}

func (a *Jazz) compAccentNoteAt(slot int, active []bool, interval int) int {
	if slot >= len(active) || !active[slot] {
		return -1
	}
	return a.compRootless(slot, interval)
}

func (a *Jazz) compAccentAnd2At(slot int) int {
	if !a.allowCompDialogue(slot, false) {
		return -1
	}
	return a.compAccentNoteAt(slot, a.accentAnd2, 9)
}

func (a *Jazz) compAccentBeat4At(slot int) int {
	if !a.allowCompDialogue(slot, true) {
		return -1
	}
	return a.compAccentNoteAt(slot, a.accentBeat4, 7)
}

func (a *Jazz) compAccentAnd4At(slot int) int {
	if !a.allowCompDialogue(slot, true) {
		return -1
	}
	return a.compAccentNoteAt(slot, a.accentAnd4, 5)
}

func (a *Jazz) allowCompDialogue(bar int, lateHalf bool) bool {
	if a.section.LeadLevel == 0 || a.saxOn == nil || !*a.saxOn {
		return true
	}
	front, back := a.saxActivityInBar(bar)
	if lateHalf {
		return !back
	}
	return !front
}

func (a *Jazz) saxActivityInBar(bar int) (front, back bool) {
	base := bar * 2
	return a.saxPlanCodeAt(base) != jazzPlanRest, a.saxPlanCodeAt(base+1) != jazzPlanRest
}

// saxNoteAt resolves one slot of the phrase plan. The plan includes explicit
// rests and chord/color-tone targets, producing short phrases instead of
// isolated spot notes.
func (a *Jazz) saxNoteAt(slot int) int {
	if len(a.progression) == 0 {
		return 72
	}
	chordIdx := (slot / 2) % len(a.progression)
	chord := a.progression[chordIdx]
	next := a.progression[(chordIdx+1)%len(a.progression)]
	switch a.saxPlanCodeAt(slot) {
	case jazzPlanRest:
		return -1
	case jazzPlanRoot:
		return clampMidiToRange(a.currentRoot()+chord.tones[0]+12, 64, 86)
	case jazzPlanThird:
		return clampMidiToRange(a.currentRoot()+chord.tones[1]+12, 64, 86)
	case jazzPlanFifth:
		return clampMidiToRange(a.currentRoot()+chord.tones[2]+12, 64, 86)
	case jazzPlanSeventh:
		return clampMidiToRange(a.currentRoot()+chord.tones[3]+12, 64, 86)
	case jazzPlanNinth:
		return clampMidiToRange(a.currentRoot()+chord.tones[0]+14, 64, 86)
	case jazzPlanApproachAbove:
		return clampMidiToRange(a.currentRoot()+next.tones[0]+13, 64, 86)
	case jazzPlanApproachBelow:
		return clampMidiToRange(a.currentRoot()+next.tones[0]+11, 64, 86)
	case jazzPlanSuspendFourth:
		return clampMidiToRange(a.currentRoot()+chord.tones[0]+17, 64, 86)
	case jazzPlanResolveThird:
		return clampMidiToRange(a.currentRoot()+chord.tones[1]+12, 64, 86)
	case jazzPlanAnticipateNextRoot:
		return clampMidiToRange(a.currentRoot()+next.tones[0]+12, 64, 86)
	default:
		return clampMidiToRange(a.currentRoot()+chord.tones[1]+12, 64, 86)
	}
}

func (a *Jazz) makeBassPlan(totalBeats int) []int {
	out := make([]int, totalBeats)
	for bar := 0; bar < len(a.progression); bar++ {
		base := bar * 4
		chord := a.progression[bar]
		next := a.progression[(bar+1)%len(a.progression)]
		out[base] = jazzPlanRoot
		if next.rootSemi >= chord.rootSemi {
			out[base+1] = jazzPlanThird
			out[base+2] = jazzPlanFifth
			out[base+3] = jazzPlanApproachBelow
		} else {
			out[base+1] = jazzPlanFifth
			out[base+2] = jazzPlanSeventh
			out[base+3] = jazzPlanApproachAbove
		}
	}
	return out
}

func (a *Jazz) reharmonizeProgression(base []jazzChord) []jazzChord {
	out := append([]jazzChord(nil), base...)
	if len(out) == 0 {
		return out
	}
	for i := range out {
		next := out[(i+1)%len(out)]
		switch {
		case next.quality == "m7" && a.rng.Float64() < 0.28:
			out[i] = a.secondaryDominantOf(next)
		case jazzIsDominant(out[i]) && a.rng.Float64() < 0.22:
			out[i] = jazzTritoneSub(out[i])
		case out[i].quality == "maj7" && next.rootSemi == 9 && a.rng.Float64() < 0.25:
			out[i] = jazzMaj7(8, "Abmaj7")
		}
	}
	if len(out) >= 2 && out[len(out)-1].quality != "7" && a.rng.Float64() < 0.35 {
		out[len(out)-1] = a.secondaryDominantOf(out[0])
	}
	return out
}

func (a *Jazz) secondaryDominantOf(target jazzChord) jazzChord {
	root := wrapPitchClass(target.rootSemi + 7)
	label := pitchClassLabel(root) + "7"
	return jazzDom7(root, label)
}

func jazzTritoneSub(chord jazzChord) jazzChord {
	root := wrapPitchClass(chord.rootSemi + 6)
	label := pitchClassLabel(root) + "7"
	return jazzDom7(root, label)
}

func jazzIsDominant(chord jazzChord) bool {
	return chord.quality == "7"
}

func (a *Jazz) buildCompLine(interval, numBars int) []int {
	out := make([]int, numBars)
	voicing := jazzVoicingA
	prev := 0
	for i := 0; i < numBars; i++ {
		if i > 0 && a.progression[i].rootSemi != a.progression[i-1].rootSemi && a.rng.Float64() < 0.75 {
			if voicing == jazzVoicingA {
				voicing = jazzVoicingB
			} else {
				voicing = jazzVoicingA
			}
		}
		chord := a.progression[i%len(a.progression)]
		var rel int
		switch interval {
		case 3:
			rel = chord.tones[1]
		case 5:
			rel = chord.tones[2]
		case 7:
			rel = chord.tones[3]
		default:
			rel = chord.tones[0] + 14
		}
		target := a.currentRoot() + rel
		if voicing == jazzVoicingB {
			switch interval {
			case 3, 5:
				target += 12
			default:
				target -= 12
			}
		}
		if prev == 0 {
			out[i] = clampMidiToRange(target, 58, 76)
		} else {
			out[i] = voiceLeadNearest(prev, target, []int{0}, 58, 76)
		}
		prev = out[i]
	}
	return out
}

func (a *Jazz) makeCompAccentPlans(numBars int) ([]bool, []bool, []bool) {
	and2 := make([]bool, numBars)
	beat4 := make([]bool, numBars)
	and4 := make([]bool, numBars)
	cells := []struct {
		and2  bool
		beat4 bool
		and4  bool
	}{
		{and2: true, beat4: false, and4: false},
		{and2: true, beat4: true, and4: false},
		{and2: false, beat4: false, and4: true},
		{and2: false, beat4: true, and4: true},
	}
	for i := 0; i < numBars; i++ {
		cell := cells[a.rng.Intn(len(cells))]
		and2[i] = cell.and2
		beat4[i] = cell.beat4
		and4[i] = cell.and4
	}
	return and2, beat4, and4
}

func (a *Jazz) makeSaxPlan(numSlots int) []int {
	return trimOrRepeatPhrase(a.saxMotifs.A, numSlots, jazzPlanRest)
}

func (a *Jazz) saxPlanCodeAt(slot int) int {
	phrase := a.saxMotifs.PhraseFor(a.section.Kind)
	if len(phrase) == 0 {
		phrase = a.saxPlan
	}
	if len(phrase) == 0 {
		return jazzPlanRest
	}
	slot = ((slot % len(phrase)) + len(phrase)) % len(phrase)
	return phrase[slot]
}

func (a *Jazz) makeSaxMotifs() MotifMemory {
	callTemplates := [][]int{
		{jazzPlanRest, jazzPlanNinth, jazzPlanApproachBelow, jazzPlanResolveThird},
		{jazzPlanThird, jazzPlanRest, jazzPlanSuspendFourth, jazzPlanResolveThird},
		{jazzPlanRest, jazzPlanSeventh, jazzPlanApproachAbove, jazzPlanAnticipateNextRoot},
	}
	call := callTemplates[a.rng.Intn(len(callTemplates))]
	answer := []int{jazzPlanThird, jazzPlanNinth, jazzPlanApproachBelow, jazzPlanResolveThird}
	aPhrase := stitchPhrase(call, answer)
	aPrime := sequencePhrase(aPhrase, map[int]int{
		jazzPlanNinth:              jazzPlanSeventh,
		jazzPlanSeventh:            jazzPlanNinth,
		jazzPlanApproachBelow:      jazzPlanApproachAbove,
		jazzPlanAnticipateNextRoot: jazzPlanApproachBelow,
	})
	bPhrase := stitchPhrase(
		[]int{jazzPlanRest, jazzPlanSeventh, jazzPlanApproachAbove, jazzPlanAnticipateNextRoot},
		[]int{jazzPlanThird, jazzPlanSuspendFourth, jazzPlanApproachBelow, jazzPlanResolveThird},
	)
	cadence := stitchPhrase(aPhrase[:4], []int{jazzPlanThird, jazzPlanNinth, jazzPlanApproachBelow, jazzPlanRoot})
	outro := stitchPhrase([]int{jazzPlanRest, jazzPlanThird, jazzPlanResolveThird, jazzPlanRoot})
	return MotifMemory{
		A:       aPhrase,
		Aprime:  aPrime,
		B:       bPhrase,
		Cadence: cadence,
		Outro:   outro,
	}
}

// scheduleNextDrift picks when the next macro key-drift will fire.
// 4–8 minutes between drifts, deterministic from rng.
func (a *Jazz) scheduleNextDrift() {
	mins := 4.0 + 4.0*a.rng.Float64()
	step := a.barSamples * 4
	if step <= 0 {
		step = 4 * 44100
	}
	a.nextDriftAt = scheduleQuantizedAfter(a.samplesElapsed, mins*60.0, step)
}

// applyMacroMutations is called once per render block. Currently handles
// section toggles and key drift.
func (a *Jazz) applyMacroMutations(prev int64) {
	a.applyArrangement()
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
	if a.form.SectionBoundaryCrossed(prev, a.samplesElapsed) {
		a.applyArrangement()
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
	prev := a.samplesElapsed
	a.core.renderInto(left, right)
	a.samplesElapsed += int64(len(left))
	a.applyMacroMutations(prev)
}

func (a *Jazz) applyArrangement() {
	a.section = a.form.SectionAt(a.samplesElapsed)
	if a.saxOn != nil {
		*a.saxOn = a.section.LeadLevel > 0 && a.section.Kind != FormOutro
	}
	if a.core == nil {
		return
	}
	comp := SectionSceneFor(a.section, RoleComp)
	bass := SectionSceneFor(a.section, RoleBass)
	lead := SectionSceneFor(a.section, RoleLead)
	drums := SectionSceneFor(a.section, RoleDrums)
	a.core.setReverbSend(0, SectionCC(48, comp.ReverbDelta))
	a.core.setReverbSend(1, SectionCC(18, bass.ReverbDelta))
	a.core.setReverbSend(2, SectionCC(86, lead.ReverbDelta))
	a.core.setReverbSend(drumChannel, SectionCC(42, drums.ReverbDelta))
	a.core.setChannelCutoff(0, SectionCC(96, comp.BrightnessDelta))
	a.core.setChannelCutoff(1, SectionCC(88, bass.BrightnessDelta))
	a.core.setChannelCutoff(2, SectionCC(110, lead.BrightnessDelta))
	a.core.setChannelCutoff(drumChannel, SectionCC(92, drums.BrightnessDelta))
	a.core.setChannelExpression(0, SectionCC(108, comp.ExpressionDelta))
	a.core.setChannelExpression(1, SectionCC(104, bass.ExpressionDelta))
	a.core.setChannelExpression(2, SectionCC(110, lead.ExpressionDelta))
	a.core.setChannelExpression(drumChannel, SectionCC(100, drums.ExpressionDelta))
}

func (a *Jazz) SectionGain() float64 {
	return SectionMixProfileFor(a.section).Gain
}

func (a *Jazz) DebugStatus() DebugStatus {
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

func (a *Jazz) crashNoteAt(slot int) int {
	bar := slot % len(a.progression)
	if bar == 0 || a.section.Kind == FormCadence {
		return drumCrash
	}
	if (bar+1)%4 == 0 && a.section.Kind != FormBreakdown {
		return drumCrash
	}
	return -1
}

func (a *Jazz) currentBar() int {
	if a.barSamples <= 0 || len(a.progression) == 0 {
		return 0
	}
	return sampleBarIndex(a.samplesElapsed, a.barSamples) % len(a.progression)
}

func (a *Jazz) ghostSnareNoteAt(slot int) int {
	bar := slot % len(a.progression)
	if a.section.Kind == FormCadence {
		return jazzSnareBrushed
	}
	if bar%4 == 1 || (bar+1)%4 == 0 {
		return jazzSnareBrushed
	}
	return -1
}

func (a *Jazz) kickNoteAt(slot int) int {
	if len(a.progression) == 0 {
		return jazzKickKey
	}
	bar := (slot / 2) % len(a.progression)
	beat := slot % 2
	if beat == 0 {
		return jazzKickKey
	}
	if a.section.Kind == FormCadence || (bar+1)%4 == 0 {
		return jazzKickKey
	}
	if a.section.Kind == FormB && bar%2 == 0 {
		return jazzKickKey
	}
	if bar%4 == 0 {
		return jazzKickKey
	}
	return -1
}

func (a *Jazz) snareCompNoteAt(slot int) int {
	if len(a.progression) == 0 {
		return jazzSnareBrushed
	}
	bar := slot % len(a.progression)
	if a.section.Kind == FormIntro && bar%2 == 1 {
		return -1
	}
	if a.section.Kind == FormBreakdown && (bar+1)%4 != 0 {
		return -1
	}
	if a.section.Kind == FormCadence || (bar+1)%4 == 0 || bar%4 == 1 {
		return jazzSnareBrushed
	}
	return -1
}

func (a *Jazz) brushFillNoteAt(slot int) int {
	if len(a.progression) == 0 {
		return -1
	}
	bar := (slot / 4) % len(a.progression)
	step := slot % 4
	if a.section.Kind != FormCadence && (bar+1)%4 != 0 {
		return -1
	}
	if a.section.Kind == FormCadence {
		if step >= 1 {
			return jazzSnareBrushed
		}
		return -1
	}
	if step >= 2 {
		return jazzSnareBrushed
	}
	return -1
}

func (a *Jazz) ListeningMarkers() []ListeningMarker {
	return a.form.ListeningMarkers(2)
}

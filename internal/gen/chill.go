package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"

	"github.com/mrbrutti/termus/internal/synth"
)

var _ Algorithm = (*Chill)(nil)
var _ SF2Reverberator = (*Chill)(nil)

// Chill is a lofi-style algorithm: an Electric Piano 2 (chorused Rhodes)
// arpeggiating extended jazz chords (m7, dom7, maj7), an acoustic bass
// walking through the chord on each beat, and a sparse vibraphone melody
// on top. The harmony is a ii-V-I-VI jazz turnaround in a major key,
// looping every 12.8 s (~75 BPM, 4 beats per chord, 4 chords).
//
// For hours-long listening:
//   - per-track mutation: ~15% chance per slot transition to re-roll one
//     note from the current chord's tone set
//   - macro key-drift: every 4–7 minutes the entire key transposes ±1..2
//     semitones, taking effect gradually as mutations roll in
//   - progression also drifts: occasionally one chord in the progression
//     is swapped with a relative substitute (vi for I, IV for ii, etc.)
//
// The algorithm auto-installs a short SyntheticRoomIR by default for the
// "in the room" lofi feel; --ir overrides.
type Chill struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
	rng  *rand.Rand

	rootMidi  int // base tonic (a MAJOR-key root)
	keyOffset int

	// Track the active chord progression so mutators can re-roll within it.
	progression []chillChord

	samplesElapsed int64
	nextDriftAt    int64
}

// chillChord is one chord in the loop: a list of MIDI offsets (relative to
// rootMidi+keyOffset) plus a label kept for debugging.
type chillChord struct {
	tones []int  // semitones from key tonic; chord-tones in canonical order
	label string // e.g. "ii7", "V7"
}

// scaleMajor is the major scale (degrees of the diatonic scale in semitones).
// Used by Chill since lofi is typically rooted in a major key.
var scaleMajor = []int{0, 2, 4, 5, 7, 9, 11}

// chillProgressions: each is a 4-chord turnaround that loops every 12.8 s
// (4 beats × 4 chords at ~75 BPM). Each chord is a 4-note 7th voicing,
// expressed in semitones from the major-key tonic. ii-V-I-VI is the classic
// jazz/lofi turnaround.
var chillProgressions = [][]chillChord{
	{
		// ii-V-I-VI in major
		{tones: []int{2, 5, 9, 12}, label: "ii7"},   // Dm7 in C: D F A C
		{tones: []int{7, 11, 14, 17}, label: "V7"},  // G7:        G B D F
		{tones: []int{0, 4, 7, 11}, label: "Imaj7"}, // Cmaj7:     C E G B
		{tones: []int{9, 12, 16, 19}, label: "vi7"}, // Am7:       A C E G
	},
	{
		// IV-V-iii-vi
		{tones: []int{5, 9, 12, 16}, label: "IVmaj7"},
		{tones: []int{7, 11, 14, 17}, label: "V7"},
		{tones: []int{4, 7, 11, 14}, label: "iii7"},
		{tones: []int{9, 12, 16, 19}, label: "vi7"},
	},
	{
		// Imaj7-IVmaj7-iii7-VImaj7 (more wistful)
		{tones: []int{0, 4, 7, 11}, label: "Imaj7"},
		{tones: []int{5, 9, 12, 16}, label: "IVmaj7"},
		{tones: []int{4, 7, 11, 14}, label: "iii7"},
		{tones: []int{9, 13, 16, 20}, label: "VImaj7"}, // tonicizes vi → uses major 3rd
	},
}

// NewChill constructs the algorithm. Caller must call Seed before Next.
func NewChill(sf *meltysynth.SoundFont) *Chill { return &Chill{sf: sf} }

func (a *Chill) Name() string { return "chill" }

func (a *Chill) currentRoot() int { return a.rootMidi + a.keyOffset }

func (a *Chill) Seed(seedVal int64) {
	a.rng = rand.New(rand.NewSource(seedVal)) //nolint:gosec
	// Tonic in C3..F#3 — comfortable EP register, bass falls into mid range.
	a.rootMidi = 48 + a.rng.Intn(7)
	a.keyOffset = 0
	a.samplesElapsed = 0
	a.scheduleNextDrift()

	core, err := newSF2Core(a.sf, 3.2, seedVal)
	if err != nil {
		a.core = nil
		return
	}
	// Channel layout:
	//   0 — Electric Piano 2 / chorused Rhodes (#5)  arpeggio
	//   1 — Acoustic Bass (#32)                      walking bass
	//   2 — Vibraphone (#11)                         sparse top melody
	core.setProgram(0, 5)
	core.setProgram(1, 32)
	core.setProgram(2, 11)

	// Pick a progression.
	a.progression = chillProgressions[a.rng.Intn(len(chillProgressions))]

	// Tempo: ~75 BPM. 4 beats per chord × 4 chords = 16 beats per cycle.
	const bpm = 75.0
	beatSec := 60.0 / bpm
	cycleSec := beatSec * float64(4*len(a.progression))
	// numBeats = 4 chords × 4 beats = 16
	numBeats := 4 * len(a.progression)

	// EP arpeggio: 16 notes per cycle, one per beat. Cycle through chord
	// tones [root, 3rd, 5th, 7th] for each chord (4 beats × 4 chords).
	epNotes := make([]int, numBeats)
	for i := range epNotes {
		epNotes[i] = a.epNoteAt(i)
	}
	epMutate := func(slot int, _ int) int { return a.epNoteAt(slot) }

	// Bass: walking pattern [root, fifth, third, seventh] across the 4 beats
	// of each chord. Same 16-note total length.
	bassNotes := make([]int, numBeats)
	for i := range bassNotes {
		bassNotes[i] = a.bassNoteAt(i)
	}
	bassMutate := func(slot int, _ int) int { return a.bassNoteAt(slot) }

	// Vibraphone melody: sparse — one note per chord (4 notes per cycle).
	// Picks a random chord tone in the high register, occasionally a 9th
	// (chord_root + 14) for color. Each note lasts the full chord.
	vibeNotes := make([]int, len(a.progression))
	for i := range vibeNotes {
		vibeNotes[i] = a.vibeNoteAt(i)
	}
	vibeMutate := func(slot int, _ int) int { return a.vibeNoteAt(slot) }

	core.addTrack(SF2Track{
		Channel: 0, Velocity: 78, Notes: epNotes,
		PeriodSec: cycleSec, Phase01: 0,
		MutationRate: 0.18, MutateOne: epMutate,
	})
	core.addTrack(SF2Track{
		Channel: 1, Velocity: 94, Notes: bassNotes,
		PeriodSec: cycleSec, Phase01: 0,
		MutationRate: 0.10, MutateOne: bassMutate,
	})
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 70, Notes: vibeNotes,
		PeriodSec: cycleSec, Phase01: 0,
		MutationRate: 0.30, MutateOne: vibeMutate,
	})

	// Soft room reverb by default — the "in a small studio" lofi feel.
	core.setConvolutionIR(synth.SyntheticRoomIR(0.10), 0.45)

	a.core = core
}

// epNoteAt returns the EP arpeggio note for the given beat slot (0..numBeats-1).
// Chord changes every 4 beats; within each chord, walks 0→1→2→1 through the
// chord tones so the figure feels like a rocking arpeggio rather than a
// dry up-down sweep.
func (a *Chill) epNoteAt(slot int) int {
	chordIdx := (slot / 4) % len(a.progression)
	beat := slot % 4
	tonePattern := []int{0, 1, 2, 1}
	c := a.progression[chordIdx]
	// EP voicing in the +24 register (above bass, below vibe).
	return a.currentRoot() + c.tones[tonePattern[beat]] + 24
}

// bassNoteAt returns the walking-bass note for the given beat slot.
// Pattern per chord: root, fifth, third, seventh — a classic chord-tone walk
// that outlines the harmony while still feeling melodic.
func (a *Chill) bassNoteAt(slot int) int {
	chordIdx := (slot / 4) % len(a.progression)
	beat := slot % 4
	bassPattern := []int{0, 2, 1, 3} // root, 5th, 3rd, 7th
	c := a.progression[chordIdx]
	return a.currentRoot() + c.tones[bassPattern[beat]] - 12
}

// vibeNoteAt returns one melody note per chord. 70% chord tone, 30% chord
// extension (9th or 13th relative to the chord root) for jazzy color.
func (a *Chill) vibeNoteAt(slot int) int {
	chordIdx := slot % len(a.progression)
	c := a.progression[chordIdx]
	chordRoot := a.currentRoot() + c.tones[0]
	if a.rng.Float64() < 0.30 {
		// Color tone: 9th (chord root + 14) or 13th (chord root + 21).
		if a.rng.Float64() < 0.5 {
			return chordRoot + 14
		}
		return chordRoot + 21
	}
	// Chord tone in high register.
	return a.currentRoot() + c.tones[a.rng.Intn(len(c.tones))] + 36
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
}

package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"

	"github.com/mrbrutti/termus/internal/synth"
)

// Compile-time assertion that *SF2 implements Algorithm.
var _ Algorithm = (*SF2)(nil)

// SF2 is the hi-fi algorithm: a real chord progression in a minor key, played
// by four sampled instruments (acoustic grand piano arpeggio, string ensemble
// pad, warm synth pad, acoustic bass) via the go-meltysynth SoundFont
// synthesizer. The chord-based scheduling and voice-leading rules give the
// output a "composed" feel that the random-walk algorithms cannot match,
// while the sampled timbres provide the realism that pure synthesis lacks.
//
// The caller must inject a *meltysynth.SoundFont before calling Seed.
type SF2 struct {
	rng *rand.Rand

	sf  *meltysynth.SoundFont
	syn *meltysynth.Synthesizer

	// Music-theory state.
	rootMidi  int        // MIDI number of the tonic, low register (e.g. 45 = A2)
	chords    []chord    // the progression
	chordIdx  int        // current chord
	chordHold int64      // samples remaining on the current chord
	chordLen  int64      // total samples per chord (constant within a session)

	// Beat-level state for the piano arpeggio.
	beatIdx     int   // 0..3 within the current chord
	beatHold    int64 // samples remaining on the current beat
	beatLen     int64

	t int64 // total samples since Seed

	// Active notes per MIDI channel, so we can NoteOff them on chord change.
	activePad    []int // strings + pad combined
	activeBass   int   // single bass note, or -1
	activePiano  int   // last arpeggio note, or -1

	// Scratch for Render — go-meltysynth uses []float32, we use []float64.
	bufF32L []float32
	bufF32R []float32
}

// chord is a set of MIDI note numbers (one per chord tone, in canonical
// root-3rd-5th-7th order). The notes are absolute MIDI, already in the
// register we want for the pad/strings.
type chord struct {
	rootDeg int   // scale degree of the chord root (0..6)
	tones   []int // MIDI note numbers of chord tones (4 notes typically)
}

// Common minor-key progressions (Roman numerals applied to natural minor).
// Each progression is a list of scale degrees (0=i, 1=ii°, 2=III, 3=iv, 4=v,
// 5=VI, 6=VII) — the root of each chord in the progression.
var minorProgressions = [][]int{
	{0, 5, 2, 6}, // i - VI - III - VII (Andalusian-ish)
	{0, 4, 5, 2}, // i - v - VI - III
	{0, 3, 6, 2}, // i - iv - VII - III
	{0, 5, 3, 6}, // i - VI - iv - VII
	{0, 2, 5, 3}, // i - III - VI - iv
}

// NewSF2 constructs the algorithm bound to the given soundfont.
// Caller must call Seed before Next.
func NewSF2(sf *meltysynth.SoundFont) *SF2 {
	return &SF2{sf: sf}
}

func (s *SF2) Name() string { return "sf2-progression" }

func (s *SF2) Seed(seedVal int64) {
	s.rng = rand.New(rand.NewSource(seedVal)) //nolint:gosec
	// Tonic in A1..G2 range — gives a warm bass register without going subsonic.
	s.rootMidi = 33 + s.rng.Intn(8) // A1..F2

	// Build the chord progression.
	prog := minorProgressions[s.rng.Intn(len(minorProgressions))]
	s.chords = make([]chord, len(prog))
	for i, deg := range prog {
		s.chords[i] = buildMinorChord(s.rootMidi, deg)
	}

	// Tempo: 4 to 6 seconds per chord (slow ambient).
	chordSec := 4.0 + 2.0*s.rng.Float64()
	s.chordLen = int64(chordSec * float64(synth.SampleRate))
	s.chordHold = 0 // fire chord 0 immediately on first Next() call
	s.chordIdx = 0

	// 4 beats per chord for the piano arpeggio.
	s.beatLen = s.chordLen / 4
	s.beatHold = 0
	s.beatIdx = 0

	s.activeBass = -1
	s.activePiano = -1
	s.activePad = nil
	s.t = 0

	// Build the synthesizer.
	settings := meltysynth.NewSynthesizerSettings(synth.SampleRate)
	settings.EnableReverbAndChorus = true
	settings.MaximumPolyphony = 64
	syn, err := meltysynth.NewSynthesizer(s.sf, settings)
	if err != nil {
		// Construction errors are unrecoverable. Reset to a safe state and
		// the caller will hear silence — better than panicking on the audio
		// thread.
		s.syn = nil
		return
	}
	s.syn = syn

	// Channel layout:
	//   0 — Acoustic Grand Piano (GM #0)        for melodic arpeggio
	//   1 — String Ensemble 1 (GM #48)          for the chord pad
	//   2 — Warm Pad (GM #89)                   for the lower pad layer
	//   3 — Acoustic Bass (GM #32)              for the bass note
	const ccProgramChange = 0xC0
	s.syn.ProcessMidiMessage(0, ccProgramChange, 0, 0)
	s.syn.ProcessMidiMessage(1, ccProgramChange, 48, 0)
	s.syn.ProcessMidiMessage(2, ccProgramChange, 89, 0)
	s.syn.ProcessMidiMessage(3, ccProgramChange, 32, 0)
}

// buildMinorChord returns a 4-note chord built on the given scale degree
// of the natural minor scale anchored at rootMidi.
//
// For each degree, we use the natural minor diatonic triad with the 7th
// added (root, 3rd, 5th, 7th — picked from the parent scale).
func buildMinorChord(rootMidi, deg int) chord {
	// Natural minor scale intervals: 0, 2, 3, 5, 7, 8, 10.
	root := rootMidi + scaleMinor[deg]
	third := rootMidi + scaleMinor[(deg+2)%7]
	if (deg+2) >= 7 {
		third += 12
	}
	fifth := rootMidi + scaleMinor[(deg+4)%7]
	if (deg+4) >= 7 {
		fifth += 12
	}
	seventh := rootMidi + scaleMinor[(deg+6)%7]
	if (deg+6) >= 7 {
		seventh += 12
	}
	return chord{
		rootDeg: deg,
		tones:   []int{root, third, fifth, seventh},
	}
}

func (s *SF2) Next(left, right []float64) {
	if s.syn == nil {
		// Soundfont didn't initialize — render silence.
		for i := range left {
			left[i] = 0
			right[i] = 0
		}
		return
	}
	n := len(left)
	if cap(s.bufF32L) < n {
		s.bufF32L = make([]float32, n)
		s.bufF32R = make([]float32, n)
	}
	s.bufF32L = s.bufF32L[:n]
	s.bufF32R = s.bufF32R[:n]

	// Walk through the block, firing events at the appropriate sample
	// boundaries and rendering in between.
	pos := 0
	for pos < n {
		// How many samples until the next musical event?
		// (Either a chord change or a beat tick, whichever comes first.)
		ahead := n - pos
		if s.chordHold < int64(ahead) {
			ahead = int(s.chordHold)
		}
		if s.beatHold < int64(ahead) {
			ahead = int(s.beatHold)
		}
		if ahead == 0 {
			// Fire whichever events are due.
			if s.chordHold == 0 {
				s.fireChord()
				s.chordHold = s.chordLen
				s.beatHold = 0
				s.beatIdx = 0
			}
			if s.beatHold == 0 {
				s.fireBeat()
				s.beatHold = s.beatLen
				s.beatIdx++
				if s.beatIdx >= 4 {
					s.beatIdx = 0
				}
			}
			continue
		}
		// Render `ahead` samples.
		s.syn.Render(s.bufF32L[pos:pos+ahead], s.bufF32R[pos:pos+ahead])
		s.chordHold -= int64(ahead)
		s.beatHold -= int64(ahead)
		s.t += int64(ahead)
		pos += ahead
	}

	// Convert float32 → float64, apply soft-clip for safety, and write out.
	// 3.5 master gain brings go-meltysynth's conservative levels up to match
	// our pure-synth algorithms; tanh prevents any peaks from clipping.
	for i := 0; i < n; i++ {
		left[i] = synth.SoftClip(float64(s.bufF32L[i]) * 3.5)
		right[i] = synth.SoftClip(float64(s.bufF32R[i]) * 3.5)
	}
}

// fireChord releases the previous pad/bass notes and gates the new ones.
func (s *SF2) fireChord() {
	// Release previous pad and bass.
	for _, k := range s.activePad {
		s.syn.NoteOff(1, int32(k))
		s.syn.NoteOff(2, int32(k))
	}
	if s.activeBass >= 0 {
		s.syn.NoteOff(3, int32(s.activeBass))
	}

	c := s.chords[s.chordIdx]
	// Pad: play 3rd, 5th, 7th in the +12..+24 register (raised to avoid mud).
	pad := []int{c.tones[1] + 12, c.tones[2] + 12, c.tones[3] + 12}
	s.activePad = pad
	for _, k := range pad {
		s.syn.NoteOn(1, int32(k), 80) // strings
		s.syn.NoteOn(2, int32(k), 64) // warm pad
	}
	// Bass: the root, low.
	bass := c.tones[0]
	s.syn.NoteOn(3, int32(bass), 100)
	s.activeBass = bass

	// Advance to the next chord on the next fire.
	s.chordIdx = (s.chordIdx + 1) % len(s.chords)
}

// fireBeat plays a piano note from the current chord — simple arpeggio
// pattern (root, 3rd, 5th, 3rd) — in a higher register.
func (s *SF2) fireBeat() {
	if s.activePiano >= 0 {
		s.syn.NoteOff(0, int32(s.activePiano))
	}
	// chordIdx has already been incremented by fireChord, so the "current"
	// chord during a beat is at (chordIdx - 1 + len) % len.
	cur := (s.chordIdx - 1 + len(s.chords)) % len(s.chords)
	c := s.chords[cur]
	pattern := []int{0, 1, 2, 1} // chord-tone indices for the 4 beats
	key := c.tones[pattern[s.beatIdx]] + 24 // two octaves up for the piano line
	s.syn.NoteOn(0, int32(key), 96)
	s.activePiano = key
}

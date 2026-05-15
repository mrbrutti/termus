package gen

import (
	"math"
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"

	"github.com/mrbrutti/termus/internal/synth"
)

// Compile-time assertions: gen.SF2 satisfies both gen.Algorithm and
// gen.SF2Reverberator so the --ir flag works against it.
var _ Algorithm = (*SF2)(nil)
var _ SF2Reverberator = (*SF2)(nil)

// SF2 is the hi-fi algorithm: a minor-key A-B chord progression with modal
// interchange (borrowed chords from parallel major), voiced for piano,
// strings, warm pad, acoustic bass, and a sparse flute lead, all rendered
// through a SoundFont via go-meltysynth. The chord-based scheduling, the
// borrowed-chord harmony, and the voice-leading rules on the pad give the
// output a "composed" feel; the sampled timbres give the realism.
//
// Output passes through a small mastering chain: low-shelf warming, a touch
// of high-shelf air, and a soft-knee stereo compressor before final
// soft-clip.
//
// The caller must inject a *meltysynth.SoundFont before calling Seed.
type SF2 struct {
	rng *rand.Rand

	sf  *meltysynth.SoundFont
	syn *meltysynth.Synthesizer

	// Music-theory state.
	rootMidi  int     // MIDI number of the tonic, low register
	chords    []chord // the full A-B progression (8 chords)
	chordIdx  int     // current chord
	chordHold int64
	chordLen  int64

	// Beat-level state for the piano arpeggio (4 beats per chord).
	beatIdx  int
	beatHold int64
	beatLen  int64

	// Melody-level state — flute notes fire every 1..3 beats, sometimes rest.
	melodyHold int64
	melodyKey  int // currently sounding flute key, or -1 if rested

	// Currently sounding pad voices (3 voices: 3rd, 5th, 7th of the chord)
	// stored as absolute MIDI numbers so voice-leading can compute the
	// nearest-pitch move on the next chord.
	padVoices [3]int

	// Single notes per channel (we always replace these on the next fire).
	activeBass  int
	activePiano int

	t int64

	// Effects bus (per-channel; mono signal duplicated for L/R since
	// go-meltysynth already produces stereo output from the synth).
	eqLowL, eqLowR   *synth.LowShelf
	eqHighL, eqHighR *synth.HighShelf
	comp             *synth.StereoCompressor

	// Optional convolution reverb on the master bus.
	convL, convR synth.RealtimeConvolver
	convWet      float64

	// Scratch buffers — go-meltysynth uses []float32, we use []float64.
	bufF32L []float32
	bufF32R []float32
}

// SetReverbIR installs a convolution reverb on the master bus. The IR is
// shared across both channels (one instance per channel, both seeded from
// the same IR). wet is the mix level in [0, 1]; pass nil/empty ir or wet ≤ 0
// to disable.
func (s *SF2) SetReverbIR(ir []float64, wet float64) {
	if len(ir) == 0 || wet <= 0 {
		s.convL = nil
		s.convR = nil
		s.convWet = 0
		return
	}
	if wet > 1 {
		wet = 1
	}
	// Cube-root normalization keeps long, dense IRs perceptually
	// comparable to short ones without going inaudibly quiet.
	var sumSq float64
	for _, x := range ir {
		sumSq += x * x
	}
	norm := 1.0
	if sumSq > 0 {
		norm = math.Pow(1.0/sumSq, 1.0/3.0)
	}
	scaled := make([]float64, len(ir))
	for i, x := range ir {
		scaled[i] = x * norm
	}
	// Short IRs use direct convolution (zero latency); long IRs use the
	// FFT-based partitioned convolver. Threshold and block size match
	// sf2Core for consistency.
	const fftThreshold = 1024
	const fftBlockSize = 512
	if len(scaled) <= fftThreshold {
		s.convL = synth.NewConvolver(scaled)
		s.convR = synth.NewConvolver(scaled)
	} else {
		s.convL = synth.NewFFTConvolver(scaled, fftBlockSize)
		s.convR = synth.NewFFTConvolver(scaled, fftBlockSize)
	}
	s.convWet = wet
}

// chord is a set of MIDI note numbers in canonical root-3rd-5th-7th order.
type chord struct {
	tones []int
}

// chordQual selects the chord quality used when a progression specifies a
// scale-degree root. The qualities cover everything we need for minor-key
// material with modal interchange from parallel major.
type chordQual int

const (
	qualMinor      chordQual = iota // 1, b3, 5, b7 — the diatonic minor i/iv/v
	qualMajor                       // 1, 3, 5, b7 — diatonic III/VI/VII; or borrowed I/IV/V
	qualMinor7                      // 1, b3, 5, b7
	qualMajor7                      // 1, 3, 5, 7
	qualDominant7                   // 1, 3, 5, b7 — borrowed V from harmonic minor
)

// progressionStep names one chord by its position relative to the tonic, plus
// the quality. Degrees are 0..6 of natural minor: 0=i, 1=ii°, 2=III, 3=iv,
// 4=v, 5=VI, 6=VII. Quality lets us "borrow" from parallel major (e.g. major V
// instead of minor v, major IV instead of minor iv).
type progressionStep struct {
	degree int
	qual   chordQual
}

// progressions: each is an 8-chord A-B form. The A section is purely
// diatonic minor; the B section introduces a borrowed chord for color. These
// were hand-picked to feel coherent rather than randomly generated, since the
// "wrong" progression is much more noticeable than a "wrong" random note.
var progressions = [][]progressionStep{
	// 1. Andalusian (i-VII-VI-V) with a borrowed dominant V in the B section.
	{
		{0, qualMinor}, {6, qualMajor}, {5, qualMajor}, {4, qualMinor},
		{0, qualMinor}, {6, qualMajor}, {5, qualMajor}, {4, qualDominant7},
	},
	// 2. i-III-VI-VII / iv-i-V-i with a borrowed V at the end.
	{
		{0, qualMinor}, {2, qualMajor}, {5, qualMajor}, {6, qualMajor},
		{3, qualMinor}, {0, qualMinor}, {4, qualDominant7}, {0, qualMinor},
	},
	// 3. i-iv-VII-III / VI-bII-V-i  (Phrygian flavor in the B section).
	{
		{0, qualMinor}, {3, qualMinor}, {6, qualMajor}, {2, qualMajor},
		{5, qualMajor}, {3, qualMinor}, {4, qualDominant7}, {0, qualMinor},
	},
	// 4. i-VI-III-VII (A) / IV-i-VII-i  (IV is borrowed from parallel major).
	{
		{0, qualMinor}, {5, qualMajor}, {2, qualMajor}, {6, qualMajor},
		{3, qualMajor}, {0, qualMinor}, {6, qualMajor}, {0, qualMinor},
	},
}

// NewSF2 constructs the algorithm bound to the given soundfont.
// Caller must call Seed before Next.
func NewSF2(sf *meltysynth.SoundFont) *SF2 {
	return &SF2{sf: sf}
}

func (s *SF2) Name() string { return "sf2-progression" }

func (s *SF2) Seed(seedVal int64) {
	s.rng = rand.New(rand.NewSource(seedVal)) //nolint:gosec
	s.rootMidi = 33 + s.rng.Intn(8) // A1..F2

	// Pick a progression.
	prog := progressions[s.rng.Intn(len(progressions))]
	s.chords = make([]chord, len(prog))
	for i, step := range prog {
		s.chords[i] = buildChordAt(s.rootMidi, step.degree, step.qual)
	}

	// Tempo: 3.5..5.5 seconds per chord (slow ambient).
	chordSec := 3.5 + 2.0*s.rng.Float64()
	s.chordLen = int64(chordSec * float64(synth.SampleRate))
	s.beatLen = s.chordLen / 4

	s.chordHold = 0
	s.beatHold = 0
	s.melodyHold = 0
	s.chordIdx = 0
	s.beatIdx = 0

	s.activeBass = -1
	s.activePiano = -1
	s.melodyKey = -1
	s.padVoices = [3]int{-1, -1, -1}
	s.t = 0

	// Effects bus.
	s.eqLowL = synth.NewLowShelf(180, 2.5, 0.707)
	s.eqLowR = synth.NewLowShelf(180, 2.5, 0.707)
	s.eqHighL = synth.NewHighShelf(7500, 3.0, 0.707)
	s.eqHighR = synth.NewHighShelf(7500, 3.0, 0.707)
	// Master bus compressor: threshold -14 dB, 3:1, fairly fast attack to
	// catch piano transients, slow release for ambient breathing.
	s.comp = synth.NewStereoCompressor(-14, 3.0, 8, 250, 6, 4)

	// Build the synthesizer.
	settings := meltysynth.NewSynthesizerSettings(synth.SampleRate)
	settings.EnableReverbAndChorus = true
	settings.MaximumPolyphony = 96
	syn, err := meltysynth.NewSynthesizer(s.sf, settings)
	if err != nil {
		s.syn = nil
		return
	}
	s.syn = syn

	// Channel layout:
	//   0 — Acoustic Grand Piano (GM #0)     piano arpeggio
	//   1 — String Ensemble 1   (GM #48)     chord pad
	//   2 — Warm Pad            (GM #89)     lower pad layer
	//   3 — Acoustic Bass       (GM #32)     bass note
	//   4 — Flute               (GM #73)     melody lead
	const ccProgramChange = 0xC0
	const ccControlChange = 0xB0
	const ccBrightness = 74
	s.syn.ProcessMidiMessage(0, ccProgramChange, 0, 0)
	s.syn.ProcessMidiMessage(1, ccProgramChange, 48, 0)
	s.syn.ProcessMidiMessage(2, ccProgramChange, 89, 0)
	s.syn.ProcessMidiMessage(3, ccProgramChange, 32, 0)
	s.syn.ProcessMidiMessage(4, ccProgramChange, 73, 0)
	// Per-channel base cutoffs (CC 74) — same darkening pattern used by
	// other SF2 algorithms. Piano + strings + flute kept bright for clarity;
	// warm pad slightly darkened; bass moderate.
	s.syn.ProcessMidiMessage(0, ccControlChange, ccBrightness, 80) // piano
	s.syn.ProcessMidiMessage(1, ccControlChange, ccBrightness, 76) // strings
	s.syn.ProcessMidiMessage(2, ccControlChange, ccBrightness, 64) // warm pad
	s.syn.ProcessMidiMessage(3, ccControlChange, ccBrightness, 64) // bass
	s.syn.ProcessMidiMessage(4, ccControlChange, ccBrightness, 96) // flute melody
}

// buildChordAt returns a 4-note chord rooted on the given scale degree of
// natural minor (relative to rootMidi), with the requested quality. The
// chord tones are returned in absolute MIDI numbers in the bass register
// (caller raises octaves as needed).
func buildChordAt(rootMidi, deg int, q chordQual) chord {
	chordRoot := rootMidi + scaleMinor[deg]
	// Intervals (in semitones) from the chord root for each quality.
	var intervals [4]int
	switch q {
	case qualMinor, qualMinor7:
		intervals = [4]int{0, 3, 7, 10}
	case qualMajor:
		intervals = [4]int{0, 4, 7, 10}
	case qualMajor7:
		intervals = [4]int{0, 4, 7, 11}
	case qualDominant7:
		intervals = [4]int{0, 4, 7, 10}
	default:
		intervals = [4]int{0, 3, 7, 10}
	}
	return chord{
		tones: []int{
			chordRoot + intervals[0],
			chordRoot + intervals[1],
			chordRoot + intervals[2],
			chordRoot + intervals[3],
		},
	}
}

func (s *SF2) Next(left, right []float64) {
	if s.syn == nil {
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

	pos := 0
	for pos < n {
		ahead := n - pos
		if s.chordHold < int64(ahead) {
			ahead = int(s.chordHold)
		}
		if s.beatHold < int64(ahead) {
			ahead = int(s.beatHold)
		}
		if s.melodyHold < int64(ahead) {
			ahead = int(s.melodyHold)
		}
		if ahead == 0 {
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
			if s.melodyHold == 0 {
				s.fireMelody()
			}
			continue
		}
		s.syn.Render(s.bufF32L[pos:pos+ahead], s.bufF32R[pos:pos+ahead])
		s.chordHold -= int64(ahead)
		s.beatHold -= int64(ahead)
		s.melodyHold -= int64(ahead)
		s.t += int64(ahead)
		pos += ahead
	}

	// Master bus: gain → EQ → optional convolution wet bus → compressor → soft-clip.
	for i := 0; i < n; i++ {
		l := float64(s.bufF32L[i]) * 3.5
		r := float64(s.bufF32R[i]) * 3.5
		l = s.eqLowL.Tick(l)
		r = s.eqLowR.Tick(r)
		l = s.eqHighL.Tick(l)
		r = s.eqHighR.Tick(r)
		if s.convL != nil {
			wetL := s.convL.Tick(l)
			wetR := s.convR.Tick(r)
			l += wetL * s.convWet
			r += wetR * s.convWet
		}
		l, r = s.comp.Tick(l, r)
		left[i] = synth.SoftClip(l)
		right[i] = synth.SoftClip(r)
	}
}

// fireChord advances to the next chord, applies voice-leading to the pad
// voices, and re-gates the bass.
func (s *SF2) fireChord() {
	c := s.chords[s.chordIdx]

	// Release the previous pad voices.
	for _, k := range s.padVoices {
		if k >= 0 {
			s.syn.NoteOff(1, int32(k))
			s.syn.NoteOff(2, int32(k))
		}
	}
	// Release the previous bass.
	if s.activeBass >= 0 {
		s.syn.NoteOff(3, int32(s.activeBass))
	}

	// Voice-leading for the pad: candidates for the new chord are the 3rd,
	// 5th, and 7th of the chord, in the +12..+24 register from the chord
	// root. For each previous voice, pick the closest unused candidate.
	candidates := []int{
		c.tones[1] + 12,
		c.tones[2] + 12,
		c.tones[3] + 12,
	}
	used := [3]bool{}
	var newVoices [3]int
	for v := 0; v < 3; v++ {
		// If we have no previous voice (initial state), just take candidate v.
		if s.padVoices[v] < 0 {
			newVoices[v] = candidates[v]
			used[v] = true
			continue
		}
		// Find unused candidate closest in pitch to the previous voice.
		best := -1
		bestDist := 1 << 30
		for j, cand := range candidates {
			if used[j] {
				continue
			}
			d := cand - s.padVoices[v]
			if d < 0 {
				d = -d
			}
			if d < bestDist {
				bestDist = d
				best = j
			}
		}
		newVoices[v] = candidates[best]
		used[best] = true
	}

	for _, k := range newVoices {
		s.syn.NoteOn(1, int32(k), 78) // strings
		s.syn.NoteOn(2, int32(k), 62) // warm pad
	}
	s.padVoices = newVoices

	// Bass: chord root.
	bass := c.tones[0]
	s.syn.NoteOn(3, int32(bass), 100)
	s.activeBass = bass

	s.chordIdx = (s.chordIdx + 1) % len(s.chords)
}

// fireBeat plays a piano note from the current chord using a fixed arpeggio
// pattern. The "current" chord is chordIdx-1 because chordIdx was advanced.
func (s *SF2) fireBeat() {
	if s.activePiano >= 0 {
		s.syn.NoteOff(0, int32(s.activePiano))
	}
	cur := (s.chordIdx - 1 + len(s.chords)) % len(s.chords)
	c := s.chords[cur]
	pattern := []int{0, 1, 2, 1} // chord-tone indices for the 4 beats
	key := c.tones[pattern[s.beatIdx]] + 24
	s.syn.NoteOn(0, int32(key), 92)
	s.activePiano = key
}

// fireMelody schedules the flute lead: a sparse, slow-moving melodic line
// drawn from the current chord tones (with occasional non-chord passing
// tones from the parent minor scale). Each note lasts 1..3 beats, with a
// ~30% chance of resting on the next slot.
func (s *SF2) fireMelody() {
	if s.melodyKey >= 0 {
		s.syn.NoteOff(4, int32(s.melodyKey))
		s.melodyKey = -1
	}
	// Decide whether to play or rest.
	if s.rng.Float64() < 0.30 {
		s.melodyHold = s.beatLen * int64(1+s.rng.Intn(2))
		return
	}
	cur := (s.chordIdx - 1 + len(s.chords)) % len(s.chords)
	c := s.chords[cur]
	// Pick a note: 70% chord tone, 30% scale tone.
	var key int
	if s.rng.Float64() < 0.70 {
		// Chord tone: one of root/3rd/5th/7th, in the +36..+48 register.
		idx := s.rng.Intn(4)
		key = c.tones[idx] + 36
	} else {
		// Scale tone: a degree of the parent natural minor, in the same
		// register as the chord-tone choice. Bias toward degrees near a
		// chord tone for stepwise motion.
		deg := s.rng.Intn(7)
		key = s.rootMidi + scaleMinor[deg] + 48
	}
	s.syn.NoteOn(4, int32(key), 84)
	s.melodyKey = key
	s.melodyHold = s.beatLen * int64(1+s.rng.Intn(3)) // 1..3 beats
}

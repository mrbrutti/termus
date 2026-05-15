package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

var _ Algorithm = (*SF2Glass)(nil)
var _ SF2Reverberator = (*SF2Glass)(nil)

// SF2Glass is the bells algorithm — Boards of Canada / Music for Airports
// style sparse bell phrases. Tubular bells lead the texture; celesta and
// glockenspiel add upper-register sparkle; a music-box layer adds occasional
// nostalgic ornaments; a soft choir/strings pad and sub-bass pedal hold the
// harmony underneath.
//
// All melodic content is pentatonic so notes never clash, and trigger rates
// are long + incommensurate so the bell phrases never quite repeat.
//
// Preferred SF: fairy-tale (celesta, music box, tubular bells, glockenspiel
// all in one bank).
type SF2Glass struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
	rng  *rand.Rand

	rootMidi  int
	keyOffset int

	// Pentatonic scale used by every melodic track. Major or minor pentatonic
	// is seed-chosen.
	scale []int

	// 2 chord centers cycled slowly — same pentatonic but transposed up a 5th
	// or down a 4th for a "second key area" feeling.
	chordOffsets    []int
	currentChordIdx int

	samplesElapsed int64
	nextChordAt    int64
	nextSectionAt  int64
	musicBoxOn     *bool

	bellPhrase    []int
	celestaPhrase []int
}

// majorPentatonic: 0, 2, 4, 7, 9 (the "happy" pentatonic).
// minorPentatonic: 0, 3, 5, 7, 10 (the "sad" pentatonic).
var (
	majorPentatonic = []int{0, 2, 4, 7, 9}
	minorPentatonic = []int{0, 3, 5, 7, 10}
)

func NewSF2Glass(sf *meltysynth.SoundFont) *SF2Glass { return &SF2Glass{sf: sf} }

func (a *SF2Glass) Name() string { return "bells" }

func (a *SF2Glass) currentRoot() int { return a.rootMidi + a.keyOffset }

func (a *SF2Glass) Seed(seedVal int64) {
	a.rng = rand.New(rand.NewSource(seedVal)) //nolint:gosec
	a.rootMidi = 48 + a.rng.Intn(7) // C3..F#3
	a.keyOffset = 0
	if a.rng.Float64() < 0.55 {
		a.scale = majorPentatonic
	} else {
		a.scale = minorPentatonic
	}
	// Two harmonic centers — tonic and either +7 (dominant) or +5 (subdom).
	if a.rng.Float64() < 0.5 {
		a.chordOffsets = []int{0, 7}
	} else {
		a.chordOffsets = []int{0, 5}
	}
	a.currentChordIdx = 0
	a.samplesElapsed = 0
	a.scheduleNextChord()

	musicBoxStart := true
	a.musicBoxOn = &musicBoxStart
	a.scheduleNextSection()

	// Master gain raised (3.0 → 3.6) — previous output was ~7 dB below the
	// other genres; bells should sit comfortably in the mix.
	core, err := newSF2Core(a.sf, 3.6, seedVal)
	if err != nil {
		a.core = nil
		return
	}

	// Channel layout:
	//   0 — Tubular Bells   (program 14)  main bell motif
	//   1 — Celesta         (program 8)   upper-mid sparkle
	//   2 — Glockenspiel    (program 9)   high register
	//   3 — Music Box       (program 10)  occasional ornament
	//   4 — Warm Pad        (program 89)  soft bed
	//   5 — Choir Aahs      (program 52)  vocal bed
	//   6 — Synth Bass 2    (program 39)  sub-bass pedal
	core.setProgram(0, 14)
	core.setProgram(1, 8)
	core.setProgram(2, 9)
	core.setProgram(3, 10)
	core.setProgram(4, 89)
	core.setProgram(5, 52)
	core.setProgram(6, 39)
	core.setPan(0, 64)
	core.setPan(1, 80)
	core.setPan(2, 96)
	core.setPan(3, 32)
	core.setPan(4, 72)
	core.setPan(5, 56)
	core.setPan(6, 64)

	// Bells/celesta/glock kept very bright for sparkle; pad darkened.
	core.setChannelCutoff(0, 120)
	core.setChannelCutoff(1, 120)
	core.setChannelCutoff(2, 120)
	core.setChannelCutoff(3, 110)
	core.setChannelCutoff(4, 60)
	core.setChannelCutoff(5, 76)
	core.setChannelCutoff(6, 50)

	// Slow pad LFO so the bed breathes underneath the bells.
	core.addFilterLFO(4, 1.0/16.0, 60, 24)
	core.addFilterLFO(5, 1.0/23.0, 72, 20)

	// Reverb: everyone in halo except bass.
	core.setReverbSend(0, 120)
	core.setReverbSend(1, 120)
	core.setReverbSend(2, 120)
	core.setReverbSend(3, 110)
	core.setReverbSend(4, 96)
	core.setReverbSend(5, 100)
	core.setReverbSend(6, 30)
	core.setChorusSend(4, 32)
	core.setChorusSend(5, 28)

	// --- Tubular bells: a single coherent 8-note pentatonic phrase, looped
	// on a long period. Listeners recognize a melody, not random strikes.
	a.bellPhrase = a.makePhraseInBellRegister(0)
	core.addTrack(SF2Track{
		Channel: 0, Velocity: 72, Notes: a.bellPhrase,
		PeriodSec: 21.7, Phase01: a.rng.Float64(),
		MutationRate: 0.10,
		MutateOne: func(slot int, _ int) int {
			return a.bellPhrase[slot%len(a.bellPhrase)]
		},
		VelocityJitter: 16, TimingJitterSec: 0.10,
	})

	// --- Celesta counter-phrase: a second coherent phrase one octave up,
	// on a different period — answers the tubular bell motif.
	a.celestaPhrase = a.makePhraseInBellRegister(12)
	core.addTrack(SF2Track{
		Channel: 1, Velocity: 56, Notes: a.celestaPhrase,
		PeriodSec: 29.7, Phase01: a.rng.Float64(),
		MutationRate: 0.12,
		MutateOne: func(slot int, _ int) int {
			return a.celestaPhrase[slot%len(a.celestaPhrase)]
		},
		VelocityJitter: 14, TimingJitterSec: 0.10,
	})

	// --- Glockenspiel: 1 voice, very high, very sparse — only one note per
	// long period, just the chord's brightest tone.
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 48, Notes: []int{a.bellNote(0, 24)},
		PeriodSec: 37.3, Phase01: a.rng.Float64(),
		MutationRate: 0.40,
		MutateOne:    func(_ int, _ int) int { return a.bellNote(0, 24) },
		VelocityJitter: 12, TimingJitterSec: 0.12,
		Enabled: a.musicBoxOn,
	})

	// --- Warm pad bed: 2 voices spread, sustained, slow retrigger.
	for ti, period := range []float64{31.3, 43.7} {
		voice := ti
		core.addTrack(SF2Track{
			Channel: 4, Velocity: 44, Notes: []int{a.padNote(voice)},
			PeriodSec: period, Phase01: a.rng.Float64(),
			MutationRate: 0.30,
			MutateOne:    func(_ int, _ int) int { return a.padNote(voice) },
			VelocityJitter: 6, TimingJitterSec: 0.08,
		})
	}

	// --- Choir aahs: 1 voice in upper register, very slow.
	core.addTrack(SF2Track{
		Channel: 5, Velocity: 40, Notes: []int{a.padNote(1) + 12},
		PeriodSec: 53.9, Phase01: a.rng.Float64(),
		MutationRate: 0.35,
		MutateOne:    func(_ int, _ int) int { return a.padNote(1) + 12 },
		VelocityJitter: 6, TimingJitterSec: 0.10,
	})

	// --- Sub-bass pedal: holds the chord root.
	core.addTrack(SF2Track{
		Channel: 6, Velocity: 60, Notes: []int{a.bassRoot()},
		PeriodSec: 51.3, Phase01: 0,
		MutationRate: 0.50,
		MutateOne:    func(_ int, _ int) int { return a.bassRoot() },
		VelocityJitter: 4, TimingJitterSec: 0.05,
	})

	a.core = core
}

// makePhraseInBellRegister builds an 8-note coherent melodic phrase in the
// chosen pentatonic scale, placed into the bell register with optional
// octave-bump offset. The shape is one of melodicPhrases — a coherent
// contour beats per-note randomness for any motivic instrument.
func (a *SF2Glass) makePhraseInBellRegister(octaveBump int) []int {
	contour := pickMelodicPhrase(a.rng)
	chordOff := a.chordOffsets[a.currentChordIdx]
	root := a.currentRoot() + chordOff + 12 + octaveBump
	notes := applyPhraseToScale(contour, a.scale, root, 0, 0)
	for i, k := range notes {
		for k < 60 {
			k += 12
		}
		for k > 96 {
			k -= 12
		}
		notes[i] = k
	}
	return notes
}

// bellNote returns a pentatonic-scale MIDI key for a bell voice. voice picks
// which scale degree (cycled), bumpSemis adds an octave-shift offset for
// register placement.
func (a *SF2Glass) bellNote(voice, bumpSemis int) int {
	deg := a.scale[a.rng.Intn(len(a.scale))]
	chordOff := a.chordOffsets[a.currentChordIdx]
	key := a.currentRoot() + chordOff + deg + 12 + bumpSemis
	// Spread voices across octaves.
	key += 12 * voice
	for key < 60 {
		key += 12
	}
	for key > 96 {
		key -= 12
	}
	return key
}

// padNote returns a pentatonic chord-tone key in the pad register (around
// C4–C5).
func (a *SF2Glass) padNote(voice int) int {
	deg := a.scale[voice%len(a.scale)]
	chordOff := a.chordOffsets[a.currentChordIdx]
	key := a.currentRoot() + chordOff + deg + 12
	for key < 60 {
		key += 12
	}
	for key > 76 {
		key -= 12
	}
	return key
}

// bassRoot returns the chord root in the bass register.
func (a *SF2Glass) bassRoot() int {
	chordOff := a.chordOffsets[a.currentChordIdx]
	key := a.currentRoot() + chordOff
	for key > 48 {
		key -= 12
	}
	for key < 30 {
		key += 12
	}
	return key
}

func (a *SF2Glass) scheduleNextChord() {
	// 40-70 s per chord — slow but noticeable harmonic shifts.
	secs := 40.0 + 30.0*a.rng.Float64()
	a.nextChordAt = a.samplesElapsed + int64(secs*44100)
}

func (a *SF2Glass) scheduleNextSection() {
	secs := 90.0 + 90.0*a.rng.Float64()
	a.nextSectionAt = a.samplesElapsed + int64(secs*44100)
}

func (a *SF2Glass) advance() {
	if a.samplesElapsed >= a.nextChordAt {
		a.currentChordIdx = (a.currentChordIdx + 1) % len(a.chordOffsets)
		a.scheduleNextChord()
	}
	if a.samplesElapsed >= a.nextSectionAt {
		*a.musicBoxOn = !*a.musicBoxOn
		a.scheduleNextSection()
	}
}

func (a *SF2Glass) SetReverbIR(ir []float64, wet float64) {
	if a.core != nil {
		a.core.setConvolutionIR(ir, wet)
	}
}

func (a *SF2Glass) Next(left, right []float64) {
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

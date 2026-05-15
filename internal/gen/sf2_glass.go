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

	bellContour         []int
	bellStartDegree     int
	celestaContour      []int
	celestaStartDegree  int
	musicBoxContour     []int
	musicBoxStartDegree int
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
	a.rootMidi = 48 + a.rng.Intn(7)           // C3..F#3
	a.keyOffset = 0
	// Boards of Canada "Roygbiv" canonical recipe: major pentatonic only,
	// avoiding 4th and 7th (which sound un-bell-like). 60% major pent for
	// the Roygbiv brightness, 40% minor pent for the moodier BoC tracks.
	if a.rng.Float64() < 0.60 {
		a.scale = majorPentatonic
	} else {
		a.scale = minorPentatonic
	}
	// Roygbiv progression: I-IV-I-V (4 chords per cycle). The 4th-chord cycle
	// is what makes Boards of Canada read as bells rather than abstract
	// pentatonic noodling.
	a.chordOffsets = []int{0, 5, 0, 7}
	a.currentChordIdx = 0
	a.samplesElapsed = 0
	a.bellContour = append([]int(nil), pickMelodicPhrase(a.rng)...)
	a.bellStartDegree = a.rng.Intn(2)
	a.celestaContour = append([]int(nil), pickMelodicPhrase(a.rng)...)
	a.celestaStartDegree = 1 + a.rng.Intn(2)
	a.musicBoxContour = append([]int(nil), melodicPhrases[5][:4]...)
	a.musicBoxStartDegree = a.rng.Intn(3)
	a.scheduleNextChord()

	musicBoxStart := true
	a.musicBoxOn = &musicBoxStart
	a.scheduleNextSection()

	// Master gain raised aggressively (3.0 → 4.2) — bell content is sparse
	// by nature so the long silences pull the RMS down even with reasonable
	// peaks. Need significantly more gain than a denser genre would.
	core, err := newSF2Core(a.sf, 4.2, seedVal)
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

	// --- Tubular bells: a coherent pentatonic phrase resolved against the
	// current chord at fire time, so the motif survives while the harmony moves.
	bellSlots := make([]int, len(a.bellContour))
	core.addTrack(SF2Track{
		Channel: 0, Velocity: 72, Notes: bellSlots,
		PeriodSec: 21.7, Phase01: a.rng.Float64(),
		ResolveNote:    func(slot int, _ int) int { return a.phraseNoteAt(slot, a.bellContour, a.bellStartDegree, 12, 60, 96) },
		VelocityJitter: 16, TimingJitterSec: 0.10,
	})

	// --- Celesta counter-phrase: a second coherent phrase one octave up,
	// on a different period — answers the tubular bell motif.
	celestaSlots := make([]int, len(a.celestaContour))
	core.addTrack(SF2Track{
		Channel: 1, Velocity: 56, Notes: celestaSlots,
		PeriodSec: 29.7, Phase01: a.rng.Float64(),
		ResolveNote: func(slot int, _ int) int {
			return a.phraseNoteAt(slot, a.celestaContour, a.celestaStartDegree, 24, 72, 96)
		},
		VelocityJitter: 14, TimingJitterSec: 0.10,
	})

	// --- Glockenspiel: 1 voice, very high, very sparse — only one note per
	// long period, just the chord's brightest tone.
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 48, Notes: []int{a.bellNote(0, 24)},
		PeriodSec: 37.3, Phase01: a.rng.Float64(),
		MutationRate:   0.40,
		MutateOne:      func(_ int, _ int) int { return a.bellNote(0, 24) },
		ResolveNote:    func(_ int, _ int) int { return a.bellNote(0, 24) },
		VelocityJitter: 12, TimingJitterSec: 0.12,
	})

	// --- Music box ornament: nostalgic answering fragment that drops in/out
	// by section. Uses explicit rests so the layer feels like a phrase rather
	// than a constant ostinato.
	musicBoxSlots := make([]int, len(a.musicBoxContour)+2)
	core.addTrack(SF2Track{
		Channel: 3, Velocity: 44, Notes: musicBoxSlots,
		PeriodSec: 43.1, Phase01: a.rng.Float64(),
		ResolveNote: func(slot int, _ int) int {
			if slot%len(musicBoxSlots) >= len(a.musicBoxContour) {
				return -1
			}
			return a.phraseNoteAt(slot, a.musicBoxContour, a.musicBoxStartDegree, 12, 72, 92)
		},
		VelocityJitter: 12, TimingJitterSec: 0.12,
		Enabled: a.musicBoxOn,
	})

	// --- Warm pad bed: 2 voices spread, sustained, slow retrigger.
	for ti, period := range []float64{31.3, 43.7} {
		voice := ti
		core.addTrack(SF2Track{
			Channel: 4, Velocity: 44, Notes: []int{a.padNote(voice)},
			PeriodSec: period, Phase01: a.rng.Float64(),
			MutationRate:   0.30,
			MutateOne:      func(_ int, _ int) int { return a.padNote(voice) },
			ResolveNote:    func(_ int, _ int) int { return a.padNote(voice) },
			VelocityJitter: 6, TimingJitterSec: 0.08,
		})
	}

	// --- Choir aahs: 1 voice in upper register, very slow.
	core.addTrack(SF2Track{
		Channel: 5, Velocity: 40, Notes: []int{a.padNote(1) + 12},
		PeriodSec: 53.9, Phase01: a.rng.Float64(),
		MutationRate:   0.35,
		MutateOne:      func(_ int, _ int) int { return a.padNote(1) + 12 },
		ResolveNote:    func(_ int, _ int) int { return a.padNote(1) + 12 },
		VelocityJitter: 6, TimingJitterSec: 0.10,
	})

	// --- Sub-bass pedal: holds the chord root.
	core.addTrack(SF2Track{
		Channel: 6, Velocity: 60, Notes: []int{a.bassRoot()},
		PeriodSec: 51.3, Phase01: 0,
		MutationRate:   0.50,
		MutateOne:      func(_ int, _ int) int { return a.bassRoot() },
		ResolveNote:    func(_ int, _ int) int { return a.bassRoot() },
		VelocityJitter: 4, TimingJitterSec: 0.05,
	})

	a.core = core
}

func (a *SF2Glass) phraseNoteAt(slot int, contour []int, startDegree, octaveBump, low, high int) int {
	chordOff := a.chordOffsets[a.currentChordIdx]
	root := a.currentRoot() + chordOff + 12 + octaveBump
	key := scaleNoteAt(contour, slot, a.scale, root, startDegree, 0)
	return clampMidiToRange(key, low, high)
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

package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

var _ Algorithm = (*SF2Drone)(nil)
var _ SF2Reverberator = (*SF2Drone)(nil)

// SF2Drone is a Stars-of-the-Lid / William Basinski style algorithm: long,
// slow, evolving textures with very minimal harmonic motion. There's no
// rhythm and no melodic line — just sustained voices that swell, drift, and
// gradually shift their relationship to each other.
//
//   - A sub-bass pedal that holds the chord root for the entire chord cycle
//     (90–150 s before the chord changes).
//   - Two bowed-glass / synth-string layers in the middle register, on long
//     incommensurate periods (37 s, 53 s, 71 s) so swells overlap unevenly.
//   - A choir aahs layer in the upper register, sparse.
//   - A "shimmer" FM-EP overtone layer that catches the upper partials.
//   - Filter LFOs on each pad at very slow rates (15–35 s per cycle) so the
//     texture breathes.
//
// Harmony moves through 3 chord centers built from quartal voicings (stacked
// 4ths) — broader and more "open" than triadic voicings.
//
// Preferred SF: fm-dx (DX-style EPs / metallic bells + sustained FM textures).
type SF2Drone struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
	rng  *rand.Rand

	rootMidi  int
	keyOffset int

	chords          []droneChord
	currentChordIdx int

	samplesElapsed int64
	nextChordAt    int64
}

// droneChord is one harmonic center as a set of semitone offsets from the
// key center. Voicings are quartal (stacked 4ths) for the broad, open feel.
type droneChord struct {
	tones []int
	label string
}

// droneCycles: 3 chord centers each, quartal-voiced. Slow modal drift.
var droneCycles = [][]droneChord{
	// Quartal triad on tonic / on 4th / on 5th — classic minimalist drift.
	{
		{tones: []int{0, 5, 10, 14}, label: "Q(i)"},
		{tones: []int{5, 10, 14, 19}, label: "Q(iv)"},
		{tones: []int{7, 12, 16, 21}, label: "Q(v)"},
	},
	// Modal: tonic / Mixolydian VII / sub-mediant
	{
		{tones: []int{0, 4, 7, 11}, label: "Imaj7"},
		{tones: []int{10, 14, 17, 21}, label: "bVII"},
		{tones: []int{8, 12, 15, 19}, label: "VImaj"},
	},
	// Two-chord cycle (very Stars-of-the-Lid): i / bVI alternating.
	{
		{tones: []int{0, 3, 7, 10}, label: "i"},
		{tones: []int{8, 12, 15, 19}, label: "bVI"},
	},
}

func NewSF2Drone(sf *meltysynth.SoundFont) *SF2Drone { return &SF2Drone{sf: sf} }

func (a *SF2Drone) Name() string { return "drone" }

func (a *SF2Drone) currentRoot() int { return a.rootMidi + a.keyOffset }

func (a *SF2Drone) Seed(seedVal int64) {
	a.rng = rand.New(rand.NewSource(seedVal)) //nolint:gosec
	a.rootMidi = 30 + a.rng.Intn(7) // Bb1..E2 — very low pedal
	a.keyOffset = 0
	a.samplesElapsed = 0
	a.currentChordIdx = 0
	a.chords = droneCycles[a.rng.Intn(len(droneCycles))]
	a.scheduleNextChord()

	core, err := newSF2Core(a.sf, 3.6, seedVal)
	if err != nil {
		a.core = nil
		return
	}

	// Channel layout:
	//   0 — Bowed Glass     (program 92)  primary drone bed (mid)
	//   1 — Synth Strings 1 (program 50)  pad layer (mid)
	//   2 — Choir Aahs      (program 52)  vocal bed (upper)
	//   3 — Electric Piano 1(program 4)   FM shimmer overtones
	//   4 — Synth Bass 1    (program 38)  sub-bass pedal
	core.setProgram(0, 92)
	core.setProgram(1, 50)
	core.setProgram(2, 52)
	core.setProgram(3, 4)
	core.setProgram(4, 38)
	core.setPan(0, 48)
	core.setPan(1, 80)
	core.setPan(2, 64)
	core.setPan(3, 96)
	core.setPan(4, 64)

	// All darkened — drone aesthetic is "muted, foggy" not bright.
	core.setChannelCutoff(0, 70)
	core.setChannelCutoff(1, 64)
	core.setChannelCutoff(2, 76)
	core.setChannelCutoff(3, 88) // FM shimmer brighter than the others
	core.setChannelCutoff(4, 50)

	// Very slow filter LFOs — 22 s, 28 s, 37 s. Different on each so they
	// never sync up and the texture has a constantly-moving spectral profile.
	core.addFilterLFO(0, 1.0/22.0, 70, 24)
	core.addFilterLFO(1, 1.0/28.0, 60, 28)
	core.addFilterLFO(2, 1.0/37.0, 72, 22)

	// Massive reverb sends — drones live in the reverb.
	core.setReverbSend(0, 120)
	core.setReverbSend(1, 110)
	core.setReverbSend(2, 120)
	core.setReverbSend(3, 100)
	core.setReverbSend(4, 40) // bass stays present
	core.setChorusSend(0, 56)
	core.setChorusSend(1, 48)
	core.setChorusSend(2, 32)

	// --- Bowed glass drone bed: 2 voices on long incommensurate periods.
	// Fewer voices, longer periods = more space + clearer harmonic identity.
	for ti, period := range []float64{53.3, 79.1} {
		voice := ti
		core.addTrack(SF2Track{
			Channel: 0, Velocity: 52, Notes: []int{a.droneTone(voice, 0)},
			PeriodSec: period, Phase01: a.rng.Float64(),
			MutationRate: 0.20,
			MutateOne:    func(_ int, _ int) int { return a.droneTone(voice, 0) },
			VelocityJitter: 8, TimingJitterSec: 0.15,
		})
	}

	// --- Synth strings parallel layer: 2 voices, slightly higher register.
	for ti, period := range []float64{61.7, 89.3} {
		voice := ti
		core.addTrack(SF2Track{
			Channel: 1, Velocity: 46, Notes: []int{a.droneTone(voice, 12)},
			PeriodSec: period, Phase01: a.rng.Float64(),
			MutationRate: 0.20,
			MutateOne:    func(_ int, _ int) int { return a.droneTone(voice, 12) },
			VelocityJitter: 6, TimingJitterSec: 0.18,
		})
	}

	// --- Choir aahs: 1 voice in the upper register, very sparse.
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 44, Notes: []int{a.droneTone(1, 24)},
		PeriodSec: 71.7, Phase01: a.rng.Float64(),
		MutationRate: 0.30,
		MutateOne:    func(_ int, _ int) int { return a.droneTone(1, 24) },
		VelocityJitter: 8, TimingJitterSec: 0.20,
	})

	// --- FM EP shimmer: a single high voice that catches upper partials of
	// the chord. Very long period, infrequent retrigger.
	core.addTrack(SF2Track{
		Channel: 3, Velocity: 38, Notes: []int{a.droneTone(2, 36)},
		PeriodSec: 91.1, Phase01: a.rng.Float64(),
		MutationRate: 0.40,
		MutateOne:    func(_ int, _ int) int { return a.droneTone(2, 36) },
		VelocityJitter: 10, TimingJitterSec: 0.25,
	})

	// --- Sub-bass pedal: holds the chord root the entire chord cycle.
	core.addTrack(SF2Track{
		Channel: 4, Velocity: 56, Notes: []int{a.bassRoot()},
		PeriodSec: 60.0, Phase01: 0,
		MutationRate: 0.60,
		MutateOne:    func(_ int, _ int) int { return a.bassRoot() },
		VelocityJitter: 4, TimingJitterSec: 0.05,
	})

	a.core = core
}

func (a *SF2Drone) droneTone(voice, bumpSemis int) int {
	if len(a.chords) == 0 {
		return 60
	}
	c := a.chords[a.currentChordIdx]
	idx := voice % len(c.tones)
	key := a.currentRoot() + c.tones[idx] + 24 + bumpSemis
	for key < 36 {
		key += 12
	}
	for key > 96 {
		key -= 12
	}
	return key
}

func (a *SF2Drone) bassRoot() int {
	if len(a.chords) == 0 {
		return 36
	}
	c := a.chords[a.currentChordIdx]
	key := a.currentRoot() + c.tones[0]
	for key > 42 {
		key -= 12
	}
	for key < 24 {
		key += 12
	}
	return key
}

func (a *SF2Drone) scheduleNextChord() {
	// 90–150 s per chord — even slower than ambient.
	secs := 90.0 + 60.0*a.rng.Float64()
	a.nextChordAt = a.samplesElapsed + int64(secs*44100)
}

func (a *SF2Drone) advance() {
	if a.samplesElapsed >= a.nextChordAt {
		a.currentChordIdx = (a.currentChordIdx + 1) % len(a.chords)
		a.scheduleNextChord()
	}
}

func (a *SF2Drone) SetReverbIR(ir []float64, wet float64) {
	if a.core != nil {
		a.core.setConvolutionIR(ir, wet)
	}
}

func (a *SF2Drone) Next(left, right []float64) {
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

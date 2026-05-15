package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

var _ Algorithm = (*Ambient)(nil)
var _ SF2Reverberator = (*Ambient)(nil)

// Ambient is a "Music for Airports" style algorithm. The old version layered
// 5 voices on incommensurate loops but didn't have the canonical bell-voice
// motif or any harmonic skeleton — pads were just modal scale notes picked
// at random. This rewrite gives it:
//
//   - A 4-chord modal "slow harmony" cycle, one chord every 45–75 seconds,
//     so the texture changes color but never stops feeling still.
//   - Two sustained pad layers (strings + warm pad) playing chord tones in
//     long-period loops with slow filter-cutoff LFOs for breathing.
//   - A choir "aahs" layer on the upper chord tones — the signature Eno
//     "voice in the pad" sound.
//   - A sparse tubular-bell motif: pentatonic chord tones triggered every
//     8–14 seconds on long incommensurate loops, drenched in reverb.
//   - A celesta sparkle layer (optional via section toggle).
//   - A sub-bass pedal that follows the chord root.
//
// There is no rhythm. Reverb sends are very high on every voice except the
// bass, which stays present.
//
// Preferred SF: arachno (D-50/M1/MU/Fairlight blend — perfect for retro pads).
type Ambient struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
	rng  *rand.Rand

	rootMidi  int
	keyOffset int

	// Chord cycle (modal — usually i / III / VI / iv in minor, or I / IV / vi
	// / V in major). Slot index per chord rotates very slowly.
	chords          []ambientChord
	currentChordIdx int

	samplesElapsed int64
	nextChordAt    int64
	nextSectionAt  int64

	bellsOn   *bool
	celestaOn *bool

	// Coherent 8-note bell phrase, pre-computed at Seed time so the bell
	// motif reads as a tune rather than a random sequence of strikes.
	bellPhrase []int
}

// ambientChord is one harmonic center: a root offset from the key center,
// plus the chord tones (semitone offsets from THE chord root).
type ambientChord struct {
	rootSemi int
	tones    []int // 3 or 4 chord tones from chord root
	label    string
}

// Modal ambient cycles: each is 4 chords, drifting slowly through related
// modes. Chord changes happen every 45–75 s so the listener feels harmonic
// motion without "song" structure.
var ambientCycles = [][]ambientChord{
	// Aeolian (natural minor) drift: i — III — VI — iv
	{
		{rootSemi: 0, tones: []int{0, 3, 7}, label: "i"},
		{rootSemi: 3, tones: []int{0, 4, 7}, label: "III"},
		{rootSemi: 8, tones: []int{0, 4, 7}, label: "VI"},
		{rootSemi: 5, tones: []int{0, 3, 7}, label: "iv"},
	},
	// Ionian (major) drift: I — V — vi — IV (50s changes slowed to ambient)
	{
		{rootSemi: 0, tones: []int{0, 4, 7, 11}, label: "Imaj7"},
		{rootSemi: 7, tones: []int{0, 4, 7}, label: "V"},
		{rootSemi: 9, tones: []int{0, 3, 7}, label: "vi"},
		{rootSemi: 5, tones: []int{0, 4, 7, 11}, label: "IVmaj7"},
	},
	// Dorian: i — IV — bIII — v (modal, no leading tone)
	{
		{rootSemi: 0, tones: []int{0, 3, 7}, label: "i_dor"},
		{rootSemi: 5, tones: []int{0, 4, 7}, label: "IV_dor"},
		{rootSemi: 3, tones: []int{0, 4, 7}, label: "bIII"},
		{rootSemi: 7, tones: []int{0, 3, 7}, label: "v"},
	},
	// Mixolydian: I — bVII — IV — I (very Eno-2/1)
	{
		{rootSemi: 0, tones: []int{0, 4, 7, 10}, label: "I7"},
		{rootSemi: 10, tones: []int{0, 4, 7}, label: "bVII"},
		{rootSemi: 5, tones: []int{0, 4, 7}, label: "IV"},
		{rootSemi: 0, tones: []int{0, 4, 7, 10}, label: "I7"},
	},
}

// NewAmbient constructs the algorithm. Caller must call Seed before Next.
func NewAmbient(sf *meltysynth.SoundFont) *Ambient { return &Ambient{sf: sf} }

func (a *Ambient) Name() string { return "ambient" }

func (a *Ambient) currentRoot() int { return a.rootMidi + a.keyOffset }

func (a *Ambient) Seed(seedVal int64) {
	a.rng = rand.New(rand.NewSource(seedVal)) //nolint:gosec
	// Root in the bass register so chord tones can stack upward through the
	// pad/choir/bell registers.
	a.rootMidi = 36 + a.rng.Intn(7) // C2..F#2
	a.keyOffset = 0
	a.samplesElapsed = 0
	a.currentChordIdx = 0
	a.chords = ambientCycles[a.rng.Intn(len(ambientCycles))]
	a.scheduleNextChord()

	bellsStart := true
	celestaStart := true
	a.bellsOn = &bellsStart
	a.celestaOn = &celestaStart
	a.scheduleNextSection()

	core, err := newSF2Core(a.sf, 3.2, seedVal)
	if err != nil {
		a.core = nil
		return
	}

	// Channel layout — all sustained, sparse trigger rate. Mapping:
	//   0 — String Ensemble 1   (program 48)  pad bed
	//   1 — Warm Pad            (program 89)  pad bed (parallel)
	//   2 — Choir Aahs          (program 52)  vocal pad
	//   3 — Tubular Bells       (program 14)  bell motif (sparse)
	//   4 — Celesta             (program 8)   high sparkle (sparse)
	//   5 — Synth Bass 1        (program 38)  sub-bass pedal
	core.setProgram(0, 48)
	core.setProgram(1, 89)
	core.setProgram(2, 52)
	core.setProgram(3, 14)
	core.setProgram(4, 8)
	core.setProgram(5, 38)
	core.setPan(0, 40)
	core.setPan(1, 88)
	core.setPan(2, 64)
	core.setPan(3, 92)
	core.setPan(4, 36)
	core.setPan(5, 64)

	// Brightness: pads darkened (the "warm wash"), choir mid, bells & celesta
	// full bright for the halo effect.
	core.setChannelCutoff(0, 72)
	core.setChannelCutoff(1, 64)
	core.setChannelCutoff(2, 80)
	core.setChannelCutoff(3, 110)
	core.setChannelCutoff(4, 110)
	core.setChannelCutoff(5, 56)

	// Filter LFOs on the pads — gives the texture a slow "breathing" quality.
	// Different rates so the two pads pulse out of sync.
	core.addFilterLFO(0, 1.0/14.0, 70, 22)
	core.addFilterLFO(1, 1.0/19.0, 60, 26)
	core.addFilterLFO(2, 1.0/11.0, 78, 18) // choir slightly faster

	// Reverb sends: everyone wet except the bass.
	core.setReverbSend(0, 100)
	core.setReverbSend(1, 96)
	core.setReverbSend(2, 110)
	core.setReverbSend(3, 120) // bells in halo
	core.setReverbSend(4, 120) // celesta in halo
	core.setReverbSend(5, 30)
	core.setChorusSend(1, 48)
	core.setChorusSend(2, 32)

	// --- String pad: 2-voice spread. Periods chosen from Eno's documented
	// Music for Airports loop lengths (19.8s, 25.7s) — coprime in seconds
	// so they never realign within the listener's attention window.
	for ti, period := range []float64{19.8, 25.7} {
		voice := ti
		core.addTrack(SF2Track{
			Channel: 0, Velocity: 56, Notes: []int{a.padNote(voice, 0)},
			PeriodSec: period, Phase01: a.rng.Float64(),
			MutationRate: 0.30,
			MutateOne:    func(_ int, _ int) int { return a.padNote(voice, 0) },
			VelocityJitter: 8, TimingJitterSec: 0.05,
		})
	}
	// Warm pad layered in parallel — also documented Eno periods.
	for ti, period := range []float64{23.6, 31.0} {
		voice := ti
		core.addTrack(SF2Track{
			Channel: 1, Velocity: 50, Notes: []int{a.padNote(voice, 0)},
			PeriodSec: period, Phase01: a.rng.Float64(),
			MutationRate: 0.30,
			MutateOne:    func(_ int, _ int) int { return a.padNote(voice, 0) },
			VelocityJitter: 6, TimingJitterSec: 0.05,
		})
	}

	// --- Choir aahs: single upper-register voice on a 29.2s period (an
	// Eno-documented loop length).
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 48, Notes: []int{a.choirNote(0)},
		PeriodSec: 29.2, Phase01: a.rng.Float64(),
		MutationRate: 0.35,
		MutateOne:    func(_ int, _ int) int { return a.choirNote(0) },
		VelocityJitter: 8, TimingJitterSec: 0.06,
	})

	// --- Tubular bell motif: literal Eno "2/1" recipe. Three voices, each
	// holding ONE pitch of the current chord, on incommensurate loop
	// periods (25.7s, 31.0s, 36.1s) so they never realign. This is the
	// canonical Music for Airports voices-on-tape-loops technique that the
	// listener recognizes as ambient — not a melodic phrase, but three
	// pitches that overlap unpredictably with long fadeouts.
	a.bellPhrase = a.makeBellPhrase()
	bellPeriods := []float64{25.7, 31.0, 36.1}
	for ti, period := range bellPeriods {
		voice := ti
		core.addTrack(SF2Track{
			Channel: 3, Velocity: 64, Notes: []int{a.bellPhrase[voice%len(a.bellPhrase)]},
			PeriodSec: period, Phase01: a.rng.Float64(),
			MutationRate: 0.20,
			MutateOne: func(_ int, _ int) int {
				return a.bellPhrase[voice%len(a.bellPhrase)]
			},
			VelocityJitter: 14, TimingJitterSec: 0.30, // tape-loop slippage feel
			Enabled: a.bellsOn,
		})
	}

	// --- Celesta sparkle: very high register, very sparse — single voice on
	// a long period.
	core.addTrack(SF2Track{
		Channel: 4, Velocity: 44, Notes: []int{a.celestaNote()},
		PeriodSec: 53.7, Phase01: a.rng.Float64(),
		MutationRate: 0.40,
		MutateOne:    func(_ int, _ int) int { return a.celestaNote() },
		VelocityJitter: 12, TimingJitterSec: 0.10,
		Enabled: a.celestaOn,
	})

	// --- Sub-bass pedal: very slow rate, holds the chord root in the very
	// bottom of the register so the texture has a foundation.
	core.addTrack(SF2Track{
		Channel: 5, Velocity: 60, Notes: []int{a.bassRoot()},
		PeriodSec: 41.7, Phase01: 0,
		MutationRate: 0.50,
		MutateOne:    func(_ int, _ int) int { return a.bassRoot() },
		VelocityJitter: 6, TimingJitterSec: 0.03,
	})

	a.core = core
}

// padNote returns one MIDI key for the pad/strings, picking chord tone
// `voice` (0..2) from the current chord, in the pad register (around C4–B5).
func (a *Ambient) padNote(voice, _ int) int {
	if len(a.chords) == 0 {
		return 60
	}
	c := a.chords[a.currentChordIdx]
	idx := voice % len(c.tones)
	key := a.currentRoot() + c.rootSemi + c.tones[idx]
	// Spread voices across octaves so a 3-tone chord covers the pad register.
	octaveBump := 24 + 12*voice
	key += octaveBump
	for key < 60 {
		key += 12
	}
	for key > 84 {
		key -= 12
	}
	return key
}

// choirNote returns an upper chord-tone key in the soprano range (C5–C6).
func (a *Ambient) choirNote(voice int) int {
	if len(a.chords) == 0 {
		return 72
	}
	c := a.chords[a.currentChordIdx]
	// Prefer 3rd and 5th over root for the choir (more "vocal").
	pick := []int{1, 2, 1, 2, 0}
	idx := pick[voice%len(pick)]
	if idx >= len(c.tones) {
		idx = len(c.tones) - 1
	}
	key := a.currentRoot() + c.rootSemi + c.tones[idx] + 36
	for key < 72 {
		key += 12
	}
	for key > 84 {
		key -= 12
	}
	return key
}

// makeBellPhrase builds the bell-motif's coherent 8-note melodic contour
// using the current chord's tones, transposed into the bell register
// (C5–C7). Picks a stylized contour (peak-and-fall by default for the Eno
// 2/1 vibe) and uses chord-tone scale degrees so every note resolves.
func (a *Ambient) makeBellPhrase() []int {
	if len(a.chords) == 0 {
		return []int{72}
	}
	// Pentatonic of the chord: root, 3rd (or 4th if no 3rd), 5th, 7th (or 9th).
	c := a.chords[a.currentChordIdx]
	scale := make([]int, 0, len(c.tones))
	for _, t := range c.tones {
		scale = append(scale, t)
	}
	// Peak-and-fall contour — Music for Airports "2/1" feel.
	phrase := melodicPhrases[2] // {0, 2, 4, 6, 4, 2, 0, -2}
	root := a.currentRoot() + c.rootSemi + 36
	notes := applyPhraseToScale(phrase, scale, root, 0, 0)
	// Clamp into the bell register.
	for i, k := range notes {
		for k < 72 {
			k += 12
		}
		for k > 96 {
			k -= 12
		}
		notes[i] = k
	}
	return notes
}

// celestaNote returns a very high chord-tone key (C6–C7).
func (a *Ambient) celestaNote() int {
	if len(a.chords) == 0 {
		return 84
	}
	c := a.chords[a.currentChordIdx]
	idx := a.rng.Intn(len(c.tones))
	key := a.currentRoot() + c.rootSemi + c.tones[idx] + 48
	for key < 84 {
		key += 12
	}
	for key > 96 {
		key -= 12
	}
	return key
}

// bassRoot returns the chord's root in the bass register (around C2).
func (a *Ambient) bassRoot() int {
	if len(a.chords) == 0 {
		return 36
	}
	c := a.chords[a.currentChordIdx]
	key := a.currentRoot() + c.rootSemi
	for key > 48 {
		key -= 12
	}
	for key < 24 {
		key += 12
	}
	return key
}

func (a *Ambient) scheduleNextChord() {
	// 45-75 seconds per chord — very slow harmonic motion.
	secs := 45.0 + 30.0*a.rng.Float64()
	a.nextChordAt = a.samplesElapsed + int64(secs*44100)
}

func (a *Ambient) scheduleNextSection() {
	// 2–4 min between section toggles (which ornaments are on).
	secs := 120.0 + 120.0*a.rng.Float64()
	a.nextSectionAt = a.samplesElapsed + int64(secs*44100)
}

func (a *Ambient) advance() {
	if a.samplesElapsed >= a.nextChordAt {
		a.currentChordIdx = (a.currentChordIdx + 1) % len(a.chords)
		a.scheduleNextChord()
	}
	if a.samplesElapsed >= a.nextSectionAt {
		// Toggle one of the ornament layers.
		if a.rng.Float64() < 0.5 {
			*a.bellsOn = !*a.bellsOn
		} else {
			*a.celestaOn = !*a.celestaOn
		}
		a.scheduleNextSection()
	}
}

// SetReverbIR installs a convolution reverb on the master bus.
func (a *Ambient) SetReverbIR(ir []float64, wet float64) {
	if a.core != nil {
		a.core.setConvolutionIR(ir, wet)
	}
}

func (a *Ambient) Next(left, right []float64) {
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

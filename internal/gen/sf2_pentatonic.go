package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

var _ Algorithm = (*SF2Pentatonic)(nil)

// SF2Pentatonic is the pentatonic-walk algorithm rendered through a
// SoundFont. The random walk through pentatonic minor sounds especially
// good on piano + orchestral harp — both are immediately associated with
// "music box" / "lullaby" textures.
type SF2Pentatonic struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
	rng  *rand.Rand

	samplesElapsed int64
	nextSwapAt     int64
}

func NewSF2Pentatonic(sf *meltysynth.SoundFont) *SF2Pentatonic {
	return &SF2Pentatonic{sf: sf}
}

func (a *SF2Pentatonic) Name() string { return "pentatonic-sf2" }

func (a *SF2Pentatonic) Seed(seedVal int64) {
	rng := rand.New(rand.NewSource(seedVal)) //nolint:gosec
	a.rng = rng
	rootMidi := 36 + rng.Intn(12)
	a.samplesElapsed = 0
	a.scheduleNextSwap()

	core, err := newSF2Core(a.sf, 3.2, seedVal)
	if err != nil {
		a.core = nil
		return
	}
	core.setProgram(0, 0)  // Acoustic Grand Piano  (center-left)
	core.setProgram(1, 46) // Orchestral Harp        (center-right)
	core.setProgram(2, 10) // Music Box              (high sparkle, hard right)
	core.setProgram(3, 32) // Acoustic Bass          (low foundation, center)
	core.setPan(0, 50)
	core.setPan(1, 78)
	core.setPan(2, 100)
	core.setPan(3, 64)

	// Pentatonic-sf2 leans lullaby. Darken the piano to "music box" register
	// (CC 74 ≈ 56), keep harp clear, music box natural bright, bass moderate.
	core.setChannelCutoff(0, 56)  // piano — music-box-y darken
	core.setChannelCutoff(1, 80)  // orchestral harp
	core.setChannelCutoff(2, 70)  // music box
	core.setChannelCutoff(3, 64)  // bass

	core.setReverbSend(0, 70)  // piano: moderate
	core.setReverbSend(1, 80)  // harp: a bit more for that "music box" halo
	core.setReverbSend(2, 110) // music box: drenched
	core.setReverbSend(3, 35)  // bass: dry, hold the low end
	core.setChorusSend(0, 24)
	core.setChorusSend(1, 24)

	// Same walk-based note generation as gen.Pentatonic.
	for i, period := range pentaLoopPeriods {
		count := 6 + rng.Intn(5)
		notes := pentatonicWalk(rng, rootMidi, count)
		ch := int32(i % 2)
		var vel int32 = 80
		if ch == 1 {
			vel = 68
		}
		root := rootMidi
		mutate := func(_ int, prev int) int {
			closest := findClosestScalePitch(prev, root, scalePentatonicMinor)
			step := walkStep(rng, closest.degreeIdx, len(scalePentatonicMinor))
			return root + scalePentatonicMinor[step] + closest.octaveOffset*12
		}
		core.addTrack(SF2Track{
			Channel: ch, Velocity: vel, Notes: notes,
			PeriodSec: period, Phase01: rng.Float64(),
			MutationRate: 0.15, MutateOne: mutate,
			VelocityJitter: 8, TimingJitterSec: 0.015,
		})
	}

	// Acoustic bass: slow walk through pentatonic notes one octave below,
	// 5 notes per ~25s cycle. Gives a foundation without competing rhythmically.
	bassNotes := pentatonicWalk(rng, rootMidi-12, 5)
	bassMutate := func(_ int, prev int) int {
		closest := findClosestScalePitch(prev, rootMidi-12, scalePentatonicMinor)
		step := walkStep(rng, closest.degreeIdx, len(scalePentatonicMinor))
		return (rootMidi - 12) + scalePentatonicMinor[step] + closest.octaveOffset*12
	}
	core.addTrack(SF2Track{
		Channel: 3, Velocity: 78, Notes: bassNotes,
		PeriodSec: 22.0, Phase01: rng.Float64(),
		MutationRate: 0.10, MutateOne: bassMutate,
		VelocityJitter: 5, TimingJitterSec: 0.018,
	})

	// Sub-audible shaker on the GM drum channel. GM percussion key 70 is
	// Maracas, key 82 is Shaker — both work for a soft hand-percussion bed.
	// Eight quiet hits per 16-second cycle (one every 2 s). Very low velocity
	// so it's atmosphere, not rhythm. Skip the drum channel's bank-select
	// dance; meltysynth treats channel 9 as percussion automatically.
	const drumCh = 9
	core.setProgram(drumCh, 0)
	core.setPan(drumCh, 64)
	core.setReverbSend(drumCh, 60)
	shakerNotes := make([]int, 8)
	for i := range shakerNotes {
		shakerNotes[i] = 82 // GM Shaker
	}
	core.addTrack(SF2Track{
		Channel: drumCh, Velocity: 40, Notes: shakerNotes,
		PeriodSec: 16.0, Phase01: 0,
		VelocityJitter: 14, TimingJitterSec: 0.020,
	})

	// Music box sparkle: very sparse, high register. 3 notes per ~30s cycle.
	mbNotes := make([]int, 3)
	for j := range mbNotes {
		deg := scalePentatonicMinor[rng.Intn(len(scalePentatonicMinor))]
		mbNotes[j] = rootMidi + deg + 48 // four octaves above root
	}
	mbMutate := func(_ int, _ int) int {
		deg := scalePentatonicMinor[rng.Intn(len(scalePentatonicMinor))]
		return rootMidi + deg + 48
	}
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 60, Notes: mbNotes,
		PeriodSec: 30.0, Phase01: rng.Float64(),
		MutationRate: 0.20, MutateOne: mbMutate,
		VelocityJitter: 12, TimingJitterSec: 0.025,
	})

	a.core = core
}

// scalePitchLoc describes where a MIDI note sits in a scale anchored at
// rootMidi: which degree (0..len(scale)-1) and how many octaves above the
// root. Used by mutation to find a "nearby" scale note to walk from.
type scalePitchLoc struct {
	degreeIdx    int
	octaveOffset int
}

// findClosestScalePitch returns the closest scale-anchored pitch to the given
// MIDI note. Useful for mutation closures that want to walk from "wherever
// we currently are" rather than re-rolling from scratch.
func findClosestScalePitch(midi, rootMidi int, scale []int) scalePitchLoc {
	rel := midi - rootMidi
	// Decompose into (octave, semitone within octave) where semitone is in 0..11.
	octave := rel / 12
	semi := rel % 12
	if semi < 0 {
		semi += 12
		octave--
	}
	bestIdx := 0
	bestDist := 1 << 30
	for i, s := range scale {
		d := semi - s
		if d < 0 {
			d = -d
		}
		if d < bestDist {
			bestDist = d
			bestIdx = i
		}
	}
	return scalePitchLoc{degreeIdx: bestIdx, octaveOffset: octave}
}

// pentatonicWalk produces a random walk through the pentatonic-minor scale,
// matching gen.Pentatonic's logic. Extracted here as a shared helper.
func pentatonicWalk(rng *rand.Rand, rootMidi, count int) []int {
	scale := scalePentatonicMinor
	idx := rng.Intn(len(scale))
	octave := 12 * (2 + rng.Intn(3))
	notes := make([]int, count)
	notes[0] = rootMidi + scale[idx] + octave
	for i := 1; i < count; i++ {
		idx = walkStep(rng, idx, len(scale))
		if rng.Float64() < 0.18 {
			if rng.Float64() < 0.5 {
				octave += 12
			} else {
				octave -= 12
			}
			if octave < 12 {
				octave = 12
			}
			if octave > 60 {
				octave = 60
			}
		}
		notes[i] = rootMidi + scale[idx] + octave
	}
	return notes
}

// SetReverbIR installs a convolution reverb on the master bus.
func (a *SF2Pentatonic) SetReverbIR(ir []float64, wet float64) {
	if a.core != nil {
		a.core.setConvolutionIR(ir, wet)
	}
}

func (a *SF2Pentatonic) Next(left, right []float64) {
	if a.core == nil {
		for i := range left {
			left[i] = 0
			right[i] = 0
		}
		return
	}
	a.core.renderInto(left, right)
	a.samplesElapsed += int64(len(left))
	if a.samplesElapsed >= a.nextSwapAt {
		a.swapOneInstrument()
		a.scheduleNextSwap()
	}
}

var pentaChannelAlternatives = map[int32][]int32{
	0: {0, 1, 4, 5},        // Acoustic Grand (default), Bright Piano, EP1, EP2
	1: {46, 24, 25, 105},   // Orchestral Harp (default), Nylon, Steel, Banjo
	2: {10, 8, 9},          // Music Box (default), Celesta, Glockenspiel
	3: {32, 33, 87},        // Acoustic Bass (default), Electric Bass, Lead Bass
}

func (a *SF2Pentatonic) scheduleNextSwap() {
	secs := 220.0 + 180.0*a.rng.Float64()
	a.nextSwapAt = a.samplesElapsed + int64(secs*44100)
}

func (a *SF2Pentatonic) swapOneInstrument() {
	channels := []int32{0, 1, 2, 3}
	ch := channels[a.rng.Intn(len(channels))]
	a.core.programSwap(ch, pentaChannelAlternatives[ch], a.rng)
}

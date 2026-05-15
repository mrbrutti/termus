package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

var _ Algorithm = (*SF2Drone)(nil)

// SF2Drone is the drone-bed algorithm rendered through a SoundFont. Long
// sustained notes on string ensemble and slow strings, with a high-register
// flute shimmer voice — close in feel to the Eno-meets-Stars-of-the-Lid
// orchestrations.
type SF2Drone struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
	rng  *rand.Rand

	samplesElapsed int64
	nextSwapAt     int64
}

func NewSF2Drone(sf *meltysynth.SoundFont) *SF2Drone {
	return &SF2Drone{sf: sf}
}

func (a *SF2Drone) Name() string { return "drone-sf2" }

func (a *SF2Drone) Seed(seedVal int64) {
	rng := rand.New(rand.NewSource(seedVal)) //nolint:gosec
	a.rng = rng
	rootMidi := 24 + rng.Intn(7) // C1..F#1
	a.samplesElapsed = 0
	a.scheduleNextSwap()

	core, err := newSF2Core(a.sf, 3.0, seedVal)
	if err != nil {
		a.core = nil
		return
	}
	core.setProgram(0, 48) // String Ensemble 1 (left)
	core.setProgram(1, 49) // String Ensemble 2 slow (right)
	core.setProgram(2, 73) // Flute (shimmer, high right)
	core.setProgram(3, 53) // Choir Voice "Oohs" (warmth, center)
	core.setProgram(4, 32) // Acoustic Bass (foundation, center)
	core.setPan(0, 38)
	core.setPan(1, 90)
	core.setPan(2, 100)
	core.setPan(3, 64)
	core.setPan(4, 64)

	// Per-channel base cutoffs. Drone wants smooth sustained tones — slight
	// darkening on the strings/choir helps remove brittleness on long held
	// notes. Flute kept bright since it's the high-register shimmer.
	core.setChannelCutoff(0, 72)  // string ensemble 1
	core.setChannelCutoff(1, 68)  // string ensemble 2 slow
	core.setChannelCutoff(2, 88)  // flute shimmer
	core.setChannelCutoff(3, 64)  // choir warmth
	core.setChannelCutoff(4, 76)  // bass

	// Filter LFOs on the sustained string and choir layers. Different rates
	// per channel so they breathe out of phase with each other.
	core.addFilterLFO(0, 1.0/14.0, 62, 28)
	core.addFilterLFO(1, 1.0/19.0, 60, 26)
	core.addFilterLFO(3, 1.0/11.0, 70, 30)

	// Drone is the wettest algorithm — everything except the bass sits deep
	// in a cathedral. Strings + choir get full wet; flute shimmer drenched
	// for the "from far away" halo; bass kept drier so the low end has body.
	core.setReverbSend(0, 110)
	core.setReverbSend(1, 110)
	core.setReverbSend(2, 120) // flute shimmer
	core.setReverbSend(3, 100) // choir
	core.setReverbSend(4, 40)  // bass — drier
	core.setChorusSend(0, 50)
	core.setChorusSend(1, 50)
	core.setChorusSend(3, 56)

	// Bed voices on long periods. Mutation is gentle here — drone wants
	// to feel stable; abrupt note changes would betray the aesthetic.
	bedMutate := func(_ int, _ int) int {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		octave := 12 * (1 + rng.Intn(3))
		return rootMidi + degree + octave
	}
	for _, period := range droneLoopPeriods {
		notes := make([]int, 3+rng.Intn(3))
		for j := range notes {
			notes[j] = bedMutate(0, 0)
		}
		phase := rng.Float64()
		core.addTrack(SF2Track{
			Channel: 0, Velocity: 64, Notes: notes,
			PeriodSec: period, Phase01: phase,
			MutationRate: 0.05, MutateOne: bedMutate,
			VelocityJitter: 4, TimingJitterSec: 0.020,
		})
		core.addTrack(SF2Track{
			Channel: 1, Velocity: 48, Notes: notes,
			PeriodSec: period, Phase01: phase,
			VelocityJitter: 4, TimingJitterSec: 0.020,
		})
	}

	// Shimmer voice on a 19s period — flute in the high register.
	shimmerMutate := func(_ int, _ int) int {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		octave := 12 * (3 + rng.Intn(2))
		return rootMidi + degree + octave
	}
	shimmerNotes := make([]int, 4+rng.Intn(3))
	for j := range shimmerNotes {
		shimmerNotes[j] = shimmerMutate(0, 0)
	}
	core.addTrack(SF2Track{
		Channel: 2, Velocity: 70, Notes: shimmerNotes,
		PeriodSec: shimmerPeriod, Phase01: rng.Float64(),
		MutationRate: 0.15, MutateOne: shimmerMutate,
		VelocityJitter: 10, TimingJitterSec: 0.025,
	})

	// Choir voice "oohs" in mid-register for human warmth. Slow cycle so
	// the same chord-tone-ish notes hold for a long time.
	choirMutate := func(_ int, _ int) int {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		return rootMidi + degree + 12 // one octave above root
	}
	choirNotes := make([]int, 3)
	for j := range choirNotes {
		choirNotes[j] = choirMutate(0, 0)
	}
	core.addTrack(SF2Track{
		Channel: 3, Velocity: 54, Notes: choirNotes,
		PeriodSec: 37.0, Phase01: rng.Float64(),
		MutationRate: 0.08, MutateOne: choirMutate,
		VelocityJitter: 4, TimingJitterSec: 0.020,
	})

	// Acoustic bass: very slow walk through scale tones at the root register.
	bassMutate := func(_ int, _ int) int {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		return rootMidi + degree // base octave (rootMidi is already low C1-F#1)
	}
	bassNotes := make([]int, 3)
	for j := range bassNotes {
		bassNotes[j] = bassMutate(0, 0)
	}
	core.addTrack(SF2Track{
		Channel: 4, Velocity: 76, Notes: bassNotes,
		PeriodSec: 41.0, Phase01: rng.Float64(),
		MutationRate: 0.06, MutateOne: bassMutate,
		VelocityJitter: 4, TimingJitterSec: 0.020,
	})

	a.core = core
}

// SetReverbIR installs a convolution reverb on the master bus.
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
	a.core.renderInto(left, right)
	a.samplesElapsed += int64(len(left))
	if a.samplesElapsed >= a.nextSwapAt {
		a.swapOneInstrument()
		a.scheduleNextSwap()
	}
}

var droneChannelAlternatives = map[int32][]int32{
	0: {48, 49, 50, 51, 91}, // String Ensemble 1 (default), 2, Synth Strings 1, 2, Choir
	1: {49, 48, 50, 91},     // Slow Strings (default), String Ens 1, Synth Strings, Choir
	2: {73, 74, 75, 76, 88}, // Flute (default), Recorder, Pan Flute, Blown Bottle, New Age Pad
	3: {53, 52, 54, 91},     // Choir Voice "Oohs" (default), Aahs, Synth Voice, Choir Pad
	4: {32, 33, 38, 87},     // Acoustic Bass (default), Electric Bass, Synth Bass 1, Lead Bass
}

func (a *SF2Drone) scheduleNextSwap() {
	secs := 240.0 + 180.0*a.rng.Float64()
	a.nextSwapAt = a.samplesElapsed + int64(secs*44100)
}

func (a *SF2Drone) swapOneInstrument() {
	channels := []int32{0, 1, 2, 3, 4}
	ch := channels[a.rng.Intn(len(channels))]
	a.core.programSwap(ch, droneChannelAlternatives[ch], a.rng)
}

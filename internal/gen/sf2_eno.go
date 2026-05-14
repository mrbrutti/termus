package gen

import (
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

// Compile-time assertion.
var _ Algorithm = (*SF2Eno)(nil)

// SF2Eno is the eno-drift algorithm rendered through a SoundFont. Same
// scheduling structure as Eno (5 pad voices on incommensurate loops + 2 lead
// voices on shorter loops) but voices are mapped to sampled instruments —
// strings + warm pad for the bed, acoustic grand piano for the leads.
type SF2Eno struct {
	sf   *meltysynth.SoundFont
	core *sf2Core
	rng  *rand.Rand

	samplesElapsed int64
	nextSwapAt     int64

	// Section state: ornament layers (celesta sparkle, french horn) toggle
	// on/off every ~2–4 minutes. Two booleans give 4 sections naturally
	// (both off, just celesta, just horn, both on), giving 8–16 minutes of
	// cyclic form variation before patterns can repeat.
	celestaOn *bool
	hornOn    *bool
	nextSectionAt int64
}

// NewSF2Eno builds the algorithm. Caller must call Seed before Next.
func NewSF2Eno(sf *meltysynth.SoundFont) *SF2Eno {
	return &SF2Eno{sf: sf}
}

func (a *SF2Eno) Name() string { return "eno-sf2" }

func (a *SF2Eno) Seed(seedVal int64) {
	rng := rand.New(rand.NewSource(seedVal)) //nolint:gosec
	a.rng = rng
	rootMidi := 36 + rng.Intn(12)
	a.samplesElapsed = 0
	a.scheduleNextSwap()
	// Start in the fullest section (both ornaments on). The first section
	// transition usually mutes one of them, so the listener hears a clear
	// "the celesta dropped out" event.
	on1 := true
	on2 := true
	a.celestaOn = &on1
	a.hornOn = &on2
	a.scheduleNextSection()

	core, err := newSF2Core(a.sf, 3.5, seedVal)
	if err != nil {
		a.core = nil
		return
	}
	core.setProgram(0, 48) // String Ensemble 1     (left)
	core.setProgram(1, 89) // Warm Pad              (right)
	core.setProgram(2, 0)  // Acoustic Grand Piano  (center, leads)
	core.setProgram(3, 8)  // Celesta               (high sparkle, right)
	core.setProgram(4, 60) // French Horn           (low foundation, center)
	core.setPan(0, 38)
	core.setPan(1, 90)
	core.setPan(2, 64)
	core.setPan(3, 100)
	core.setPan(4, 64)

	// Slow filter modulation on the sustained pads — gives the texture a
	// breathing quality. ~12 s cycle on strings, ~17 s on the warm pad
	// (different rates so the two pads don't pulse in sync).
	core.addFilterLFO(0, 1.0/12.0, 64, 26)
	core.addFilterLFO(1, 1.0/17.0, 64, 30)
	core.addFilterLFO(4, 1.0/9.0, 70, 18)

	// Per-channel reverb/chorus send. Strings + pad sit IN the reverb (high
	// send), piano leads have moderate, celesta sparkle has lots of reverb
	// for the "halo" effect, french horn slightly dry to anchor the low end.
	core.setReverbSend(0, 90) // strings: deeply wet
	core.setReverbSend(1, 80) // warm pad: wet
	core.setReverbSend(2, 60) // piano leads: moderate
	core.setReverbSend(3, 110) // celesta: drenched, halo effect
	core.setReverbSend(4, 50)  // french horn: drier, holds the low end
	core.setChorusSend(0, 32)
	core.setChorusSend(1, 48) // warm pad benefits most from chorus

	// Slow pad bed — same logic as gen.Eno.Seed but with two tracks per
	// musical voice so strings and warm pad layer together.
	padMutate := func(_ int, _ int) int {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		octave := 12 * (2 + rng.Intn(3))
		return rootMidi + degree + octave
	}
	for _, period := range loopPeriods {
		notes := make([]int, 2+rng.Intn(3))
		for j := range notes {
			notes[j] = padMutate(0, 0)
		}
		phase := rng.Float64()
		core.addTrack(SF2Track{
			Channel: 0, Velocity: 70, Notes: notes,
			PeriodSec: period, Phase01: phase,
			MutationRate: 0.08, MutateOne: padMutate,
			VelocityJitter: 6, TimingJitterSec: 0.012,
		})
		core.addTrack(SF2Track{
			Channel: 1, Velocity: 56, Notes: notes,
			PeriodSec: period, Phase01: phase,
			VelocityJitter: 4, TimingJitterSec: 0.008,
		})
	}

	// Lead voices — shorter periods, more notes, higher register, piano.
	leadMutate := func(_ int, _ int) int {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		octave := 12 * (3 + rng.Intn(2))
		return rootMidi + degree + octave
	}
	for _, period := range leadPeriods {
		notes := make([]int, 4+rng.Intn(3))
		for j := range notes {
			notes[j] = leadMutate(0, 0)
		}
		core.addTrack(SF2Track{
			Channel: 2, Velocity: 92, Notes: notes,
			PeriodSec: period, Phase01: rng.Float64(),
			MutationRate: 0.18, MutateOne: leadMutate,
			VelocityJitter: 12, TimingJitterSec: 0.018,
		})
	}

	// Celesta: sparse high-register sparkle, very slow trigger rate.
	celesta := func(_ int, _ int) int {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		return rootMidi + degree + 48 // +4 octaves
	}
	celestaNotes := make([]int, 4)
	for j := range celestaNotes {
		celestaNotes[j] = celesta(0, 0)
	}
	core.addTrack(SF2Track{
		Channel: 3, Velocity: 56, Notes: celestaNotes,
		PeriodSec: 27.0, Phase01: rng.Float64(),
		MutationRate: 0.20, MutateOne: celesta,
		VelocityJitter: 12, TimingJitterSec: 0.018,
		Enabled: a.celestaOn,
	})

	// French horn: warm low sustained voice on chord roots. Cycles slowly
	// (chord change every ~25s) — anchors the harmony without intruding on
	// the pad bed.
	horn := func(_ int, _ int) int {
		degree := scaleMinor[rng.Intn(len(scaleMinor))]
		return rootMidi + degree // low register
	}
	hornNotes := make([]int, 3)
	for j := range hornNotes {
		hornNotes[j] = horn(0, 0)
	}
	core.addTrack(SF2Track{
		Channel: 4, Velocity: 50, Notes: hornNotes,
		PeriodSec: 25.0, Phase01: rng.Float64(),
		MutationRate: 0.10, MutateOne: horn,
		VelocityJitter: 6, TimingJitterSec: 0.012,
		Enabled: a.hornOn,
	})

	a.core = core
}

// SetReverbIR installs a convolution reverb on the master bus.
func (a *SF2Eno) SetReverbIR(ir []float64, wet float64) {
	if a.core != nil {
		a.core.setConvolutionIR(ir, wet)
	}
}

func (a *SF2Eno) Next(left, right []float64) {
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
	if a.samplesElapsed >= a.nextSectionAt {
		a.advanceSection()
		a.scheduleNextSection()
	}
}

// advanceSection randomly flips one of the ornament-layer enable flags.
// This produces "verse / chorus / bridge" feel over many minutes: sometimes
// just strings + pad + piano, sometimes with celesta sparkle, sometimes
// with horn warmth, sometimes the full ensemble.
func (a *SF2Eno) advanceSection() {
	if a.rng.Float64() < 0.5 {
		*a.celestaOn = !*a.celestaOn
	} else {
		*a.hornOn = !*a.hornOn
	}
}

func (a *SF2Eno) scheduleNextSection() {
	secs := 120.0 + 120.0*a.rng.Float64() // 2–4 min
	a.nextSectionAt = a.samplesElapsed + int64(secs*44100)
}

// enoChannelAlternatives lists GM programs that are musically compatible
// substitutes for each instrument role in the algorithm. Picked so swapping
// at any time stays inside the "Music for Airports" character.
var enoChannelAlternatives = map[int32][]int32{
	0: {48, 49, 50, 51, 92},      // String Ensemble 1 (default), 2, Synth Strings 1, 2, Bowed Glass
	1: {89, 88, 91, 95},          // Warm Pad (default), New Age, Choir, Sweep Pad
	2: {0, 1, 4, 5, 8},           // Acoustic Grand (default), Bright, EP1, EP2, Celesta
	3: {8, 9, 10, 14},            // Celesta (default), Glockenspiel, Music Box, Tubular Bells
	4: {60, 61, 68, 71},          // French Horn (default), Brass Section, Oboe, Clarinet
}

func (a *SF2Eno) scheduleNextSwap() {
	// 3–5 minutes between swaps. With 5 channels, on average each channel
	// gets swapped every ~20 min — slow enough to be noticed but not
	// fatiguing.
	secs := 180.0 + 120.0*a.rng.Float64()
	a.nextSwapAt = a.samplesElapsed + int64(secs*44100)
}

func (a *SF2Eno) swapOneInstrument() {
	channels := []int32{0, 1, 2, 3, 4}
	ch := channels[a.rng.Intn(len(channels))]
	a.core.programSwap(ch, enoChannelAlternatives[ch], a.rng)
}

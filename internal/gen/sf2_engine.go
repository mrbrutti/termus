package gen

import (
	"math"
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"

	"github.com/mrbrutti/termus/internal/synth"
)

// SF2Reverberator is the optional interface implemented by SF2-mode
// algorithms that can have a convolution reverb installed on their master
// bus. Used by the --ir flag to inject a real impulse response.
type SF2Reverberator interface {
	Algorithm
	SetReverbIR(ir []float64, wet float64)
}

// SF2Track configures one cycling-note track inside an sf2Core engine.
// Multiple tracks can share a MIDI channel (e.g. several voices all on
// piano) — events from each track will sum naturally inside go-meltysynth's
// polyphony.
//
// Mutation: if MutationRate > 0 and MutateOne != nil, the engine will
// occasionally re-roll one of the Notes (never the one currently sounding)
// using MutateOne. This makes the music gradually evolve over time rather
// than looping the same figure indefinitely.
//
// Humanization: VelocityJitter randomizes velocity per NoteOn so dynamics
// aren't deadly flat. TimingJitterSec adds a random ±N-second offset to
// each NoteOn's fire time so the rhythm doesn't sound machine-precise.
// Typical values: Velocity 4–14, Timing 0.005–0.020 (5–20 ms).
type SF2Track struct {
	Channel   int32 // MIDI channel 0..15
	Velocity  int32 // 0..127
	Notes     []int // MIDI keys, cycled
	PeriodSec float64
	Phase01   float64 // 0..1 phase offset within the period

	MutationRate float64
	MutateOne    func(slot int, prev int) int

	VelocityJitter  int32   // ±range added to Velocity per NoteOn
	TimingJitterSec float64 // ±range added to fire time per NoteOn (seconds)
}

// sf2TrackState is the runtime state for one track. The scheduler uses
// absolute fire times (samples since Seed) rather than per-cycle phase math
// so timing jitter can be applied naturally — the next-fire sample is just
// the slot boundary plus a random offset, with no special cases for "fired
// early" vs "fired late".
type sf2TrackState struct {
	cfg           SF2Track
	periodSamples int64
	phaseOffset   int64
	notesLen      int64 // len(cfg.Notes) cached as int64 to avoid conversion in hot path
	curSlot       int   // last slot we fired (-1 before first fire)
	curKey        int   // currently sounding key, or -1
	nextFireT     int64 // absolute sample at which we should fire the next NoteOn
}

// sf2Core is a shared SoundFont rendering engine. Each SF2-mode algorithm
// constructs one of these, registers its tracks, and uses it to render audio
// while the algorithm's own logic (chord progressions, walks, Markov chains,
// etc.) decides what notes go on which tracks.
//
// The engine handles:
//   - SoundFont synthesizer construction
//   - Per-track NoteOn/NoteOff scheduling at slot boundaries
//   - Master bus: gain, low-shelf + high-shelf EQ, soft-knee stereo
//     compressor, soft-clip safety
//   - float32 ↔ float64 conversion at the go-meltysynth boundary
type sf2Core struct {
	syn        *meltysynth.Synthesizer
	tracks     []*sf2TrackState
	t          int64
	masterGain float64
	rng        *rand.Rand // for mutation; lives on the audio goroutine only

	eqLowL, eqLowR   *synth.LowShelf
	eqHighL, eqHighR *synth.HighShelf
	comp             *synth.StereoCompressor

	// Optional master-bus low-pass — for the "muffled tape" feel of lofi.
	// nil disables it; the chain falls through without filtering.
	lpL, lpR *synth.Lowpass

	// Optional tape-hiss noise floor at the master bus output. 0 disables.
	hissLevel float64

	// Optional convolution reverb. When non-nil, applied in parallel with
	// the dry signal at convWet mix level. Each channel has its own
	// instance, both seeded from the same IR. nil disables convolution.
	convL, convR synth.RealtimeConvolver
	convWet      float64

	bufF32L []float32
	bufF32R []float32
}

// newSF2Core constructs the engine and the master bus. masterGain compensates
// for go-meltysynth's conservative internal levels. mutationSeed seeds the
// mutation RNG; pass the same value Seed() uses for note generation to keep
// the mutation sequence deterministic with the rest of the algorithm.
func newSF2Core(sf *meltysynth.SoundFont, masterGain float64, mutationSeed int64) (*sf2Core, error) {
	settings := meltysynth.NewSynthesizerSettings(synth.SampleRate)
	settings.EnableReverbAndChorus = true
	settings.MaximumPolyphony = 96
	syn, err := meltysynth.NewSynthesizer(sf, settings)
	if err != nil {
		return nil, err
	}
	return &sf2Core{
		syn:        syn,
		masterGain: masterGain,
		// Distinct seed offset so the mutation stream doesn't correlate with
		// the note-generation stream (which would make mutations feel
		// "predictable" alongside the initial figure).
		rng:        rand.New(rand.NewSource(mutationSeed ^ 0x6D757461)), //nolint:gosec // ASCII "muta"
		eqLowL:     synth.NewLowShelf(180, 2.5, 0.707),
		eqLowR:     synth.NewLowShelf(180, 2.5, 0.707),
		eqHighL:    synth.NewHighShelf(7500, 3.0, 0.707),
		eqHighR:    synth.NewHighShelf(7500, 3.0, 0.707),
		comp:       synth.NewStereoCompressor(-14, 3.0, 8, 250, 6, 4),
	}, nil
}

// setProgram changes the GM program on a MIDI channel.
func (e *sf2Core) setProgram(channel int32, program int32) {
	const ccProgramChange = 0xC0
	e.syn.ProcessMidiMessage(channel, ccProgramChange, program, 0)
}

// programSwap rotates the channel to a different GM program, chosen from
// the given list. Used by algorithms' macro mutation timer to occasionally
// swap timbres for long-form variety. Currently playing notes keep their
// old timbre (program changes only affect subsequent NoteOns).
func (e *sf2Core) programSwap(channel int32, alternatives []int32, rng *rand.Rand) {
	if len(alternatives) == 0 || rng == nil {
		return
	}
	e.setProgram(channel, alternatives[rng.Intn(len(alternatives))])
}

// setPan positions a MIDI channel in the stereo field. pan is 0..127 where
// 0 is full-left, 64 is center, 127 is full-right (MIDI CC 10 standard).
func (e *sf2Core) setPan(channel int32, pan int32) {
	const ccControlChange = 0xB0
	const ccPan = 10
	if pan < 0 {
		pan = 0
	}
	if pan > 127 {
		pan = 127
	}
	e.syn.ProcessMidiMessage(channel, ccControlChange, ccPan, pan)
}

// setMasterLowpass installs a stereo low-pass filter at the end of the master
// bus, giving the output a "muffled tape" character. Pass 0 hz to disable.
func (e *sf2Core) setMasterLowpass(cutoffHz, q float64) {
	if cutoffHz <= 0 {
		e.lpL = nil
		e.lpR = nil
		return
	}
	e.lpL = synth.NewLowpass(cutoffHz, q)
	e.lpR = synth.NewLowpass(cutoffHz, q)
}

// setTapeHiss installs a low-amplitude noise floor on the master output.
// level is the linear amplitude (e.g. 0.005 ≈ -46 dBFS, just audible
// behind quiet passages, inaudible behind loud ones).
func (e *sf2Core) setTapeHiss(level float64) {
	if level < 0 {
		level = 0
	}
	e.hissLevel = level
}

// setConvolutionIR installs a convolution reverb on the master bus. The IR is
// shared across both channels (mono → stereo via two independent convolver
// instances seeded from the same IR). wet is the mix level in [0, 1]; pass
// nil ir or 0 wet to disable.
func (e *sf2Core) setConvolutionIR(ir []float64, wet float64) {
	if len(ir) == 0 || wet <= 0 {
		e.convL = nil
		e.convR = nil
		e.convWet = 0
		return
	}
	if wet > 1 {
		wet = 1
	}
	// Normalize so convolved output stays in a reasonable level range —
	// otherwise a long, dense IR sums to a much louder signal than the dry.
	var sumSq float64
	for _, x := range ir {
		sumSq += x * x
	}
	norm := 1.0
	if sumSq > 0 {
		// Cube root: fully normalizing a dense 1-second IR makes it
		// inaudibly quiet; this gives a perceptually balanced level.
		norm = math.Pow(1.0/sumSq, 1.0/3.0)
	}
	scaled := make([]float64, len(ir))
	for i, x := range ir {
		scaled[i] = x * norm
	}
	// Pick an implementation based on IR length. For short IRs (≤1024 samples
	// ≈ 23 ms at 44.1 kHz) direct convolution is faster than FFT-based and
	// adds zero latency. For longer IRs the FFT version is essential —
	// direct convolution's O(N) per-sample cost becomes prohibitive.
	const fftThreshold = 1024
	const fftBlockSize = 512
	if len(scaled) <= fftThreshold {
		e.convL = synth.NewConvolver(scaled)
		e.convR = synth.NewConvolver(scaled)
	} else {
		e.convL = synth.NewFFTConvolver(scaled, fftBlockSize)
		e.convR = synth.NewFFTConvolver(scaled, fftBlockSize)
	}
	e.convWet = wet
}

// addTrack registers a cycling-note track with the engine. nextFireT is
// initialized to 0, which (combined with curSlot=-1) means "fire the slot
// matching the current time on the very next render call."
func (e *sf2Core) addTrack(t SF2Track) {
	period := int64(t.PeriodSec * float64(synth.SampleRate))
	if period < 1 {
		period = 1
	}
	state := &sf2TrackState{
		cfg:           t,
		periodSamples: period,
		phaseOffset:   int64(t.Phase01 * float64(period)),
		notesLen:      int64(len(t.Notes)),
		curSlot:       -1,
		curKey:        -1,
		nextFireT:     0,
	}
	e.tracks = append(e.tracks, state)
}

// samplesUntilNextSlot returns how many samples until this track's next
// NoteOn should fire. Returns 0 if the next fire is overdue.
func (s *sf2TrackState) samplesUntilNextSlot(t int64) int64 {
	if t >= s.nextFireT {
		return 0
	}
	return s.nextFireT - t
}

// fireTransition fires the NoteOn for the slot the track is currently in
// (computed from `t`), then schedules the next fire time at the next slot
// boundary plus a random timing-jitter offset.
//
// Velocity is jittered if VelocityJitter > 0. Also optionally re-rolls one
// OTHER slot's note so the cycled material gradually evolves.
func (s *sf2TrackState) fireTransition(t int64, syn *meltysynth.Synthesizer, rng *rand.Rand) {
	// Compute the slot we're firing — based on t, not the previous slot.
	// With timing jitter the fire time may be just past the natural slot
	// boundary, so the new slot index is what's at time t.
	phase := (t + s.phaseOffset) % s.periodSamples
	newSlot := int(phase * s.notesLen / s.periodSamples)
	if newSlot < 0 {
		newSlot = 0
	}
	if int64(newSlot) >= s.notesLen {
		newSlot = int(s.notesLen) - 1
	}

	if s.curKey >= 0 {
		syn.NoteOff(s.cfg.Channel, int32(s.curKey))
	}
	key := s.cfg.Notes[newSlot]
	vel := s.cfg.Velocity
	if s.cfg.VelocityJitter > 0 && rng != nil {
		offset := int32(rng.Intn(int(2*s.cfg.VelocityJitter)+1)) - s.cfg.VelocityJitter
		vel += offset
		if vel < 1 {
			vel = 1
		}
		if vel > 127 {
			vel = 127
		}
	}
	syn.NoteOn(s.cfg.Channel, int32(key), vel)
	s.curSlot = newSlot
	s.curKey = key

	// Schedule the next fire = natural boundary + timing jitter.
	// Natural boundary of the next slot is at the smallest phase value
	// where (phase * notesLen / periodSamples) >= (newSlot + 1), i.e.
	// ceil((newSlot+1) * periodSamples / notesLen).
	nextSlotStart := (int64(newSlot+1)*s.periodSamples + s.notesLen - 1) / s.notesLen
	naturalBoundary := t + (nextSlotStart - phase)

	jitter := int64(0)
	if s.cfg.TimingJitterSec > 0 && rng != nil {
		jSamples := int64(s.cfg.TimingJitterSec * float64(synth.SampleRate))
		if jSamples > 0 {
			jitter = rng.Int63n(2*jSamples+1) - jSamples
		}
	}
	s.nextFireT = naturalBoundary + jitter
	if s.nextFireT <= t {
		// Jitter pulled the fire time into the past — clamp to "very next
		// sample" so we always make forward progress.
		s.nextFireT = t + 1
	}

	if s.cfg.MutateOne != nil && s.cfg.MutationRate > 0 && len(s.cfg.Notes) > 1 && rng != nil {
		if rng.Float64() < s.cfg.MutationRate {
			victim := rng.Intn(len(s.cfg.Notes) - 1)
			if victim >= newSlot {
				victim++
			}
			s.cfg.Notes[victim] = s.cfg.MutateOne(victim, s.cfg.Notes[victim])
		}
	}
}

// renderInto fills the stereo float64 buffer by alternately rendering audio
// chunks and firing track events at slot boundaries, then applying the
// master bus.
func (e *sf2Core) renderInto(left, right []float64) {
	n := len(left)
	if cap(e.bufF32L) < n {
		e.bufF32L = make([]float32, n)
		e.bufF32R = make([]float32, n)
	}
	e.bufF32L = e.bufF32L[:n]
	e.bufF32R = e.bufF32R[:n]

	pos := 0
	for pos < n {
		// Find the smallest number of samples until the next event across
		// all tracks. Render that many samples, fire events, repeat.
		ahead := int64(n - pos)
		for _, s := range e.tracks {
			d := s.samplesUntilNextSlot(e.t)
			if d < ahead {
				ahead = d
			}
		}
		if ahead > 0 {
			e.syn.Render(e.bufF32L[pos:pos+int(ahead)], e.bufF32R[pos:pos+int(ahead)])
			e.t += ahead
			pos += int(ahead)
		}
		if pos < n {
			for _, s := range e.tracks {
				if s.samplesUntilNextSlot(e.t) == 0 {
					s.fireTransition(e.t, e.syn, e.rng)
				}
			}
		}
	}

	// Master bus: gain → EQ → optional conv wet → optional LP → optional hiss
	//              → compressor → soft-clip.
	for i := 0; i < n; i++ {
		l := float64(e.bufF32L[i]) * e.masterGain
		r := float64(e.bufF32R[i]) * e.masterGain
		l = e.eqLowL.Tick(l)
		r = e.eqLowR.Tick(r)
		l = e.eqHighL.Tick(l)
		r = e.eqHighR.Tick(r)
		if e.convL != nil {
			wetL := e.convL.Tick(l)
			wetR := e.convR.Tick(r)
			l += wetL * e.convWet
			r += wetR * e.convWet
		}
		if e.lpL != nil {
			l = e.lpL.Tick(l)
			r = e.lpR.Tick(r)
		}
		if e.hissLevel > 0 {
			// Stereo-decorrelated white noise — independent samples per channel.
			l += (e.rng.Float64()*2 - 1) * e.hissLevel
			r += (e.rng.Float64()*2 - 1) * e.hissLevel
		}
		l, r = e.comp.Tick(l, r)
		left[i] = synth.SoftClip(l)
		right[i] = synth.SoftClip(r)
	}
}

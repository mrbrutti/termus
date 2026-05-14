package gen

import (
	"github.com/sinshu/go-meltysynth/meltysynth"

	"github.com/mrbrutti/termus/internal/synth"
)

// SF2Track configures one cycling-note track inside an sf2Core engine.
// Multiple tracks can share a MIDI channel (e.g. several voices all on
// piano) — events from each track will sum naturally inside go-meltysynth's
// polyphony.
type SF2Track struct {
	Channel   int32 // MIDI channel 0..15
	Velocity  int32 // 0..127
	Notes     []int // MIDI keys, cycled
	PeriodSec float64
	Phase01   float64 // 0..1 phase offset within the period
}

// sf2TrackState is the runtime state for one track.
type sf2TrackState struct {
	cfg           SF2Track
	periodSamples int64
	phaseOffset   int64
	notesLen      int64 // len(cfg.Notes) cached as int64 to avoid conversion in hot path
	curSlot       int
	curKey        int // currently sounding key, or -1
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

	eqLowL, eqLowR   *synth.LowShelf
	eqHighL, eqHighR *synth.HighShelf
	comp             *synth.StereoCompressor

	bufF32L []float32
	bufF32R []float32
}

// newSF2Core constructs the engine and the master bus. masterGain compensates
// for go-meltysynth's conservative internal levels.
func newSF2Core(sf *meltysynth.SoundFont, masterGain float64) (*sf2Core, error) {
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

// addTrack registers a cycling-note track with the engine.
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
	}
	e.tracks = append(e.tracks, state)
}

// slotAt returns the slot index for the given absolute sample time. Uses the
// `int(phase * notesLen / period)` formula (same as gen.Eno's existing code)
// because integer division consistently rounds toward zero — there's no
// "phantom" slot at the end of the period that division-by-slotLen would
// produce.
func (s *sf2TrackState) slotAt(t int64) int {
	phase := (t + s.phaseOffset) % s.periodSamples
	return int(phase * s.notesLen / s.periodSamples)
}

// samplesUntilNextSlot returns how many samples until this track's slot
// changes again. Returns 0 if the slot has already changed and we need to
// fire an event now.
func (s *sf2TrackState) samplesUntilNextSlot(t int64) int64 {
	phase := (t + s.phaseOffset) % s.periodSamples
	slot := int(phase * s.notesLen / s.periodSamples)
	if slot != s.curSlot {
		return 0
	}
	// Next slot starts at the smallest phase d where
	//   (phase + d) * notesLen / periodSamples >= slot + 1
	// i.e.  phase + d >= ceil((slot+1) * periodSamples / notesLen)
	nextSlotStart := (int64(slot+1)*s.periodSamples + s.notesLen - 1) / s.notesLen
	return nextSlotStart - phase
}

// fireTransition sends NoteOff for the currently sounding key (if any) and
// NoteOn for the slot's note.
func (s *sf2TrackState) fireTransition(t int64, syn *meltysynth.Synthesizer) {
	newSlot := s.slotAt(t)
	if s.curKey >= 0 {
		syn.NoteOff(s.cfg.Channel, int32(s.curKey))
	}
	key := s.cfg.Notes[newSlot]
	syn.NoteOn(s.cfg.Channel, int32(key), s.cfg.Velocity)
	s.curSlot = newSlot
	s.curKey = key
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
					s.fireTransition(e.t, e.syn)
				}
			}
		}
	}

	// Master bus: gain → EQ → compressor → soft-clip.
	for i := 0; i < n; i++ {
		l := float64(e.bufF32L[i]) * e.masterGain
		r := float64(e.bufF32R[i]) * e.masterGain
		l = e.eqLowL.Tick(l)
		r = e.eqLowR.Tick(r)
		l = e.eqHighL.Tick(l)
		r = e.eqHighR.Tick(r)
		l, r = e.comp.Tick(l, r)
		left[i] = synth.SoftClip(l)
		right[i] = synth.SoftClip(r)
	}
}

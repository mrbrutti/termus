package gen

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/sinshu/go-meltysynth/meltysynth"

	"github.com/mrbrutti/termus/internal/synth"
)

const ccControlChange = 0xB0

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
	Notes     []int // MIDI keys, cycled. Negative values are treated as rests.
	PeriodSec float64
	Phase01   float64 // 0..1 phase offset within the period

	MutationRate float64
	MutateOne    func(slot int, prev int) int

	// ResolveNote, when non-nil, is called right before firing a slot. It can
	// remap a stored note to the current harmonic context without waiting for a
	// background mutation pass. Returning a negative value turns the slot into
	// a rest for this fire.
	ResolveNote func(slot int, prev int) int

	// Gate is the portion of the slot for which the note should be held.
	// 1.0 means "hold until the next slot boundary", 0.5 means "release
	// halfway through", and values >1.0 let the note ring into a following
	// rest or until the next note-on.
	Gate float64
	// ResolveGate overrides Gate for a particular slot.
	ResolveGate func(slot int, key int) float64
	// Legato keeps a note alive until the next slot transition if no earlier
	// release occurred, giving monophonic tracks connected phrasing.
	Legato bool
	// TieRepeats keeps an already-sounding note alive when the next slot
	// resolves to the same MIDI key, instead of re-articulating it.
	TieRepeats bool
	// OverlapSec lets a previous note ring briefly after the next note-on,
	// producing a small slur instead of an immediate cutover.
	OverlapSec float64
	// ReleaseSec adds a small fixed tail after the gate-calculated duration.
	ReleaseSec float64

	// ResolveVelocity can reshape a track's base velocity per slot before
	// random jitter is applied.
	ResolveVelocity func(slot int, key int, base int32) int32
	// ResolveExpression optionally supplies a per-note expression contour for
	// the channel (MIDI CC 11). It is best used on channels with one musical
	// voice, such as solo leads or single pads.
	ResolveExpression func(slot int, key int) SF2ExpressionCurve
	// ResolveModWheel optionally shapes MIDI CC 1 per note for vibrato /
	// modulation depth on SoundFonts that map mod-wheel musically.
	ResolveModWheel func(slot int, key int) SF2ExpressionCurve
	// ResolveBrightness optionally reshapes MIDI CC 74 per note so attacks
	// can bloom brighter than the sustain.
	ResolveBrightness func(slot int, key int) SF2ExpressionCurve
	// ResolveDetuneCents optionally applies a small per-note pitch bend,
	// expressed in cents and interpreted against the default +/-2 semitone
	// MIDI bend range.
	ResolveDetuneCents func(slot int, key int) int32

	VelocityJitter  int32   // ±range added to Velocity per NoteOn
	TimingJitterSec float64 // ±range added to fire time per NoteOn (seconds)

	// SwingAmount applies a systematic offset on odd-indexed slots,
	// pushing them later by SwingAmount * slotLength. 0 = straight time
	// (default). 0.12–0.18 = classic hip-hop / lofi shuffle. Distinct
	// from TimingJitterSec which is random; swing is deterministic and
	// applies to every odd slot.
	SwingAmount float64
	// ResolveTimingOffsetSec adds a deterministic per-slot timing offset on
	// top of swing, before random jitter. Negative values pull the slot
	// earlier; positive values lay it back.
	ResolveTimingOffsetSec func(slot int) float64

	// FireProbability, if > 0, randomly skips firing the slot's NoteOn.
	// 1.0 = always fire (default behavior when 0 — left unset acts as 1).
	// 0.8 = fire 80% of the time (drum ghost-note variety).
	// 0.5 = fire half the time (rhythmic sparsification).
	// The track still advances its slot pointer + nextFireT, so the rhythm
	// stays locked to the grid — only the audibility of each hit varies.
	FireProbability float64

	// Enabled, if non-nil, is checked by the engine on every slot
	// transition. When *Enabled is false, the track silently advances
	// through its slots without firing NoteOn — used by algorithms with
	// "section" structure to drop/add layers over time.
	Enabled *bool

	// OnFire is called by the engine right after each successful NoteOn
	// (i.e. not when the track was disabled). Used by sidechain
	// ducking — the chill kick uses this to gate-trigger the master-bus
	// duck envelope.
	OnFire func()
}

// SF2ExpressionCurve describes a simple channel-expression gesture across one
// note. Start is sent at NoteOn, Peak later in the note, and End near release.
type SF2ExpressionCurve struct {
	Start    int32
	Peak     int32
	End      int32
	PeakAt01 float64 // 0..1 fraction of the note duration; defaults to 0.35
}

type sf2EventSink interface {
	NoteOn(channel int32, key int32, velocity int32)
	NoteOff(channel int32, key int32)
	ProcessMidiMessage(channel int32, command int32, data1 int32, data2 int32)
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
	releaseT      int64
	overlapKey    int
	overlapOffT   int64
	exprNextT     int64
	exprStage     int
	exprCurve     SF2ExpressionCurve
	modNextT      int64
	modStage      int
	modCurve      SF2ExpressionCurve
	brightNextT   int64
	brightStage   int
	brightCurve   SF2ExpressionCurve
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
	primary       *sf2RenderEngine
	engines       []*sf2RenderEngine
	enginesByName map[string]*sf2RenderEngine
	channelEngine map[int32]*sf2RenderEngine
	tracks        []*sf2TrackState
	t             int64
	masterGain    float64
	rng           *rand.Rand // for mutation; lives on the audio goroutine only

	eqLowL, eqLowR   *synth.LowShelf
	eqHighL, eqHighR *synth.HighShelf
	comp             *synth.StereoCompressor

	// Optional master-bus low-pass — for the "muffled tape" feel of lofi.
	// nil disables it; the chain falls through without filtering.
	lpL, lpR *synth.Lowpass

	// Optional tape-hiss noise floor at the master bus output. 0 disables.
	hissLevel float64

	// Optional tape saturation amount, 0..1. Adds a soft odd-harmonic
	// curve to the master before final soft-clip — emulates tape
	// magnetization behavior.
	tapeSatAmount float64

	// Optional vinyl-crackle: chance per sample (e.g. 1e-4) of firing
	// a brief noise burst, and the duration in samples of each pop.
	crackleProb     float64
	crackleAmp      float64
	cracklePopSamps int64
	cracklePopLeft  int64 // remaining samples on current pop
	cracklePopVal   float64

	// Optional shared-type wow/flutter pitch modulator — lofi only.
	// Applied before the EQ so it modulates the full mix. nil = bypass.
	wowFlutter *synth.WowFlutter

	// Optional shared-type Tape saturator — replaces the legacy
	// tapeSatAmount path when non-nil. nil = fall through to legacy inline.
	sharedTape *synth.Tape

	// Optional shared-type Vinyl noise/crackle — replaces the legacy
	// crackleProb path when non-nil. nil = fall through to legacy inline.
	sharedVinyl *synth.Vinyl

	// Optional convolution reverb. When non-nil, applied in parallel with
	// the dry signal at convWet mix level. Each channel has its own
	// instance, both seeded from the same IR. nil disables convolution.
	convL, convR synth.RealtimeConvolver
	convWet      float64

	// Filter-cutoff LFOs — one per channel that wants modulation. The engine
	// emits MIDI CC 74 on each at ~20 Hz control rate.
	lfos []*filterLFO

	// Optional sidechain duck envelope — multiplied with the master output.
	// Triggered externally (e.g. by chill's kick OnFire). 1.0 = no duck,
	// dives to duckFloor on trigger, exponentially recovers.
	duckValue       float64
	duckFloor       float64 // bottom of the duck, e.g. 0.55 = -5 dB
	duckAttackCoef  float64 // per-sample multiplier during attack
	duckReleaseCoef float64 // per-sample multiplier during release
	duckState       int     // 0=idle, 1=attacking, 2=releasing

	bufF32L []float32
	bufF32R []float32

	capture       *midiCapture
	bootstrapMIDI []capturedMIDIMessage
}

type sf2RenderEngine struct {
	name string
	sf   *meltysynth.SoundFont
	syn  *meltysynth.Synthesizer
	bufL []float32
	bufR []float32
}

// configureSidechain installs a duck envelope. floorDB is the depth of the
// duck (negative dB, e.g. -5). attackMs is how fast it dives (1–5 ms is
// snappy; 8–20 ms is more "musical"). releaseMs is the recovery time
// (typically 150–400 ms for a noticeable but not annoying pump).
func (e *sf2Core) configureSidechain(floorDB, attackMs, releaseMs float64) {
	e.duckFloor = math.Pow(10, floorDB/20)
	e.duckValue = 1.0
	e.duckAttackCoef = math.Exp(-1.0 / (attackMs * 0.001 * float64(synth.SampleRate)))
	e.duckReleaseCoef = math.Exp(-1.0 / (releaseMs * 0.001 * float64(synth.SampleRate)))
	e.duckState = 0
}

// triggerDuck starts a new duck cycle. Called by an algorithm's OnFire
// callback (usually on a kick hit). Safe to call from the audio thread —
// it just resets the state machine.
func (e *sf2Core) triggerDuck() {
	if e.duckAttackCoef == 0 {
		return // not configured
	}
	e.duckState = 1 // attacking
}

// stepDuck advances the duck envelope by one sample and returns the
// current attenuation (1.0 = no duck).
func (e *sf2Core) stepDuck() float64 {
	if e.duckAttackCoef == 0 {
		return 1.0
	}
	switch e.duckState {
	case 1:
		// Exponential approach toward duckFloor.
		e.duckValue = e.duckAttackCoef*(e.duckValue-e.duckFloor) + e.duckFloor
		if e.duckValue <= e.duckFloor+0.005 {
			e.duckValue = e.duckFloor
			e.duckState = 2
		}
	case 2:
		// Exponential recovery toward 1.0.
		e.duckValue = e.duckReleaseCoef*(e.duckValue-1.0) + 1.0
		if e.duckValue >= 0.999 {
			e.duckValue = 1.0
			e.duckState = 0
		}
	}
	return e.duckValue
}

// newSF2Core constructs the engine and the master bus. masterGain compensates
// for go-meltysynth's conservative internal levels. mutationSeed seeds the
// mutation RNG; pass the same value Seed() uses for note generation to keep
// the mutation sequence deterministic with the rest of the algorithm.
func newSF2Core(sf *meltysynth.SoundFont, masterGain float64, mutationSeed int64) (*sf2Core, error) {
	primary, err := newSF2RenderEngine("primary", sf)
	if err != nil {
		return nil, err
	}
	core := &sf2Core{
		primary:       primary,
		engines:       []*sf2RenderEngine{primary},
		enginesByName: map[string]*sf2RenderEngine{"primary": primary},
		channelEngine: make(map[int32]*sf2RenderEngine),
		masterGain:    masterGain,
		// Distinct seed offset so the mutation stream doesn't correlate with
		// the note-generation stream (which would make mutations feel
		// "predictable" alongside the initial figure).
		rng:     rand.New(rand.NewSource(mutationSeed ^ 0x6D757461)), //nolint:gosec // ASCII "muta"
		eqLowL:  synth.NewLowShelf(180, 2.5, 0.707),
		eqLowR:  synth.NewLowShelf(180, 2.5, 0.707),
		eqHighL: synth.NewHighShelf(7500, 3.0, 0.707),
		eqHighR: synth.NewHighShelf(7500, 3.0, 0.707),
		comp:    synth.NewStereoCompressor(-14, 3.0, 8, 250, 6, 4),
	}
	runtime := currentSF2Runtime()
	if runtime.strategy == "max" {
		for name, loaded := range runtime.fonts {
			if loaded == nil {
				continue
			}
			if existing := core.engineBySoundFont(loaded); existing != nil {
				core.enginesByName[name] = existing
				if loaded == sf {
					core.primary = existing
				}
				continue
			}
			engine, err := newSF2RenderEngine(name, loaded)
			if err != nil {
				return nil, err
			}
			core.engines = append(core.engines, engine)
			core.enginesByName[name] = engine
			if loaded == sf {
				core.primary = engine
			}
		}
		if core.primary != nil {
			core.enginesByName["primary"] = core.primary
		}
	}
	return core, nil
}

func newSF2RenderEngine(name string, sf *meltysynth.SoundFont) (*sf2RenderEngine, error) {
	settings := meltysynth.NewSynthesizerSettings(synth.SampleRate)
	settings.EnableReverbAndChorus = true
	settings.MaximumPolyphony = 96
	syn, err := meltysynth.NewSynthesizer(sf, settings)
	if err != nil {
		return nil, err
	}
	return &sf2RenderEngine{name: name, sf: sf, syn: syn}, nil
}

func (e *sf2Core) engineBySoundFont(sf *meltysynth.SoundFont) *sf2RenderEngine {
	for _, engine := range e.engines {
		if engine != nil && engine.sf == sf {
			return engine
		}
	}
	return nil
}

type sf2SynthSink struct {
	core *sf2Core
}

func (s sf2SynthSink) NoteOn(channel int32, key int32, velocity int32) {
	s.core.noteOn(channel, key, velocity)
}

func (s sf2SynthSink) NoteOff(channel int32, key int32) {
	s.core.noteOff(channel, key)
}

func (s sf2SynthSink) ProcessMidiMessage(channel int32, command int32, data1 int32, data2 int32) {
	s.core.processMIDI(channel, command, data1, data2)
}

func (e *sf2Core) startMIDICapture() {
	e.capture = &midiCapture{
		events: append([]capturedMIDIMessage(nil), e.bootstrapMIDI...),
	}
}

func (e *sf2Core) finishMIDICapture() []capturedMIDIMessage {
	if e.capture == nil {
		return nil
	}
	events := e.capture.events
	e.capture = nil
	return events
}

func (e *sf2Core) recordMIDI(sample int64, status, data1, data2 byte) {
	msg := capturedMIDIMessage{
		Sample: sample,
		Status: status,
		Data1:  data1,
		Data2:  data2,
	}
	if e.capture != nil {
		e.capture.events = append(e.capture.events, msg)
		return
	}
	if e.t == 0 {
		e.bootstrapMIDI = append(e.bootstrapMIDI, msg)
	}
}

func (e *sf2Core) processMIDIAt(sample int64, channel int32, command int32, data1 int32, data2 int32) {
	e.engineForChannel(channel).syn.ProcessMidiMessage(channel, command, data1, data2)
	e.recordMIDI(sample, byte(command)|byte(channel&0x0F), byte(data1), byte(data2))
}

func (e *sf2Core) processMIDI(channel int32, command int32, data1 int32, data2 int32) {
	e.processMIDIAt(e.t, channel, command, data1, data2)
}

func (e *sf2Core) noteOn(channel int32, key int32, velocity int32) {
	e.engineForChannel(channel).syn.NoteOn(channel, key, velocity)
	e.recordMIDI(e.t, 0x90|byte(channel&0x0F), byte(key), byte(velocity))
}

func (e *sf2Core) noteOff(channel int32, key int32) {
	e.engineForChannel(channel).syn.NoteOff(channel, key)
	e.recordMIDI(e.t, 0x80|byte(channel&0x0F), byte(key), 0)
}

func (e *sf2Core) engineForChannel(channel int32) *sf2RenderEngine {
	if engine, ok := e.channelEngine[channel]; ok && engine != nil {
		return engine
	}
	return e.primary
}

func (e *sf2Core) routeChannelPreset(channel int32, preset string) {
	if preset == "" {
		delete(e.channelEngine, channel)
		return
	}
	engine, ok := e.enginesByName[preset]
	if !ok || engine == nil || len(e.engines) <= 1 {
		return
	}
	e.channelEngine[channel] = engine
}

func (e *sf2Core) usingMaxPalette() bool {
	return len(e.engines) > 1
}

// setProgram changes the GM program on a MIDI channel.
func (e *sf2Core) setProgram(channel int32, program int32) {
	const ccProgramChange = 0xC0
	e.processMIDI(channel, ccProgramChange, program, 0)
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

// setReverbSend sets the channel's send level to the synthesizer's internal
// reverb (MIDI CC 91). 0 = dry, 127 = full wet. Used per-instrument so e.g.
// a lead voice can be drenched in reverb while the bass stays dry.
func (e *sf2Core) setReverbSend(channel, level int32) {
	const ccControlChange = 0xB0
	const ccReverbSend = 91
	if level < 0 {
		level = 0
	}
	if level > 127 {
		level = 127
	}
	e.processMIDI(channel, ccControlChange, ccReverbSend, level)
}

// setChorusSend sets the channel's send to the internal chorus (CC 93).
// Same conventions as setReverbSend. Good for thickening pads or for
// adding the classic "Rhodes chorus" feel.
func (e *sf2Core) setChorusSend(channel, level int32) {
	const ccControlChange = 0xB0
	const ccChorusSend = 93
	if level < 0 {
		level = 0
	}
	if level > 127 {
		level = 127
	}
	e.processMIDI(channel, ccControlChange, ccChorusSend, level)
}

// filterLFO holds the state for a slow CC-74 (filter cutoff) LFO on one
// MIDI channel. Sustained pads especially benefit — without modulation
// they sound static, with it they "breathe."
type filterLFO struct {
	channel    int32
	rateHz     float64 // typical 0.05–0.25 Hz (4–20 second cycle)
	depth      int32   // ±range of CC74 values; total swing = 2*depth
	center     int32   // 0..127, midpoint of the swing
	phaseSamps int64   // running phase counter in samples
	periodSamp int64   // samples per full cycle
	emitEvery  int64   // emit a CC74 every N samples (e.g. ~ 50 ms)
	emitAcc    int64
}

// addFilterLFO installs a slow filter-cutoff LFO on the given channel.
// rateHz is the LFO frequency (a 4–20 second cycle is typical for ambient).
// center 64 is mid-range; depth 30 swings ±30 around center (34..94).
func (e *sf2Core) addFilterLFO(channel int32, rateHz, center, depth float64) {
	if rateHz <= 0 {
		return
	}
	period := int64(float64(synth.SampleRate) / rateHz)
	if period < 1 {
		period = 1
	}
	lfo := &filterLFO{
		channel:    channel,
		rateHz:     rateHz,
		depth:      int32(depth),
		center:     int32(center),
		periodSamp: period,
		emitEvery:  int64(float64(synth.SampleRate) * 0.05), // 20 Hz control rate
	}
	e.lfos = append(e.lfos, lfo)
}

// updateLFOs advances every installed filter LFO by samples and emits MIDI
// CC 74 messages at the configured rate. Called once per render block.
func (e *sf2Core) updateLFOs(samples int64) {
	const ccControlChange = 0xB0
	const ccBrightness = 74 // CC 74 = filter cutoff / brightness
	for _, l := range e.lfos {
		l.phaseSamps = (l.phaseSamps + samples) % l.periodSamp
		l.emitAcc += samples
		if l.emitAcc < l.emitEvery {
			continue
		}
		l.emitAcc = 0
		// LFO value in -1..1 via sine.
		theta := 2 * math.Pi * float64(l.phaseSamps) / float64(l.periodSamp)
		v := math.Sin(theta)
		cc := int32(float64(l.center) + v*float64(l.depth))
		if cc < 0 {
			cc = 0
		}
		if cc > 127 {
			cc = 127
		}
		e.processMIDIAt(e.t, l.channel, ccControlChange, ccBrightness, cc)
	}
}

// setChannelCutoff sets a static value for MIDI CC 74 (filter cutoff /
// brightness) on the given channel. Use this to darken specific instruments
// down to genre-appropriate territory — lofi piano and Rhodes especially
// benefit from a very low static cutoff (CC 74 ≈ 30–40) which dramatically
// muffles the upper harmonics in the way that classic lofi recordings do.
//
// If you also installed a filter LFO on the same channel via addFilterLFO,
// the LFO's emits will overwrite this static value — the LFO's center
// parameter takes over the same role.
func (e *sf2Core) setChannelCutoff(channel, value int32) {
	const ccControlChange = 0xB0
	const ccBrightness = 74
	if value < 0 {
		value = 0
	}
	if value > 127 {
		value = 127
	}
	e.processMIDI(channel, ccControlChange, ccBrightness, value)
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
	e.processMIDI(channel, ccControlChange, ccPan, pan)
}

func (e *sf2Core) setChannelExpression(channel, value int32) {
	const ccControlChange = 0xB0
	const ccExpression = 11
	if value < 0 {
		value = 0
	}
	if value > 127 {
		value = 127
	}
	e.processMIDI(channel, ccControlChange, ccExpression, value)
}

func (e *sf2Core) setPitchBend(channel, value int32) {
	const pitchBend = 0xE0
	if value < 0 {
		value = 0
	}
	if value > 16383 {
		value = 16383
	}
	lsb := value & 0x7F
	msb := (value >> 7) & 0x7F
	e.processMIDI(channel, pitchBend, lsb, msb)
}

// setMasterEQ overrides the default master-bus shelf EQ. The engine builds
// in a +2.5 dB low shelf at 180 Hz and a +3 dB high shelf at 7.5 kHz as
// sensible defaults — algorithms can pass different parameters to dial in
// their specific character (e.g. chill wants a high-shelf CUT, not boost,
// so the tape lowpass isn't fighting an air-boost above its corner).
func (e *sf2Core) setMasterEQ(lowHz, lowDB, highHz, highDB float64) {
	e.eqLowL = synth.NewLowShelf(lowHz, lowDB, 0.707)
	e.eqLowR = synth.NewLowShelf(lowHz, lowDB, 0.707)
	e.eqHighL = synth.NewHighShelf(highHz, highDB, 0.707)
	e.eqHighR = synth.NewHighShelf(highHz, highDB, 0.707)
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

// setTapeSaturation installs a soft-saturation shaper on the master bus
// before the final SoftClip. amount in 0..1: 0 = none, 1 = strong tape-y
// compression of peaks. Approximates analog-tape magnetization curve via a
// scaled tanh. For lofi character, 0.20–0.45 sounds right.
func (e *sf2Core) setTapeSaturation(amount float64) {
	if amount < 0 {
		amount = 0
	}
	if amount > 1 {
		amount = 1
	}
	e.tapeSatAmount = amount
}

// setVinylCrackle installs a vinyl-record-style crackle layer on the master
// output. popsPerSec is the average pop rate (e.g. 12 = somewhat dusty
// record, 40 = very dusty). amp is the linear amplitude of each pop
// (e.g. 0.06). popMs is the duration each pop holds before decaying
// (typical real-vinyl pop: 0.5–2 ms).
func (e *sf2Core) setVinylCrackle(popsPerSec, amp, popMs float64) {
	if popsPerSec <= 0 || amp <= 0 {
		e.crackleProb = 0
		e.crackleAmp = 0
		return
	}
	e.crackleProb = popsPerSec / float64(synth.SampleRate)
	e.crackleAmp = amp
	e.cracklePopSamps = int64(popMs * 0.001 * float64(synth.SampleRate))
	if e.cracklePopSamps < 1 {
		e.cracklePopSamps = 1
	}
}

// stepCrackle returns the current crackle sample (or 0 when idle). Each
// pop is a brief burst at random sign with exponential decay over
// cracklePopSamps samples.
func (e *sf2Core) stepCrackle() float64 {
	if e.crackleProb == 0 {
		return 0
	}
	if e.cracklePopLeft > 0 {
		v := e.cracklePopVal
		e.cracklePopLeft--
		// Decay toward zero for a click-and-fade rather than a square pulse.
		e.cracklePopVal *= 0.6
		return v
	}
	if e.rng.Float64() < e.crackleProb {
		// Start a new pop with random sign + random amplitude within ±amp.
		sign := 1.0
		if e.rng.Float64() < 0.5 {
			sign = -1.0
		}
		e.cracklePopVal = sign * e.crackleAmp * (0.4 + 0.6*e.rng.Float64())
		e.cracklePopLeft = e.cracklePopSamps - 1
		return e.cracklePopVal
	}
	return 0
}

// setWowFlutter installs a WowFlutter pitch modulator on the master bus.
// Applied before the EQ so the full mix is pitch-modulated. nil or a zero
// config disables it. Intended for lofi only — other styles leave this nil.
func (e *sf2Core) setWowFlutter(cfg synth.WowFlutterConfig) {
	e.wowFlutter = synth.NewWowFlutter(float64(synth.SampleRate), cfg)
}

// setSharedTape installs a synth.Tape saturator on the master bus, replacing
// the legacy inline tapeSatAmount path. Call with DriveDB 0 to disable.
func (e *sf2Core) setSharedTape(cfg synth.TapeConfig) {
	if cfg.DriveDB == 0 {
		e.sharedTape = nil
		return
	}
	e.sharedTape = synth.NewTape(cfg)
}

// setSharedVinyl installs a synth.Vinyl noise/crackle generator on the master
// bus, replacing the legacy inline crackleProb path. Nil disables it.
func (e *sf2Core) setSharedVinyl(sampleRate float64, cfg synth.VinylConfig) {
	e.sharedVinyl = synth.NewVinyl(sampleRate, cfg)
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

// setNamedConvolutionIR loads an IR preset by name from synth.IRLibrary and
// installs it on the master bus at the given wet level. Returns an error if
// the name doesn't exist. The IR is generated deterministically for the given
// seed; pass 0 to use the package default.
//
// This entry point is intended for SP5/SP6 style wiring that selects reverb
// by character name (e.g. "jazz_club") rather than by raw IR buffer.
func (e *sf2Core) setNamedConvolutionIR(name string, sampleRate float64, seed int64, wet float64) error {
	preset := synth.IRByName(name)
	if preset == nil {
		return fmt.Errorf("sf2Core: IR preset %q not found in library", name)
	}
	ir := preset.Generate(sampleRate, seed)
	e.setConvolutionIR(ir, wet)
	return nil
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
		overlapKey:    -1,
		nextFireT:     0,
	}
	e.tracks = append(e.tracks, state)
}

// samplesUntilNextSlot returns how many samples until this track's next
// NoteOn should fire. Returns 0 if the next fire is overdue.
func (s *sf2TrackState) samplesUntilNextEvent(t int64) int64 {
	if t >= s.nextFireT {
		return 0
	}
	ahead := s.nextFireT - t
	if s.releaseT > 0 {
		if t >= s.releaseT {
			return 0
		}
		if d := s.releaseT - t; d < ahead {
			ahead = d
		}
	}
	if s.overlapOffT > 0 {
		if t >= s.overlapOffT {
			return 0
		}
		if d := s.overlapOffT - t; d < ahead {
			ahead = d
		}
	}
	if s.exprNextT > 0 {
		if t >= s.exprNextT {
			return 0
		}
		if d := s.exprNextT - t; d < ahead {
			ahead = d
		}
	}
	if s.modNextT > 0 {
		if t >= s.modNextT {
			return 0
		}
		if d := s.modNextT - t; d < ahead {
			ahead = d
		}
	}
	if s.brightNextT > 0 {
		if t >= s.brightNextT {
			return 0
		}
		if d := s.brightNextT - t; d < ahead {
			ahead = d
		}
	}
	return ahead
}

// fireTransition fires the NoteOn for the slot the track is currently in
// (computed from `t`), then schedules the next fire time at the next slot
// boundary plus a random timing-jitter offset.
//
// Velocity is jittered if VelocityJitter > 0. Also optionally re-rolls notes
// so the cycled material gradually evolves.
func (s *sf2TrackState) fireTransition(t int64, syn sf2EventSink, rng *rand.Rand) {
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

	// Skip firing this slot for either of:
	//   • track is section-disabled
	//   • FireProbability set and we rolled a skip (drum ghost-note variety)
	skip := false
	if s.cfg.Enabled != nil && !*s.cfg.Enabled {
		skip = true
	}
	if !skip && s.cfg.FireProbability > 0 && s.cfg.FireProbability < 1 && rng != nil {
		if rng.Float64() >= s.cfg.FireProbability {
			skip = true
		}
	}
	key := s.cfg.Notes[newSlot]
	if s.cfg.ResolveNote != nil {
		key = s.cfg.ResolveNote(newSlot, key)
		s.cfg.Notes[newSlot] = key
	}
	if key < 0 {
		skip = true
	}

	slotLen := s.periodSamples / s.notesLen
	if slotLen < 1 {
		slotLen = 1
	}
	hold := s.holdSamples(newSlot, key, slotLen)

	if skip {
		s.curSlot = newSlot
		// Let any currently ringing note release naturally; skips no longer
		// force a hard note-off at the slot boundary.
	} else {
		tie := s.cfg.Legato && s.cfg.TieRepeats && s.curKey >= 0 && s.curKey == key
		if tie {
			s.curSlot = newSlot
			if nextRelease := t + hold; nextRelease > s.releaseT {
				s.releaseT = nextRelease
			}
			if s.cfg.ResolveExpression != nil {
				s.updateTiedExpression(newSlot, key, t)
			}
			if s.cfg.ResolveModWheel != nil {
				s.updateTiedControlCurve(newSlot, key, t, &s.modCurve, &s.modStage, &s.modNextT, s.cfg.ResolveModWheel)
			}
			if s.cfg.ResolveBrightness != nil {
				s.updateTiedControlCurve(newSlot, key, t, &s.brightCurve, &s.brightStage, &s.brightNextT, s.cfg.ResolveBrightness)
			}
		} else {
			if s.curKey >= 0 {
				s.releaseCurrentForNext(t, syn, slotLen)
			}
			vel := s.cfg.Velocity
			if s.cfg.ResolveVelocity != nil {
				vel = s.cfg.ResolveVelocity(newSlot, key, vel)
			}
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
			s.releaseT = t + hold
			s.exprNextT = 0
			s.exprStage = 0
			s.modNextT = 0
			s.modStage = 0
			s.brightNextT = 0
			s.brightStage = 0
			if s.cfg.ResolveExpression != nil {
				s.startExpression(newSlot, key, hold, t, syn)
			}
			if s.cfg.ResolveModWheel != nil {
				s.startControlCurve(newSlot, key, hold, t, syn, 1, 0, &s.modCurve, &s.modStage, &s.modNextT, s.cfg.ResolveModWheel)
			}
			if s.cfg.ResolveBrightness != nil {
				s.startControlCurve(newSlot, key, hold, t, syn, 74, 96, &s.brightCurve, &s.brightStage, &s.brightNextT, s.cfg.ResolveBrightness)
			}
			if s.cfg.ResolveDetuneCents != nil {
				cents := s.cfg.ResolveDetuneCents(newSlot, key)
				syn.ProcessMidiMessage(s.cfg.Channel, 0xE0, pitchBendLSB(cents), pitchBendMSB(cents))
			}
			if s.cfg.OnFire != nil {
				s.cfg.OnFire()
			}
		}
	}

	// Schedule the next fire = natural boundary + swing + timing jitter.
	// Natural boundary of the next slot is at the smallest phase value
	// where (phase * notesLen / periodSamples) >= (newSlot + 1), i.e.
	// ceil((newSlot+1) * periodSamples / notesLen).
	nextSlotStart := (int64(newSlot+1)*s.periodSamples + s.notesLen - 1) / s.notesLen
	naturalBoundary := t + (nextSlotStart - phase)
	slotLen = s.periodSamples / s.notesLen

	// Swing: odd-indexed slots fire later. This is the systematic shuffle
	// that turns "straight 8ths" into the lofi/hip-hop groove.
	if s.cfg.SwingAmount > 0 {
		nextSlotIdx := newSlot + 1
		if int64(nextSlotIdx) >= s.notesLen {
			nextSlotIdx = 0
		}
		if nextSlotIdx%2 == 1 {
			naturalBoundary += int64(s.cfg.SwingAmount * float64(slotLen))
		}
		if s.cfg.ResolveTimingOffsetSec != nil {
			naturalBoundary += secondsToSamples(s.cfg.ResolveTimingOffsetSec(nextSlotIdx))
		}
	}
	if s.cfg.SwingAmount == 0 && s.cfg.ResolveTimingOffsetSec != nil {
		nextSlotIdx := newSlot + 1
		if int64(nextSlotIdx) >= s.notesLen {
			nextSlotIdx = 0
		}
		naturalBoundary += secondsToSamples(s.cfg.ResolveTimingOffsetSec(nextSlotIdx))
	}

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

	if s.cfg.MutateOne != nil && s.cfg.MutationRate > 0 && len(s.cfg.Notes) > 0 && rng != nil {
		if rng.Float64() < s.cfg.MutationRate {
			victim := 0
			if len(s.cfg.Notes) == 1 {
				victim = 0
			} else {
				victim = rng.Intn(len(s.cfg.Notes) - 1)
				if victim >= newSlot {
					victim++
				}
			}
			s.cfg.Notes[victim] = s.cfg.MutateOne(victim, s.cfg.Notes[victim])
		}
	}
}

func (s *sf2TrackState) handleDueEvents(t int64, syn sf2EventSink, rng *rand.Rand) {
	if s.overlapOffT > 0 && t >= s.overlapOffT {
		if s.overlapKey >= 0 {
			syn.NoteOff(s.cfg.Channel, int32(s.overlapKey))
			s.overlapKey = -1
		}
		s.overlapOffT = 0
	}
	if s.releaseT > 0 && t >= s.releaseT {
		if s.curKey >= 0 {
			syn.NoteOff(s.cfg.Channel, int32(s.curKey))
			s.curKey = -1
			syn.ProcessMidiMessage(s.cfg.Channel, 0xE0, pitchBendLSB(0), pitchBendMSB(0))
		}
		s.releaseT = 0
		s.exprNextT = 0
		s.exprStage = 0
		s.modNextT = 0
		s.modStage = 0
		s.brightNextT = 0
		s.brightStage = 0
	}
	if s.exprNextT > 0 && t >= s.exprNextT {
		s.handleControlCurve(t, syn, 11, &s.exprCurve, &s.exprStage, &s.exprNextT)
	}
	if s.modNextT > 0 && t >= s.modNextT {
		s.handleControlCurve(t, syn, 1, &s.modCurve, &s.modStage, &s.modNextT)
	}
	if s.brightNextT > 0 && t >= s.brightNextT {
		s.handleControlCurve(t, syn, 74, &s.brightCurve, &s.brightStage, &s.brightNextT)
	}
	if s.nextFireT <= t {
		s.fireTransition(t, syn, rng)
	}
}

func (s *sf2TrackState) holdSamples(slot, key int, slotLen int64) int64 {
	gate := s.cfg.Gate
	if gate <= 0 {
		gate = 1.0
	}
	if s.cfg.ResolveGate != nil {
		if g := s.cfg.ResolveGate(slot, key); g > 0 {
			gate = g
		}
	}
	hold := int64(gate * float64(slotLen))
	hold += secondsToSamples(s.cfg.ReleaseSec)
	if hold < 1 {
		hold = 1
	}
	if s.cfg.Legato && hold < slotLen {
		hold = slotLen
	}
	return hold
}

func (s *sf2TrackState) releaseCurrentForNext(t int64, syn sf2EventSink, slotLen int64) {
	if s.curKey < 0 {
		return
	}
	if s.overlapKey >= 0 {
		syn.NoteOff(s.cfg.Channel, int32(s.overlapKey))
		s.overlapKey = -1
		s.overlapOffT = 0
	}
	overlap := secondsToSamples(s.cfg.OverlapSec)
	if overlap > slotLen/2 {
		overlap = slotLen / 2
	}
	if overlap > 0 {
		s.overlapKey = s.curKey
		s.overlapOffT = t + overlap
	} else {
		syn.NoteOff(s.cfg.Channel, int32(s.curKey))
	}
	s.curKey = -1
	s.releaseT = 0
	s.exprNextT = 0
	s.exprStage = 0
	s.modNextT = 0
	s.modStage = 0
	s.brightNextT = 0
	s.brightStage = 0
}

func (s *sf2TrackState) startExpression(slot, key int, hold, t int64, syn sf2EventSink) {
	s.startControlCurve(slot, key, hold, t, syn, 11, 96, &s.exprCurve, &s.exprStage, &s.exprNextT, s.cfg.ResolveExpression)
}

func (s *sf2TrackState) updateTiedExpression(slot, key int, t int64) {
	s.updateTiedControlCurve(slot, key, t, &s.exprCurve, &s.exprStage, &s.exprNextT, s.cfg.ResolveExpression)
}

func (s *sf2TrackState) startControlCurve(slot, key int, hold, t int64, syn sf2EventSink, control, defaultStart int32, curve *SF2ExpressionCurve, stage *int, nextT *int64, resolve func(int, int) SF2ExpressionCurve) {
	*curve = resolve(slot, key)
	if curve.Start <= 0 {
		curve.Start = defaultStart
	}
	if curve.Peak <= 0 {
		curve.Peak = curve.Start
	}
	if curve.End <= 0 {
		curve.End = curve.Peak
	}
	if curve.PeakAt01 <= 0 || curve.PeakAt01 >= 1 {
		curve.PeakAt01 = 0.35
	}
	syn.ProcessMidiMessage(s.cfg.Channel, ccControlChange, control, curve.Start)
	if curve.Peak != curve.Start {
		*stage = 1
		*nextT = t + int64(float64(hold)*curve.PeakAt01)
		if *nextT <= t {
			*nextT = t + 1
		}
	} else if curve.End != curve.Peak && s.releaseT > t+1 {
		*stage = 2
		*nextT = s.releaseT - 1
	}
}

func (s *sf2TrackState) updateTiedControlCurve(slot, key int, t int64, curve *SF2ExpressionCurve, stage *int, nextT *int64, resolve func(int, int) SF2ExpressionCurve) {
	if resolve == nil {
		return
	}
	nextCurve := resolve(slot, key)
	if nextCurve.End <= 0 {
		nextCurve.End = curve.End
	}
	curve.End = nextCurve.End
	if nextCurve.Peak > curve.Peak {
		curve.Peak = nextCurve.Peak
		if s.releaseT > t+1 {
			*stage = 1
			*nextT = t + int64(float64(s.releaseT-t)*nextCurve.PeakAt01)
			if *nextT <= t {
				*nextT = t + 1
			}
		}
		return
	}
	if s.releaseT > t+1 {
		*stage = 2
		*nextT = s.releaseT - 1
	}
}

func (s *sf2TrackState) handleControlCurve(t int64, syn sf2EventSink, control int32, curve *SF2ExpressionCurve, stage *int, nextT *int64) {
	switch *stage {
	case 1:
		syn.ProcessMidiMessage(s.cfg.Channel, ccControlChange, control, curve.Peak)
		if curve.End != curve.Peak && s.releaseT > t+1 {
			*stage = 2
			*nextT = s.releaseT - 1
		} else {
			*nextT = 0
			*stage = 0
		}
	case 2:
		syn.ProcessMidiMessage(s.cfg.Channel, ccControlChange, control, curve.End)
		*nextT = 0
		*stage = 0
	}
}

func pitchBendValue(cents int32) int32 {
	if cents > 200 {
		cents = 200
	}
	if cents < -200 {
		cents = -200
	}
	return 8192 + cents*8192/200
}

func pitchBendLSB(cents int32) int32 {
	return pitchBendValue(cents) & 0x7F
}

func pitchBendMSB(cents int32) int32 {
	return (pitchBendValue(cents) >> 7) & 0x7F
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
	sink := sf2SynthSink{core: e}
	for pos < n {
		// Find the smallest number of samples until the next event across
		// all tracks. Render that many samples, fire events, repeat.
		ahead := int64(n - pos)
		for _, s := range e.tracks {
			d := s.samplesUntilNextEvent(e.t)
			if d < ahead {
				ahead = d
			}
		}
		if ahead > 0 {
			if len(e.lfos) > 0 {
				e.updateLFOs(ahead)
			}
			renderLen := int(ahead)
			for i := 0; i < renderLen; i++ {
				e.bufF32L[pos+i] = 0
				e.bufF32R[pos+i] = 0
			}
			for _, engine := range e.engines {
				if engine == nil || engine.syn == nil {
					continue
				}
				if cap(engine.bufL) < renderLen {
					engine.bufL = make([]float32, renderLen)
					engine.bufR = make([]float32, renderLen)
				}
				engine.bufL = engine.bufL[:renderLen]
				engine.bufR = engine.bufR[:renderLen]
				engine.syn.Render(engine.bufL, engine.bufR)
				for i := 0; i < renderLen; i++ {
					e.bufF32L[pos+i] += engine.bufL[i]
					e.bufF32R[pos+i] += engine.bufR[i]
				}
			}
			e.t += ahead
			pos += renderLen
		}
		if pos < n {
			for _, s := range e.tracks {
				if s.samplesUntilNextEvent(e.t) == 0 {
					s.handleDueEvents(e.t, sink, e.rng)
				}
			}
		}
	}

	// Master bus: [wow/flutter] → gain → EQ → optional conv wet → optional LP
	//              → optional hiss → sidechain duck → tape sat → vinyl crackle
	//              → compressor → soft-clip.
	//
	// WowFlutter (lofi only) sits before gain so it modulates the full mix
	// at a consistent level; its fractional-delay line output is stable in
	// amplitude so inserting before gain is equivalent to after.
	for i := 0; i < n; i++ {
		l := float64(e.bufF32L[i])
		r := float64(e.bufF32R[i])
		// Optional wow/flutter pitch modulator — lofi only. Applied first so
		// the pitch modulation affects the complete pre-gain signal.
		if e.wowFlutter != nil {
			l, r = e.wowFlutter.Tick(l, r)
		}
		l *= e.masterGain
		r *= e.masterGain
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
		// Sidechain duck — multiply both channels by the duck envelope's
		// current attenuation. Idle = 1.0 (no effect).
		if e.duckAttackCoef != 0 {
			duck := e.stepDuck()
			l *= duck
			r *= duck
		}
		// Tape saturation — prefer the shared synth.Tape type when installed
		// (lofi uses it); fall back to the legacy inline path otherwise.
		if e.sharedTape != nil {
			l = e.sharedTape.Tick(l)
			r = e.sharedTape.Tick(r)
		} else if e.tapeSatAmount > 0 {
			drive := 1.0 + 1.5*e.tapeSatAmount
			l = math.Tanh(l*drive) / drive
			r = math.Tanh(r*drive) / drive
		}
		// Vinyl crackle — prefer the shared synth.Vinyl type when installed
		// (lofi uses it); fall back to the legacy inline path otherwise.
		if e.sharedVinyl != nil {
			vL, vR := e.sharedVinyl.Tick()
			l += vL
			r += vR
		} else if e.crackleProb > 0 {
			c := e.stepCrackle()
			l += c
			r += c
		}
		l, r = e.comp.Tick(l, r)
		left[i] = synth.SoftClip(l)
		right[i] = synth.SoftClip(r)
	}
}

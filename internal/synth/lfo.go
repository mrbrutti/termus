package synth

import (
	"math"
)

// LFOShape enumerates the modulator waveforms.
type LFOShape int

const (
	LFOSine       LFOShape = iota
	LFOTriangle            // bipolar triangle wave
	LFOSampleHold          // new uniform value at each positive-going phase crossing
	LFORandomWalk          // 1/f-ish brownian motion clamped to [-1, 1]
)

// LFOConfig configures an LFO. Output is in [-Depth, +Depth].
type LFOConfig struct {
	Shape     LFOShape
	RateHz    float64
	Depth     float64 // amplitude multiplier; default 1.0
	PhaseRad  float64 // initial phase in radians
	FadeInSec float64 // 0 = no fade-in
	DelaySec  float64 // delay before LFO output starts (vibrato delay)
	Seed      int64   // for SampleHold / RandomWalk
}

// LFO is a single-channel modulator. Call NewLFO then Tick once per audio
// sample; the returned value is the LFO output scaled to [-Depth, +Depth]
// subject to fade-in / delay.
type LFO struct {
	sampleRate float64
	cfg        LFOConfig

	phase       float64 // [0, 1)
	inc         float64 // phase increment per sample
	delaySamps  int     // samples remaining in delay period
	fadeSamps   int     // total fade-in length in samples
	fadeCurrent int     // samples elapsed since fade started

	// SampleHold state
	holdVal    float64
	prevPhase  float64
	rng        lcgRand

	// RandomWalk state
	walkVal float64
}

// lcgRand is a minimal 64-bit LCG pseudo-random generator that is seeded
// deterministically so LFO outputs are reproducible.
type lcgRand struct {
	state uint64
}

func newLCG(seed int64) lcgRand {
	return lcgRand{state: uint64(seed) ^ 0x3243f6a8885a308d}
}

// nextFloat returns a value in [0, 1).
func (r *lcgRand) nextFloat() float64 {
	r.state = r.state*6364136223846793005 + 1442695040888963407
	return float64(r.state>>11) / (1 << 53)
}

// NewLFO creates a ready-to-tick LFO.
func NewLFO(sampleRate float64, cfg LFOConfig) *LFO {
	l := &LFO{
		sampleRate: sampleRate,
		cfg:        cfg,
		rng:        newLCG(cfg.Seed),
	}
	l.reset()
	return l
}

func (l *LFO) reset() {
	l.inc = l.cfg.RateHz / l.sampleRate
	l.phase = l.cfg.PhaseRad / (2 * math.Pi)
	// Normalise phase to [0, 1).
	l.phase -= math.Floor(l.phase)

	l.delaySamps = int(l.cfg.DelaySec * l.sampleRate)
	l.fadeSamps = int(l.cfg.FadeInSec * l.sampleRate)
	l.fadeCurrent = 0

	l.prevPhase = l.phase
	l.holdVal = 0
	l.walkVal = 0
	l.rng = newLCG(l.cfg.Seed)
}

// Reset resets phase and all delay/fade counters to the initial state.
func (l *LFO) Reset() {
	l.reset()
}

// Tick advances the LFO by one sample and returns the current output.
func (l *LFO) Tick() float64 {
	// Delay period: output silence.
	if l.delaySamps > 0 {
		l.delaySamps--
		l.advancePhase()
		return 0
	}

	raw := l.rawValue()
	l.advancePhase()

	// Apply fade-in envelope.
	env := 1.0
	if l.fadeSamps > 0 && l.fadeCurrent < l.fadeSamps {
		env = float64(l.fadeCurrent) / float64(l.fadeSamps)
		l.fadeCurrent++
	}

	depth := l.cfg.Depth
	if depth == 0 {
		depth = 1.0
	}
	return raw * env * depth
}

// advancePhase increments the phase accumulator, handling SampleHold crossing
// detection before the increment so prevPhase / phase bracket the crossing.
func (l *LFO) advancePhase() {
	l.prevPhase = l.phase
	l.phase += l.inc
	if l.phase >= 1 {
		l.phase -= 1
		// Positive-going crossing — pick new SampleHold value.
		if l.cfg.Shape == LFOSampleHold {
			l.holdVal = l.rng.nextFloat()*2 - 1
		}
	}
}

// rawValue returns the unscaled, un-enveloped LFO sample for the current phase.
func (l *LFO) rawValue() float64 {
	switch l.cfg.Shape {
	case LFOSine:
		return math.Sin(l.phase * 2 * math.Pi)

	case LFOTriangle:
		p := l.phase
		if p < 0.25 {
			return 4 * p
		} else if p < 0.75 {
			return 2 - 4*p
		}
		return 4*p - 4

	case LFOSampleHold:
		return l.holdVal

	case LFORandomWalk:
		const sigma = 0.05
		l.walkVal += sigma * (l.rng.nextFloat() - 0.5)
		if l.walkVal > 1 {
			l.walkVal = 1
		}
		if l.walkVal < -1 {
			l.walkVal = -1
		}
		return l.walkVal
	}
	return 0
}

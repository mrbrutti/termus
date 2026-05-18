// internal/synth/personality.go
//
// Per-voice "personality" DSP primitives: NoiseBurst and PitchSag.
// These provide the pre-attack and post-release transients (key clicks,
// breath noise, mallet thumps, release tails) and pitch-sag envelopes that
// give synthesized voices acoustic character. Not wired into algorithms yet —
// these are standalone, unit-testable primitives (SP4).
package synth

import (
	"math"
	"math/rand"
)

// ──────────────────────────────────────────────────────────────────────────────
// NoiseColor
// ──────────────────────────────────────────────────────────────────────────────

// NoiseColor controls the spectral shape of a NoiseBurst.
type NoiseColor int

const (
	NoiseWhite     NoiseColor = iota
	NoiseLowpass              // one-pole LP at CutoffHz
	NoiseBandpass             // simple bandpass via LP + HP
	NoiseHighpass             // one-pole HP at CutoffHz
)

// ──────────────────────────────────────────────────────────────────────────────
// NoiseBurst
// ──────────────────────────────────────────────────────────────────────────────

// NoiseBurst is a one-shot filtered-noise envelope used for key clicks,
// mallet thumps, breath onsets, and finger releases. Call Trigger() to fire;
// the burst decays to silence and stays silent until the next Trigger().
type NoiseBurst struct {
	cfg    NoiseBurstConfig
	rng    *rand.Rand
	active bool

	// envelope state
	envPhase    int     // sample counter
	attackSamps int
	decaySamps  int
	envValue    float64

	// one-pole biquad filter state (used for all color modes)
	// LP / HP one-pole state
	lpZ float64 // one-pole LP memory
	hpZ float64 // one-pole HP memory (stores previous output)
	hpX float64 // one-pole HP memory (stores previous input)

	// For bandpass: chain LP then HP
	lpZbp float64
	hpZbp float64
	hpXbp float64
}

// NoiseBurstConfig configures a NoiseBurst.
type NoiseBurstConfig struct {
	Color      NoiseColor
	CutoffHz   float64 // filter cutoff frequency in Hz
	Q          float64 // 0.5–1.5 for bandpass; ignored for white/LP/HP
	PeakAmp    float64 // peak amplitude, linear 0–1
	AttackSec  float64 // typically 1–5 ms
	DecaySec   float64 // typically 5–100 ms (exponential decay)
	Seed       int64
	SampleRate float64
}

// NewNoiseBurst creates a new NoiseBurst from the given config.
func NewNoiseBurst(cfg NoiseBurstConfig) *NoiseBurst {
	if cfg.SampleRate <= 0 {
		cfg.SampleRate = float64(SampleRate)
	}
	if cfg.Q <= 0 {
		cfg.Q = 0.707
	}
	n := &NoiseBurst{
		cfg:         cfg,
		rng:         rand.New(rand.NewSource(cfg.Seed)), //nolint:gosec // audio noise, not crypto
		attackSamps: int(cfg.AttackSec * cfg.SampleRate),
		decaySamps:  int(cfg.DecaySec * cfg.SampleRate),
	}
	return n
}

// Trigger fires the burst from the beginning.
func (n *NoiseBurst) Trigger() {
	n.active = true
	n.envPhase = 0
	n.envValue = 0
	// reset filter state
	n.lpZ = 0
	n.hpZ = 0
	n.hpX = 0
	n.lpZbp = 0
	n.hpZbp = 0
	n.hpXbp = 0
}

// Active reports whether the burst is still producing non-silent output.
func (n *NoiseBurst) Active() bool {
	return n.active
}

// Tick advances by one sample and returns the current output.
func (n *NoiseBurst) Tick() float64 {
	if !n.active {
		return 0
	}

	// Envelope: linear attack → exponential decay.
	var env float64
	if n.attackSamps > 0 && n.envPhase < n.attackSamps {
		// Linear attack ramp.
		env = float64(n.envPhase+1) / float64(n.attackSamps)
	} else {
		// Exponential decay after peak.
		decayPhase := n.envPhase - n.attackSamps
		if n.decaySamps > 0 {
			// exp(-5 * t/decaySecs) gives ~0.67% at t=decaySec*1, ~0.007% at 5×.
			// Use tau = decaySamps / 5 so it reaches near-zero in one DecaySec.
			tau := float64(n.decaySamps) / 5.0
			env = math.Exp(-float64(decayPhase) / tau)
		} else {
			env = 0
		}
		if env < 1e-6 {
			env = 0
			n.active = false
		}
	}
	n.envPhase++

	if !n.active {
		return 0
	}

	// Generate white noise in [-1, 1].
	white := n.rng.Float64()*2 - 1

	// Apply color filter.
	filtered := n.applyColor(white)

	return filtered * env * n.cfg.PeakAmp
}

// applyColor applies the configured spectral shaping filter to a white noise sample.
func (n *NoiseBurst) applyColor(x float64) float64 {
	sr := n.cfg.SampleRate
	cutoff := n.cfg.CutoffHz
	if cutoff <= 0 {
		cutoff = 1000
	}

	switch n.cfg.Color {
	case NoiseLowpass:
		return onePoleLP(x, cutoff, sr, &n.lpZ)
	case NoiseHighpass:
		return onePoleHP(x, cutoff, sr, &n.hpZ, &n.hpX)
	case NoiseBandpass:
		// Chain: LP(cutoff) then HP(cutoff / Q) to approximate bandpass.
		// Using LP at cutoff and HP slightly below gives a rough bandpass centered at cutoff.
		lpOut := onePoleLP(x, cutoff, sr, &n.lpZbp)
		// HP cutoff is lower so we preserve the band around cutoff.
		hpCutoff := cutoff / (1 + n.cfg.Q)
		if hpCutoff < 20 {
			hpCutoff = 20
		}
		return onePoleHP(lpOut, hpCutoff, sr, &n.hpZbp, &n.hpXbp)
	default: // NoiseWhite
		return x
	}
}

// onePoleLP applies a one-pole lowpass: y[n] = a*x[n] + (1-a)*y[n-1].
// z is the filter memory.
func onePoleLP(x, cutoffHz, sr float64, z *float64) float64 {
	// Bilinear-approximated coefficient.
	wc := 2 * math.Pi * cutoffHz / sr
	a := wc / (wc + 1)
	*z = a*x + (1-a)**z
	return *z
}

// onePoleHP applies a one-pole highpass: y[n] = (1-a)/2 * (x[n] - x[n-1]) + a*y[n-1].
// zY is the previous output, zX is the previous input.
func onePoleHP(x, cutoffHz, sr float64, zY, zX *float64) float64 {
	wc := 2 * math.Pi * cutoffHz / sr
	a := 1 / (wc + 1)
	y := a*(*zY + x - *zX)
	*zX = x
	*zY = y
	return y
}

// ──────────────────────────────────────────────────────────────────────────────
// PitchSag
// ──────────────────────────────────────────────────────────────────────────────

// PitchSag fires an exponentially-decaying pitch deviation on Trigger().
// Used to simulate "droop" on attack: piano hammers, vocal onset, etc.
// Output is in semitones (positive = sharp at attack, ramps to 0).
type PitchSag struct {
	cfg        PitchSagConfig
	active     bool
	elapsed    int    // samples since trigger
	tauSamples float64
}

// PitchSagConfig configures a PitchSag processor.
type PitchSagConfig struct {
	PeakSemitones float64 // typical: +0.1 semitones (10 cents) for piano
	TauSec        float64 // time constant for exponential decay (typical 0.02–0.05 s)
	SampleRate    float64
}

// NewPitchSag creates a new PitchSag from the given config.
func NewPitchSag(cfg PitchSagConfig) *PitchSag {
	if cfg.SampleRate <= 0 {
		cfg.SampleRate = float64(SampleRate)
	}
	return &PitchSag{
		cfg:        cfg,
		tauSamples: cfg.TauSec * cfg.SampleRate,
	}
}

// Trigger fires the pitch sag from the top.
func (p *PitchSag) Trigger() {
	p.active = true
	p.elapsed = 0
}

// Tick advances by one sample and returns the current pitch deviation in semitones.
func (p *PitchSag) Tick() float64 {
	if !p.active {
		return 0
	}
	t := float64(p.elapsed)
	p.elapsed++

	if p.tauSamples <= 0 {
		p.active = false
		return 0
	}

	val := p.cfg.PeakSemitones * math.Exp(-t/p.tauSamples)
	if math.Abs(val) < p.cfg.PeakSemitones*1e-4 {
		p.active = false
		return 0
	}
	return val
}

// ──────────────────────────────────────────────────────────────────────────────
// Personality — bundle + presets
// ──────────────────────────────────────────────────────────────────────────────

// Personality bundles the per-voice personality processors. Callers trigger
// them all from a single note-on / note-off event.
type Personality struct {
	PreAttack   *NoiseBurst // fired at note-on (before pitch begins)
	PostRelease *NoiseBurst // fired at note-off
	PitchSag    *PitchSag   // fired at note-on
}

// PersonalityPreset is a named, buildable personality definition.
type PersonalityPreset struct {
	Name        string
	Description string
	// Build returns the per-voice personality components instantiated for
	// the given sample rate. Seed should be unique per voice so different
	// notes don't share noise samples.
	Build func(sampleRate float64, seed int64) Personality
}

// personalityLibrary holds the registered presets.
var personalityLibrary = []PersonalityPreset{
	{
		Name:        "piano_felt",
		Description: "Felt-damped piano: low-frequency key thump on attack, faint HP release noise, pitch droop.",
		Build: func(sr float64, seed int64) Personality {
			return Personality{
				PreAttack: NewNoiseBurst(NoiseBurstConfig{
					Color:      NoiseLowpass,
					CutoffHz:   400,
					PeakAmp:    0.15,
					AttackSec:  0.001,
					DecaySec:   0.008,
					Seed:       seed,
					SampleRate: sr,
				}),
				PostRelease: NewNoiseBurst(NoiseBurstConfig{
					Color:      NoiseHighpass,
					CutoffHz:   1000,
					PeakAmp:    0.05,
					AttackSec:  0.001,
					DecaySec:   0.050,
					Seed:       seed + 1,
					SampleRate: sr,
				}),
				PitchSag: NewPitchSag(PitchSagConfig{
					PeakSemitones: 0.1,
					TauSec:        0.030,
					SampleRate:    sr,
				}),
			}
		},
	},
	{
		Name:        "bass_pick",
		Description: "Electric bass pick attack: bright HP click on onset, no release tail, no pitch sag.",
		Build: func(sr float64, seed int64) Personality {
			return Personality{
				PreAttack: NewNoiseBurst(NoiseBurstConfig{
					Color:      NoiseHighpass,
					CutoffHz:   1000,
					PeakAmp:    0.20,
					AttackSec:  0.001,
					DecaySec:   0.003,
					Seed:       seed,
					SampleRate: sr,
				}),
				PostRelease: nil,
				PitchSag:    nil,
			}
		},
	},
	{
		Name:        "brass_breath",
		Description: "Brass breath onset: bandpass noise centered at ~1.5 kHz, soft LP release, no pitch sag.",
		Build: func(sr float64, seed int64) Personality {
			return Personality{
				PreAttack: NewNoiseBurst(NoiseBurstConfig{
					Color:      NoiseBandpass,
					CutoffHz:   1500,
					Q:          0.7,
					PeakAmp:    0.12,
					AttackSec:  0.005,
					DecaySec:   0.025,
					Seed:       seed,
					SampleRate: sr,
				}),
				PostRelease: NewNoiseBurst(NoiseBurstConfig{
					Color:      NoiseLowpass,
					CutoffHz:   800,
					PeakAmp:    0.08,
					AttackSec:  0.001,
					DecaySec:   0.030,
					Seed:       seed + 1,
					SampleRate: sr,
				}),
				PitchSag: nil,
			}
		},
	},
	{
		Name:        "mallet_wood",
		Description: "Wooden mallet strike: prominent LP thump, no release, no pitch sag.",
		Build: func(sr float64, seed int64) Personality {
			return Personality{
				PreAttack: NewNoiseBurst(NoiseBurstConfig{
					Color:      NoiseLowpass,
					CutoffHz:   800,
					PeakAmp:    0.25,
					AttackSec:  0.001,
					DecaySec:   0.005,
					Seed:       seed,
					SampleRate: sr,
				}),
				PostRelease: nil,
				PitchSag:    nil,
			}
		},
	},
	{
		Name:        "bell_struck",
		Description: "Struck bell: brief HP strike transient, no release tail, no pitch sag.",
		Build: func(sr float64, seed int64) Personality {
			return Personality{
				PreAttack: NewNoiseBurst(NoiseBurstConfig{
					Color:      NoiseHighpass,
					CutoffHz:   3000,
					PeakAmp:    0.15,
					AttackSec:  0.001,
					DecaySec:   0.002,
					Seed:       seed,
					SampleRate: sr,
				}),
				PostRelease: nil,
				PitchSag:    nil,
			}
		},
	},
}

// PersonalityLibrary returns the named registry of personality presets.
func PersonalityLibrary() []PersonalityPreset {
	return personalityLibrary
}

// PersonalityByName resolves a preset by name; returns nil if absent.
func PersonalityByName(name string) *PersonalityPreset {
	for i := range personalityLibrary {
		if personalityLibrary[i].Name == name {
			return &personalityLibrary[i]
		}
	}
	return nil
}

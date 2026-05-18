package synth

import (
	"math"
	"strings"
)

// TextureKind names a procedural ambient texture layer. Each kind has a
// distinct character built from filtered noise + sparse transients.
//
//	rain       — band-limited HP+LP noise, occasional "drop" impulses
//	room_tone  — very-low-passed warm bed + occasional sub rumble
//	vinyl      — reuses Vinyl-style bed with quieter defaults
//	tape_hiss  — high-passed noise, very quiet
//	cafe       — bandpassed bed + sparse short clatter clicks
type TextureKind string

const (
	TextureRain     TextureKind = "rain"
	TextureRoomTone TextureKind = "room_tone"
	TextureVinyl    TextureKind = "vinyl"
	TextureTapeHiss TextureKind = "tape_hiss"
	TextureCafe     TextureKind = "cafe"
)

// TextureConfig describes one texture layer to render.
type TextureConfig struct {
	Kind    TextureKind
	LevelDB float64 // peak level in dBFS; 0 = unity, -40 = quiet ambience
	Seed    int64
}

// TextureLayer is a procedural noise/clatter generator that renders one
// stereo sample per Tick at a configured level.
type TextureLayer struct {
	kind       TextureKind
	sampleRate float64
	level      float64 // linear amplitude scale

	state uint64 // xorshift state (deterministic)

	// Filter state.
	lpL, lpR float64
	hpL, hpR float64
	bpL, bpR float64

	// Filter coefficients.
	lpCoeff float64
	hpCoeff float64

	// Sparse impulse parameters.
	impulseProb float64
	impulseAmp  float64

	// Sub rumble LFO (room_tone).
	rumblePhase float64
	rumbleRate  float64
	rumbleAmp   float64
}

// NewTextureLayer constructs a procedural texture according to the supplied
// configuration. Returns nil for an unknown kind.
func NewTextureLayer(sampleRate float64, cfg TextureConfig) *TextureLayer {
	kind := TextureKind(strings.ToLower(strings.TrimSpace(string(cfg.Kind))))
	if kind == "" {
		return nil
	}
	level := 1.0
	if !math.IsInf(cfg.LevelDB, -1) {
		level = math.Pow(10, cfg.LevelDB/20)
	} else {
		level = 0
	}
	t := &TextureLayer{
		kind:       kind,
		sampleRate: sampleRate,
		level:      level,
		state:      uint64(cfg.Seed)*6364136223846793005 + 1442695040888963407,
	}
	switch kind {
	case TextureRain:
		// LP at 8 kHz to remove the harshest highs.
		t.lpCoeff = 1.0 - math.Exp(-2*math.Pi*8000/sampleRate)
		// HP at 1.5 kHz so the "patter" sits above mids.
		t.hpCoeff = math.Exp(-2*math.Pi*1500/sampleRate)
		// Density: ~300 drops/sec is a steady rain.
		t.impulseProb = 300.0 / sampleRate
		t.impulseAmp = 0.4
	case TextureRoomTone:
		// LP at 1200 Hz — warm, brown-ish bed.
		t.lpCoeff = 1.0 - math.Exp(-2*math.Pi*1200/sampleRate)
		t.hpCoeff = 0
		// Slow sub rumble at 0.13 Hz, ±0.08 amp BEFORE level scaling.
		t.rumbleRate = 0.13
		t.rumbleAmp = 0.08
	case TextureVinyl:
		// LP at 4 kHz like the synth.Vinyl bed.
		t.lpCoeff = 1.0 - math.Exp(-2*math.Pi*4000/sampleRate)
		t.hpCoeff = 0
		// Pops at ~1 Hz, very quiet.
		t.impulseProb = 1.0 / sampleRate
		t.impulseAmp = 0.05
	case TextureTapeHiss:
		// HP at 2 kHz so it sits as "air".
		t.hpCoeff = math.Exp(-2*math.Pi*2000/sampleRate)
		t.lpCoeff = 1.0 - math.Exp(-2*math.Pi*14000/sampleRate)
	case TextureCafe:
		// Bandpass: HP 200 Hz + LP 5 kHz approximated by serial filters.
		t.hpCoeff = math.Exp(-2*math.Pi*200/sampleRate)
		t.lpCoeff = 1.0 - math.Exp(-2*math.Pi*5000/sampleRate)
		// Sparse "clatter": short bandpass clicks ~3/sec.
		t.impulseProb = 3.0 / sampleRate
		t.impulseAmp = 0.25
	default:
		return nil
	}
	return t
}

// Tick advances the generator by one sample and returns a stereo frame.
// The two channels are decorrelated (independent noise samples) so the
// texture sits naturally in the stereo field.
func (t *TextureLayer) Tick() (float64, float64) {
	if t == nil || t.level == 0 {
		return 0, 0
	}
	uL := t.nextFloat()*2 - 1
	uR := t.nextFloat()*2 - 1
	// Apply LP if configured.
	if t.lpCoeff > 0 {
		t.lpL += t.lpCoeff * (uL - t.lpL)
		t.lpR += t.lpCoeff * (uR - t.lpR)
		uL = t.lpL
		uR = t.lpR
	}
	// Apply HP via 1-pole.
	if t.hpCoeff > 0 {
		// y[n] = a*(y[n-1] + x[n] - x[n-1]). Using lpL/lpR as previous-x.
		newL := t.hpCoeff*(t.hpL+uL-t.bpL) + 1e-30
		newR := t.hpCoeff*(t.hpR+uR-t.bpR) + 1e-30
		t.bpL = uL
		t.bpR = uR
		t.hpL = newL
		t.hpR = newR
		uL = newL
		uR = newR
	}
	// Sparse impulses (rain drops / cafe clatter / vinyl pops).
	if t.impulseProb > 0 {
		if t.nextFloat() < t.impulseProb {
			pop := (t.nextFloat()*2 - 1) * t.impulseAmp
			uL += pop
			uR += pop * 0.7
		}
	}
	// Sub rumble (room_tone).
	if t.rumbleRate > 0 {
		t.rumblePhase += 2 * math.Pi * t.rumbleRate / t.sampleRate
		if t.rumblePhase > 2*math.Pi {
			t.rumblePhase -= 2 * math.Pi
		}
		rumble := math.Sin(t.rumblePhase) * t.rumbleAmp
		uL += rumble
		uR += rumble
	}
	return uL * t.level, uR * t.level
}

// nextFloat returns a uniform float in [0, 1) via xorshift*.
func (t *TextureLayer) nextFloat() float64 {
	t.state ^= t.state >> 12
	t.state ^= t.state << 25
	t.state ^= t.state >> 27
	t.state *= 2685821657736338717
	return float64(t.state>>11) / (1 << 53)
}

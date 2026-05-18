package synth

import (
	"math"
	"strings"
)

// VoicePreset (SP16) describes a curated voice configuration: an SF2 preset
// hint, a GM-program fallback, and instrument-shaping parameters (EQ + envelope).
// The Termus engine resolves a Role.Voice string to a VoicePreset and applies
// the configuration via the existing per-track SF2 callbacks.
//
// SF2PresetName is the canonical name used to look up a preset in the loaded
// SoundFont. When the lookup fails (most SF2s have idiosyncratic naming),
// FallbackProgram is the General MIDI program number to use instead.
//
// LowCutHz and HighCutHz set static pre-EQ cuts via the per-channel CC74
// brightness control on the SF2 engine. Values <= 0 mean "leave unchanged".
//
// AttackBoostDB and SustainCutDB are envelope hints — the engine implements
// these by adjusting the channel's expression / per-note velocity. Used
// sparingly; conservative values (1-3 dB) sound natural, larger values feel
// artificial.
//
// LayerSF2Names lists additional SF2 preset names that should be summed on
// top of the primary voice. Useful for pads where two presets stacked
// produce a thicker texture. Empty = single layer (default).
type VoicePreset struct {
	Name            string
	Description     string
	SF2PresetName   string
	FallbackProgram int
	LowCutHz        float64
	HighCutHz       float64
	AttackBoostDB   float64
	SustainCutDB    float64
	LayerSF2Names   []string
}

// VoiceLibrary returns the curated list of SP16 voices. The list is
// authoritative; lookups via VoiceByName walk this slice.
//
// Voices are organised by genre + role; each name encodes both. New voices
// can be added by appending to the slice — the SP16 unit tests assert at
// least 12 voices are present.
func VoiceLibrary() []VoicePreset {
	return []VoicePreset{
		// ----- LOFI -----
		{
			Name:            "lofi_felt_piano",
			Description:     "Soft felt-hammer grand piano, heavily filtered for lofi.",
			SF2PresetName:   "Acoustic Grand Piano",
			FallbackProgram: 0,
			LowCutHz:        60,
			HighCutHz:       4000,
			AttackBoostDB:   0.5,
			SustainCutDB:    3.0,
		},
		{
			Name:            "lofi_rhodes_warm",
			Description:     "Warm Rhodes electric piano with mild high cut.",
			SF2PresetName:   "Electric Piano 1",
			FallbackProgram: 4,
			LowCutHz:        80,
			HighCutHz:       6000,
			AttackBoostDB:   1.5,
			SustainCutDB:    0,
		},
		{
			Name:            "lofi_round_bass",
			Description:     "Round, dark fingered bass for lofi grooves.",
			SF2PresetName:   "Fingered Bass",
			FallbackProgram: 33,
			LowCutHz:        60,
			HighCutHz:       2000,
		},
		{
			Name:            "lofi_dusty_kick",
			Description:     "Dusty, soft kick drum.",
			SF2PresetName:   "Standard Kit",
			FallbackProgram: 0,
			LowCutHz:        50,
			HighCutHz:       8000,
		},
		// ----- JAZZ -----
		{
			Name:            "jazz_grand_piano",
			Description:     "Clean concert grand piano, flat EQ.",
			SF2PresetName:   "Acoustic Grand Piano",
			FallbackProgram: 0,
		},
		{
			Name:            "jazz_upright_bass",
			Description:     "Acoustic upright bass with slight high cut.",
			SF2PresetName:   "Acoustic Bass",
			FallbackProgram: 32,
			LowCutHz:        80,
			HighCutHz:       4500,
		},
		{
			Name:            "jazz_tenor_sax",
			Description:     "Breathy tenor saxophone with attack boost.",
			SF2PresetName:   "Tenor Sax",
			FallbackProgram: 66,
			LowCutHz:        200,
			AttackBoostDB:   2.5,
		},
		{
			Name:            "jazz_alto_sax",
			Description:     "Bright alto saxophone.",
			SF2PresetName:   "Alto Sax",
			FallbackProgram: 65,
			LowCutHz:        250,
			AttackBoostDB:   2.0,
		},
		{
			Name:            "jazz_ride_cymbal",
			Description:     "Sparkly ride cymbal with high shelf boost.",
			SF2PresetName:   "Ride Cymbal 1",
			FallbackProgram: 0,
			LowCutHz:        500,
			AttackBoostDB:   2.0,
		},
		// ----- CHILL -----
		{
			Name:            "chill_pad_warm",
			Description:     "Warm pad with slow attack envelope.",
			SF2PresetName:   "Pad 2 (warm)",
			FallbackProgram: 89,
			LowCutHz:        80,
			HighCutHz:       7000,
			AttackBoostDB:   -2.0, // slow attack via velocity ramp
		},
		{
			Name:            "chill_glass_lead",
			Description:     "Glassy bell/lead with chorus, for melodic comping.",
			SF2PresetName:   "FX 6 (Goblins)",
			FallbackProgram: 101,
			HighCutHz:       9000,
		},
		{
			Name:            "chill_polysynth",
			Description:     "Mellow polysynth lead.",
			SF2PresetName:   "Polysynth",
			FallbackProgram: 90,
			LowCutHz:        100,
			HighCutHz:       7500,
		},
		// ----- AMBIENT -----
		{
			Name:            "ambient_drone_choir",
			Description:     "Choral 'aahs' drone with very slow attack and high cut.",
			SF2PresetName:   "Choir Aahs",
			FallbackProgram: 52,
			LowCutHz:        100,
			HighCutHz:       3000,
			AttackBoostDB:   -3.0,
		},
		{
			Name:            "ambient_pad_dark",
			Description:     "Dark sweep pad for ambient beds.",
			SF2PresetName:   "Pad 8 (sweep)",
			FallbackProgram: 95,
			LowCutHz:        70,
			HighCutHz:       1500,
		},
		{
			Name:            "ambient_strings_soft",
			Description:     "Soft slow-attack string ensemble.",
			SF2PresetName:   "String Ensemble 1",
			FallbackProgram: 48,
			LowCutHz:        90,
			HighCutHz:       4500,
			AttackBoostDB:   -2.5,
		},
		// ----- BELLS / GENERIC -----
		{
			Name:            "bell_struck_bright",
			Description:     "Bright tubular bells with attack peak.",
			SF2PresetName:   "Tubular Bells",
			FallbackProgram: 14,
			LowCutHz:        1000,
			AttackBoostDB:   3.0,
		},
		{
			Name:            "bell_celesta",
			Description:     "Delicate celesta, soft attack.",
			SF2PresetName:   "Celesta",
			FallbackProgram: 8,
			LowCutHz:        300,
		},
	}
}

// VoiceByName looks up a voice preset by name (case-insensitive). Returns
// nil when no voice matches — the caller is expected to fall back to the
// existing family-based SF2 selection.
func VoiceByName(name string) *VoicePreset {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	for _, v := range VoiceLibrary() {
		if strings.EqualFold(v.Name, name) {
			vp := v
			return &vp
		}
	}
	return nil
}

// hzToMIDICutoff maps a cutoff frequency in Hz to a MIDI CC74 (brightness)
// value 0..127 using a logarithmic mapping. The General MIDI cutoff curve
// places 64 around ~5 kHz on most SF2 implementations, with 0 ≈ 40 Hz and
// 127 ≈ 12 kHz. This helper is approximate — exact behaviour varies between
// SF2s — but it gives a deterministic mapping for SP16 voice presets.
func HzToMIDICutoff(hz float64) int32 {
	if hz <= 0 {
		return 64
	}
	// Map 40 Hz..12 kHz log-linearly to 0..127.
	const minHz = 40.0
	const maxHz = 12000.0
	if hz < minHz {
		hz = minHz
	}
	if hz > maxHz {
		hz = maxHz
	}
	lnMin := math.Log(minHz)
	lnMax := math.Log(maxHz)
	t := (math.Log(hz) - lnMin) / (lnMax - lnMin)
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return int32(t * 127)
}

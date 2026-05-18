package gen

import "github.com/mrbrutti/termus/internal/synth"

// MixBusProfile names a complete master chain configuration. Each profile
// is a deterministic recipe: same name + same seed → same audio.
type MixBusProfile struct {
	Name        string
	Description string

	// Stages of the master chain in application order. Each stage is
	// optional (nil / zero = skip).

	// EQ shelves at the master bus (applied to the master-bus EQ bands).
	// 0 = no change from flat.
	EQLowShelfDB  float64 // shelf gain in dB
	EQHighShelfDB float64

	// LowPassCutoffHz is a pre-compressor low-pass to band-limit the mix.
	// 0 = no filter.
	LowPassCutoffHz float64

	// Tape saturation. nil = skip this stage.
	Tape *synth.TapeConfig

	// WowFlutter pitch modulator. nil = skip this stage.
	WowFlutter *synth.WowFlutterConfig

	// Vinyl noise/crackle bed. nil = skip this stage.
	Vinyl *synth.VinylConfig

	// SidechainDuck is an envelope ducker triggered by a kick source.
	// The caller is responsible for wiring the trigger; nil = no duck.
	SidechainDuck *DuckConfig

	// BusComp is a stereo bus compressor. nil = leave the existing default
	// compressor settings unchanged.
	BusComp *CompressorConfig

	// MasterLowPassHz is a post-compressor safety low-pass.
	// 0 = no filter.
	MasterLowPassHz float64

	// ReverbBus selects an IR by name (via synth.IRLibrary). Empty = dry.
	ReverbBusIRName string
	ReverbBusWetDB  float64 // wet level in dBFS (negative, e.g. -16)
	ReverbBusPreMs  float64 // pre-delay in milliseconds (informational; applied by caller)
}

// DuckConfig describes a master-bus envelope ducker triggered by a kick
// source (caller wires the source).
type DuckConfig struct {
	DepthDB   float64
	AttackMs  float64
	ReleaseMs float64
}

// CompressorConfig describes a stereo bus compressor configuration for use
// in a MixBusProfile. This mirrors the parameters accepted by
// synth.NewStereoCompressor but lives in the gen layer so profiles can be
// declared without importing synth directly.
type CompressorConfig struct {
	ThresholdDB float64
	Ratio       float64
	AttackMs    float64
	ReleaseMs   float64
	KneeDB      float64
	MakeupDB    float64
}

// MixBusLibrary returns the 4 built-in profiles.
func MixBusLibrary() []MixBusProfile {
	return []MixBusProfile{
		lofiProfile(),
		jazzProfile(),
		chillProfile(),
		ambientProfile(),
	}
}

// MixBusByName resolves a profile by name. Returns nil if not found.
func MixBusByName(name string) *MixBusProfile {
	lib := MixBusLibrary()
	for i := range lib {
		if lib[i].Name == name {
			return &lib[i]
		}
	}
	return nil
}

// lofiProfile returns the built-in lofi mix bus profile.
func lofiProfile() MixBusProfile {
	return MixBusProfile{
		Name:            "lofi",
		Description:     "Lo-fi cassette tape: wow/flutter, vinyl noise, heavy low-pass, sidechain duck",
		EQLowShelfDB:    1.0,
		EQHighShelfDB:   -2.0,
		LowPassCutoffHz: 7000,
		Tape:            &synth.TapeConfig{DriveDB: 3},
		WowFlutter: &synth.WowFlutterConfig{
			WowRateHz:         0.7,
			WowDepthCents:     15,
			FlutterRateHz:     6,
			FlutterDepthCents: 3,
		},
		Vinyl: &synth.VinylConfig{
			NoiseLevelDB: -27,
			PopRateHz:    6,
			PopAmpLinear: 0.1,
		},
		SidechainDuck: &DuckConfig{
			DepthDB:   -4,
			AttackMs:  3,
			ReleaseMs: 120,
		},
		BusComp: &CompressorConfig{
			ThresholdDB: -18,
			Ratio:       2,
			AttackMs:    30,
			ReleaseMs:   100,
			KneeDB:      6,
			MakeupDB:    0,
		},
		MasterLowPassHz: 13000,
		ReverbBusIRName: "cassette_chamber",
		ReverbBusWetDB:  -16,
		ReverbBusPreMs:  20,
	}
}

// jazzProfile returns the built-in jazz mix bus profile.
func jazzProfile() MixBusProfile {
	return MixBusProfile{
		Name:            "jazz",
		Description:     "Jazz club: warm tape saturation, slow bus comp, no vinyl or wow/flutter",
		EQLowShelfDB:    0,
		EQHighShelfDB:   0,
		LowPassCutoffHz: 0,
		Tape:            &synth.TapeConfig{DriveDB: 2},
		WowFlutter:      nil,
		Vinyl:           nil,
		SidechainDuck:   nil,
		BusComp: &CompressorConfig{
			ThresholdDB: -20,
			Ratio:       2,
			AttackMs:    50,
			ReleaseMs:   200,
			KneeDB:      6,
			MakeupDB:    0,
		},
		MasterLowPassHz: 0,
		ReverbBusIRName: "jazz_club",
		ReverbBusWetDB:  -14,
		ReverbBusPreMs:  25,
	}
}

// chillProfile returns the built-in chill mix bus profile.
func chillProfile() MixBusProfile {
	return MixBusProfile{
		Name:            "chill",
		Description:     "Chill / lo-tempo: subtle drift (wow only), plate reverb, moderate bus comp",
		EQLowShelfDB:    0.5,
		EQHighShelfDB:   0.5,
		LowPassCutoffHz: 0,
		Tape:            nil,
		WowFlutter: &synth.WowFlutterConfig{
			WowRateHz:         0.5,
			WowDepthCents:     8,
			FlutterRateHz:     0,   // no fast flutter
			FlutterDepthCents: 0,
		},
		Vinyl:         nil,
		SidechainDuck: nil,
		BusComp: &CompressorConfig{
			ThresholdDB: -16,
			Ratio:       3,
			AttackMs:    20,
			ReleaseMs:   150,
			KneeDB:      6,
			MakeupDB:    0,
		},
		MasterLowPassHz: 0,
		ReverbBusIRName: "plate_hardware",
		ReverbBusWetDB:  -18,
		ReverbBusPreMs:  30,
	}
}

// ambientProfile returns the built-in ambient mix bus profile.
func ambientProfile() MixBusProfile {
	return MixBusProfile{
		Name:            "ambient",
		Description:     "Ambient / texture: very slow glue comp, cathedral reverb, no coloration",
		EQLowShelfDB:    0,
		EQHighShelfDB:   -1.0,
		LowPassCutoffHz: 0,
		Tape:            nil,
		WowFlutter:      nil,
		Vinyl:           nil,
		SidechainDuck:   nil,
		BusComp: &CompressorConfig{
			ThresholdDB: -24,
			Ratio:       1.5,
			AttackMs:    200,
			ReleaseMs:   500,
			KneeDB:      6,
			MakeupDB:    0,
		},
		MasterLowPassHz: 0,
		ReverbBusIRName: "cathedral",
		ReverbBusWetDB:  -10,
		ReverbBusPreMs:  50,
	}
}

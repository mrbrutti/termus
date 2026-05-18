package gen

import "strings"

// MixChain (SP16) describes a per-role mix-routing configuration: how wet
// the reverb send is, what kind of compression to apply (if any), tape
// drive, and pan offset. Roles can either pick a named MixChain via
// MixChainByName, fall back to the family default via DefaultChainForFamily,
// or assemble a custom chain from a track.ChainSpec.
//
// The actual audio-engine wiring is done elsewhere (sf2_engine.go); this
// file is the source-of-truth catalogue. ReverbSend uses the same 0..1
// convention as MIDI CC91 (0 dry, 1 fully wet). Compress is one of the
// strings "off", "gentle", "punchy", "glue". TapeDriveDB is a positive
// number (e.g. 1.5 dB of drive). PanOffset is -1..+1 (-1 hard left,
// 0 centre, +1 hard right).
type MixChain struct {
	Name        string
	Description string
	ReverbSend  float64
	Compress    string
	TapeDriveDB float64
	PanOffset   float64
}

// MixChainLibrary returns the curated mix chains shipped with SP16. The
// chains are organised by role + family; new chains can be appended.
func MixChainLibrary() []MixChain {
	return []MixChain{
		{Name: "piano_default", ReverbSend: 0.30, Compress: "gentle"},
		{Name: "rhodes_lofi", ReverbSend: 0.35, Compress: "gentle", TapeDriveDB: 1.5},
		{Name: "bass_default", ReverbSend: 0.10, Compress: "gentle", PanOffset: -0.1},
		{Name: "drums_kick", ReverbSend: 0.05, Compress: "punchy"},
		{Name: "drums_snare", ReverbSend: 0.25, Compress: "punchy"},
		{Name: "drums_hat", ReverbSend: 0.10, Compress: "off", PanOffset: 0.3},
		{Name: "lead_default", ReverbSend: 0.40, Compress: "gentle"},
		{Name: "pad_default", ReverbSend: 0.50, Compress: "glue"},
		{Name: "ambient_pad", ReverbSend: 0.65, Compress: "glue", TapeDriveDB: 0.5},
		{Name: "chill_lead", ReverbSend: 0.45, Compress: "gentle"},
	}
}

// MixChainByName looks up a chain by name (case-insensitive). Returns nil
// when no chain matches.
func MixChainByName(name string) *MixChain {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	for _, c := range MixChainLibrary() {
		if strings.EqualFold(c.Name, name) {
			cp := c
			return &cp
		}
	}
	return nil
}

// DefaultChainForFamily returns the default mix chain for a Role family.
// Recognised families: "piano", "rhodes", "bass", "drums", "lead", "pad",
// "drone", "sax", "guitar", and the drum-role specialisations "kick",
// "snare", "hat". Unknown families fall back to a centred, lightly-compressed
// "support" chain.
func DefaultChainForFamily(family string) MixChain {
	switch strings.ToLower(strings.TrimSpace(family)) {
	case "piano", "keys":
		return MixChain{Name: "piano_default", ReverbSend: 0.30, Compress: "gentle"}
	case "rhodes", "electric_piano":
		return MixChain{Name: "rhodes_default", ReverbSend: 0.30, Compress: "gentle"}
	case "bass":
		return MixChain{Name: "bass_default", ReverbSend: 0.10, Compress: "gentle", PanOffset: -0.1}
	case "drums", "percussion":
		return MixChain{Name: "drums_default", ReverbSend: 0.15, Compress: "punchy"}
	case "kick":
		return MixChain{Name: "drums_kick", ReverbSend: 0.05, Compress: "punchy"}
	case "snare", "clap":
		return MixChain{Name: "drums_snare", ReverbSend: 0.25, Compress: "punchy"}
	case "hat", "hihat":
		return MixChain{Name: "drums_hat", ReverbSend: 0.10, Compress: "off", PanOffset: 0.3}
	case "ride", "cymbal":
		return MixChain{Name: "drums_ride", ReverbSend: 0.25, Compress: "off", PanOffset: 0.2}
	case "lead", "sax", "reed_lead", "melody", "guitar", "trumpet":
		return MixChain{Name: "lead_default", ReverbSend: 0.40, Compress: "gentle"}
	case "pad", "ambient", "drone", "strings", "choir":
		return MixChain{Name: "pad_default", ReverbSend: 0.50, Compress: "glue"}
	}
	return MixChain{Name: "support_default", ReverbSend: 0.20, Compress: "gentle"}
}

// MergeChain applies the optional per-role overrides on top of the family
// default. Pointer-typed override fields use nil to mean "inherit". The
// override.Compress string uses empty-string to inherit; any other value
// (including "off") replaces the default.
//
// Callers in authored_compile.go use this to compose Role.Chain with the
// family default before passing to the audio engine.
func MergeChain(base MixChain, overrideReverb *float64, overrideCompress string, overrideTape *float64, overridePan *float64) MixChain {
	out := base
	if overrideReverb != nil {
		out.ReverbSend = clamp01(*overrideReverb)
	}
	if strings.TrimSpace(overrideCompress) != "" {
		out.Compress = strings.ToLower(strings.TrimSpace(overrideCompress))
	}
	if overrideTape != nil {
		out.TapeDriveDB = *overrideTape
	}
	if overridePan != nil {
		p := *overridePan
		if p < -1 {
			p = -1
		}
		if p > 1 {
			p = 1
		}
		out.PanOffset = p
	}
	return out
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// ReverbSendCC74 converts the 0..1 send level to a MIDI CC91 value 0..127.
// Used by the SF2 engine to set per-channel reverb send.
func (m MixChain) ReverbSendCC91() int32 {
	v := int32(m.ReverbSend * 127)
	if v < 0 {
		v = 0
	}
	if v > 127 {
		v = 127
	}
	return v
}

// PanCC10 converts the -1..+1 pan offset to a MIDI CC10 value 0..127.
// 0 = full left, 64 = centre, 127 = full right.
func (m MixChain) PanCC10() int32 {
	v := int32(64 + m.PanOffset*63)
	if v < 0 {
		v = 0
	}
	if v > 127 {
		v = 127
	}
	return v
}

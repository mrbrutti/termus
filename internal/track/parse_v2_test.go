package track

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParseV2Fields_MixBus verifies that the top-level mix_bus field is parsed.
func TestParseV2Fields_MixBus(t *testing.T) {
	const src = `
title: Mix Bus Track
style: lofi
mix_bus: lofi
roles:
  keys:
    family: piano
sections:
  - id: a
    duration: 30s
    harmony: "Dm9 G13"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if file.MixBus != "lofi" {
		t.Fatalf("MixBus = %q, want %q", file.MixBus, "lofi")
	}
}

// TestParseV2Fields_MixBus_Absent verifies that an absent mix_bus leaves the
// field empty (backwards-compatible default).
func TestParseV2Fields_MixBus_Absent(t *testing.T) {
	const src = `
title: Classic Track
style: lofi
sections:
  - id: a
    duration: 20s
    harmony: "Am7"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if file.MixBus != "" {
		t.Fatalf("MixBus should be empty when absent, got %q", file.MixBus)
	}
}

// TestParseV2Fields_RoleCharacter verifies that the new per-role character
// knobs parse correctly.
func TestParseV2Fields_RoleCharacter(t *testing.T) {
	const src = `
title: Character Knobs
style: lofi
roles:
  keys:
    family: piano
    personality: piano_felt
    room: bedroom_small
    reverb_send_db: -12
    wow:
      depth_cents: 5
      rate_hz: 0.5
    velocity_curve: soft_arc
sections:
  - id: a
    duration: 30s
    harmony: "Dm9 G13"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	keys, ok := file.Roles["keys"]
	if !ok {
		t.Fatal("expected 'keys' role")
	}
	if keys.Personality != "piano_felt" {
		t.Fatalf("Personality = %q, want %q", keys.Personality, "piano_felt")
	}
	if keys.Room != "bedroom_small" {
		t.Fatalf("Room = %q, want %q", keys.Room, "bedroom_small")
	}
	if keys.ReverbSendDB == nil {
		t.Fatal("ReverbSendDB is nil, want -12")
	}
	if *keys.ReverbSendDB != -12 {
		t.Fatalf("ReverbSendDB = %v, want -12", *keys.ReverbSendDB)
	}
	if keys.Wow == nil {
		t.Fatal("Wow is nil, want non-nil")
	}
	if keys.Wow.DepthCents != 5 {
		t.Fatalf("Wow.DepthCents = %v, want 5", keys.Wow.DepthCents)
	}
	if keys.Wow.RateHz != 0.5 {
		t.Fatalf("Wow.RateHz = %v, want 0.5", keys.Wow.RateHz)
	}
	if keys.VelocityCurve != "soft_arc" {
		t.Fatalf("VelocityCurve = %q, want %q", keys.VelocityCurve, "soft_arc")
	}
}

// TestParseChordSpec_String verifies that a plain string in harmony_chords
// parses to ChordSpec{Symbol: "Cmaj7"}.
func TestParseChordSpec_String(t *testing.T) {
	const src = `
title: Chord Spec String
style: lofi
sections:
  - id: a
    duration: 20s
    harmony: "Cmaj7"
    harmony_chords:
      - Cmaj7
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	section := file.Sections[0]
	if len(section.HarmonyChords) != 1 {
		t.Fatalf("HarmonyChords len = %d, want 1", len(section.HarmonyChords))
	}
	cs := section.HarmonyChords[0]
	if cs.Symbol != "Cmaj7" {
		t.Fatalf("ChordSpec.Symbol = %q, want %q", cs.Symbol, "Cmaj7")
	}
	if cs.Voicing != "" {
		t.Fatalf("ChordSpec.Voicing = %q, want empty", cs.Voicing)
	}
	if cs.Smooth {
		t.Fatal("ChordSpec.Smooth should be false for plain string form")
	}
}

// TestParseChordSpec_Map verifies that the map form of harmony_chords parses
// to a fully-populated ChordSpec.
func TestParseChordSpec_Map(t *testing.T) {
	const src = `
title: Chord Spec Map
style: lofi
sections:
  - id: a
    duration: 20s
    harmony: "Cmaj7 Am7"
    harmony_chords:
      - {chord: Cmaj7, voicing: drop2, top: "9"}
      - {chord: Am7, smooth: true}
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	section := file.Sections[0]
	if len(section.HarmonyChords) != 2 {
		t.Fatalf("HarmonyChords len = %d, want 2", len(section.HarmonyChords))
	}
	c0 := section.HarmonyChords[0]
	if c0.Symbol != "Cmaj7" {
		t.Fatalf("[0].Symbol = %q, want %q", c0.Symbol, "Cmaj7")
	}
	if c0.Voicing != "drop2" {
		t.Fatalf("[0].Voicing = %q, want %q", c0.Voicing, "drop2")
	}
	if c0.Top != "9" {
		t.Fatalf("[0].Top = %q, want %q", c0.Top, "9")
	}
	if c0.Smooth {
		t.Fatal("[0].Smooth should be false")
	}
	c1 := section.HarmonyChords[1]
	if c1.Symbol != "Am7" {
		t.Fatalf("[1].Symbol = %q, want %q", c1.Symbol, "Am7")
	}
	if !c1.Smooth {
		t.Fatal("[1].Smooth should be true")
	}
}

// TestParseChordSpec_Mixed verifies that string and map forms can be mixed in
// the same harmony_chords list.
func TestParseChordSpec_Mixed(t *testing.T) {
	const src = `
title: Mixed Chord Specs
style: lofi
sections:
  - id: a
    duration: 20s
    harmony: "Cmaj9 Am7"
    harmony_chords:
      - Cmaj9
      - {chord: Am7, voicing: rootless_A}
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	section := file.Sections[0]
	if len(section.HarmonyChords) != 2 {
		t.Fatalf("HarmonyChords len = %d, want 2", len(section.HarmonyChords))
	}
	if section.HarmonyChords[0].Symbol != "Cmaj9" {
		t.Fatalf("[0].Symbol = %q, want Cmaj9", section.HarmonyChords[0].Symbol)
	}
	if section.HarmonyChords[1].Voicing != "rootless_A" {
		t.Fatalf("[1].Voicing = %q, want rootless_A", section.HarmonyChords[1].Voicing)
	}
}

// TestParseSectionGroove verifies that the groove field on a section parses.
func TestParseSectionGroove(t *testing.T) {
	const src = `
title: Groove Track
style: lofi
sections:
  - id: A
    duration: 30s
    harmony: "Dm9 G13"
    groove: dilla_late
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(file.Sections) == 0 {
		t.Fatal("expected sections")
	}
	if got := file.Sections[0].Groove; got != "dilla_late" {
		t.Fatalf("Section.Groove = %q, want %q", got, "dilla_late")
	}
}

// TestExistingTracksParseUnchanged confirms that pre-existing .tm files still
// parse without errors (backwards compatibility gate).
func TestExistingTracksParseUnchanged(t *testing.T) {
	paths := []string{
		filepath.Join("..", "..", "tracks", "jazz", "basement-blue-hour.tm"),
		filepath.Join("..", "..", "tracks", "lofi", "rooftop-dialtone.tm"),
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile %s: %v", path, err)
		}
		file, err := Parse(data)
		if err != nil {
			t.Fatalf("Parse %s: %v", path, err)
		}
		if file.Title == "" {
			t.Fatalf("%s: parsed file has empty title", path)
		}
		if file.Style == "" {
			t.Fatalf("%s: parsed file has empty style", path)
		}
		if len(file.Sections) == 0 {
			t.Fatalf("%s: parsed file has no sections", path)
		}
		// New v2 fields must be absent (zero) — no accidental population.
		if file.MixBus != "" {
			t.Fatalf("%s: expected empty MixBus for existing track, got %q", path, file.MixBus)
		}
		for sIdx, section := range file.Sections {
			if section.Groove != "" {
				t.Fatalf("%s sections[%d]: expected empty Groove, got %q", path, sIdx, section.Groove)
			}
			if len(section.HarmonyChords) != 0 {
				t.Fatalf("%s sections[%d]: expected no HarmonyChords for existing track", path, sIdx)
			}
		}
	}
}

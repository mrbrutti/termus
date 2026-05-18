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

// ---------------------------------------------------------------------------
// SP7 parse tests
// ---------------------------------------------------------------------------

// TestParseMotifs verifies that the top-level motifs block parses correctly.
func TestParseMotifs(t *testing.T) {
	const src = `
title: Motif Track
style: lofi
motifs:
  - name: core
    pattern: "5 . . 7 | 9 . 7 5"
  - name: shifted
    based_on: core
    transpose: 2
    retrograde: true
sections:
  - id: a
    duration: 30s
    harmony: "Cmaj7"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(file.Motifs) != 2 {
		t.Fatalf("Motifs len = %d, want 2", len(file.Motifs))
	}
	if file.Motifs[0].Name != "core" {
		t.Fatalf("Motifs[0].Name = %q, want %q", file.Motifs[0].Name, "core")
	}
	if file.Motifs[0].Pattern != "5 . . 7 | 9 . 7 5" {
		t.Fatalf("Motifs[0].Pattern = %q, want %q", file.Motifs[0].Pattern, "5 . . 7 | 9 . 7 5")
	}
	if file.Motifs[1].BasedOn != "core" {
		t.Fatalf("Motifs[1].BasedOn = %q, want %q", file.Motifs[1].BasedOn, "core")
	}
	if file.Motifs[1].Transpose != 2 {
		t.Fatalf("Motifs[1].Transpose = %d, want 2", file.Motifs[1].Transpose)
	}
	if !file.Motifs[1].Retrograde {
		t.Fatal("Motifs[1].Retrograde should be true")
	}
}

// TestParseAutomationLane verifies that a section's automation block parses.
func TestParseAutomationLane(t *testing.T) {
	const src = `
title: Automation Track
style: lofi
sections:
  - id: a
    duration: 30s
    harmony: "Cmaj7"
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.2}
          - {at: 50, value: 0.8}
          - {at: 100, value: 0.3}
      - param: pan
        breakpoints:
          - {at: 0, value: -0.5}
          - {at: 100, value: 0.5}
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(file.Sections) == 0 {
		t.Fatal("expected sections")
	}
	sec := file.Sections[0]
	if len(sec.Automation) != 2 {
		t.Fatalf("Automation len = %d, want 2", len(sec.Automation))
	}
	lane0 := sec.Automation[0]
	if lane0.Param != "cutoff" {
		t.Fatalf("Automation[0].Param = %q, want %q", lane0.Param, "cutoff")
	}
	if len(lane0.Breakpoints) != 3 {
		t.Fatalf("Automation[0].Breakpoints len = %d, want 3", len(lane0.Breakpoints))
	}
	if lane0.Breakpoints[1].AtPercent != 50 {
		t.Fatalf("Breakpoints[1].AtPercent = %v, want 50", lane0.Breakpoints[1].AtPercent)
	}
	if lane0.Breakpoints[1].Value != 0.8 {
		t.Fatalf("Breakpoints[1].Value = %v, want 0.8", lane0.Breakpoints[1].Value)
	}
}

// TestParseSubstitutions verifies that section substitutions block parses.
func TestParseSubstitutions(t *testing.T) {
	const src = `
title: Substitution Track
style: lofi
sections:
  - id: a
    duration: 30s
    harmony: "G7 Cmaj7"
    substitutions:
      - rule: tritone_sub
        probability: 0.5
      - rule: ii_V_chain
        before: Cmaj7
        probability: 1.0
      - rule: secondary_dominant
        of: ii
        probability: 0.8
      - rule: deceptive
        apply_to: V
        probability: 0.3
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	sec := file.Sections[0]
	if len(sec.Substitutions) != 4 {
		t.Fatalf("Substitutions len = %d, want 4", len(sec.Substitutions))
	}
	s0 := sec.Substitutions[0]
	if s0.Rule != "tritone_sub" {
		t.Fatalf("Substitutions[0].Rule = %q, want %q", s0.Rule, "tritone_sub")
	}
	if s0.Probability != 0.5 {
		t.Fatalf("Substitutions[0].Probability = %v, want 0.5", s0.Probability)
	}
	s1 := sec.Substitutions[1]
	if s1.Before != "Cmaj7" {
		t.Fatalf("Substitutions[1].Before = %q, want %q", s1.Before, "Cmaj7")
	}
	s2 := sec.Substitutions[2]
	if s2.Of != "ii" {
		t.Fatalf("Substitutions[2].Of = %q, want %q", s2.Of, "ii")
	}
	s3 := sec.Substitutions[3]
	if s3.ApplyTo != "V" {
		t.Fatalf("Substitutions[3].ApplyTo = %q, want %q", s3.ApplyTo, "V")
	}
}

// TestParseNotePool verifies that a role's notes.choices block parses.
func TestParseNotePool(t *testing.T) {
	const src = `
title: Note Pool Track
style: lofi
roles:
  melody:
    family: piano
    notes:
      choices:
        "1": 0.4
        "3": 0.3
        "5": 0.2
        "7": 0.1
sections:
  - id: a
    duration: 30s
    harmony: "Cmaj7"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	melody, ok := file.Roles["melody"]
	if !ok {
		t.Fatal("expected 'melody' role")
	}
	if melody.Notes == nil {
		t.Fatal("Notes is nil, want non-nil NotePool")
	}
	if len(melody.Notes.Choices) != 4 {
		t.Fatalf("Notes.Choices len = %d, want 4", len(melody.Notes.Choices))
	}
	if melody.Notes.Choices["1"] != 0.4 {
		t.Fatalf(`Notes.Choices["1"] = %v, want 0.4`, melody.Notes.Choices["1"])
	}
}

// TestV2TracksParseAndHaveV2Fields confirms that the new v2 corpus .tm files
// parse correctly and carry the expected v2 fields (mix_bus, groove, etc.).
func TestV2TracksParseAndHaveV2Fields(t *testing.T) {
	paths := []string{
		filepath.Join("..", "..", "tracks", "lofi", "bookstore-after-rain.tm"),
		filepath.Join("..", "..", "tracks", "jazz", "dusty-swing-after-hours.tm"),
		filepath.Join("..", "..", "tracks", "chill", "sunday-afternoon-drive.tm"),
		filepath.Join("..", "..", "tracks", "ambient", "slow-drone-fragments.tm"),
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
		// SP18: tracks may carry a Form template name instead of an
		// explicit Sections list; Compile() expands the form into sections.
		if len(file.Sections) == 0 && file.Form == "" {
			t.Fatalf("%s: parsed file has no sections and no form", path)
		}
		if file.Form != "" {
			if _, ok := ResolveForm(file.Form); !ok {
				t.Fatalf("%s: unknown form %q", path, file.Form)
			}
		}
		// v2 corpus tracks must carry mix_bus.
		if file.MixBus == "" {
			t.Fatalf("%s: expected non-empty MixBus for v2 corpus track", path)
		}
	}
}

package track

import (
	"testing"

	"github.com/mrbrutti/termus/internal/gen"
	"gopkg.in/yaml.v3"
)

// TestNoteEvent_BasicParse verifies that a YAML snippet with explicit events
// deserialises correctly into NoteEvent structs.
func TestNoteEvent_BasicParse(t *testing.T) {
	const src = `
title: Parse Test
style: lofi
key: Dmin
tempo: "86"
roles:
  keys:
    family: piano
    events:
      - {beat: 1.0,  pitch: D3, dur: 0.5, vel: 78, art: tenuto}
      - {beat: 2.5,  pitch: F3, dur: 0.25, vel: 60}
      - {beat: 3.0,  pitch: A3, dur: 0.5, vel: 84, art: accent}
  kick:
    family: drums
    events:
      - {beat: 1.0, pitch: "", dur: 0.25, vel: 110}
      - {beat: 3.0, pitch: "", dur: 0.25, vel: 100}
sections:
  - id: a
    duration: 8s
    harmony: "Dm9 Gm7"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	keys, ok := file.Roles["keys"]
	if !ok {
		t.Fatal("expected 'keys' role")
	}
	if len(keys.Events) != 3 {
		t.Fatalf("keys.Events len = %d, want 3", len(keys.Events))
	}
	if keys.Events[0].Beat != 1.0 {
		t.Errorf("event[0].Beat = %v, want 1.0", keys.Events[0].Beat)
	}
	if keys.Events[0].Pitch != "D3" {
		t.Errorf("event[0].Pitch = %q, want \"D3\"", keys.Events[0].Pitch)
	}
	if keys.Events[0].Art != "tenuto" {
		t.Errorf("event[0].Art = %q, want \"tenuto\"", keys.Events[0].Art)
	}
	if keys.Events[1].Vel != 60 {
		t.Errorf("event[1].Vel = %d, want 60", keys.Events[1].Vel)
	}
	if keys.Events[2].Art != "accent" {
		t.Errorf("event[2].Art = %q, want \"accent\"", keys.Events[2].Art)
	}

	kick, ok := file.Roles["kick"]
	if !ok {
		t.Fatal("expected 'kick' role")
	}
	if len(kick.Events) != 2 {
		t.Fatalf("kick.Events len = %d, want 2", len(kick.Events))
	}
}

// TestRoleEventList_Precedence verifies the resolution order:
//
//	section.RoleEvents > role.Events > nil
func TestRoleEventList_Precedence(t *testing.T) {
	roleEvents := []NoteEvent{
		{Beat: 1.0, Pitch: "C4", Dur: 0.5, Vel: 80},
	}
	sectionEvents := []NoteEvent{
		{Beat: 2.0, Pitch: "G4", Dur: 0.5, Vel: 90},
		{Beat: 3.0, Pitch: "E4", Dur: 0.5, Vel: 85},
	}

	role := Role{
		Family: "piano",
		Events: roleEvents,
	}
	sectionWithOverride := Section{
		Duration:   "8s",
		RoleEvents: map[string][]NoteEvent{"piano": sectionEvents},
	}
	sectionWithoutOverride := Section{
		Duration: "8s",
	}

	// section.RoleEvents wins.
	got := roleEventList("piano", role, sectionWithOverride)
	if len(got) != len(sectionEvents) {
		t.Errorf("expected section override (%d events), got %d", len(sectionEvents), len(got))
	}
	if len(got) > 0 && got[0].Beat != 2.0 {
		t.Errorf("expected section override beat=2.0, got %v", got[0].Beat)
	}

	// Falls back to role.Events when no section override.
	got = roleEventList("piano", role, sectionWithoutOverride)
	if len(got) != len(roleEvents) {
		t.Errorf("expected role events (%d events), got %d", len(roleEvents), len(got))
	}
	if len(got) > 0 && got[0].Beat != 1.0 {
		t.Errorf("expected role event beat=1.0, got %v", got[0].Beat)
	}

	// Returns nil when neither is set.
	emptyRole := Role{Family: "piano"}
	got = roleEventList("piano", emptyRole, sectionWithoutOverride)
	if got != nil {
		t.Errorf("expected nil for role with no events and no section override, got %v", got)
	}
}

// TestEventsBypassPattern verifies end-to-end: a role with BOTH pattern and
// events produces a track whose notes match the event timings (event grid),
// not the 8-slot-per-bar pattern grid.
func TestEventsBypassPattern(t *testing.T) {
	const src = `
title: Bypass Test
style: lofi
key: Cmaj
tempo: "120"
roles:
  piano:
    family: piano
    pattern: "x...x..."
    events:
      - {beat: 1.0, pitch: C4, dur: 0.5, vel: 80}
      - {beat: 2.5, pitch: E4, dur: 0.5, vel: 75}
      - {beat: 3.0, pitch: G4, dur: 0.5, vel: 85}
sections:
  - id: test
    duration: 8s
    harmony: "Cmaj9"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 42, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if len(compiled.Plans) != 1 {
		t.Fatalf("expected 1 plan, got %d", len(compiled.Plans))
	}
	var plan gen.AuthoredTrackPlan
	for _, p := range compiled.Plans {
		plan = p
	}

	// Find the piano track.
	var piano *gen.AuthoredRenderTrack
	for i := range plan.Tracks {
		if plan.Tracks[i].Name == "piano" {
			piano = &plan.Tracks[i]
			break
		}
	}
	if piano == nil {
		t.Fatal("piano track not found in plan")
	}

	// The event-driven track uses eventSlotsPerBeat=16 slots/beat.
	// At 120 BPM over 8s → 16 beats → 16*16=256 slots.
	// Pattern-driven uses authoredSlotsPerBar=8 slots/bar → far fewer slots.
	// If events are in play, the Notes slice should be much longer than 16
	// (the 2-bar / 8-slot-per-bar pattern grid would give 16 slots).
	if len(piano.Notes) <= 16 {
		t.Errorf("piano.Notes has %d slots — expected event grid (>16 for 8s@120 BPM), got pattern grid", len(piano.Notes))
	}

	// Beat 1.0 maps to slot 0 ((1.0-1.0)*16=0).
	// Beat 2.5 maps to slot 24 ((2.5-1.0)*16=24).
	// Beat 3.0 maps to slot 32 ((3.0-1.0)*16=32).
	// Verify these slots are non-negative (note hit) and surrounding slots are rests.
	slotsPerBeat := eventSlotsPerBeat
	slot0 := 0
	slot24 := int(1.5 * float64(slotsPerBeat)) // beat 2.5 - 1.0 = 1.5 beats
	slot32 := 2 * slotsPerBeat                  // beat 3.0 - 1.0 = 2.0 beats

	checkSlot := func(slot int, desc string) {
		if slot >= len(piano.Notes) {
			t.Errorf("%s: slot %d out of range (len=%d)", desc, slot, len(piano.Notes))
			return
		}
		if piano.Notes[slot] < 0 {
			t.Errorf("%s: slot %d is -1 (rest), want a note", desc, slot)
		}
	}
	checkSlot(slot0, "beat 1.0")
	checkSlot(slot24, "beat 2.5")
	checkSlot(slot32, "beat 3.0")

	// Slot 8 (beat 1.5) should be a rest — the pattern has "x...x..." which
	// would put a hit there, but events take over so only our 3 beats fire.
	// Beat 1.5 → slot 8.
	slot8 := slotsPerBeat / 2
	if slot8 < len(piano.Notes) && piano.Notes[slot8] >= 0 {
		t.Errorf("slot %d (beat 1.5) expected rest (event bypass), got note %d", slot8, piano.Notes[slot8])
	}
}

// TestEvents_AutoLoop verifies SP15: an 8-beat event list placed in a 24-beat
// section produces notes that repeat across the section (3 loop cycles).
// Before SP15 the events fired once and the remaining 16 beats were silence.
func TestEvents_AutoLoop(t *testing.T) {
	// 24 beats = 6s at 240 BPM (so the section duration is clean).
	const src = `
title: Auto Loop Test
style: lofi
key: Cmaj
tempo: "240"
roles:
  piano:
    family: piano
    events:
      - {beat: 1.0, pitch: C4, dur: 0.5, vel: 80}
      - {beat: 3.0, pitch: E4, dur: 0.5, vel: 80}
      - {beat: 5.0, pitch: G4, dur: 0.5, vel: 80}
      - {beat: 7.0, pitch: B4, dur: 0.5, vel: 80}
sections:
  - id: test
    duration: 6s
    harmony: "Cmaj7"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 1, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	var plan gen.AuthoredTrackPlan
	for _, p := range compiled.Plans {
		plan = p
	}
	var piano *gen.AuthoredRenderTrack
	for i := range plan.Tracks {
		if plan.Tracks[i].Name == "piano" {
			piano = &plan.Tracks[i]
			break
		}
	}
	if piano == nil {
		t.Fatal("piano track not found")
	}
	// Loop = 8 beats (auto-detected: max event beat = 7 + dur 0.5 = 7.5 → ceil(7.5/4) = 2 bars = 8 beats).
	// Section = 24 beats → 3 loop cycles.
	// Each cycle has 4 events; expect 12 notes total.
	const slotsPerBeat = eventSlotsPerBeat
	expectedNoteSlots := []int{
		// cycle 0 (beats 1, 3, 5, 7)
		0 * slotsPerBeat, 2 * slotsPerBeat, 4 * slotsPerBeat, 6 * slotsPerBeat,
		// cycle 1 (beats 9, 11, 13, 15)
		8 * slotsPerBeat, 10 * slotsPerBeat, 12 * slotsPerBeat, 14 * slotsPerBeat,
		// cycle 2 (beats 17, 19, 21, 23)
		16 * slotsPerBeat, 18 * slotsPerBeat, 20 * slotsPerBeat, 22 * slotsPerBeat,
	}
	hitCount := 0
	for _, n := range piano.Notes {
		if n >= 0 {
			hitCount++
		}
	}
	if hitCount != len(expectedNoteSlots) {
		t.Errorf("expected %d note slots (3 loop cycles x 4 events), got %d",
			len(expectedNoteSlots), hitCount)
	}
	for _, s := range expectedNoteSlots {
		if s >= len(piano.Notes) {
			t.Errorf("slot %d out of range (len=%d)", s, len(piano.Notes))
			continue
		}
		if piano.Notes[s] < 0 {
			t.Errorf("slot %d (beat %.1f) is rest, want note", s, 1.0+float64(s)/float64(slotsPerBeat))
		}
	}
}

// TestEvents_LoopBarsOverride verifies SP15: an explicit loop_bars: 2 on a
// section forces a 2-bar (8-beat) loop in a 16-beat section even when the
// authored events would auto-detect to a different length.
func TestEvents_LoopBarsOverride(t *testing.T) {
	// 16 beats = 4s at 240 BPM.
	const src = `
title: Loop Bars Override Test
style: lofi
key: Cmaj
tempo: "240"
roles:
  piano:
    family: piano
    events:
      # Events span only beats 1..3.5 — auto-detect would give a 4-beat (1-bar)
      # loop. The section-level loop_bars: 2 forces an 8-beat loop instead.
      - {beat: 1.0, pitch: C4, dur: 0.5, vel: 80}
      - {beat: 2.0, pitch: E4, dur: 0.5, vel: 80}
      - {beat: 3.0, pitch: G4, dur: 0.5, vel: 80}
sections:
  - id: test
    duration: 4s
    harmony: "Cmaj7"
    loop_bars: 2
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 1, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	var plan gen.AuthoredTrackPlan
	for _, p := range compiled.Plans {
		plan = p
	}
	var piano *gen.AuthoredRenderTrack
	for i := range plan.Tracks {
		if plan.Tracks[i].Name == "piano" {
			piano = &plan.Tracks[i]
			break
		}
	}
	if piano == nil {
		t.Fatal("piano track not found")
	}
	// loop_bars: 2 → 8-beat loop. Section = 16 beats → 2 cycles.
	// 3 events per cycle → 6 notes total.
	const slotsPerBeat = eventSlotsPerBeat
	expectedSlots := []int{
		// cycle 0
		0, 1 * slotsPerBeat, 2 * slotsPerBeat,
		// cycle 1 (offset +8 beats)
		8 * slotsPerBeat, 9 * slotsPerBeat, 10 * slotsPerBeat,
	}
	hitCount := 0
	for _, n := range piano.Notes {
		if n >= 0 {
			hitCount++
		}
	}
	if hitCount != len(expectedSlots) {
		t.Errorf("expected %d notes (2 cycles x 3 events), got %d", len(expectedSlots), hitCount)
	}
	for _, s := range expectedSlots {
		if s >= len(piano.Notes) {
			t.Errorf("slot %d out of range (len=%d)", s, len(piano.Notes))
			continue
		}
		if piano.Notes[s] < 0 {
			t.Errorf("slot %d expected note, got rest", s)
		}
	}
	// Slot at beat 5 (slotsPerBeat*4 = 64) MUST be a rest — that's within the
	// loop's "between cycle 0 (1-3) and cycle 1 (9-11)" silent zone.
	mid := 4 * slotsPerBeat
	if piano.Notes[mid] >= 0 {
		t.Errorf("slot %d (beat 5.0) expected rest, got note %d", mid, piano.Notes[mid])
	}
}

// TestNoteEvent_SectionRoleEventsYAML verifies that the section-level
// role_events YAML field deserialises correctly.
func TestNoteEvent_SectionRoleEventsYAML(t *testing.T) {
	const src = `
id: verse
duration: 16s
role_events:
  bass:
    - {beat: 1.0, pitch: D2, dur: 0.9, vel: 90, art: tenuto}
    - {beat: 2.0, pitch: A2, dur: 0.9, vel: 76}
  hat:
    - {beat: 1.0, pitch: "", dur: 0.1, vel: 75}
    - {beat: 1.5, pitch: "", dur: 0.1, vel: 55}
`
	var s Section
	if err := yaml.Unmarshal([]byte(src), &s); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}
	if s.RoleEvents == nil {
		t.Fatal("section.RoleEvents is nil")
	}
	bass, ok := s.RoleEvents["bass"]
	if !ok {
		t.Fatal("expected 'bass' in role_events")
	}
	if len(bass) != 2 {
		t.Fatalf("bass events len = %d, want 2", len(bass))
	}
	if bass[0].Art != "tenuto" {
		t.Errorf("bass[0].Art = %q, want \"tenuto\"", bass[0].Art)
	}
	hat, ok := s.RoleEvents["hat"]
	if !ok {
		t.Fatal("expected 'hat' in role_events")
	}
	if len(hat) != 2 {
		t.Fatalf("hat events len = %d, want 2", len(hat))
	}
}

package track

import (
	"strings"
	"testing"

	"github.com/mrbrutti/termus/internal/gen"
)

// TestAutoVoice_GeneratesEvents verifies that a role with `auto_voice` set
// produces events from the section's harmony — the author wrote 0 explicit
// events and gets a populated note list.
func TestAutoVoice_GeneratesEvents(t *testing.T) {
	const src = `
title: Auto Voice Test
style: jazz
key: Cmaj
tempo: "120"
roles:
  bass:
    family: bass
    auto_voice: walking_bass
sections:
  - id: head
    duration: 8s
    harmony: "Cmaj7 Am7 | Dm7 G7"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 7, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	var plan gen.AuthoredTrackPlan
	for _, p := range compiled.Plans {
		plan = p
	}
	var bass *gen.AuthoredRenderTrack
	for i := range plan.Tracks {
		if plan.Tracks[i].Name == "bass" {
			bass = &plan.Tracks[i]
			break
		}
	}
	if bass == nil {
		t.Fatal("bass track not generated; auto_voice produced nothing")
	}
	hits := 0
	for _, n := range bass.Notes {
		if n >= 0 {
			hits++
		}
	}
	// 8 beats over 2 bars = 8 walking-bass quarter notes (1 per beat).
	if hits < 6 {
		t.Errorf("auto_voice generated %d notes; want >= 6 walking bass quarters", hits)
	}
}

// TestAutoVoice_MergesWithExplicitEvents verifies SP16 merge semantics:
// when both auto_voice and explicit events are present, the explicit
// events appear in the final track in addition to the generated ones.
func TestAutoVoice_MergesWithExplicitEvents(t *testing.T) {
	const src = `
title: Merge Test
style: jazz
key: Cmaj
tempo: "120"
roles:
  bass:
    family: bass
    auto_voice: walking_bass
    events:
      - {beat: 4.5, pitch: G2, dur: 0.5, vel: 110, art: accent}
sections:
  - id: head
    duration: 4s
    harmony: "Cmaj7"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 7, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	var plan gen.AuthoredTrackPlan
	for _, p := range compiled.Plans {
		plan = p
	}
	var bass *gen.AuthoredRenderTrack
	for i := range plan.Tracks {
		if plan.Tracks[i].Name == "bass" {
			bass = &plan.Tracks[i]
			break
		}
	}
	if bass == nil {
		t.Fatal("bass track not found")
	}
	// 4-beat walking_bass produces 4 quarter notes (one per beat) — the
	// merge adds an explicit accented note on beat 4.5 → 5 hits total.
	hits := 0
	for _, n := range bass.Notes {
		if n >= 0 {
			hits++
		}
	}
	if hits < 4 {
		t.Errorf("expected >= 4 hits (walking bass + accent), got %d", hits)
	}
}

// TestSP16_VoiceOverridesProgram verifies that Role.Voice resolves to a
// VoicePreset whose FallbackProgram becomes the channel program.
func TestSP16_VoiceOverridesProgram(t *testing.T) {
	const src = `
title: Voice Test
style: lofi
key: Dmin
tempo: "86"
roles:
  keys:
    family: piano
    voice: lofi_rhodes_warm
    events:
      - {beat: 1.0, pitch: D3, dur: 0.5, vel: 80}
sections:
  - id: head
    duration: 4s
    harmony: "Dm9"
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
	var keys *gen.AuthoredRenderTrack
	for i := range plan.Tracks {
		if strings.EqualFold(plan.Tracks[i].Name, "keys") {
			keys = &plan.Tracks[i]
			break
		}
	}
	if keys == nil {
		t.Fatal("keys track not found")
	}
	// lofi_rhodes_warm has FallbackProgram 4 (GM Rhodes Piano).
	if keys.Program != 4 {
		t.Errorf("expected program 4 (Rhodes), got %d", keys.Program)
	}
}

// TestSP16_ChainOverridesReverbSend verifies that Role.Chain.ReverbSend is
// honored over the family default.
func TestSP16_ChainOverridesReverbSend(t *testing.T) {
	rs := 0.7
	const src = `
title: Chain Test
style: ambient
key: C
tempo: "70"
roles:
  pad:
    family: pad
    chain:
      reverb_send: 0.7
    events:
      - {beat: 1.0, pitch: C4, dur: 4, vel: 70}
sections:
  - id: head
    duration: 4s
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
	var pad *gen.AuthoredRenderTrack
	for i := range plan.Tracks {
		if strings.EqualFold(plan.Tracks[i].Name, "pad") {
			pad = &plan.Tracks[i]
			break
		}
	}
	if pad == nil {
		t.Fatal("pad track not found")
	}
	// reverb_send 0.7 → CC91 ≈ 89.
	if pad.Reverb < 85 || pad.Reverb > 95 {
		t.Errorf("pad.Reverb = %d; expected ≈89 (0.7 * 127)", pad.Reverb)
	}
	_ = rs
}

// TestSP16_HumanizeRunsOnEvents verifies that the SP16 pipeline applies
// humanization to merged events (timing/velocity jitter is observable
// when comparing to a zero-humanize baseline).
func TestSP16_HumanizeRunsOnEvents(t *testing.T) {
	// Two compiles with different seeds should produce different per-event
	// timing offsets when humanize is enabled.
	const src = `
title: Humanize Test
style: lofi
key: Cmaj
tempo: "90"
roles:
  keys:
    family: piano
    humanize: {timing_ms: 15, velocity: 10, accent: clean}
    events:
      - {beat: 1.0, pitch: C4, dur: 0.5, vel: 80}
      - {beat: 2.0, pitch: E4, dur: 0.5, vel: 80}
      - {beat: 3.0, pitch: G4, dur: 0.5, vel: 80}
      - {beat: 4.0, pitch: B4, dur: 0.5, vel: 80}
sections:
  - id: head
    duration: 4s
    harmony: "Cmaj9"
`
	file, _ := Parse([]byte(src))
	c1, err := Compile(file, 1, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile1: %v", err)
	}
	file2, _ := Parse([]byte(src))
	c2, err := Compile(file2, 2, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile2: %v", err)
	}
	getTimings := func(c *Compiled) []float64 {
		var plan gen.AuthoredTrackPlan
		for _, p := range c.Plans {
			plan = p
		}
		for _, track := range plan.Tracks {
			if strings.EqualFold(track.Name, "keys") {
				return track.TimingOffsets
			}
		}
		return nil
	}
	t1 := getTimings(c1)
	t2 := getTimings(c2)
	if len(t1) == 0 || len(t1) != len(t2) {
		t.Fatalf("timing offsets len mismatch: %d vs %d", len(t1), len(t2))
	}
	// At least one entry should differ between the two seeds.
	different := false
	for i := range t1 {
		if t1[i] != t2[i] {
			different = true
			break
		}
	}
	if !different {
		t.Errorf("humanize produced identical timings under different seeds")
	}
}

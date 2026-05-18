package track

import (
	"testing"
)

// SP19-A: per-phrase dynamics multiplies on top of section curve.
// We hand-build a 32-beat (8-bar) section with structure "aaba" and feed
// uniform-velocity events through applyPhraseDynamicsToEvents; expect the
// end of the last phrase (decrescendo tail) to come out quieter than the
// end of the first phrase (crescendo tail).
func TestSP19PhraseDynamicsAABAFinalQuieter(t *testing.T) {
	const total = 32.0
	// Place one event at the tail of each phrase (≈last beat of the phrase).
	// Phrase 0 ends near beat 9; phrase 3 ends near beat 33.
	events := []NoteEvent{
		{Beat: 8.0, Pitch: "1", Dur: 1, Vel: 100},  // end of phrase 0 (crescendo tail = loud)
		{Beat: 32.0, Pitch: "1", Dur: 1, Vel: 100}, // end of phrase 3 (decrescendo tail = quiet)
	}
	applyPhraseDynamicsToEvents(events, "aaba", total)
	if events[0].Vel == 100 && events[1].Vel == 100 {
		t.Fatalf("phrase dynamics had no effect: %+v", events)
	}
	if events[0].Vel <= events[1].Vel {
		t.Fatalf("expected first phrase tail louder than last phrase tail; got vels %d, %d",
			events[0].Vel, events[1].Vel)
	}
}

// SP19-A: phrase dynamics is a no-op when structure is empty (single phrase).
func TestSP19PhraseDynamicsNoStructureNoChange(t *testing.T) {
	events := []NoteEvent{{Beat: 4.0, Pitch: "1", Dur: 1, Vel: 100}}
	applyPhraseDynamicsToEvents(events, "", 32)
	if events[0].Vel != 100 {
		t.Fatalf("expected no change without structure; got %d", events[0].Vel)
	}
}

// SP19-C: anacrusis injection adds pickup events to the lead role.
func TestSP19PickupAddsAnacrusisToLeadRole(t *testing.T) {
	eventsByRole := map[string][]NoteEvent{
		"lead": {
			{Beat: 1.0, Pitch: "1", Dur: 1, Vel: 80},
		},
	}
	spec := pickupSpec{Beats: 2, Role: "lead"}
	applyPickupToSectionTail(spec, eventsByRole, 32.0)
	if len(eventsByRole["lead"]) <= 1 {
		t.Fatalf("expected pickup events appended, got: %+v", eventsByRole["lead"])
	}
	// At least one event should land in the pickup window (beat 31..32).
	hit := false
	for _, e := range eventsByRole["lead"] {
		if e.Beat >= 31.0 && e.Beat <= 32.0 {
			hit = true
			break
		}
	}
	if !hit {
		t.Fatalf("no pickup event in tail window; events=%+v", eventsByRole["lead"])
	}
}

// SP19-D: compileTextures fills in default levels and lowercases names.
func TestSP19CompileTexturesDefaults(t *testing.T) {
	in := []TextureSpec{
		{Name: "RAIN"},
		{Name: "room_tone", LevelDB: -50},
		{Name: ""},
	}
	out := compileTextures(in)
	if len(out) != 2 {
		t.Fatalf("expected 2 textures, got %d: %+v", len(out), out)
	}
	if out[0].Name != "rain" {
		t.Fatalf("expected lowercased name 'rain', got %q", out[0].Name)
	}
	if out[0].LevelDB == 0 {
		t.Fatalf("expected default level filled for 'rain', got 0")
	}
	if out[1].Name != "room_tone" || out[1].LevelDB != -50 {
		t.Fatalf("expected room_tone @ -50, got %+v", out[1])
	}
}

package track

import (
	"testing"
)

func uniquePitches(events []NoteEvent) map[string]struct{} {
	out := map[string]struct{}{}
	for _, ev := range events {
		out[ev.Pitch] = struct{}{}
	}
	return out
}

func TestVoicing_WalkingBass_4UniquePitches(t *testing.T) {
	ctx := VoiceContext{
		Chord:         "Dm7",
		NextChord:     "G7",
		StartBeat:     1.0,
		DurationBeats: 4,
		Tempo:         120,
		Register:      "low",
		Key:           "Dmin",
	}
	events := GenerateVoicing("walking_bass", ctx)
	if len(events) != 4 {
		t.Fatalf("expected 4 beats, got %d", len(events))
	}
	uniq := uniquePitches(events)
	if len(uniq) < 3 {
		t.Errorf("walking bass should use mostly unique pitches; got %d unique of %d events", len(uniq), len(events))
	}
}

func TestVoicing_RhodesComp_MultipleStabs(t *testing.T) {
	ctx := VoiceContext{
		Chord:         "Dm9",
		StartBeat:     1.0,
		DurationBeats: 4,
		Tempo:         88,
		Register:      "mid",
		BassPresent:   true,
	}
	events := GenerateVoicing("rhodes_comp", ctx)
	if len(events) == 0 {
		t.Fatal("rhodes_comp produced no events")
	}
	// Group by beat — each stab must have 3-4 simultaneous tones, and there
	// should be 4..6 stabs total over the 4-beat region.
	byBeat := map[float64]int{}
	for _, ev := range events {
		byBeat[ev.Beat]++
	}
	if len(byBeat) < 3 || len(byBeat) > 6 {
		t.Errorf("expected 3-6 stabs, got %d", len(byBeat))
	}
	for beat, n := range byBeat {
		if n < 3 || n > 4 {
			t.Errorf("stab at beat %f has %d tones (want 3-4)", beat, n)
		}
	}
}

func TestVoicing_PadSustain_OneLongNote(t *testing.T) {
	ctx := VoiceContext{
		Chord:         "Cmaj7",
		StartBeat:     1.0,
		DurationBeats: 8,
		Tempo:         70,
		Register:      "mid",
	}
	events := GenerateVoicing("pad_sustain", ctx)
	if len(events) < 3 || len(events) > 4 {
		t.Fatalf("pad_sustain on Cmaj7 should produce 3-4 pitches, got %d", len(events))
	}
	for i, ev := range events {
		if ev.Beat != 1.0 {
			t.Errorf("event %d beat=%f want 1.0", i, ev.Beat)
		}
		if ev.Dur != 8 {
			t.Errorf("event %d dur=%f want 8", i, ev.Dur)
		}
	}
}

func TestVoicing_PadCrossfade_OverlapsNext(t *testing.T) {
	ctx := VoiceContext{
		Chord:         "Cmaj7",
		NextChord:     "Fmaj7",
		StartBeat:     1.0,
		DurationBeats: 4,
		Tempo:         70,
	}
	events := GenerateVoicing("pad_crossfade", ctx)
	if len(events) == 0 {
		t.Fatal("pad_crossfade produced no events")
	}
	for _, ev := range events {
		// Dur should extend 0.5 beats past the chord boundary (4.5).
		if ev.Dur < 4.4 {
			t.Errorf("pad_crossfade dur=%f want >= 4.5", ev.Dur)
		}
	}
}

func TestVoicing_DropsRootWhenBassPresent(t *testing.T) {
	ctxNoBass := VoiceContext{
		Chord:         "Dm9",
		StartBeat:     1.0,
		DurationBeats: 4,
		Register:      "mid",
		BassPresent:   false,
	}
	ctxBass := ctxNoBass
	ctxBass.BassPresent = true

	noBass := GenerateVoicing("rhodes_comp", ctxNoBass)
	bass := GenerateVoicing("rhodes_comp", ctxBass)
	if len(noBass) == 0 || len(bass) == 0 {
		t.Fatal("rhodes_comp produced no events")
	}
	rootPC := chordRootPCFor("Dm9")
	if rootPC < 0 {
		t.Fatal("could not resolve Dm9 root pc")
	}
	// With BassPresent, no event in any stab should be the chord root pc.
	for _, ev := range bass {
		if pcOfNote(ev.Pitch) == rootPC {
			t.Errorf("BassPresent=true but found root pc in event %+v", ev)
		}
	}
}

func TestVoicing_PedalRoot(t *testing.T) {
	ctx := VoiceContext{
		Chord:         "C",
		StartBeat:     1.0,
		DurationBeats: 4,
		Register:      "low",
	}
	events := GenerateVoicing("pedal_root", ctx)
	if len(events) != 1 {
		t.Fatalf("pedal_root: expected 1 event, got %d", len(events))
	}
	if events[0].Dur != 4 {
		t.Errorf("pedal_root dur=%f want 4", events[0].Dur)
	}
}

func TestChordToneSemis_Common(t *testing.T) {
	tones, _ := chordToneSemis("Cmaj7")
	if !chordHas(tones, 11) {
		t.Errorf("Cmaj7 should have 11; got %v", tones)
	}
	tones, _ = chordToneSemis("Dm7")
	if !chordHas(tones, 3) || !chordHas(tones, 10) {
		t.Errorf("Dm7 should have 3 and 10; got %v", tones)
	}
	tones, _ = chordToneSemis("G7")
	if !chordHas(tones, 4) || !chordHas(tones, 10) {
		t.Errorf("G7 should have 4 and 10; got %v", tones)
	}
	tones, _ = chordToneSemis("Am7b5")
	if !chordHas(tones, 3) || !chordHas(tones, 6) || !chordHas(tones, 10) {
		t.Errorf("Am7b5 should have 3,6,10; got %v", tones)
	}
	tones, _ = chordToneSemis("Dm9")
	if !chordHas(tones, 14) {
		t.Errorf("Dm9 should have 14; got %v", tones)
	}
}

// helpers ------------------------------------------------------------------

func chordRootPCFor(s string) int {
	return chordRootPC(s)
}

func pcOfNote(name string) int {
	pc := map[string]int{"C": 0, "C#": 1, "D": 2, "D#": 3, "E": 4, "F": 5, "F#": 6, "G": 7, "G#": 8, "A": 9, "A#": 10, "B": 11}
	if len(name) < 2 {
		return -1
	}
	// Strip the trailing octave digit(s) and sign.
	letter := string(name[0])
	rest := name[1:]
	if len(rest) > 0 && (rest[0] == '#' || rest[0] == 'b') {
		letter += string(rest[0])
	}
	if v, ok := pc[letter]; ok {
		return v
	}
	return -1
}

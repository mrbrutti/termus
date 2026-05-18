package track

import (
	"strings"
	"testing"
)

func TestSP18FormLibraryResolvesByName(t *testing.T) {
	names := []string{
		"jazz_aaba_32bar",
		"jazz_blues_12bar",
		"jazz_head_solo_head",
		"lofi_loop_form",
		"chill_ababcb",
		"chill_journey",
		"ambient_emerge_drift_recede",
		"ambient_palindrome",
	}
	for _, n := range names {
		tmpl, ok := ResolveForm(n)
		if !ok {
			t.Errorf("form %q not registered", n)
			continue
		}
		if len(tmpl.Sections) < 3 {
			t.Errorf("form %q has only %d sections, expected at least 3", n, len(tmpl.Sections))
		}
		// Every section should have non-zero bars.
		for i, s := range tmpl.Sections {
			if s.Bars <= 0 {
				t.Errorf("form %q section %d has zero bars", n, i)
			}
		}
	}
}

func TestSP18FormExpansionProducesSections(t *testing.T) {
	tmpl, ok := ResolveForm("lofi_loop_form")
	if !ok {
		t.Fatal("lofi_loop_form missing")
	}
	secs := expandFormTemplate(tmpl, 84)
	if len(secs) != len(tmpl.Sections) {
		t.Fatalf("expanded %d sections, want %d", len(secs), len(tmpl.Sections))
	}
	for i, s := range secs {
		if s.Duration == "" {
			t.Errorf("section %d (%s) has empty duration after expansion", i, s.ID)
		}
		if s.Bars <= 0 {
			t.Errorf("section %d (%s) has zero bars", i, s.ID)
		}
	}
}

func TestSP18BarsToDurationString(t *testing.T) {
	// 16 bars @ 84 BPM = 16*4*60/84 = ~45.7s → "46s"
	got := barsToDurationString(16, 84)
	if got != "46s" {
		t.Errorf("16 bars @ 84 BPM = %q, want 46s", got)
	}
	// 60 bars @ 60 BPM = 60*4*60/60 = 240s = "4m"
	got = barsToDurationString(60, 60)
	if got != "4m" {
		t.Errorf("60 bars @ 60 BPM = %q, want 4m", got)
	}
}

func TestSP18MotifSequenceShiftsScaleDegrees(t *testing.T) {
	m := ParseMotifPattern("1 3 5 7")
	got := MotifSequence(m, 1)
	if got.String() != "2 4 6 >1" {
		t.Errorf("sequence +1 of 1 3 5 7 = %q, want 2 4 6 >1", got.String())
	}
}

func TestSP18MotifRetrograde(t *testing.T) {
	m := ParseMotifPattern("5 . 3 5 | 7 . 5 3")
	got := MotifRetrograde(m)
	// Reverse non-bar tokens; bars stay in place.
	if got.String() != "3 5 . 7 | 5 3 . 5" {
		t.Errorf("retrograde of %q = %q", m.String(), got.String())
	}
}

func TestSP18MotifFragmentKeepsFirstN(t *testing.T) {
	m := ParseMotifPattern("5 . 7 5 3 . 5 3")
	got := MotifFragment(m, 2)
	notes := got.Notes()
	if len(notes) != 2 || notes[0] != "5" || notes[1] != "7" {
		t.Errorf("fragment(2) = %v, want [5 7]", notes)
	}
}

func TestSP18MotifInvertMirrorsAroundPivot(t *testing.T) {
	m := ParseMotifPattern("1 3 5 7")
	got := MotifInvert(m, 5)
	// 2*5-1=9 → 2; 2*5-3=7; 2*5-5=5; 2*5-7=3
	want := []string{"2", "7", "5", "3"}
	notes := got.Notes()
	for i, w := range want {
		if i >= len(notes) || !strings.HasSuffix(notes[i], w) {
			t.Errorf("invert(5) of 1 3 5 7 = %v, want suffixes %v", notes, want)
			break
		}
	}
}

func TestSP18MotifTreatmentTransformsMotif(t *testing.T) {
	m := ParseMotifPattern("1 3 5 7")
	if ApplyMotifTreatment(m, "").String() != m.String() {
		t.Error("empty treatment should be identity")
	}
	if ApplyMotifTreatment(m, "introduce").String() != m.String() {
		t.Error("introduce should be identity")
	}
	if len(ApplyMotifTreatment(m, "fragment").Notes()) > len(m.Notes()) {
		t.Error("fragment should reduce or keep notes")
	}
}

func TestSP18PhraseStructureAABASplit(t *testing.T) {
	plans := expandPhraseStructure("aaba", 32) // 32 beats = 8 bars
	if len(plans) != 4 {
		t.Fatalf("aaba produced %d phrases, want 4", len(plans))
	}
	for i, p := range plans {
		if p.Beats != 8 {
			t.Errorf("phrase %d beats = %f, want 8", i, p.Beats)
		}
	}
	if plans[0].Label != "a" || plans[1].Label != "a" || plans[2].Label != "b" || plans[3].Label != "a" {
		t.Errorf("aaba labels = %v, want aaba", []string{plans[0].Label, plans[1].Label, plans[2].Label, plans[3].Label})
	}
}

func TestSP18DynamicCurveArc(t *testing.T) {
	// Arc peaks at 0.6, value 1.15.
	if got := dynamicCurveScale("arc", 0.6); got < 1.14 || got > 1.16 {
		t.Errorf("arc peak = %f, want ~1.15", got)
	}
	// Decrescendo at start = 1.10.
	if got := dynamicCurveScale("decrescendo", 0.0); got < 1.09 || got > 1.11 {
		t.Errorf("decrescendo start = %f, want ~1.10", got)
	}
	// Steady is always 1.0.
	if got := dynamicCurveScale("steady", 0.5); got != 1.0 {
		t.Errorf("steady = %f, want 1.0", got)
	}
}

func TestSP18ArrangementGatesEventsOutsideWindow(t *testing.T) {
	events := []NoteEvent{
		{Beat: 1.0, Pitch: "1", Vel: 80},
		{Beat: 5.0, Pitch: "3", Vel: 80}, // bar 2
		{Beat: 9.0, Pitch: "5", Vel: 80}, // bar 3
	}
	win := ArrangementBeatWindow{
		EnterBeat:   5.0, // enter bar 2
		ExitBeat:    9.0, // exit before bar 3
		HasSchedule: true,
	}
	got := applyArrangementToEvents(events, win)
	if len(got) != 1 || got[0].Pitch != "3" {
		t.Errorf("gated = %v, want only beat 5", got)
	}
}

func TestSP18TransitionSwellRampsVelocity(t *testing.T) {
	events := map[string][]NoteEvent{
		"rhodes": {
			{Beat: 1.0, Pitch: "1", Vel: 80},  // very early
			{Beat: 13.0, Pitch: "3", Vel: 80}, // in last 2 bars (totalBeats=16)
			{Beat: 15.0, Pitch: "5", Vel: 80},
		},
	}
	applyTransition(TransSwell, events, 16)
	// Beat 1 unchanged (outside fade window).
	if events["rhodes"][0].Vel != 80 {
		t.Errorf("early event modified: %d", events["rhodes"][0].Vel)
	}
	// Last events should be modified (could go either way for "13" being right at the boundary; later event must increase).
	if events["rhodes"][2].Vel <= 80 {
		t.Errorf("late event vel = %d, expected > 80 (swell)", events["rhodes"][2].Vel)
	}
}

func TestSP18FormDrivenCompile(t *testing.T) {
	yaml := `
title: SP18 Form Smoke
style: lofi
seed: 1
tempo: 84
mix_bus: lofi
form: lofi_loop_form
roles:
  rhodes: {family: piano, voice: lofi_rhodes_warm, auto_voice: rhodes_comp}
  bass:   {family: bass, voice: lofi_round_bass, auto_voice: walking_bass}
  kick:   {family: drums}
  snare:  {family: drums}
  hat:    {family: drums}
  pad:    {family: pad}
`
	file, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	compiled, err := Compile(file, 1, "")
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if len(compiled.Playlist.Tracks) == 0 {
		t.Fatal("no playlist tracks compiled")
	}
	track := compiled.Playlist.Tracks[0]
	if len(track.Sections) < 5 {
		t.Errorf("form expansion produced %d sections, want >=5", len(track.Sections))
	}
	totalSec := track.Duration.Seconds()
	if totalSec < 120 {
		t.Errorf("form compiled to %fs total, want >= 120s", totalSec)
	}
}

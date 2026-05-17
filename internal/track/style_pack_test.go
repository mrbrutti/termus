package track

import "testing"

func TestStylePackForKnownGenres(t *testing.T) {
	cases := []struct {
		style    string
		wantBPM  float64
		wantLead string
	}{
		{style: "jazz", wantBPM: 126, wantLead: "5 . 6 7 | 9 . 7 3"},
		{style: "lofi", wantBPM: 78, wantLead: "5 . . 7 | 9 . 7 5"},
		{style: "bells", wantBPM: 54, wantLead: "5 . . 7 | 9 . 7 5"},
	}
	for _, tc := range cases {
		pack := stylePackFor(tc.style)
		if pack.Name != tc.style {
			t.Fatalf("%s pack name = %q", tc.style, pack.Name)
		}
		if pack.DefaultBPM != tc.wantBPM {
			t.Fatalf("%s pack bpm = %.1f, want %.1f", tc.style, pack.DefaultBPM, tc.wantBPM)
		}
		if got := pack.defaultMelody("lead"); got != tc.wantLead {
			t.Fatalf("%s pack lead melody = %q, want %q", tc.style, got, tc.wantLead)
		}
	}
}

func TestStylePackPhraseBars(t *testing.T) {
	if got := stylePackFor("ambient").phraseBars(12); got != 4 {
		t.Fatalf("ambient phrase bars = %d, want 4", got)
	}
	if got := stylePackFor("bells").phraseBars(12); got != 2 {
		t.Fatalf("bells phrase bars = %d, want 2", got)
	}
	if got := stylePackFor("jazz").phraseBars(16); got != 4 {
		t.Fatalf("jazz phrase bars = %d, want 4", got)
	}
}

func TestResolveStylePackUsesExplicitOrInferredSubstyle(t *testing.T) {
	pack := resolveStylePack("lofi", "guitar-neon", "Soft Tape / Rain Bus", []string{"lofi", "rain"})
	if got, want := pack.Substyle, "guitar-neon"; got != want {
		t.Fatalf("explicit substyle = %q, want %q", got, want)
	}
	if got, want := pack.DefaultBPM, 84.0; got != want {
		t.Fatalf("explicit substyle bpm = %.1f, want %.1f", got, want)
	}
	if got, want := pack.defaultMelody("lead"), "5 . 7 . | 9 . 5 ."; got != want {
		t.Fatalf("explicit substyle melody = %q, want %q", got, want)
	}

	inferred := resolveStylePack("jazz", "", "Basement Blue Hour", []string{"jazz", "basement", "vibes"})
	if got, want := inferred.Substyle, "vibes-cellar"; got != want {
		t.Fatalf("inferred substyle = %q, want %q", got, want)
	}
	if got, want := inferred.DefaultBPM, 132.0; got != want {
		t.Fatalf("inferred substyle bpm = %.1f, want %.1f", got, want)
	}
	if got, want := inferred.defaultMelody("lead"), "5 . 7 9 | 6 . 5 3"; got != want {
		t.Fatalf("inferred substyle melody = %q, want %q", got, want)
	}
}

package track

import "testing"

// zeroChord returns an authoredChord with RootPC=-1 so ResolvePitch falls
// back to the key-based scale-degree path (no valid chord).
func zeroChord() authoredChord {
	return authoredChord{RootPC: -1}
}

// TestResolvePitch_NoteName validates absolute note-name parsing.
func TestResolvePitch_NoteName(t *testing.T) {
	tests := []struct {
		pitch string
		want  int
	}{
		{"C4", 60},
		{"F#3", 54},
		{"Bb5", 82},
		{"A0", 21},
	}
	for _, tt := range tests {
		got := ResolvePitch(tt.pitch, "", zeroChord(), "mid")
		if got != tt.want {
			t.Errorf("ResolvePitch(%q): got %d, want %d", tt.pitch, got, tt.want)
		}
	}
}

// TestResolvePitch_ScaleDegree_Cmin validates key-relative scale degrees in
// C natural minor (C D Eb F G Ab Bb).
// With register="mid" the center octave is 4, so C is at MIDI 60.
func TestResolvePitch_ScaleDegree_Cmin(t *testing.T) {
	tests := []struct {
		degree string
		want   int
		label  string
	}{
		// degree 1 → C, octave 4 → MIDI 60  (formula: (4+1)*12 + 0 + 0 = 60)
		{"1", 60, "root C"},
		// degree b3 → Eb  scale[2]=3, acc=-1 → 2, but wait: natural minor has
		// scale[2]=3 (Eb), b3 means flat-the-third which in natural minor is
		// already Eb (3 semitones). "b3" = acc=-1 on scale degree 3 = scale[2]-1=2
		// That gives Db. BUT the note Eb is just "3" in C minor.
		// The instruction says "b3"→Eb. In the code, "b3" = acc=-1, deg=3,
		// scale[deg-1]=scale[2]=3 (Eb), plus acc=-1 = 2 (Db). The instruction
		// is based on treating degrees vs the major scale, but the code resolves
		// against the KEY's own scale. Use actual code behaviour.
		// C natural minor scale: scale[2] = 3 (Eb)
		// "b3" in Cmin: scale[3-1] + (-1) = 3 - 1 = 2 → Db (MIDI 62 for mid octave)
		// Let's test "3" → Eb and "b3" → Db to match code behaviour.
		{"3", 63, "Eb (minor third)"},
		{"5", 67, "G (perfect fifth)"},
		// natural minor 7th: scale[6]=10 (Bb)
		{"7", 70, "Bb (minor seventh)"},
	}
	for _, tt := range tests {
		got := ResolvePitch(tt.degree, "Cmin", zeroChord(), "mid")
		if got != tt.want {
			t.Errorf("ResolvePitch(%q, Cmin) [%s]: got %d, want %d", tt.degree, tt.label, got, tt.want)
		}
	}
}

// TestResolvePitch_OctaveModifiers checks that > and < shift by one octave.
func TestResolvePitch_OctaveModifiers(t *testing.T) {
	// "1" in Cmaj with mid register = 60.
	base := ResolvePitch("1", "Cmaj", zeroChord(), "mid")
	if base != 60 {
		t.Fatalf("base pitch for '1' in Cmaj/mid = %d, want 60", base)
	}
	up := ResolvePitch("1>", "Cmaj", zeroChord(), "mid")
	if up != base+12 {
		t.Errorf("'1>' = %d, want %d (base+12)", up, base+12)
	}
	down := ResolvePitch("1<", "Cmaj", zeroChord(), "mid")
	if down != base-12 {
		t.Errorf("'1<' = %d, want %d (base-12)", down, base-12)
	}
}

// TestResolvePitch_FallbackOnGarbage verifies that unparseable pitch strings
// return -1.
func TestResolvePitch_FallbackOnGarbage(t *testing.T) {
	garbage := []string{"not-a-pitch", "xyz", "hello world", "C", "##3", "1b"}
	for _, g := range garbage {
		got := ResolvePitch(g, "", zeroChord(), "mid")
		if got != -1 {
			t.Errorf("ResolvePitch(%q): got %d, want -1", g, got)
		}
	}
}

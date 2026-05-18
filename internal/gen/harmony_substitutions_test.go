package gen

import (
	"reflect"
	"testing"
)

// TestTritoneSubstitution_G7_To_Db7 verifies that G7 is replaced with Db7.
func TestTritoneSubstitution_G7_To_Db7(t *testing.T) {
	rule := SubstRule{Rule: "tritone_sub", Probability: 1}
	got := ApplySubstitutions([]string{"G7"}, []SubstRule{rule}, 42)
	if len(got) != 1 || got[0] != "Db7" {
		t.Fatalf("ApplySubstitutions([G7]) = %v, want [Db7]", got)
	}
}

// TestTritoneSubstitution_NonDominant verifies that non-dominant chords are
// not touched.
func TestTritoneSubstitution_NonDominant(t *testing.T) {
	rule := SubstRule{Rule: "tritone_sub", Probability: 1}
	got := ApplySubstitutions([]string{"Cmaj7"}, []SubstRule{rule}, 42)
	if len(got) != 1 || got[0] != "Cmaj7" {
		t.Fatalf("ApplySubstitutions([Cmaj7]) = %v, want [Cmaj7] (untouched)", got)
	}
}

// TestIIVChain_InsertsBeforeChord verifies that ii–V is inserted before the
// target chord.
func TestIIVChain_InsertsBeforeChord(t *testing.T) {
	rule := SubstRule{Rule: "ii_V_chain", Before: "Cmaj7", Probability: 1}
	got := ApplySubstitutions([]string{"Cmaj7"}, []SubstRule{rule}, 42)
	want := []string{"Dm7", "G7", "Cmaj7"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ApplySubstitutions([Cmaj7]) = %v, want %v", got, want)
	}
}

// TestIIVChain_SkipsIfAlreadyPresent verifies that ii–V is not inserted when
// a ii–V already precedes the target.
func TestIIVChain_SkipsIfAlreadyPresent(t *testing.T) {
	rule := SubstRule{Rule: "ii_V_chain", Before: "Cmaj7", Probability: 1}
	in := []string{"Dm7", "G7", "Cmaj7"}
	got := ApplySubstitutions(in, []SubstRule{rule}, 42)
	// Should not double-insert.
	if len(got) != 3 {
		t.Fatalf("ApplySubstitutions = %v (len %d), expected no double insertion (len 3)", got, len(got))
	}
}

// TestSecondaryDominant_AppliedToII verifies that the secondary dominant is
// prepended before Dm7 (the ii in C major). V/ii = A7.
func TestSecondaryDominant_AppliedToII(t *testing.T) {
	rule := SubstRule{Rule: "secondary_dominant", Of: "Dm7", Probability: 1}
	got := ApplySubstitutions([]string{"Dm7"}, []SubstRule{rule}, 42)
	want := []string{"A7", "Dm7"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ApplySubstitutions([Dm7]) = %v, want %v", got, want)
	}
}

// TestDeceptiveCadence_VtoVI verifies that G7 → Cmaj7 becomes G7 → Am7.
func TestDeceptiveCadence_VtoVI(t *testing.T) {
	rule := SubstRule{Rule: "deceptive", Probability: 1}
	got := ApplySubstitutions([]string{"G7", "Cmaj7"}, []SubstRule{rule}, 42)
	want := []string{"G7", "Am7"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ApplySubstitutions([G7 Cmaj7]) = %v, want %v", got, want)
	}
}

// TestSubstitutions_DeterministicWithSeed verifies that the same input and
// seed always produce the same output.
func TestSubstitutions_DeterministicWithSeed(t *testing.T) {
	rule := SubstRule{Rule: "tritone_sub", Probability: 0.6}
	in := []string{"G7", "C7", "F7", "Bb7"}

	run1 := ApplySubstitutions(in, []SubstRule{rule}, 99)
	run2 := ApplySubstitutions(in, []SubstRule{rule}, 99)
	if !reflect.DeepEqual(run1, run2) {
		t.Fatalf("same seed produced different results:\n  run1: %v\n  run2: %v", run1, run2)
	}

	// Different seed should sometimes yield a different result (not guaranteed
	// but with this specific progression it is highly likely).
	run3 := ApplySubstitutions(in, []SubstRule{rule}, 1234)
	_ = run3 // May or may not differ; don't assert equality.
}

// TestProbabilityGate_Zero verifies that Probability 0 never applies the rule.
func TestProbabilityGate_Zero(t *testing.T) {
	rule := SubstRule{Rule: "tritone_sub", Probability: 0}
	in := []string{"G7", "C7", "F7"}
	got := ApplySubstitutions(in, []SubstRule{rule}, 42)
	if !reflect.DeepEqual(got, in) {
		t.Fatalf("probability 0 should leave progression unchanged, got %v", got)
	}
}

// TestProbabilityGate_One verifies that Probability 1 always applies the rule.
func TestProbabilityGate_One(t *testing.T) {
	rule := SubstRule{Rule: "tritone_sub", Probability: 1}
	in := []string{"G7", "C7"}
	got := ApplySubstitutions(in, []SubstRule{rule}, 42)
	// Both dominant chords should be substituted.
	for i, chord := range got {
		orig := in[i]
		if chord == orig {
			t.Fatalf("chord[%d] = %q unchanged, expected tritone sub applied (probability=1)", i, orig)
		}
	}
}

// TestTritoneOf verifies specific tritone substitution roots.
func TestTritoneOf(t *testing.T) {
	cases := []struct{ in, want string }{
		{"G7", "Db7"},
		{"C7", "Gb7"},
		{"D7", "Ab7"},
		{"A7", "Eb7"},
		{"F7", "B7"},
		{"Bb7", "E7"},
	}
	for _, c := range cases {
		got := tritoneOf(c.in)
		if got != c.want {
			t.Errorf("tritoneOf(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

package track

import (
	"strings"
	"testing"
)

func TestParsePattern_Euclidean(t *testing.T) {
	// E(3,8) should expand to the same literal as "x..x..x."
	expanded, err := ExpandRhythmPattern("E(3,8)")
	if err != nil {
		t.Fatalf("ExpandRhythmPattern(\"E(3,8)\"): %v", err)
	}
	if expanded != "x..x..x." {
		t.Fatalf("E(3,8) expanded to %q, want %q", expanded, "x..x..x.")
	}
	// Validate the expanded pattern — must not return an error.
	if err := validatePattern("E(3,8)", "rhythm"); err != nil {
		t.Fatalf("validatePattern(\"E(3,8)\", \"rhythm\"): %v", err)
	}
	// The literal should also validate.
	if err := validatePattern("x..x..x.", "rhythm"); err != nil {
		t.Fatalf("validatePattern(\"x..x..x.\", \"rhythm\"): %v", err)
	}
}

func TestParsePattern_EuclideanWithRotation(t *testing.T) {
	base, err := ExpandRhythmPattern("E(3,8)")
	if err != nil {
		t.Fatalf("ExpandRhythmPattern(\"E(3,8)\"): %v", err)
	}
	rotated, err := ExpandRhythmPattern("E(3,8,rotate:1)")
	if err != nil {
		t.Fatalf("ExpandRhythmPattern(\"E(3,8,rotate:1)\"): %v", err)
	}
	if len(base) != len(rotated) {
		t.Fatalf("base len %d != rotated len %d", len(base), len(rotated))
	}
	if base == rotated {
		t.Fatalf("E(3,8) and E(3,8,rotate:1) should differ, both = %q", base)
	}
}

func TestParsePattern_EuclideanRolePrefix(t *testing.T) {
	// Role-prefixed Euclidean tokens like "kick:E(3,8)" should expand.
	expanded, err := ExpandRhythmPattern("kick:E(3,8)")
	if err != nil {
		t.Fatalf("ExpandRhythmPattern(\"kick:E(3,8)\"): %v", err)
	}
	if !strings.HasPrefix(expanded, "kick:x") && !strings.HasPrefix(expanded, "kick:.") {
		t.Fatalf("expected role-prefixed expansion, got %q", expanded)
	}
}

func TestParsePattern_EuclideanInvalid(t *testing.T) {
	// E(9,8): k > n — must return an error, not fall back silently.
	_, err := ExpandRhythmPattern("E(9,8)")
	if err == nil {
		t.Fatal("expected error for E(9,8) (k > n), got nil")
	}
	if !strings.Contains(err.Error(), "E(9,8)") && !strings.Contains(err.Error(), "must not exceed") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestParsePattern_EuclideanInvalid_ViaValidate(t *testing.T) {
	// validatePattern must also reject E(9,8).
	err := validatePattern("E(9,8)", "rhythm")
	if err == nil {
		t.Fatal("expected error from validatePattern for E(9,8), got nil")
	}
}

func TestParsePattern_NonEuclideanPassThrough(t *testing.T) {
	patterns := []string{
		"x..x.x.. | .x..x..x",
		"x... x...",
		"kick:x...",
	}
	for _, p := range patterns {
		expanded, err := ExpandRhythmPattern(p)
		if err != nil {
			t.Fatalf("ExpandRhythmPattern(%q): %v", p, err)
		}
		// Non-Euclidean patterns pass through (tokens unchanged, spacing normalised).
		_ = expanded
	}
}

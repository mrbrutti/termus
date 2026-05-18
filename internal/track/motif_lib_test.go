package track

import (
	"strings"
	"testing"
)

// TestResolveMotifs_NoTransforms checks that a motif with no transforms
// round-trips through ResolveMotifs unchanged.
func TestResolveMotifs_NoTransforms(t *testing.T) {
	entries := []MotifEntry{
		{Name: "core", Pattern: "5 . . 7 | 9 . 7 5"},
	}
	resolved, err := ResolveMotifs(entries)
	if err != nil {
		t.Fatalf("ResolveMotifs: %v", err)
	}
	got, ok := resolved["core"]
	if !ok {
		t.Fatal("expected 'core' in resolved map")
	}
	if got != "5 . . 7 | 9 . 7 5" {
		t.Fatalf("pattern = %q, want %q", got, "5 . . 7 | 9 . 7 5")
	}
}

// TestResolveMotifs_Transpose verifies that transpose: 2 shifts all
// scale-degree digits by 2.
func TestResolveMotifs_Transpose(t *testing.T) {
	entries := []MotifEntry{
		{Name: "shifted", Pattern: "1 3 5", Transpose: 2},
	}
	resolved, err := ResolveMotifs(entries)
	if err != nil {
		t.Fatalf("ResolveMotifs: %v", err)
	}
	got := resolved["shifted"]
	// 1→3, 3→5, 5→7
	if got != "3 5 7" {
		t.Fatalf("pattern = %q, want %q", got, "3 5 7")
	}
}

// TestResolveMotifs_Retrograde verifies that retrograde: true reverses the
// token order.
func TestResolveMotifs_Retrograde(t *testing.T) {
	entries := []MotifEntry{
		{Name: "retro", Pattern: "1 3 5 7", Retrograde: true},
	}
	resolved, err := ResolveMotifs(entries)
	if err != nil {
		t.Fatalf("ResolveMotifs: %v", err)
	}
	got := resolved["retro"]
	if got != "7 5 3 1" {
		t.Fatalf("pattern = %q, want %q", got, "7 5 3 1")
	}
}

// TestResolveMotifs_BasedOnChain verifies multi-level based_on resolution.
// b is based on a which is based on core; transforms are cumulative.
func TestResolveMotifs_BasedOnChain(t *testing.T) {
	entries := []MotifEntry{
		{Name: "core", Pattern: "1 2 3"},
		{Name: "a", BasedOn: "core", Transpose: 1},     // "2 3 4"
		{Name: "b", BasedOn: "a", Retrograde: true},    // "4 3 2"
	}
	resolved, err := ResolveMotifs(entries)
	if err != nil {
		t.Fatalf("ResolveMotifs: %v", err)
	}

	if got := resolved["core"]; got != "1 2 3" {
		t.Fatalf("core = %q, want %q", got, "1 2 3")
	}
	if got := resolved["a"]; got != "2 3 4" {
		t.Fatalf("a = %q, want %q", got, "2 3 4")
	}
	if got := resolved["b"]; got != "4 3 2" {
		t.Fatalf("b = %q, want %q", got, "4 3 2")
	}
}

// TestResolveMotifs_CycleDetected verifies that a → b → a returns an error.
func TestResolveMotifs_CycleDetected(t *testing.T) {
	entries := []MotifEntry{
		{Name: "a", BasedOn: "b", Pattern: "1 2"},
		{Name: "b", BasedOn: "a", Pattern: "3 4"},
	}
	_, err := ResolveMotifs(entries)
	if err == nil {
		t.Fatal("expected error for cyclic based_on chain, got nil")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected 'cycle' in error message, got: %v", err)
	}
}

// TestRetrogradePattern_BarMarkers confirms that bar markers "|" are treated
// as tokens and reversed along with notes.
func TestRetrogradePattern_BarMarkers(t *testing.T) {
	got := retrogradePattern("1 2 | 3 4")
	if got != "4 3 | 2 1" {
		t.Fatalf("retrogradePattern = %q, want %q", got, "4 3 | 2 1")
	}
}

// TestTransposePattern_NonNumericPreserved confirms rests and bar markers are
// untouched by transposePattern.
func TestTransposePattern_NonNumericPreserved(t *testing.T) {
	got := transposePattern(". 1 | . 3", 2)
	if got != ". 3 | . 5" {
		t.Fatalf("transposePattern = %q, want %q", got, ". 3 | . 5")
	}
}

// TestValidateNotePool_OK verifies no warning for a pool that sums to 1.0.
func TestValidateNotePool_OK(t *testing.T) {
	pool := NotePool{Choices: map[string]float64{"1": 0.4, "3": 0.3, "5": 0.2, "7": 0.1}}
	if warn := ValidateNotePool(pool); warn != "" {
		t.Fatalf("expected no warning, got: %q", warn)
	}
}

// TestValidateNotePool_BadSum verifies a warning when weights don't sum to 1.
func TestValidateNotePool_BadSum(t *testing.T) {
	pool := NotePool{Choices: map[string]float64{"1": 0.4, "3": 0.3}}
	if warn := ValidateNotePool(pool); warn == "" {
		t.Fatal("expected warning for weights not summing to 1.0, got empty string")
	}
}

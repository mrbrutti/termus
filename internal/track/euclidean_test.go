package track

import (
	"strings"
	"testing"
)

func boolsToString(b []bool) string {
	buf := make([]byte, len(b))
	for i, v := range b {
		if v {
			buf[i] = 'x'
		} else {
			buf[i] = '.'
		}
	}
	return string(buf)
}

func TestEuclidean_3_8(t *testing.T) {
	// E(3,8) is the classic tresillo: x..x..x.
	pattern, err := EuclideanRhythm(3, 8, 0)
	if err != nil {
		t.Fatalf("EuclideanRhythm(3,8,0): %v", err)
	}
	if len(pattern) != 8 {
		t.Fatalf("expected length 8, got %d", len(pattern))
	}
	got := boolsToString(pattern)
	want := "x..x..x."
	if got != want {
		t.Fatalf("E(3,8) = %q, want %q", got, want)
	}
	// Verify exactly 3 onsets.
	onsets := 0
	for _, v := range pattern {
		if v {
			onsets++
		}
	}
	if onsets != 3 {
		t.Fatalf("expected 3 onsets, got %d", onsets)
	}
}

func TestEuclidean_5_16(t *testing.T) {
	// E(5,16): 5 pulses in 16 steps — Bjorklund canonical output.
	pattern, err := EuclideanRhythm(5, 16, 0)
	if err != nil {
		t.Fatalf("EuclideanRhythm(5,16,0): %v", err)
	}
	if len(pattern) != 16 {
		t.Fatalf("expected length 16, got %d", len(pattern))
	}
	onsets := 0
	for _, v := range pattern {
		if v {
			onsets++
		}
	}
	if onsets != 5 {
		t.Fatalf("expected 5 onsets, got %d", onsets)
	}
	// Bjorklund output for E(5,16) per this implementation: x..x..x..x..x...
	// (5 pulses distributed as evenly as possible across 16 steps).
	got := boolsToString(pattern)
	want := "x..x..x..x..x..."
	if got != want {
		t.Fatalf("E(5,16) = %q, want %q", got, want)
	}
}

func TestEuclidean_Rotation(t *testing.T) {
	base, err := EuclideanRhythm(3, 8, 0)
	if err != nil {
		t.Fatalf("EuclideanRhythm(3,8,0): %v", err)
	}
	rotated, err := EuclideanRhythm(3, 8, 1)
	if err != nil {
		t.Fatalf("EuclideanRhythm(3,8,1): %v", err)
	}
	if len(rotated) != 8 {
		t.Fatalf("expected length 8, got %d", len(rotated))
	}
	// Rotation by 1: each element i of rotated should equal base[(i+7)%8]
	// because we rotate right by 1 (element 0 of original goes to position 1).
	for i := range rotated {
		expected := base[(i+8-1)%8]
		if rotated[i] != expected {
			t.Fatalf("rotate mismatch at index %d: got %v want %v\nbase=%s rotated=%s",
				i, rotated[i], expected, boolsToString(base), boolsToString(rotated))
		}
	}
	// Verify onset count preserved.
	onsets := 0
	for _, v := range rotated {
		if v {
			onsets++
		}
	}
	if onsets != 3 {
		t.Fatalf("expected 3 onsets after rotation, got %d", onsets)
	}
}

func TestEuclidean_EdgeCases(t *testing.T) {
	// E(0,8): all rests
	p, err := EuclideanRhythm(0, 8, 0)
	if err != nil {
		t.Fatalf("E(0,8): %v", err)
	}
	if boolsToString(p) != "........" {
		t.Fatalf("E(0,8) = %q, want %q", boolsToString(p), "........")
	}

	// E(8,8): all hits
	p, err = EuclideanRhythm(8, 8, 0)
	if err != nil {
		t.Fatalf("E(8,8): %v", err)
	}
	if boolsToString(p) != "xxxxxxxx" {
		t.Fatalf("E(8,8) = %q, want %q", boolsToString(p), "xxxxxxxx")
	}

	// E(9,8): k > n → error
	_, err = EuclideanRhythm(9, 8, 0)
	if err == nil {
		t.Fatal("expected error for E(9,8), got nil")
	}
	if !strings.Contains(err.Error(), "must not exceed") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

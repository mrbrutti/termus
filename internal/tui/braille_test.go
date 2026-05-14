package tui

import (
	"flag"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func TestBrailleRenderSineGolden(t *testing.T) {
	const w, h = 40, 6
	samples := make([]float64, w*2) // two samples per Braille column
	for i := range samples {
		samples[i] = math.Sin(float64(i) * 2 * math.Pi / 20)
	}
	got := RenderBraille(samples, w, h)
	// Renderer must produce exactly h lines (ignoring trailing newline).
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != h {
		t.Fatalf("got %d lines, want %d", len(lines), h)
	}

	golden := filepath.Join("testdata", "sine.golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("missing golden file (run with -update once): %v", err)
	}
	if got != string(want) {
		t.Fatalf("mismatch.\nGOT:\n%s\nWANT:\n%s", got, want)
	}
}

func TestBrailleEmptySamples(t *testing.T) {
	out := RenderBraille(nil, 20, 4)
	if out == "" {
		t.Fatal("RenderBraille(nil) returned empty string")
	}
}

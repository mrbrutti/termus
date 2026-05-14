package audio

import (
	"math"
	"path/filepath"
	"testing"
)

// TestReadIRRoundtripWithWAVWriter writes a known signal through WAVWriter
// then reads it back through ReadIR and compares.
func TestReadIRRoundtripWithWAVWriter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ir.wav")
	w, err := NewWAVWriter(path, 44100, 2)
	if err != nil {
		t.Fatal(err)
	}
	// Frame 0: full-scale L only. Frame 1: full-scale R only.
	// After mono downmix: 0.5 and 0.5 (well, depending on sign; we used 1.0/L and 1.0/R, so positive mean = 0.5).
	frames := [][2]float64{
		{1.0, 0},
		{0, 1.0},
	}
	if err := w.Write(frames); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	got, err := ReadIR(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d samples, want 2", len(got))
	}
	if math.Abs(got[0]-0.5) > 0.01 {
		t.Fatalf("frame 0 mono = %g, want ~0.5", got[0])
	}
	if math.Abs(got[1]-0.5) > 0.01 {
		t.Fatalf("frame 1 mono = %g, want ~0.5", got[1])
	}
}

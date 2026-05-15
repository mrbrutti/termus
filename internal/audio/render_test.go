package audio

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

type flatAlgo struct {
	v float64
}

func (f *flatAlgo) Name() string { return "flat" }
func (f *flatAlgo) Seed(int64)   {}
func (f *flatAlgo) Next(l, r []float64) {
	for i := range l {
		l[i] = f.v
		r[i] = -f.v
	}
}

func TestRenderToWAVCreatesNestedPathAndFrames(t *testing.T) {
	path := filepath.Join(t.TempDir(), "exports", "demo.wav")
	frames, err := RenderToWAV(path, &flatAlgo{v: 0.5}, 0.1, 100)
	if err != nil {
		t.Fatal(err)
	}
	if frames != 4410 {
		t.Fatalf("frames = %d, want 4410", frames)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 44+frames*2*2 {
		t.Fatalf("file size = %d, want %d", len(data), 44+frames*2*2)
	}
	if got := string(data[0:4]); got != "RIFF" {
		t.Fatalf("header = %q, want RIFF", got)
	}
	if firstLeft := int16(binary.LittleEndian.Uint16(data[44:46])); firstLeft <= 0 {
		t.Fatalf("first left sample = %d, want positive PCM", firstLeft)
	}
	if firstRight := int16(binary.LittleEndian.Uint16(data[46:48])); firstRight >= 0 {
		t.Fatalf("first right sample = %d, want negative PCM", firstRight)
	}
}

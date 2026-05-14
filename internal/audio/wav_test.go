package audio

import (
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestWAVWriterRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.wav")
	w, err := NewWAVWriter(path, 44100, 2)
	if err != nil {
		t.Fatal(err)
	}
	// Write 1000 frames of stereo silence + 1000 frames at +0.5.
	frames := make([][2]float64, 2000)
	for i := 1000; i < 2000; i++ {
		frames[i] = [2]float64{0.5, -0.5}
	}
	if err := w.Write(frames); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	// Reopen and verify header + body shape.
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	hdr := make([]byte, 44)
	if _, err := io.ReadFull(f, hdr); err != nil {
		t.Fatal(err)
	}
	if string(hdr[0:4]) != "RIFF" || string(hdr[8:12]) != "WAVE" {
		t.Fatalf("bad RIFF header: %q %q", hdr[0:4], hdr[8:12])
	}
	chunkSize := binary.LittleEndian.Uint32(hdr[4:8])
	dataSize := binary.LittleEndian.Uint32(hdr[40:44])
	if chunkSize != dataSize+36 {
		t.Fatalf("chunk/data size mismatch: chunk=%d data=%d", chunkSize, dataSize)
	}
	if int(dataSize) != 2000*2*2 { // 2000 frames * 2 channels * 2 bytes
		t.Fatalf("data size = %d, want %d", dataSize, 2000*2*2)
	}
}

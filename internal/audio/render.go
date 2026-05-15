package audio

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
	"github.com/mrbrutti/termus/internal/synth"
)

// RenderToWAV renders an algorithm offline to a WAV file without touching the
// live speaker backend. Volume uses the same 0..100 scaling as the TUI.
func RenderToWAV(path string, algo gen.Algorithm, seconds float64, volume int) (written int, err error) {
	if seconds <= 0 {
		return 0, fmt.Errorf("seconds must be > 0, got %.3f", seconds)
	}
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return 0, err
		}
	}

	w, err := NewWAVWriter(path, synth.SampleRate, 2)
	if err != nil {
		return 0, err
	}
	defer func() {
		if closeErr := w.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	root := NewRoot(algo, scope.NewRing(64))
	root.SetVolume(volume)

	totalFrames := int(seconds * float64(synth.SampleRate))
	chunk := 4410
	frames := make([][2]float64, chunk)
	for written < totalFrames {
		n := chunk
		if remain := totalFrames - written; remain < n {
			n = remain
		}
		if _, ok := root.Stream(frames[:n]); !ok {
			return written, fmt.Errorf("audio stream ended after %d frames", written)
		}
		if err := w.Write(frames[:n]); err != nil {
			return written, err
		}
		written += n
	}
	return written, nil
}

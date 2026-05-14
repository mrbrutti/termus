package audio

import (
	"math"
	"testing"

	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
)

func TestRootProducesAudioAndFeedsScope(t *testing.T) {
	ring := scope.NewRing(2048)
	algo := gen.NewEno()
	algo.Seed(7)
	root := NewRoot(algo, ring)
	root.SetVolume(100)

	// Pull one second of audio.
	frames := make([][2]float64, 44100)
	n, ok := root.Stream(frames)
	if !ok || n != len(frames) {
		t.Fatalf("Stream returned (%d, %v), want (%d, true)", n, ok, len(frames))
	}

	var sumSq float64
	for _, f := range frames {
		sumSq += f[0]*f[0] + f[1]*f[1]
	}
	rms := math.Sqrt(sumSq / float64(2*len(frames)))
	if rms < 0.01 {
		t.Fatalf("root RMS=%g, expected > 0.01", rms)
	}

	// Scope should have received samples.
	snap := make([]float64, 64)
	ring.Snapshot(snap)
	anyNonZero := false
	for _, v := range snap {
		if v != 0 {
			anyNonZero = true
			break
		}
	}
	if !anyNonZero {
		t.Fatal("scope ring received no samples")
	}
}

func TestRootPauseSilences(t *testing.T) {
	ring := scope.NewRing(64)
	algo := gen.NewEno()
	algo.Seed(7)
	root := NewRoot(algo, ring)
	root.SetVolume(100)
	root.TogglePause()

	frames := make([][2]float64, 4096)
	root.Stream(frames)
	for i, f := range frames {
		if f[0] != 0 || f[1] != 0 {
			t.Fatalf("paused but frame %d = (%g, %g)", i, f[0], f[1])
		}
	}
}

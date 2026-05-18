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

// constantAlgo emits a constant value on both channels — used to verify the
// crossfade gain envelope in isolation.
type constantAlgo struct {
	v float64
	n string
}

func (c *constantAlgo) Name() string { return c.n }
func (c *constantAlgo) Seed(int64)   {}
func (c *constantAlgo) Next(l, r []float64) {
	for i := range l {
		l[i] = c.v
		r[i] = c.v
	}
}

type gainedAlgo struct {
	constantAlgo
	sectionGain float64
}

func (g *gainedAlgo) SectionGain() float64 { return g.sectionGain }

type statusAlgo struct {
	constantAlgo
	status gen.DebugStatus
}

func (s *statusAlgo) DebugStatus() gen.DebugStatus { return s.status }

func TestRootCrossfadeSwap(t *testing.T) {
	ring := scope.NewRing(2048)
	old := &constantAlgo{v: 1.0, n: "old"}
	root := NewRoot(old, ring)
	root.SetVolume(100)

	const fade = 4410 // 100 ms
	newer := &constantAlgo{v: -1.0, n: "new"}
	root.SwapAlgorithmFade(newer, fade)

	// Pull a buffer that covers fade-out + fade-in + a tail of steady-state.
	totalFrames := fade*2 + 1000
	frames := make([][2]float64, totalFrames)
	root.Stream(frames)

	// Frame 0: full old gain (≈1.0).
	if math.Abs(frames[0][0]-1.0) > 1e-6 {
		t.Errorf("frame 0 = %g, want ≈ 1.0", frames[0][0])
	}
	// Frame fade-1 (last fade-out frame): nearly zero.
	if math.Abs(frames[fade-1][0]) > 0.001 {
		t.Errorf("last fade-out frame %g, want ≈ 0", frames[fade-1][0])
	}
	// Frame fade (first fade-in frame): nearly zero of new sign.
	if math.Abs(frames[fade][0]) > 0.001 {
		t.Errorf("first fade-in frame %g, want ≈ 0", frames[fade][0])
	}
	// Frame fade*2 (first post-fade frame): full new value (-1.0).
	if math.Abs(frames[fade*2][0]+1.0) > 1e-6 {
		t.Errorf("first post-fade frame = %g, want ≈ -1.0", frames[fade*2][0])
	}
	// Sample inside fade-out should have positive sign (old algo) and smaller magnitude.
	mid := fade / 2
	if frames[mid][0] <= 0.0 || frames[mid][0] >= 1.0 {
		t.Errorf("mid fade-out frame = %g, expected in (0, 1)", frames[mid][0])
	}
	// Sample inside fade-in should have negative sign (new algo) and smaller magnitude.
	midIn := fade + fade/2
	if frames[midIn][0] >= 0.0 || frames[midIn][0] <= -1.0 {
		t.Errorf("mid fade-in frame = %g, expected in (-1, 0)", frames[midIn][0])
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

func TestToggleRecordFailsBeforeAudioStarts(t *testing.T) {
	ring := scope.NewRing(64)
	algo := gen.NewEno()
	algo.Seed(7)
	root := NewRoot(algo, ring)

	if _, err := root.ToggleRecord(); err == nil {
		t.Fatal("expected ToggleRecord to fail before Stream starts")
	}
}

func TestRootAppliesEffectiveOutputGain(t *testing.T) {
	ring := scope.NewRing(64)
	root := NewRoot(&gainedAlgo{
		constantAlgo: constantAlgo{v: 1.0, n: "glass-fm"},
		sectionGain:  0.8,
	}, ring)
	root.SetVolume(100)

	frames := make([][2]float64, 16)
	root.Stream(frames)
	want := 0.85 * 0.8
	if math.Abs(frames[0][0]-want) > 1e-6 {
		t.Fatalf("frame 0 = %g, want %g", frames[0][0], want)
	}
}

func TestRootPublishesDebugStatus(t *testing.T) {
	ring := scope.NewRing(64)
	root := NewRoot(&statusAlgo{
		constantAlgo: constantAlgo{v: 0.1, n: "stub"},
		status:       gen.DebugStatus{Chord: "Dm7", Bar: 2},
	}, ring)
	status := root.DebugStatus()
	if status.Chord != "Dm7" || status.Bar != 2 {
		t.Fatalf("initial status = %+v", status)
	}
}

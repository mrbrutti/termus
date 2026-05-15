package audio

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/mrbrutti/termus/internal/gen"
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

type markerAlgo struct {
	flatAlgo
	markers []gen.ListeningMarker
}

func (m *markerAlgo) ListeningMarkers() []gen.ListeningMarker {
	return append([]gen.ListeningMarker(nil), m.markers...)
}

func TestRenderToWAVCreatesNestedPathAndFrames(t *testing.T) {
	path := filepath.Join(t.TempDir(), "exports", "demo.wav")
	exact := RenderPlan{
		RequestedFrames: 4410,
		FadeStartFrame:  4410,
		FadeFrames:      0,
		TotalFrames:     4410,
	}
	frames, err := RenderToWAVWithPlan(path, &flatAlgo{v: 0.5}, exact, 100)
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

func TestPlanRenderSnapsToNearbyCadence(t *testing.T) {
	plan := PlanRender(&markerAlgo{
		flatAlgo: flatAlgo{v: 0.2},
		markers: []gen.ListeningMarker{
			{Label: "bar", Sample: 40000},
			{Label: "cadence:cadence", Sample: 50000},
			{Label: "section:outro", Sample: 400000},
		},
	}, float64(44100)/44100.0)
	if plan.FadeStartFrame != 50000 {
		t.Fatalf("fade start = %d, want cadence frame 50000", plan.FadeStartFrame)
	}
	if plan.SnapLabel != "cadence:cadence" {
		t.Fatalf("snap label = %q, want cadence:cadence", plan.SnapLabel)
	}
	if plan.TotalFrames <= plan.FadeStartFrame {
		t.Fatalf("total frames = %d, want tail beyond fade start %d", plan.TotalFrames, plan.FadeStartFrame)
	}
}

func TestPlanRenderPrefersNearbyOutro(t *testing.T) {
	plan := PlanRender(&markerAlgo{
		flatAlgo: flatAlgo{v: 0.2},
		markers: []gen.ListeningMarker{
			{Label: "cadence:cadence", Sample: 70000},
			{Label: "section:outro", Sample: 60000},
		},
	}, float64(44100)/44100.0)
	if plan.FadeStartFrame != 60000 {
		t.Fatalf("fade start = %d, want outro frame 60000", plan.FadeStartFrame)
	}
	if plan.SnapLabel != "section:outro" {
		t.Fatalf("snap label = %q, want section:outro", plan.SnapLabel)
	}
}

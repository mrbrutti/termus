package gen

import (
	"math/rand"
	"testing"
)

func TestComputeTimeLayerWindowDetectsBoundaries(t *testing.T) {
	const barSamples = int64(100)
	plan := NewEpisodePlan(rand.New(rand.NewSource(1)), barSamples, "jazz") //nolint:gosec
	plan.ensureBars(64)

	barWindow := ComputeTimeLayerWindow(&plan, 10, barSamples+10)
	if !barWindow.BarChanged {
		t.Fatalf("expected bar boundary when crossing first bar")
	}
	if barWindow.SectionChanged {
		t.Fatalf("did not expect section boundary on a single-bar step")
	}

	sectionBar := plan.episodes[0].Sections[0].Bars
	sectionWindow := ComputeTimeLayerWindow(&plan, int64(sectionBar)*barSamples-1, int64(sectionBar)*barSamples+1)
	if !sectionWindow.SectionChanged {
		t.Fatalf("expected section boundary at bar %d", sectionBar)
	}

	episodeBar := plan.episodes[1].StartBar
	episodeWindow := ComputeTimeLayerWindow(&plan, int64(episodeBar)*barSamples-1, int64(episodeBar)*barSamples+1)
	if !episodeWindow.EpisodeChanged {
		t.Fatalf("expected episode boundary at bar %d", episodeBar)
	}
	if !episodeWindow.SectionChanged {
		t.Fatalf("episode boundary should also surface as a section boundary")
	}
}

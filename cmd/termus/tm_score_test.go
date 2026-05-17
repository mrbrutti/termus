package main

import (
	"testing"

	"github.com/mrbrutti/termus/internal/gen"
)

func TestMergeProfilesOffsetsAroundNeutral(t *testing.T) {
	base := gen.ControlProfile{Density: 1, Brightness: 3, Motion: 2, Reverb: 2, Swing: 2, DroneDepth: 2, Tempo: 2, Phrase: 2}
	overlay := gen.ControlProfile{Density: 4, Brightness: 0, Motion: 2, Reverb: 2, Swing: 2, DroneDepth: 2, Tempo: 2, Phrase: 2}
	got := mergeProfiles(base, overlay)
	if got.Density != 3 {
		t.Fatalf("density = %d, want 3", got.Density)
	}
	if got.Brightness != 1 {
		t.Fatalf("brightness = %d, want 1", got.Brightness)
	}
}

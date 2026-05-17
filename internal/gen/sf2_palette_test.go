package gen

import "testing"

func TestMaxSF2PresetsForSpecAmbient(t *testing.T) {
	spec, ok := Resolve("ambient")
	if !ok {
		t.Fatal("ambient spec missing")
	}
	got := MaxSF2PresetsForSpec(spec)
	found := map[string]bool{}
	for _, name := range got {
		found[name] = true
	}
	for _, want := range []string{"arachno", "fairy-tale"} {
		if !found[want] {
			t.Fatalf("ambient max preset %q missing: %v", want, got)
		}
	}
	if found["general"] {
		t.Fatalf("ambient max should avoid generic fallback when a curated pool exists: %v", got)
	}
}

func TestMaxSF2PresetsForSpecLofiIncludesSharedAndAlternateBanks(t *testing.T) {
	spec, ok := Resolve("lofi")
	if !ok {
		t.Fatal("lofi spec missing")
	}
	got := MaxSF2PresetsForSpec(spec)
	found := map[string]bool{}
	for _, name := range got {
		found[name] = true
	}
	for _, want := range []string{"sgm", "tyros4", "fatboy"} {
		if !found[want] {
			t.Fatalf("lofi max preset %q missing: %v", want, got)
		}
	}
	if found["general"] {
		t.Fatalf("lofi max should avoid generic fallback when more characterful banks exist: %v", got)
	}
}

func TestMaxSF2PresetsForSpecBellsUsesCuratedPool(t *testing.T) {
	spec, ok := Resolve("bells")
	if !ok {
		t.Fatal("bells spec missing")
	}
	got := MaxSF2PresetsForSpec(spec)
	found := map[string]bool{}
	for _, name := range got {
		found[name] = true
	}
	for _, want := range []string{"fairy-tale", "arachno", "timbres-of-heaven"} {
		if !found[want] {
			t.Fatalf("bells max preset %q missing: %v", want, got)
		}
	}
	if found["general"] {
		t.Fatalf("bells max should avoid generic fallback when curated bell banks exist: %v", got)
	}
}

func TestSoftStyleMaxSelectionKeepsAmbientSupportRolesOnPrimaryPalette(t *testing.T) {
	spec, ok := Resolve("ambient")
	if !ok {
		t.Fatal("ambient spec missing")
	}
	plan := &AuthoredTrackPlan{
		Tracks: []AuthoredRenderTrack{
			{Name: "pad", Family: "pad", Tone: []string{"warm", "wide"}, Articulation: "sustain", Register: "mid", Prominence: "support", Channel: 0},
			{Name: "choir", Family: "choir", Tone: []string{"airy"}, Articulation: "sustain", Register: "high", Prominence: "support", Channel: 1},
			{Name: "bells", Family: "bells", Tone: []string{"sparkle"}, Articulation: "bloom", Register: "air", Prominence: "air", Channel: 2},
			{Name: "bass", Family: "synth_bass", Tone: []string{"warm"}, Articulation: "sustain", Register: "low", Prominence: "anchor", Channel: 4},
		},
	}
	selection := ResolveSF2SelectionForPlan(spec, plan, "max", "general")
	if selection.Primary != "arachno" {
		t.Fatalf("ambient max primary = %q, want arachno", selection.Primary)
	}
	for _, channel := range []int32{0, 1, 4} {
		if got := selection.Routes[channel]; got != selection.Primary {
			t.Fatalf("ambient support channel %d routed to %q, want primary %q", channel, got, selection.Primary)
		}
	}
	if got := selection.Routes[2]; got == "general" {
		t.Fatalf("ambient sparkle channel should stay curated, got %q", got)
	}
}

func TestSoftStyleMaxSelectionUsesLimitedBanksForBells(t *testing.T) {
	spec, ok := Resolve("bells")
	if !ok {
		t.Fatal("bells spec missing")
	}
	plan := &AuthoredTrackPlan{
		Tracks: []AuthoredRenderTrack{
			{Name: "bells", Family: "bells", Tone: []string{"glass", "sparkle"}, Articulation: "bloom", Register: "high", Prominence: "lead", Channel: 0},
			{Name: "celesta", Family: "mallet", Tone: []string{"delicate"}, Articulation: "echo", Register: "high", Prominence: "air", Channel: 1},
			{Name: "pad", Family: "pad", Tone: []string{"soft", "celestial"}, Articulation: "sustain", Register: "mid-high", Prominence: "support", Channel: 4},
			{Name: "choir", Family: "choir", Tone: []string{"airy"}, Articulation: "sustain", Register: "high", Prominence: "support", Channel: 5},
			{Name: "strings", Family: "strings", Tone: []string{"soft"}, Articulation: "sustain", Register: "mid-high", Prominence: "support", Channel: 6},
		},
	}
	selection := ResolveSF2SelectionForPlan(spec, plan, "max", "general")
	if selection.Primary != "fairy-tale" {
		t.Fatalf("bells max primary = %q, want fairy-tale", selection.Primary)
	}
	if got := selection.Routes[0]; got != "fairy-tale" {
		t.Fatalf("bells lead route = %q, want fairy-tale", got)
	}
	if got := selection.Routes[1]; got != "fairy-tale" {
		t.Fatalf("bells celesta route = %q, want fairy-tale", got)
	}
	if got := selection.Routes[4]; got != "fairy-tale" {
		t.Fatalf("bells pad route = %q, want fairy-tale", got)
	}
	if got := selection.Routes[5]; got != "fairy-tale" {
		t.Fatalf("bells choir route = %q, want fairy-tale", got)
	}
	distinct := map[string]bool{}
	for _, preset := range selection.Routes {
		distinct[preset] = true
	}
	if len(distinct) > 2 {
		t.Fatalf("bells max should stay cohesive, got %d banks: %v", len(distinct), selection.Routes)
	}
}

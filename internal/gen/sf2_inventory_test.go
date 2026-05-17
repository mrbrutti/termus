package gen

import "testing"

func TestInventoryPresetProfileIncludesRoleAwareMetadata(t *testing.T) {
	profile, ok := InventoryPresetProfile("fairy-tale")
	if !ok {
		t.Fatal("fairy-tale preset missing")
	}
	if profile.Realism == "" {
		t.Fatal("expected realism tag")
	}
	if len(profile.Blend) == 0 {
		t.Fatal("expected blend tags")
	}
	if len(profile.Articulations) == 0 {
		t.Fatal("expected articulation tags")
	}
	if len(profile.Programs) == 0 {
		t.Fatal("expected program profiles")
	}
	foundFamily := false
	for _, program := range profile.Programs {
		if program.Family == "bells" && program.Program == 14 {
			foundFamily = true
			if program.Realism == "" {
				t.Fatal("expected program realism tag")
			}
			if len(program.Blend) == 0 {
				t.Fatal("expected program blend tags")
			}
			if len(program.Articulations) == 0 {
				t.Fatal("expected program articulations")
			}
		}
	}
	if !foundFamily {
		t.Fatal("expected bells program profile in fairy-tale")
	}
}

func TestInventoryProgramProfilesExposeMultipleFamilies(t *testing.T) {
	programs := InventoryProgramProfiles("tyros4")
	if len(programs) == 0 {
		t.Fatal("expected tyros4 program profiles")
	}
	found := map[string]bool{}
	for _, program := range programs {
		found[program.Family] = true
	}
	for _, family := range []string{"reed_lead", "brass", "organ"} {
		if !found[family] {
			t.Fatalf("expected family %q in tyros4 program profiles", family)
		}
	}
}

func TestResolveSF2SelectionForPlanProChoosesProgramsByRole(t *testing.T) {
	spec, ok := Resolve("jazz")
	if !ok {
		t.Fatal("jazz spec missing")
	}
	plan := &AuthoredTrackPlan{
		Tracks: []AuthoredRenderTrack{
			{Name: "piano", Family: "acoustic_piano", Tone: []string{"warm"}, Articulation: "comp", Register: "mid", Prominence: "support", Channel: 0},
			{Name: "bass", Family: "bass", Tone: []string{"woody"}, Articulation: "walk", Register: "low", Prominence: "anchor", Channel: 1},
			{Name: "trumpet", Family: "brass", Tone: []string{"present"}, Articulation: "lyrical", Register: "mid-high", Prominence: "lead", Channel: 2},
			{Name: "drums", Family: "drums", Tone: []string{"live"}, Articulation: "swing", Register: "mid", Prominence: "anchor", Channel: 9},
		},
	}
	selection := ResolveSF2SelectionForPlan(spec, plan, "pro", "general")
	if selection.Primary == "" {
		t.Fatal("expected primary preset")
	}
	if len(selection.Programs) != 4 {
		t.Fatalf("expected 4 program assignments, got %d", len(selection.Programs))
	}
	if got := selection.Programs[0]; got != 0 {
		t.Fatalf("piano program = %d, want 0", got)
	}
	if got := selection.Programs[1]; got != 32 {
		t.Fatalf("bass program = %d, want 32", got)
	}
	if got := selection.Programs[2]; got != 56 {
		t.Fatalf("trumpet program = %d, want 56", got)
	}
	if got := selection.Programs[9]; got != 0 {
		t.Fatalf("drums program = %d, want 0", got)
	}
}

func TestResolveSF2SelectionForPlanMaxRoutesLeadAndPadSeparately(t *testing.T) {
	spec, ok := Resolve("bells")
	if !ok {
		t.Fatal("bells spec missing")
	}
	plan := &AuthoredTrackPlan{
		Tracks: []AuthoredRenderTrack{
			{Name: "bells", Family: "bells", Tone: []string{"glass", "sparkle"}, Articulation: "bloom", Register: "high", Prominence: "lead", Channel: 0},
			{Name: "pad", Family: "pad", Tone: []string{"soft", "celestial"}, Articulation: "sustain", Register: "mid-high", Prominence: "support", Channel: 4},
			{Name: "choir", Family: "choir", Tone: []string{"airy"}, Articulation: "sustain", Register: "high", Prominence: "support", Channel: 5},
		},
	}
	selection := ResolveSF2SelectionForPlan(spec, plan, "max", "general")
	if len(selection.Routes) != 3 {
		t.Fatalf("expected 3 routed channels, got %d", len(selection.Routes))
	}
	if got := selection.Programs[0]; got != 14 {
		t.Fatalf("bells program = %d, want 14", got)
	}
	if got := selection.Programs[4]; got != 89 {
		t.Fatalf("pad program = %d, want 89", got)
	}
	if got := selection.Programs[5]; got != 52 {
		t.Fatalf("choir program = %d, want 52", got)
	}
	if route := selection.Routes[0]; route == "" {
		t.Fatal("expected bells channel route")
	}
	if route := selection.Routes[4]; route == "" {
		t.Fatal("expected pad channel route")
	}
}

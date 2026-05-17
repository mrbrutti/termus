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

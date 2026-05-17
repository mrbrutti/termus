package gen

import "testing"

func TestParseChillBlueprintHarmonyAndRoles(t *testing.T) {
	got := parseChillBlueprint(TrackBlueprint{
		Harmony: "Dm9 G13 | Cmaj9 Am7",
		Roles: map[string]RoleBlueprint{
			"lead":    {Motif: "9 . b9 7 | 5 - 3 1", Active: true},
			"keys":    {Pattern: "x . . x | . x . x", Active: true},
			"texture": {Pattern: "x . . x | . x . x", Active: true},
			"drums":   {Pattern: "bd: x... x..x | sd: ..x. ..x. | hh: x.x.x.x. | fill: .... ...x", Active: true},
			"comp":    {Active: true},
		},
	})
	if !got.hasTonic || got.tonicPC != 2 {
		t.Fatalf("tonic = (%v,%d), want (true,2)", got.hasTonic, got.tonicPC)
	}
	if gotRoot := chordRootSemi(got.progression[1]); gotRoot != 5 {
		t.Fatalf("second chord root = %d, want 5", gotRoot)
	}
	if len(got.saxPhrase) == 0 || got.saxPhrase[2] != chillPlanPickupBelow {
		t.Fatalf("lead phrase did not preserve pickup-below contour: %v", got.saxPhrase)
	}
	if len(got.vibePhrase) != 4 {
		t.Fatalf("vibe phrase length = %d, want 4", len(got.vibePhrase))
	}
	if !got.roles["lead"] || !got.roles["comp"] || !got.roles["texture"] {
		t.Fatalf("roles = %#v, expected lead/comp/texture enabled", got.roles)
	}
	if !got.drums.fillHeavy || got.drums.hatDensity == 0 {
		t.Fatalf("drums = %#v, expected fill-heavy with hats", got.drums)
	}
}

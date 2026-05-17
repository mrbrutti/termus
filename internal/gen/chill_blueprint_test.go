package gen

import "testing"

func TestParseChillBlueprintHarmonyAndRoles(t *testing.T) {
	got := parseChillBlueprint(ScoreBlueprint{
		Harmony: "Dm9 G13 | Cmaj9 Am7",
		Lead:    "9 . b9 7 | 5 - 3 1",
		Comp:    "x . . x | . x . x",
		Drums:   "bd: x... x..x | sd: ..x. ..x. | hh: x.x.x.x. | fill: .... ...x",
		Arrange: "bass drums comp +lead +texture",
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

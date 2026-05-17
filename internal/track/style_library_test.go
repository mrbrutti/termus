package track

import "testing"

func TestApplyStyleLibraryAddsJazzHeadAndBridgeDefaults(t *testing.T) {
	pack := resolveStylePack("jazz", "vibes-cellar", "Basement Blue Hour", []string{"jazz", "vibes", "basement"})
	section := Section{
		ID:        "head",
		Title:     "cellar head",
		Scene:     "head clipped",
		Variation: "statement",
	}
	roles := map[string]Role{
		"lead":  {Family: "reed_lead", Motif: "5 . 6 7 | 9 . 7 3"},
		"piano": {Family: "acoustic_piano", Pattern: "x..x .x.."},
		"bass":  {Family: "bass", Pattern: "x... x..."},
		"ride":  {Family: "drums", Pattern: "x.x. x.x."},
		"snare": {Family: "drums", Pattern: ".... x..."},
	}
	section, roles = applyStyleLibrary(pack, section, roles)
	if len(section.Events) < 2 {
		t.Fatalf("expected jazz library events, got %d", len(section.Events))
	}
	if roles["lead"].Phrases["cadence"].Motif == "" {
		t.Fatal("expected jazz lead cadence motif from style library")
	}
	if roles["piano"].Phrases["answer"].Pattern == "" {
		t.Fatal("expected jazz comp answer phrase from style library")
	}

	bridge := Section{
		ID:        "bridge",
		Title:     "cellar bridge",
		Scene:     "bridge reharm",
		Variation: "sequence",
	}
	bridge, _ = applyStyleLibrary(pack, bridge, roles)
	foundStop := false
	foundStab := false
	for _, event := range bridge.Events {
		switch event.Kind {
		case "stop":
			foundStop = true
		case "stab":
			foundStab = true
		}
	}
	if !foundStop || !foundStab {
		t.Fatalf("expected bridge stop+stab, got %+v", bridge.Events)
	}
}

func TestApplyStyleLibraryAddsLofiBreakdownDefaults(t *testing.T) {
	pack := resolveStylePack("lofi", "dusty-rhodes", "Soft Tape / Rain Bus", []string{"lofi", "rain", "bus"})
	section := Section{
		ID:        "breakdown",
		Title:     "quiet block",
		Scene:     "breakdown thin",
		Variation: "subtract",
	}
	roles := map[string]Role{
		"lead":  {Family: "guitar", Motif: "5 . 6 7 | 9 . 7 5"},
		"keys":  {Family: "electric_piano", Pattern: "x..x .x.."},
		"pad":   {Family: "pad", Pattern: "x......."},
		"bass":  {Family: "bass", Pattern: "x... x..."},
		"kick":  {Family: "drums", Pattern: "x... x..."},
		"snare": {Family: "drums", Pattern: ".... x..."},
	}
	section, roles = applyStyleLibrary(pack, section, roles)
	foundDrop := false
	foundHold := false
	foundBreath := false
	for _, event := range section.Events {
		switch event.Kind {
		case "drop":
			foundDrop = true
		case "hold":
			foundHold = true
		case "breath":
			foundBreath = true
		}
	}
	if !foundDrop || !foundHold || !foundBreath {
		t.Fatalf("expected lofi breakdown defaults, got %+v", section.Events)
	}
	if roles["lead"].Phrases["release"].Motif == "" {
		t.Fatal("expected lofi melody release phrase from style library")
	}
}

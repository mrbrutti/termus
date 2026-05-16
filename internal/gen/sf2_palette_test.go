package gen

import "testing"

func TestMaxSF2PresetsForSpecAmbient(t *testing.T) {
	spec, ok := Resolve("ambient")
	if !ok {
		t.Fatal("ambient spec missing")
	}
	got := MaxSF2PresetsForSpec(spec)
	want := map[string]bool{
		"arachno":           true,
		"fairy-tale":        true,
		"fm-dx":             true,
		"merlin-symphony":   true,
		"musescore-general": true,
	}
	if len(got) != len(want) {
		t.Fatalf("ambient max preset count = %d, want %d: %v", len(got), len(want), got)
	}
	for _, name := range got {
		if !want[name] {
			t.Fatalf("ambient max preset %q not expected: %v", name, got)
		}
		delete(want, name)
	}
	if len(want) != 0 {
		t.Fatalf("ambient max preset set incomplete, still missing: %v", want)
	}
}

func TestMaxSF2PresetsForSpecLofiIncludesSharedAndAlternateBanks(t *testing.T) {
	spec, ok := Resolve("lofi")
	if !ok {
		t.Fatal("lofi spec missing")
	}
	got := MaxSF2PresetsForSpec(spec)
	want := map[string]bool{
		"fatboy":  true,
		"sgm":     true,
		"dsound4": true,
		"tyros4":  true,
	}
	if len(got) != len(want) {
		t.Fatalf("lofi max preset count = %d, want %d: %v", len(got), len(want), got)
	}
	for _, name := range got {
		if !want[name] {
			t.Fatalf("lofi max preset %q not expected: %v", name, got)
		}
		delete(want, name)
	}
	if len(want) != 0 {
		t.Fatalf("lofi max preset set incomplete, still missing: %v", want)
	}
}

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

package main

import (
	"testing"

	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/track"
)

func TestNormalizeSF2Strategy(t *testing.T) {
	tests := map[string]string{
		"single":  "single",
		"pro":     "pro",
		"optimal": "pro",
		"max":     "max",
		"":        "single",
	}
	for input, want := range tests {
		got, ok := normalizeSF2Strategy(input)
		if !ok {
			t.Fatalf("%q should be accepted", input)
		}
		if got != want {
			t.Fatalf("normalizeSF2Strategy(%q) = %q, want %q", input, got, want)
		}
	}
	if _, ok := normalizeSF2Strategy("weird"); ok {
		t.Fatal("unexpectedly accepted invalid sf2 strategy")
	}
}

func TestNeededPresetsProUsesPreferredSet(t *testing.T) {
	spec, _ := gen.Resolve("ambient")
	got := neededPresets("pro", "general", spec)
	found := map[string]bool{}
	for _, name := range got {
		found[name] = true
	}
	for _, name := range []string{"general", "arachno", "fairy-tale", "timbres-of-heaven", "sgm"} {
		if !found[name] {
			t.Fatalf("neededPresets(pro) missing %q: %v", name, got)
		}
	}
	for _, name := range []string{"dsound4", "fatboy", "musescore-general"} {
		if found[name] {
			t.Fatalf("neededPresets(pro) should stay curated, but included %q: %v", name, got)
		}
	}
}

func TestNeededPresetsMaxLoadsBeyondPreferredSet(t *testing.T) {
	spec, _ := gen.Resolve("ambient")
	got := neededPresets("max", "general", spec)
	found := map[string]bool{}
	for _, name := range got {
		found[name] = true
	}
	for _, want := range []string{"general", "arachno", "fm-dx", "fairy-tale", "timbres-of-heaven", "sgm", "tyros4", "dsound4", "fatboy", "merlin-symphony", "musescore-general"} {
		if !found[want] {
			t.Fatalf("neededPresets(max) missing %q: %v", want, got)
		}
	}
}

func TestStartupLabelShowsStationAndAlgo(t *testing.T) {
	spec, _ := gen.Resolve("jazz")
	if got := startupLabel(spec); got != "Dusty Swing · jazz" {
		t.Fatalf("startupLabel(jazz) = %q", got)
	}
}

func TestShouldOpenTrackLibraryByDefault(t *testing.T) {
	entries := []track.Entry{{ID: "lofi/demo"}}
	if !shouldOpenTrackLibraryByDefault(entries, map[string]bool{}, "", "", "", "") {
		t.Fatal("expected bare startup to prefer the authored track library")
	}
	if shouldOpenTrackLibraryByDefault(entries, map[string]bool{"algo": true}, "", "", "", "") {
		t.Fatal("explicit algo should disable default track library startup")
	}
	if shouldOpenTrackLibraryByDefault(entries, map[string]bool{}, "lofi/demo", "", "", "") {
		t.Fatal("explicit track should disable default track library startup")
	}
}

func TestShouldWarmStartupSF2SkipsDefaultTrackBrowser(t *testing.T) {
	spec, _ := gen.Resolve("jazz")
	if !shouldWarmStartupSF2(false, true, spec) {
		t.Fatal("expected live sf2 startup to warm when booting directly into a score")
	}
	if shouldWarmStartupSF2(true, true, spec) {
		t.Fatal("default track browser startup should not block on sf2 warmup")
	}
}

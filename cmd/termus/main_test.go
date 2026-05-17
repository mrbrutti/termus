package main

import (
	"testing"

	"github.com/mrbrutti/termus/internal/gen"
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

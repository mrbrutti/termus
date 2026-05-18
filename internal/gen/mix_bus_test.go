package gen

import (
	"testing"

	"github.com/mrbrutti/termus/internal/synth"
)

func TestMixBusLibraryHasFourProfiles(t *testing.T) {
	lib := MixBusLibrary()
	if got := len(lib); got != 4 {
		t.Fatalf("MixBusLibrary() returned %d profiles, want 4", got)
	}
}

func TestMixBusByNameResolves(t *testing.T) {
	names := []string{"lofi", "jazz", "chill", "ambient"}
	for _, name := range names {
		p := MixBusByName(name)
		if p == nil {
			t.Fatalf("MixBusByName(%q) returned nil, want non-nil", name)
		}
		if p.Name != name {
			t.Fatalf("MixBusByName(%q).Name = %q, want %q", name, p.Name, name)
		}
	}

	// Missing name should return nil.
	if got := MixBusByName("nonexistent"); got != nil {
		t.Fatalf("MixBusByName(%q) = %+v, want nil", "nonexistent", got)
	}
}

func TestLofiProfileHasWowFlutter(t *testing.T) {
	p := MixBusByName("lofi")
	if p == nil {
		t.Fatal("MixBusByName(\"lofi\") returned nil")
	}
	if p.WowFlutter == nil {
		t.Fatal("lofi profile: WowFlutter is nil, want non-nil")
	}
}

func TestJazzProfileHasNoWowFlutter(t *testing.T) {
	p := MixBusByName("jazz")
	if p == nil {
		t.Fatal("MixBusByName(\"jazz\") returned nil")
	}
	if p.WowFlutter != nil {
		t.Fatalf("jazz profile: WowFlutter = %+v, want nil", p.WowFlutter)
	}
}

func TestAmbientProfileHasNoVinyl(t *testing.T) {
	p := MixBusByName("ambient")
	if p == nil {
		t.Fatal("MixBusByName(\"ambient\") returned nil")
	}
	if p.Vinyl != nil {
		t.Fatalf("ambient profile: Vinyl = %+v, want nil", p.Vinyl)
	}
}

func TestAllProfilesHaveValidIRName(t *testing.T) {
	for _, p := range MixBusLibrary() {
		if p.ReverbBusIRName == "" {
			t.Errorf("profile %q: ReverbBusIRName is empty", p.Name)
			continue
		}
		preset := synth.IRByName(p.ReverbBusIRName)
		if preset == nil {
			t.Errorf("profile %q: synth.IRByName(%q) returned nil; IR not in library", p.Name, p.ReverbBusIRName)
		}
	}
}

package gen

import "testing"

func TestGrooveLibraryHasFourTemplates(t *testing.T) {
	lib := GrooveLibrary()
	if got := len(lib); got != 4 {
		t.Fatalf("GrooveLibrary() returned %d templates, want exactly 4", got)
	}
}

func TestGrooveByNameRoundtrip(t *testing.T) {
	lib := GrooveLibrary()
	for _, tmpl := range lib {
		got := GrooveByName(tmpl.Name)
		if got == nil {
			t.Fatalf("GrooveByName(%q) returned nil", tmpl.Name)
		}
		if got.Name != tmpl.Name {
			t.Fatalf("GrooveByName(%q).Name = %q, want %q", tmpl.Name, got.Name, tmpl.Name)
		}
	}
}

func TestGrooveByName_Unknown(t *testing.T) {
	got := GrooveByName("nonexistent_groove_xyz")
	if got != nil {
		t.Fatalf("expected nil for unknown groove name, got %+v", got)
	}
}

func TestStraightTemplateIsZeroOffsets(t *testing.T) {
	tmpl := GrooveByName("straight")
	if tmpl == nil {
		t.Fatal("GrooveByName(\"straight\") returned nil")
	}
	for i, offset := range tmpl.TimingOffsetsSamples {
		if offset != 0 {
			t.Errorf("straight groove: TimingOffsetsSamples[%d] = %d, want 0", i, offset)
		}
	}
	for i, mult := range tmpl.VelocityMultipliers {
		if mult != 1.0 {
			t.Errorf("straight groove: VelocityMultipliers[%d] = %f, want 1.0", i, mult)
		}
	}
}

func TestGrooveTemplateNamesAreUnique(t *testing.T) {
	lib := GrooveLibrary()
	seen := make(map[string]bool, len(lib))
	for _, tmpl := range lib {
		if seen[tmpl.Name] {
			t.Fatalf("duplicate groove name %q in library", tmpl.Name)
		}
		seen[tmpl.Name] = true
	}
}

func TestGrooveTemplateArrayLengths(t *testing.T) {
	lib := GrooveLibrary()
	for _, tmpl := range lib {
		if len(tmpl.TimingOffsetsSamples) != 16 {
			t.Errorf("groove %q: TimingOffsetsSamples has length %d, want 16",
				tmpl.Name, len(tmpl.TimingOffsetsSamples))
		}
		if len(tmpl.VelocityMultipliers) != 16 {
			t.Errorf("groove %q: VelocityMultipliers has length %d, want 16",
				tmpl.Name, len(tmpl.VelocityMultipliers))
		}
	}
}

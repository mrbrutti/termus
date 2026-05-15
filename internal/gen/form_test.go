package gen

import "testing"

func TestFormMarkersLandOnBarBoundaries(t *testing.T) {
	form := NewFormPlan(nil, 22050, "lofi")
	markers := form.ListeningMarkers(1)
	if len(markers) == 0 {
		t.Fatal("expected non-empty marker list")
	}
	for _, marker := range markers {
		if marker.Sample%22050 != 0 {
			t.Fatalf("%s at sample %d is not bar-aligned", marker.Label, marker.Sample)
		}
	}
}

func TestSectionAtAdvancesThroughPlan(t *testing.T) {
	form := NewFormPlan(nil, 100, "jazz")
	first := form.SectionAt(0)
	if first.Kind != FormIntro {
		t.Fatalf("first section = %s, want %s", first.Kind, FormIntro)
	}
	later := form.SectionAt(4 * 100)
	if later.Kind == FormIntro {
		t.Fatalf("expected section to advance after intro bars")
	}
}

func TestClassicalFormEndsWithOutro(t *testing.T) {
	form := NewFormPlan(nil, 100, "classical")
	lastBar := int64((form.TotalBars() - 1) * 100)
	if got := form.SectionAt(lastBar).Kind; got != FormOutro {
		t.Fatalf("last section = %s, want %s", got, FormOutro)
	}
}

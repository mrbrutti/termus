package gen

import "testing"

func TestJazzDrumPhraseTargetsCadenceBars(t *testing.T) {
	a := &Jazz{
		progression: []jazzChord{
			jazzMaj7(0, "I"),
			jazzMaj7(5, "IV"),
			jazzDom7(7, "V"),
			jazzMaj7(0, "I"),
		},
		section: sectionTemplate(FormCadence),
	}
	if got := a.kickNoteAt(1); got != jazzKickKey {
		t.Fatalf("kickNoteAt(1) = %d, want cadence kick", got)
	}
	if got := a.brushFillNoteAt(15); got != jazzSnareBrushed {
		t.Fatalf("brushFillNoteAt(15) = %d, want cadence fill", got)
	}
}

func TestChillDrumPhraseAddsTurnaroundFill(t *testing.T) {
	a := &Chill{
		progression: []chillChord{
			chillMaj7(0, "I"),
			chillMin7(9, "vi"),
			chillMaj7(5, "IV"),
			chillDom7(7, "V"),
		},
		section: sectionTemplate(FormA),
	}
	if got := a.kickNoteAt(7); got != drumKick {
		t.Fatalf("kickNoteAt(7) = %d, want turnaround kick", got)
	}
	if got := a.snareFillNoteAt(14); got != drumSnare {
		t.Fatalf("snareFillNoteAt(14) = %d, want turnaround fill", got)
	}
}

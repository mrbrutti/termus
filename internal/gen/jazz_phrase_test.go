package gen

import (
	"math/rand"
	"testing"
)

func TestJazzSaxMotifsSpanFourBarSentences(t *testing.T) {
	j := &Jazz{rng: rand.New(rand.NewSource(1))} //nolint:gosec
	motifs := j.makeSaxMotifs()
	if got, want := len(motifs.A), 4*jazzSaxSlotsPerBar; got != want {
		t.Fatalf("A phrase length = %d, want %d", got, want)
	}
	if got, want := len(motifs.B), 4*jazzSaxSlotsPerBar; got != want {
		t.Fatalf("B phrase length = %d, want %d", got, want)
	}
	if got, want := len(motifs.Cadence), 4*jazzSaxSlotsPerBar; got != want {
		t.Fatalf("cadence phrase length = %d, want %d", got, want)
	}
}

func TestJazzGuideTargetPrefersNextThirdOnResolution(t *testing.T) {
	j := &Jazz{rootMidi: 48}
	current := jazzDom7(7, "G7")
	next := jazzMaj7(0, "Cmaj7")
	if got, want := j.jazzGuideTarget(current, next, false), 48+next.tones[1]+12; got != want {
		t.Fatalf("guide target = %d, want %d", got, want)
	}
}

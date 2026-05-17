package gen

import (
	"math/rand"
	"testing"
)

func TestChillMotifsSpanEightBarLoops(t *testing.T) {
	c := &Chill{rng: rand.New(rand.NewSource(1))} //nolint:gosec
	vibe := c.makeVibeMotifs()
	guitar := c.makeGuitarMotifs()
	sax := c.makeSaxMotifs()
	for name, phrase := range map[string][]int{
		"vibe":   vibe.A,
		"guitar": guitar.A,
	} {
		if got, want := len(phrase), chillSupportMotifSlots; got != want {
			t.Fatalf("%s phrase length = %d, want %d", name, got, want)
		}
	}
	if got, want := len(sax.A), chillLeadMotifBars; got != want {
		t.Fatalf("sax phrase length = %d, want %d", got, want)
	}
}

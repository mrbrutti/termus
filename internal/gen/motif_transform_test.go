package gen

import (
	"math/rand"
	"testing"
)

func TestTransformNumericPhrasePreservesLength(t *testing.T) {
	src := []int{0, 2, 4, 2}
	got := transformNumericPhrase(rand.New(rand.NewSource(1)), src) //nolint:gosec
	if len(got) != len(src) {
		t.Fatalf("transform length = %d, want %d", len(got), len(src))
	}
	if phraseSignature(got) == phraseSignature(src) {
		t.Fatalf("expected transformed phrase to differ from source")
	}
}

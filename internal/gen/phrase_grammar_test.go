package gen

import "testing"

func TestPhraseShapePhrase(t *testing.T) {
	shape := PhraseShape{
		Pickup:    []int{1, 2},
		Statement: []int{3},
		Peak:      []int{4, 5},
		Release:   []int{6},
	}
	got := shape.Phrase()
	want := []int{1, 2, 3, 4, 5, 6}
	if phraseSignature(got) != phraseSignature(want) {
		t.Fatalf("phrase = %v, want %v", got, want)
	}
}

func TestBuildPhraseMotifs(t *testing.T) {
	motifs := buildPhraseMotifs(
		PhraseShape{Pickup: []int{1}, Statement: []int{2}},
		PhraseShape{Peak: []int{3}, Release: []int{4}},
		map[int]int{3: 5},
		PhraseShape{Pickup: []int{6}},
		PhraseShape{Peak: []int{7}},
		PhraseShape{Pickup: []int{8}, Release: []int{9}},
		PhraseShape{Pickup: []int{10}},
	)
	if got, want := phraseSignature(motifs.A), phraseSignature([]int{1, 2, 3, 4}); got != want {
		t.Fatalf("A signature = %s, want %s", got, want)
	}
	if got, want := phraseSignature(motifs.Aprime), phraseSignature([]int{1, 2, 5, 4}); got != want {
		t.Fatalf("Aprime signature = %s, want %s", got, want)
	}
	if got, want := phraseSignature(motifs.B), phraseSignature([]int{6, 7}); got != want {
		t.Fatalf("B signature = %s, want %s", got, want)
	}
}


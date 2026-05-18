package track

import (
	"path/filepath"
	"testing"
)

func TestLoadBundledCorpusResolvesTracks(t *testing.T) {
	corpus, err := LoadBundledCorpus()
	if err != nil {
		t.Fatalf("LoadBundledCorpus: %v", err)
	}
	entries, err := Discover(filepath.Join("..", "..", "tracks"))
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(corpus.Genres) != 6 {
		t.Fatalf("genre count = %d, want 6", len(corpus.Genres))
	}
	for style, genre := range corpus.Genres {
		if genre.Canonical == "" {
			t.Fatalf("%s canonical missing", style)
		}
		if len(genre.Corpus) < 4 {
			t.Fatalf("%s corpus too small: %d (want >=4)", style, len(genre.Corpus))
		}
		if len(genre.AB) < 2 {
			t.Fatalf("%s ab pairs too small: %d", style, len(genre.AB))
		}
		if _, ok := Resolve(entries, genre.Canonical); !ok {
			t.Fatalf("%s canonical unresolved: %s", style, genre.Canonical)
		}
		for _, id := range genre.Corpus {
			if _, ok := Resolve(entries, id); !ok {
				t.Fatalf("%s corpus unresolved: %s", style, id)
			}
		}
		for _, pair := range genre.AB {
			if _, ok := Resolve(entries, pair.A); !ok {
				t.Fatalf("%s ab unresolved A: %s", style, pair.A)
			}
			if _, ok := Resolve(entries, pair.B); !ok {
				t.Fatalf("%s ab unresolved B: %s", style, pair.B)
			}
		}
	}
}

package gen

import (
	"strings"
	"testing"
	"time"
)

func TestSameGenrePlaylistDeterministic(t *testing.T) {
	spec := AlgoSpec{Name: "ambient", Display: "Ambient", RequiresSF2: true}
	a := SameGenrePlaylist(spec, 4, 42, time.Minute)
	b := SameGenrePlaylist(spec, 4, 42, time.Minute)
	if a.Name != b.Name {
		t.Fatalf("same seed should produce same playlist name; got %q vs %q",
			a.Name, b.Name)
	}
	if len(a.Tracks) != 4 || len(b.Tracks) != 4 {
		t.Fatalf("want 4 tracks, got %d / %d", len(a.Tracks), len(b.Tracks))
	}
	for i := range a.Tracks {
		if a.Tracks[i].Seed != b.Tracks[i].Seed {
			t.Errorf("track %d: seed mismatch %d vs %d",
				i, a.Tracks[i].Seed, b.Tracks[i].Seed)
		}
		if a.Tracks[i].Spec.Name != "ambient" {
			t.Errorf("track %d: wrong spec %q", i, a.Tracks[i].Spec.Name)
		}
	}
}

func TestSameGenrePlaylistDistinctSeeds(t *testing.T) {
	spec := AlgoSpec{Name: "ambient"}
	pl := SameGenrePlaylist(spec, 5, 1, time.Minute)
	seen := map[int64]bool{}
	for _, tr := range pl.Tracks {
		if seen[tr.Seed] {
			t.Errorf("duplicate seed %d in same-genre playlist", tr.Seed)
		}
		seen[tr.Seed] = true
	}
}

func TestMixedPlaylistVaries(t *testing.T) {
	specs := []AlgoSpec{
		{Name: "ambient"}, {Name: "lofi"}, {Name: "jazz"}, {Name: "drone"},
	}
	// With 8 picks from 4 options we'd expect at least two distinct names.
	pl := MixedPlaylist(specs, 8, 99, time.Minute)
	uniq := map[string]bool{}
	for _, tr := range pl.Tracks {
		uniq[tr.Spec.Name] = true
	}
	if len(uniq) < 2 {
		t.Errorf("mixed playlist should draw varying specs; got only %v", uniq)
	}
}

func TestPlaylistNameLooksReasonable(t *testing.T) {
	pl := SameGenrePlaylist(AlgoSpec{Name: "ambient"}, 2, 7, time.Minute)
	if pl.Name == "" {
		t.Fatal("playlist name is empty")
	}
	if !strings.Contains(pl.Name, " ") {
		t.Errorf("expected multi-word name, got %q", pl.Name)
	}
}

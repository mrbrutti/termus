package gen

import (
	"fmt"
	"math/rand"
	"time"
)

// PlaylistMode describes how a playlist's tracks were generated.
type PlaylistMode int

const (
	PlaylistSameGenre PlaylistMode = iota
	PlaylistMixed
	PlaylistScore
)

// SectionStop describes one section inside a Track that has an internal
// section schedule (SP17). The Track plays seamlessly across stops; no
// crossfade is applied at stop boundaries. PlanKey is the lookup key into
// Compiled.Plans for the plan that should drive playback during this stop.
type SectionStop struct {
	Title     string
	Seed      int64
	Duration  time.Duration
	StartTime time.Duration // cumulative offset from track start
	PlanKey   string
}

// Track is one entry in a Playlist: an algorithm spec, its seed, and how long
// to play it before advancing to the next track.
//
// SP17: Tracks may carry an internal Sections schedule. When non-nil, the
// playback engine swaps the algorithm's plan at section boundaries without
// crossfading; the Track is a single seamless composition that happens to
// have multiple sections inside it. When Sections is nil/empty the Track
// behaves as the legacy single-section playlist entry (used by radio-mode
// playlists, where each entry is a different algorithm).
type Track struct {
	Spec     AlgoSpec
	Seed     int64
	Duration time.Duration
	Title    string

	// Sections, when non-empty, lists the internal section schedule for this
	// Track. The first entry's PlanKey is the plan the Track starts with;
	// subsequent stops swap the plan in-place at their StartTime offset.
	Sections []SectionStop

	// LoopForeverEvolving, when true, instructs the playback engine to loop
	// the section schedule indefinitely. Each iteration uses a different
	// seed offset so the underlying generated events vary. Combined with the
	// hour-stream listening mode this produces 1-4 hours of evolving playback
	// from the same authored composition.
	LoopForeverEvolving bool
}

// Playlist is an ordered list of Tracks with a human-readable display name.
type Playlist struct {
	Name       string
	Mode       PlaylistMode
	ListenMode ListeningMode
	Tracks     []Track
}

// SameGenrePlaylist builds a playlist of `count` tracks, all using the same
// AlgoSpec but with different seeds. Each track plays for `dur` before
// advancing. Track seeds are derived deterministically from baseSeed.
func SameGenrePlaylist(spec AlgoSpec, count int, baseSeed int64, dur time.Duration) Playlist {
	rng := rand.New(rand.NewSource(baseSeed))
	tracks := make([]Track, count)
	for i := range tracks {
		tracks[i] = Track{
			Spec:     spec,
			Seed:     baseSeed + int64(i)*1009,
			Duration: dur,
		}
	}
	return Playlist{
		Name:   generatePlaylistName(rng, spec.Label()),
		Mode:   PlaylistSameGenre,
		Tracks: tracks,
	}
}

// MixedPlaylist builds a playlist drawing randomly from genreSpecs. Each
// track gets its own seed and the same duration.
func MixedPlaylist(genreSpecs []AlgoSpec, count int, baseSeed int64, dur time.Duration) Playlist {
	rng := rand.New(rand.NewSource(baseSeed))
	tracks := make([]Track, count)
	for i := range tracks {
		spec := genreSpecs[rng.Intn(len(genreSpecs))]
		tracks[i] = Track{
			Spec:     spec,
			Seed:     baseSeed + int64(i)*1009,
			Duration: dur,
		}
	}
	return Playlist{
		Name:   generatePlaylistName(rng, ""),
		Mode:   PlaylistMixed,
		Tracks: tracks,
	}
}

// generatePlaylistName returns a stylized two-word name with an optional
// volume number — e.g. "Midnight Sessions Vol. 7". flavor, when non-empty,
// biases word choice (e.g. for "Lo-fi" we prefer hip-hop-y nouns).
func generatePlaylistName(rng *rand.Rand, flavor string) string {
	_ = flavor // reserved for future genre-aware tuning
	adj := adjectives[rng.Intn(len(adjectives))]
	noun := nouns[rng.Intn(len(nouns))]
	if rng.Intn(2) == 0 {
		return fmt.Sprintf("%s %s Vol. %d", adj, noun, 1+rng.Intn(99))
	}
	return fmt.Sprintf("%s %s", adj, noun)
}

var adjectives = []string{
	"Midnight", "Sunset", "Velvet", "Crystal", "Echo", "Drifting",
	"Endless", "Quiet", "Static", "Neon", "Distant", "Slow",
	"Wandering", "Glass", "Paper", "Silver", "Lunar", "Hidden",
	"Late", "Soft", "Faded", "Hollow", "Marble", "Indigo",
}

var nouns = []string{
	"Sessions", "Hours", "Skies", "Dreams", "Waves", "Rooms",
	"Gardens", "Fields", "Tides", "Reverie", "Cassettes", "Letters",
	"Mirrors", "Lights", "Stations", "Stories", "Pages", "Atlas",
	"Currents", "Drift", "Postcards", "Lanterns", "Patterns", "Routine",
}

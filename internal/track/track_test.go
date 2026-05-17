package track

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrbrutti/termus/internal/gen"
)

func TestCompileBuildsTrackPlaylist(t *testing.T) {
	const src = `
title: Soft Tape / Rain Bus
style: lofi
listen_mode: album-side
seed: 42
roles:
  keys:
    family: electric_piano
    pattern: "x..x .x.."
  lead:
    family: reed_lead
    motif: "5 . 6 5 | 3 . 2 1"
sections:
  - id: intro
    title: curbside intro
    duration: 90s
    harmony: "Dm9 G13 | Cmaj9 A7"
    scene: "intro sparse"
    profile:
      density: sparse
      motion: gentle
  - id: return
    title: late platform
    duration: 120s
    harmony: "Fm9 Bb13 | Ebmaj9 C7"
    scene: "return lift"
    profile:
      density: busy
      swing: groove
    roles:
      lead:
        active: true
        motif: "9 . 7 5 | 3 . 2 1"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 99, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if got, want := compiled.Playlist.Mode, gen.PlaylistScore; got != want {
		t.Fatalf("playlist mode = %v, want %v", got, want)
	}
	if got, want := compiled.Playlist.ListenMode, gen.ListeningModeAlbumSide; got != want {
		t.Fatalf("listen mode = %q, want %q", got, want)
	}
	if got, want := len(compiled.Playlist.Tracks), 2; got != want {
		t.Fatalf("track count = %d, want %d", got, want)
	}
	if compiled.Playlist.Tracks[0].Title != "curbside intro" {
		t.Fatalf("track title = %q", compiled.Playlist.Tracks[0].Title)
	}
	if len(compiled.Blueprints) != 2 {
		t.Fatalf("blueprint count = %d, want 2", len(compiled.Blueprints))
	}
}

func TestCompileRejectsBadPattern(t *testing.T) {
	const src = `
title: Broken
style: lofi
roles:
  lead:
    family: reed_lead
    motif: "5 % 3"
sections:
  - title: bad
    duration: 30s
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if _, err := Compile(file, 1, gen.ListeningModeEndless); err == nil {
		t.Fatal("expected compile error for bad melody token")
	}
}

func TestBundledTracksParseAndCompile(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "tracks", "*", "*.tm"))
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	if len(paths) < 10 {
		t.Fatalf("expected at least 10 bundled tracks, got %d", len(paths))
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile %s: %v", path, err)
		}
		file, err := Parse(data)
		if err != nil {
			t.Fatalf("Parse %s: %v", path, err)
		}
		if _, err := Compile(file, 7, gen.ListeningModeEndless); err != nil {
			t.Fatalf("Compile %s: %v", path, err)
		}
	}
}

func TestResolveAcceptsDirectPath(t *testing.T) {
	entries, err := Discover(filepath.Join("..", "..", "tracks"))
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected bundled track entries")
	}
	entry, ok := Resolve(entries, entries[0].Path)
	if !ok {
		t.Fatal("Resolve should accept direct path")
	}
	if entry.Path != entries[0].Path {
		t.Fatalf("resolved path = %q, want %q", entry.Path, entries[0].Path)
	}
}


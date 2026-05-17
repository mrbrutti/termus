package tm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrbrutti/termus/internal/gen"
)

func TestCompileBuildsScoredPlaylist(t *testing.T) {
	const src = `
title: Soft Tape / Rain Bus
listen_mode: album-side
seed: 42
globals:
  density: steady
  brightness: warm
sections:
  - title: curbside intro
    algo: lofi
    duration: 90s
    profile:
      density: sparse
      motion: gentle
    audit:
      harmony: "Dm9 G13 | Cmaj9 Am7"
      lead: "5 . 6 5 | 3 . 2 1"
      comp: "x . . x | . x . ."
      drums: "bd: x... x..x | sd: ..x. ..x."
      arrange: "bass drums comp"
  - title: late platform
    algo: jazz
    duration: 120s
    profile:
      density: busy
      swing: groove
    audit:
      harmony: "ii7 V7 | Imaj7 VI7"
      lead: "9 . 7 5 | 3 . 2 1"
      comp: "x . x . | . x . x"
      drums: "ride: x.x. x.x."
      arrange: "bass drums comp +lead"
`
	score, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(score, 99, gen.ListeningModeEndless)
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
	if len(compiled.Overrides) != 2 {
		t.Fatalf("override count = %d, want 2", len(compiled.Overrides))
	}
	if len(compiled.Blueprints) != 2 {
		t.Fatalf("blueprint count = %d, want 2", len(compiled.Blueprints))
	}
}

func TestCompileRejectsBadPattern(t *testing.T) {
	const src = `
title: Broken
sections:
  - algo: lofi
    title: bad
    duration: 30s
    audit:
      lead: "5 % 3"
`
	score, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if _, err := Compile(score, 1, gen.ListeningModeEndless); err == nil {
		t.Fatal("expected compile error for bad melody token")
	}
}

func TestSampleScoresParseAndCompile(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "scores", "*.tm"))
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("expected sample .tm files")
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile %s: %v", path, err)
		}
		score, err := Parse(data)
		if err != nil {
			t.Fatalf("Parse %s: %v", path, err)
		}
		if _, err := Compile(score, 7, gen.ListeningModeEndless); err != nil {
			t.Fatalf("Compile %s: %v", path, err)
		}
	}
}

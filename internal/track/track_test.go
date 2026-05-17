package track

import (
	"os"
	"path/filepath"
	"strings"
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
	if len(compiled.Plans) != 2 {
		t.Fatalf("plan count = %d, want 2", len(compiled.Plans))
	}
	for _, plan := range compiled.Plans {
		if len(plan.PhraseSpans) == 0 {
			t.Fatal("expected phrase spans in authored plan")
		}
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

func TestCompileAppliesSectionEvents(t *testing.T) {
	const src = `
title: Eventful
style: jazz
seed: 17
roles:
  piano:
    family: acoustic_piano
    pattern: "x..x.x.. | .x..x..x"
  kick:
    family: drums
    pattern: "x...x... | x...x..."
  snare:
    family: drums
    pattern: "....x... | ....x..."
  lead:
    family: reed_lead
    motif: "5 . 6 7 | 3 . 2 1"
sections:
  - id: head
    duration: 16s
    harmony: "Dm7 G7 | Cmaj7 A7 | Dm7 G7 | Cmaj7 Cmaj7"
    roles:
      lead:
        active: true
    events:
      - kind: fill
        bar: 2
        roles: [snare]
      - kind: drop
        bar: 3
        roles: [kick]
      - kind: pickup
        bar: 4
        roles: [lead]
        motif: "3 5 6 9"
      - kind: stab
        bar: 1
        roles: [piano]
        pattern: "x... ...."
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 17, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if len(compiled.Plans) != 1 {
		t.Fatalf("plan count = %d, want 1", len(compiled.Plans))
	}
	var plan gen.AuthoredTrackPlan
	for _, got := range compiled.Plans {
		plan = got
	}
	findTrack := func(name string) *gen.AuthoredRenderTrack {
		for i := range plan.Tracks {
			if plan.Tracks[i].Name == name {
				return &plan.Tracks[i]
			}
		}
		return nil
	}
	findPrefix := func(prefix string) *gen.AuthoredRenderTrack {
		for i := range plan.Tracks {
			if strings.HasPrefix(plan.Tracks[i].Name, prefix) {
				return &plan.Tracks[i]
			}
		}
		return nil
	}
	snare := findTrack("snare")
	if snare == nil {
		t.Fatal("expected snare track")
	}
	fillHasHit := false
	for i := 8; i < 16; i++ {
		if snare.Notes[i] >= 0 {
			fillHasHit = true
			break
		}
	}
	if !fillHasHit {
		t.Fatal("expected fill event to add a snare hit in bar 2")
	}
	kick := findTrack("kick")
	if kick == nil {
		t.Fatal("expected kick track")
	}
	for i := 16; i < 24; i++ {
		if kick.Notes[i] != -1 {
			t.Fatalf("expected dropped kick at slot %d, got %d", i, kick.Notes[i])
		}
	}
	lead := findTrack("lead")
	if lead == nil {
		t.Fatal("expected lead track")
	}
	pickupHasNote := false
	for i := 28; i < 32; i++ {
		if lead.Notes[i] >= 0 {
			pickupHasNote = true
			break
		}
	}
	if !pickupHasNote {
		t.Fatal("expected pickup event to add lead notes near the section close")
	}
	piano := findPrefix("piano-")
	if piano == nil {
		t.Fatal("expected piano voice track")
	}
	if got := piano.Notes[1]; got != -1 {
		t.Fatalf("expected stabbed piano slot 1 to be muted, got %d", got)
	}
}

func TestCompileBuildsPhraseBlocks(t *testing.T) {
	const src = `
title: Phrase Blocks
style: lofi
seed: 21
roles:
  lead:
    family: reed_lead
    motif: "5 . 6 7 | 3 . 2 1"
  keys:
    family: electric_piano
    pattern: "x..x .x.."
sections:
  - id: long
    duration: 32s
    harmony: "Dm9 G13 | Cmaj9 A7 | Bbmaj9 A7 | Dm9 G13"
    scene: "head glide"
    variation: "introduce-hook"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 21, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	var plan gen.AuthoredTrackPlan
	for _, got := range compiled.Plans {
		plan = got
	}
	if len(plan.PhraseSpans) < 2 {
		t.Fatalf("expected multiple phrase spans, got %d", len(plan.PhraseSpans))
	}
	if got, want := plan.PhraseSpans[0].Label, "statement"; got != want {
		t.Fatalf("first phrase label = %q, want %q", got, want)
	}
	if got, want := plan.PhraseSpans[len(plan.PhraseSpans)-1].Label, "release"; got != want {
		t.Fatalf("last phrase label = %q, want %q", got, want)
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
		eventCount := 0
		for _, section := range file.Sections {
			eventCount += len(section.Events)
		}
		if eventCount == 0 {
			t.Fatalf("expected curated arrangement events in %s", path)
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

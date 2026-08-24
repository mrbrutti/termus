package track

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const entryTestTM = `title: Rainy Test
style: lofi
substyle: rainy-cafe
tempo: 84 bpm
key: D minor
total_duration: 5m
textures:
  - name: rain
    level_db: -36
  - name: vinyl
    level_db: -44
sections:
  - id: intro
    duration: 30s
    harmony: "Dm9 | Gm7"
  - id: body
    duration: 90s
    harmony: "Dm9 | Bbmaj7"
`

func TestDiscoverEntryDurationsAndTextures(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "lofi")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "rainy-test.tm"), []byte(entryTestTM), 0o644); err != nil {
		t.Fatal(err)
	}
	entries, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	entry := entries[0]
	if len(entry.Structure) != 2 {
		t.Fatalf("structure = %d sections, want 2", len(entry.Structure))
	}
	if entry.Structure[0].Duration != 30*time.Second {
		t.Fatalf("intro duration = %v, want 30s", entry.Structure[0].Duration)
	}
	if entry.Structure[1].Duration != 90*time.Second {
		t.Fatalf("body duration = %v, want 90s", entry.Structure[1].Duration)
	}
	if len(entry.Textures) != 2 || entry.Textures[0] != "rain -36 dB" || entry.Textures[1] != "vinyl -44 dB" {
		t.Fatalf("textures = %v, want [rain -36 dB, vinyl -44 dB]", entry.Textures)
	}
	if entry.TotalDuration != 5*time.Minute {
		t.Fatalf("total duration = %v, want 5m", entry.TotalDuration)
	}
}

// TestDiscoverEntryNoTotalDuration covers a file that omits total_duration:
// entirely; Entry.TotalDuration should come out zero rather than defaulting
// to the summed section durations (that fallback lives in the TUI, not here).
func TestDiscoverEntryNoTotalDuration(t *testing.T) {
	entry := discoverSingleEntry(t, noTexturesTM)
	if entry.TotalDuration != 0 {
		t.Fatalf("total duration = %v, want 0", entry.TotalDuration)
	}
}

// TestResolveAdHocEntryMatchesLoadEntry covers Resolve's ad-hoc .tm path
// (input is a stat-able .tm file not already in the entries slice), which
// builds its own Entry literal separate from loadEntry's. Guards against the
// two literals drifting out of sync, as happened with TotalDuration.
func TestResolveAdHocEntryMatchesLoadEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rainy-test.tm")
	if err := os.WriteFile(path, []byte(entryTestTM), 0o644); err != nil {
		t.Fatal(err)
	}
	entry, ok := Resolve(nil, path)
	if !ok {
		t.Fatalf("Resolve(%q) = false, want true", path)
	}
	if len(entry.Structure) != 2 {
		t.Fatalf("structure = %d sections, want 2", len(entry.Structure))
	}
	if entry.Structure[0].Duration != 30*time.Second {
		t.Fatalf("intro duration = %v, want 30s", entry.Structure[0].Duration)
	}
	if entry.Structure[1].Duration != 90*time.Second {
		t.Fatalf("body duration = %v, want 90s", entry.Structure[1].Duration)
	}
	if len(entry.Textures) != 2 || entry.Textures[0] != "rain -36 dB" || entry.Textures[1] != "vinyl -44 dB" {
		t.Fatalf("textures = %v, want [rain -36 dB, vinyl -44 dB]", entry.Textures)
	}
	if entry.TotalDuration != 5*time.Minute {
		t.Fatalf("total duration = %v, want 5m", entry.TotalDuration)
	}
}

// discoverSingleEntry writes content as a .tm file under a fresh style
// subdirectory and returns the sole discovered Entry.
func discoverSingleEntry(t *testing.T, content string) Entry {
	t.Helper()
	dir := t.TempDir()
	sub := filepath.Join(dir, "lofi")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "test.tm"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	entries, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	return entries[0]
}

const barsOnlySectionTM = `title: Bars Only Test
style: lofi
tempo: 84 bpm
key: D minor
sections:
  - id: intro
    bars: 8
    harmony: "Dm9 | Gm7"
`

// TestDiscoverEntryBarsOnlySectionGetsDuration covers the case in
// buildEntrySummary where a section only sets bars: (no duration:).
// resolveSections converts bars -> a duration string using the file's
// tempo before the structure loop runs, so EntrySection.Duration should
// come out nonzero.
func TestDiscoverEntryBarsOnlySectionGetsDuration(t *testing.T) {
	entry := discoverSingleEntry(t, barsOnlySectionTM)
	if len(entry.Structure) != 1 {
		t.Fatalf("structure = %d sections, want 1", len(entry.Structure))
	}
	// 8 bars @ 84bpm, 4 beats/bar: 8*4*60/84 = 22.857s, rounded -> 23s
	// (matches barsToDurationString's rounding in form_library.go).
	want := 23 * time.Second
	if entry.Structure[0].Duration != want {
		t.Fatalf("intro duration = %v, want %v", entry.Structure[0].Duration, want)
	}
}

const textureDefaultLevelTM = `title: Texture Default Test
style: lofi
tempo: 84 bpm
key: D minor
textures:
  - name: rain
sections:
  - id: intro
    duration: 30s
    harmony: "Dm9 | Gm7"
`

// TestDiscoverEntryTextureDefaultLevel covers a texture entry that omits
// level_db: textureLabels must fall back to the same per-name default that
// compileTextures uses at playback time (authored_compile.go), so the
// track-library display matches what actually plays.
func TestDiscoverEntryTextureDefaultLevel(t *testing.T) {
	entry := discoverSingleEntry(t, textureDefaultLevelTM)
	if len(entry.Textures) != 1 || entry.Textures[0] != "rain -38 dB" {
		t.Fatalf("textures = %v, want [rain -38 dB]", entry.Textures)
	}
}

const noTexturesTM = `title: No Textures Test
style: lofi
tempo: 84 bpm
key: D minor
sections:
  - id: intro
    duration: 30s
    harmony: "Dm9 | Gm7"
`

// TestDiscoverEntryNoTextures covers a file with no textures: block at all;
// Entry.Textures should stay nil rather than an empty non-nil slice.
func TestDiscoverEntryNoTextures(t *testing.T) {
	entry := discoverSingleEntry(t, noTexturesTM)
	if entry.Textures != nil {
		t.Fatalf("textures = %v, want nil", entry.Textures)
	}
}

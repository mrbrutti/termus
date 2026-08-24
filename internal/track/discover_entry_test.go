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
}

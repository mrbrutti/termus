package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestResolveEngineForTrack_AutoReadsTM verifies that with --engine auto, the
// .tm's render_engine field decides routing.
func TestResolveEngineForTrack_AutoReadsTM(t *testing.T) {
	dir := t.TempDir()
	acePath := filepath.Join(dir, "ai.tm")
	if err := os.WriteFile(acePath, []byte(minimalAcestepTM), 0o644); err != nil {
		t.Fatalf("write tm: %v", err)
	}
	sfPath := filepath.Join(dir, "sf.tm")
	if err := os.WriteFile(sfPath, []byte(minimalSF2TM), 0o644); err != nil {
		t.Fatalf("write tm: %v", err)
	}

	cases := []struct {
		path string
		flag string
		want string
	}{
		{acePath, "auto", "acestep"},
		{acePath, "sf2", "sf2"},     // flag overrides .tm
		{acePath, "acestep", "acestep"},
		{sfPath, "auto", "sf2"},
		{sfPath, "acestep", "acestep"}, // flag forces acestep even on sf2 .tm
	}
	for _, c := range cases {
		got, err := resolveEngineForTrack(c.path, c.flag)
		if err != nil {
			t.Errorf("resolveEngineForTrack(%s, %q) error: %v", c.path, c.flag, err)
			continue
		}
		if got != c.want {
			t.Errorf("resolveEngineForTrack(%s, %q) = %q, want %q", c.path, c.flag, got, c.want)
		}
	}
}

// TestResolveEngineForTrack_RejectsUnknownEngine checks the negative path.
func TestResolveEngineForTrack_RejectsUnknownEngine(t *testing.T) {
	if _, err := resolveEngineForTrack("", "weird"); err == nil {
		t.Fatal("expected error for unknown engine")
	}
}

func TestTrackPathFromSelection_AcceptsDirectPath(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "t.tm")
	if err := os.WriteFile(p, []byte(minimalSF2TM), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := trackPathFromSelection(p)
	if err != nil {
		t.Fatalf("trackPathFromSelection: %v", err)
	}
	if got != p {
		t.Errorf("got %q, want %q", got, p)
	}
}

func TestTrackPathFromSelection_EmptyErrors(t *testing.T) {
	if _, err := trackPathFromSelection("   "); err == nil {
		t.Fatal("expected error on empty input")
	}
}

func TestHumanBytes_Formats(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{50 * 1024 * 1024, "50 MB"},
		{2 * 1024 * 1024 * 1024, "2.0 GB"},
		{9_400 * 1024 * 1024, "9.2 GB"},
	}
	for _, c := range cases {
		got := humanBytes(c.in)
		if got != c.want {
			t.Errorf("humanBytes(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

// Minimal valid .tm files for the engine-resolution tests. These are NOT
// renderable; they just need to parse and expose RenderEngine.
const minimalSF2TM = `style: lofi
title: smoke
tempo: 80
key: C major
sections:
  - id: a
    bars: 4
`

const minimalAcestepTM = `render_engine: acestep
style: lofi
title: ai-smoke
tempo: 80
key: C major
acestep:
  duration_seconds: 30
  prompt: "test prompt"
  sections:
    - id: a
      bars: 4
`

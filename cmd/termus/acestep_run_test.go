package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mrbrutti/termus/internal/acestep"
	"github.com/mrbrutti/termus/internal/tui"
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

// TestStderrSinkPrintsInstallPhaseChange verifies the headless sink writes one
// stderr line per phase transition and suppresses duplicate phases.
func TestStderrSinkPrintsInstallPhaseChange(t *testing.T) {
	var buf bytes.Buffer
	sink := newStderrSink(&buf)
	sink.Send(tui.ACEStepInstallProgressMsg{Phase: "install:python", Detail: "installing python"})
	sink.Send(tui.ACEStepInstallProgressMsg{Phase: "install:python", Detail: "still installing"})
	sink.Send(tui.ACEStepInstallProgressMsg{Phase: "install:model", Detail: "downloading model"})
	out := buf.String()
	if !strings.Contains(out, "install:python: installing python") {
		t.Fatalf("missing first phase line:\n%s", out)
	}
	if strings.Contains(out, "still installing") {
		t.Fatalf("duplicate phase should not print:\n%s", out)
	}
	if !strings.Contains(out, "install:model: downloading model") {
		t.Fatalf("missing model phase line:\n%s", out)
	}
}

// TestStderrSinkSurfacesReadyMessage covers the final ready event.
func TestStderrSinkSurfacesReadyMessage(t *testing.T) {
	var buf bytes.Buffer
	sink := newStderrSink(&buf)
	sink.Send(tui.ACEStepReadyMsg{Detail: "engine ready"})
	if !strings.Contains(buf.String(), "engine ready") {
		t.Fatalf("expected ready detail on stderr, got %q", buf.String())
	}
}

// TestInstallEventPercentMonotone ensures the install percentages we expose to
// the loader never regress as the install proceeds. The exact numbers are
// approximations; the property under test is monotonicity.
func TestInstallEventPercentMonotone(t *testing.T) {
	phases := []string{"python", "deps", "model", "done"}
	last := -1.0
	for _, p := range phases {
		got := installEventPercent(acestep.InstallEvent{Phase: p})
		if got < last {
			t.Fatalf("percent regressed at phase %q: %f < %f", p, got, last)
		}
		last = got
	}
}

// TestStatusEventPercentMonotone covers the daemon-lifecycle phases.
func TestStatusEventPercentMonotone(t *testing.T) {
	phases := []string{"checking-install", "installing", "starting-daemon", "loading-model", "ready"}
	last := -1.0
	for _, p := range phases {
		got := statusEventPercent(acestep.StatusEvent{Phase: p})
		if got < last {
			t.Fatalf("percent regressed at phase %q: %f < %f", p, got, last)
		}
		last = got
	}
}

// TestSinkInterfaceTeaProgramCompat is a compile-time check that *tea.Program
// satisfies our messageSink interface.
func TestSinkInterfaceTeaProgramCompat(t *testing.T) {
	var _ messageSink = (*tea.Program)(nil)
}

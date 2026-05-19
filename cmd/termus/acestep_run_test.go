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

// TestLoadingProgressStartsAtZero is the SP24 regression check: the bar must
// start at 0% rather than jumping to whatever the first phase's static target
// percent used to be (~0.65 for "starting-daemon" on a warm boot).
func TestLoadingProgressStartsAtZero(t *testing.T) {
	p := newLoadingProgress()
	if p.cur != 0 {
		t.Fatalf("initial percent = %f, want 0", p.cur)
	}
}

// TestLoadingProgressMonotone confirms that as new phases arrive the bar only
// ever moves forward (or stays put when a phase repeats), and stays at or
// below loadingCeiling.
func TestLoadingProgressMonotone(t *testing.T) {
	p := newLoadingProgress()
	phases := []string{
		"install:python",
		"install:deps",
		"install:model",
		"install:done",
		"status:starting-daemon",
		"status:loading-model",
		"status:ready",
	}
	last := -1.0
	for _, ph := range phases {
		got := p.observe(ph)
		if got < last {
			t.Fatalf("percent regressed at phase %q: %f < %f", ph, got, last)
		}
		if got > loadingCeiling+1e-9 {
			t.Fatalf("percent at phase %q = %f exceeded ceiling %f", ph, got, loadingCeiling)
		}
		last = got
	}
}

// TestLoadingProgressRepeatedPhaseDoesNotAdvance covers the case where the
// installer or manager emits multiple events for the same phase: the bar
// should advance once per unique phase, not per event.
func TestLoadingProgressRepeatedPhaseDoesNotAdvance(t *testing.T) {
	p := newLoadingProgress()
	first := p.observe("install:deps")
	again := p.observe("install:deps")
	if first != again {
		t.Fatalf("repeated phase advanced bar: %f -> %f", first, again)
	}
}

// TestLoadingProgressWarmBootStartsLow is the explicit SP24 scenario: when no
// install is needed, the first phase we see is something like
// "status:starting-daemon". The bar should be a single step from 0, not the
// old 0.65 target.
func TestLoadingProgressWarmBootStartsLow(t *testing.T) {
	p := newLoadingProgress()
	got := p.observe("status:starting-daemon")
	if got > 0.5 {
		t.Fatalf("warm-boot first phase percent = %f; expected something well below the old 0.65 jump", got)
	}
}

// Reference acestep so the import is still used by other tests in the file.
var _ = acestep.InstallEvent{}

// TestSinkInterfaceTeaProgramCompat is a compile-time check that *tea.Program
// satisfies our messageSink interface.
func TestSinkInterfaceTeaProgramCompat(t *testing.T) {
	var _ messageSink = (*tea.Program)(nil)
}

package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
)

func controlsTestModel() Model {
	spec, _ := gen.Resolve("ambient")
	profile := gen.DefaultControlProfile()
	m := Model{width: 118, height: 32, keyName: "Cmin", seed: 71001, algo: spec.Label()}
	m.genres = []gen.AlgoSpec{spec}
	m.musicProfile = &profile
	m.controlsVisible = true
	m.controlTab = controlTabMusic
	return m
}

func TestControlsPanelShowsPipScalesAndLadders(t *testing.T) {
	m := controlsTestModel()
	out := controlsPanel(m, 118, 32, DefaultTheme())
	if !strings.Contains(out, "CONTROL CENTER") || !strings.Contains(out, "MUSIC") {
		t.Fatalf("missing headers: %q", out)
	}
	if !strings.Contains(out, "●") || !strings.Contains(out, "○") {
		t.Fatalf("missing pip scale: %q", out)
	}
	if !strings.Contains(out, "air · lean · steady · lush · full") {
		t.Fatalf("missing density word ladder: %q", out)
	}
	if !strings.Contains(out, "nine macros") {
		t.Fatalf("missing section annotation: %q", out)
	}
	if !strings.Contains(out, "next phrase boundary") {
		t.Fatalf("missing explainer line: %q", out)
	}
	if !strings.Contains(out, "3 of 8") {
		t.Fatalf("missing section index (music is tab 3): %q", out)
	}
	if !strings.Contains(out, "▌ music") {
		t.Fatalf("missing active-section marker in rail: %q", out)
	}
}

// TestActivatingSavedLibraryClosesControlCenter guards against the SP30/SP31
// regression where opening the saved library from the control center left
// BOTH controlsVisible and libraryVisible true: View() renders the control
// center full-screen whenever controlsVisible wins (it is checked before
// libraryVisible there), but Update() checks libraryVisible first, so every
// keystroke — including "x" (delete saved seed) — silently drove the
// invisible library list underneath a frozen-looking control center.
func TestActivatingSavedLibraryClosesControlCenter(t *testing.T) {
	spec, _ := gen.Resolve("ambient")
	m := Model{width: 118, height: 32, keyName: "Cmin", seed: 71001, algo: spec.Label()}
	m.genres = []gen.AlgoSpec{spec}
	m.ring = scope.NewRing(1024)
	m.cmd = &tuiCommanderStub{}
	m.controlsVisible = true
	m.controlTab = controlTabLibrary
	m.controlRow = 0 // "saved library" row

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned unexpected type %T", updated)
	}
	if next.controlsVisible {
		t.Fatalf("controlsVisible still true after opening saved library")
	}
	if !next.libraryVisible {
		t.Fatalf("libraryVisible still false after opening saved library")
	}

	out := next.View()
	if !strings.Contains(out, "SAVED SEEDS") {
		t.Fatalf("expected library panel header in view: %q", out)
	}
	if strings.Contains(out, "CONTROL CENTER") {
		t.Fatalf("control center still rendered after opening saved library: %q", out)
	}
}

// TestControlsPanelNeverExceedsSize sweeps every width the app can run at,
// for every section, at three terminal heights. The pane is full screen now,
// so an over-wide value (a long session label, a max-width seed) or a ladder
// that will not fit has to be trimmed rather than wrapped — lipgloss's
// Width().Render() wraps silently, which the per-line assertion catches. The
// short heights matter too: h=18 leaves the debug tab's structure preview no
// room, and without vertical clipping it pushes the footer off screen.
func TestControlsPanelNeverExceedsSize(t *testing.T) {
	theme := DefaultTheme()
	sections := []controlTab{
		controlTabNow, controlTabLook, controlTabMusic, controlTabSeeds,
		controlTabLibrary, controlTabExport, controlTabAudio, controlTabDebug,
	}
	heights := []int{18, 24, 32}
	for _, section := range sections {
		for _, h := range heights {
			for w := 40; w <= 200; w++ {
				m := controlsTestModel()
				m.controlTab = section
				m.seed = 9223372036854775807
				m.seedA = &seedBookmark{Seed: 9223372036854775807}
				m.seedB = &seedBookmark{Seed: 9223372036854775806}
				m.listeningMode = "hour stream"
				m.savedSessions = []savedSessionRecord{{
					Label:  "an unusually long remembered session label for width testing",
					Algo:   "an unusually long remembered session label for width testing",
					Seed:   9223372036854775807,
					Visual: "scope",
					Theme:  "default",
				}}
				m.debug = gen.DebugStatus{Section: "development", Chord: "Dm9", Preset: "warm rhodes"}
				m.activeTrackID = "lofi/demo"
				m.tracks = []TrackNavEntry{{
					ID:           "lofi/demo",
					Style:        "lofi",
					Substyle:     "dusty-rhodes",
					Title:        "Demo Track With A Rather Long Authored Title",
					SectionCount: 3,
					EventCount:   4,
					Complexity:   "arranged",
					Ensemble:     []string{"ep", "bass", "drums", "reed"},
					Structure: []TrackNavSection{
						{ID: "development", Label: "Development", Harmony: "Dm9 G13", Events: []string{"pickup"}, RoleNames: []string{"ep", "bass"}},
						{ID: "head", Label: "Head", Harmony: "Bbmaj9 C13", Events: []string{"fill"}, RoleNames: []string{"ep", "reed", "drums"}},
					},
				}}
				m.status = "export failed: no space left on device"
				m.statusTTL = time.Now().Add(time.Minute)
				out := controlsPanel(m, w, h, theme)
				lines := strings.Split(out, "\n")
				// A wrapped row also costs a line, which would push the
				// full-screen pane past the terminal height.
				if len(lines) > h {
					t.Fatalf("controls panel is %d rows tall in section %q at w=%d h=%d, want <= %d\n%s",
						len(lines), section.label(), w, h, h, out)
				}
				for i, line := range lines {
					if got := lipgloss.Width(line); got > w {
						t.Fatalf("controls panel line %d overflows in section %q: w=%d h=%d rendered %d cols\n%q",
							i, section.label(), w, h, got, line)
					}
				}
			}
		}
	}
}

// The full-screen pane replaces bottomBar, which is the only other place
// flashStatus feedback is drawn — without the footer status line an export
// could fail silently while the control center is open.
func TestControlsPanelShowsTransientStatus(t *testing.T) {
	m := controlsTestModel()
	m.controlTab = controlTabExport
	m.status = "export failed: no space left on device"
	m.statusTTL = time.Now().Add(time.Minute)
	out := controlsPanel(m, 118, 32, DefaultTheme())
	if !strings.Contains(out, "export failed: no space left on device") {
		t.Fatalf("footer should carry the transient status:\n%s", out)
	}
	if !strings.Contains(out, "[tab] section") || !strings.Contains(out, "6 of 8") {
		t.Fatalf("status line should not displace the footer hints:\n%s", out)
	}
	m.statusTTL = time.Now().Add(-time.Minute)
	if quiet := controlsPanel(m, 118, 32, DefaultTheme()); strings.Contains(quiet, "export failed") {
		t.Fatalf("expired status should not linger:\n%s", quiet)
	}
}

func TestControlsPanelPipCountMatchesValue(t *testing.T) {
	theme := DefaultTheme()
	pips := renderPips(theme, 2, 5) // value "steady" = index 2 -> 3 filled
	if got := strings.Count(pips, "●"); got != 3 {
		t.Fatalf("filled pips = %d, want 3", got)
	}
	if got := strings.Count(pips, "○"); got != 2 {
		t.Fatalf("empty pips = %d, want 2", got)
	}
}

// TestCurrentTrackStructureSectionMatching pins the matcher that drives both
// the control center's structure preview and the track library's FORM
// highlight. The substring fallback is useful for authored ids ("head-a" vs a
// live "head") but must not fire on the generator's single-letter procedural
// kinds, or on sections that carry no id at all.
func TestCurrentTrackStructureSectionMatching(t *testing.T) {
	authored := TrackNavEntry{Structure: []TrackNavSection{
		{ID: "intro", Label: "intro"},
		{ID: "head", Label: "head"},
		{ID: "solo", Label: "solo"},
		{ID: "outro", Label: "outro"},
	}}
	if got, ok := currentTrackStructureSection(authored, "A"); ok {
		t.Fatalf("procedural kind %q must not match authored sections, got %+v", "A", got)
	}
	if got, ok := currentTrackStructureSection(authored, "B"); ok {
		t.Fatalf("procedural kind %q must not match authored sections, got %+v", "B", got)
	}
	empty := TrackNavEntry{Structure: []TrackNavSection{{}, {ID: "head", Label: "head"}}}
	if got, ok := currentTrackStructureSection(empty, "development"); ok && got.ID == "" {
		t.Fatalf("empty section id must not match %q, got %+v", "development", got)
	}
	if got, ok := currentTrackStructureSection(authored, "solo"); !ok || got.ID != "solo" {
		t.Fatalf("exact id match broke: got %+v ok=%v", got, ok)
	}
	// Normalization still bridges case and punctuation on an exact match, and
	// still bridges authored suffixes on a long-enough key.
	quoted := TrackNavEntry{Structure: []TrackNavSection{{Label: "a'"}, {ID: "head-a", Label: "Head A"}}}
	if got, ok := currentTrackStructureSection(quoted, "A'"); !ok || got.Label != "a'" {
		t.Fatalf("normalized exact match broke: got %+v ok=%v", got, ok)
	}
	if got, ok := currentTrackStructureSection(quoted, "head"); !ok || got.ID != "head-a" {
		t.Fatalf("substring fallback broke: got %+v ok=%v", got, ok)
	}
}

package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/mrbrutti/termus/internal/gen"
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

// TestControlsPanelNeverExceedsWidth sweeps every width the app can run at,
// for every section. The pane is full screen now, so an over-wide value (a
// long session label, a max-width seed) or a ladder that will not fit has to
// be trimmed rather than wrapped — lipgloss's Width().Render() wraps silently,
// which the per-line assertion catches.
func TestControlsPanelNeverExceedsWidth(t *testing.T) {
	theme := DefaultTheme()
	sections := []controlTab{
		controlTabNow, controlTabLook, controlTabMusic, controlTabSeeds,
		controlTabLibrary, controlTabExport, controlTabAudio, controlTabDebug,
	}
	for _, section := range sections {
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
			out := controlsPanel(m, w, 32, theme)
			lines := strings.Split(out, "\n")
			// A wrapped row also costs a line, which would push the
			// full-screen pane past the terminal height.
			if len(lines) > 32 {
				t.Fatalf("controls panel is %d rows tall in section %q at w=%d, want <= 32\n%s",
					len(lines), section.label(), w, out)
			}
			for i, line := range lines {
				if got := lipgloss.Width(line); got > w {
					t.Fatalf("controls panel line %d overflows in section %q: w=%d rendered %d cols\n%q",
						i, section.label(), w, got, line)
				}
			}
		}
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

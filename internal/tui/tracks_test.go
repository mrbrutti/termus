package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mrbrutti/termus/internal/gen"
)

func TestTrackPanelShowsEntries(t *testing.T) {
	m := New(nil, &tuiCommanderStub{}, "Tracks", "Cmin", 42, 70).WithTrackBrowser([]TrackNavEntry{
		{
			ID:           "lofi/soft-tape-rain-bus",
			Style:        "lofi",
			Substyle:     "dusty-rhodes",
			Title:        "Soft Tape / Rain Bus",
			Key:          "Dmin",
			Tempo:        "72",
			ListenMode:   "album-side",
			SectionCount: 3,
			Sections:     []string{"intro", "a", "outro"},
			Ensemble:     []string{"ep", "bass", "drums", "reed"},
			EventCount:   4,
			Complexity:   "arranged",
			Structure: []TrackNavSection{
				{Label: "intro", Harmony: "Dm9 G13", Events: []string{"pickup"}, RoleNames: []string{"ep", "bass"}, Duration: 30 * time.Second},
				{Label: "a", Harmony: "Bbmaj9 C13", Events: []string{"fill"}, RoleNames: []string{"ep", "reed", "drums"}, Duration: 2 * time.Minute},
			},
		},
		{ID: "jazz/dusty-swing-after-hours", Style: "jazz", Substyle: "trio-after-hours", Title: "Dusty Swing / After Hours", SectionCount: 4, EventCount: 3, Complexity: "through"},
	}, nil, true)
	panel := trackPanel(m, 90, 18, DefaultTheme())
	for _, want := range []string{"TRACK LIBRARY", "2 authored tracks · one performer", "Soft Tape / Rain Bus", "TRACKS", "dusty-rhodes", "FORM", "intro", "0:30", "2:00", "ensemble  ep · bass · drums · reed", "[t] close"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("track panel missing %q:\n%s", want, panel)
		}
	}
}

// TestTrackPanelShowsEngineBadge verifies SP25's per-row engine badge: SF2
// tracks get "[SF2]"; ACE-Step tracks get "[AI]". The badge appears both in
// the list pane and the detail pane.
func TestTrackPanelShowsEngineBadge(t *testing.T) {
	m := New(nil, &tuiCommanderStub{}, "Tracks", "Cmin", 42, 70).WithTrackBrowser([]TrackNavEntry{
		{
			ID:     "lofi/sf2-comparison",
			Style:  "lofi",
			Title:  "SF2 Comparison",
			Engine: "sf2",
		},
		{
			ID:     "lofi/ai-rainy-night",
			Style:  "lofi",
			Title:  "AI Rainy Night",
			Engine: "acestep",
		},
	}, nil, true)
	panel := trackPanel(m, 110, 20, DefaultTheme())
	for _, want := range []string{"[SF2]", "[AI]"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("track panel missing badge %q:\n%s", want, panel)
		}
	}
}

func formMapTestModel() Model {
	m := Model{width: 118, height: 32}
	m.tracks = []TrackNavEntry{{
		ID: "lofi/rainy", Style: "lofi", Substyle: "rainy-cafe", Title: "Rainy Cafe",
		Key: "D minor", Tempo: "84", ListenMode: "hour-stream",
		Description: "warm rhodes over rain",
		Ensemble:    []string{"rhodes", "bass"},
		Tags:        []string{"lofi", "rhodes"},
		Textures:    []string{"rain -36 dB", "vinyl -44 dB"},
		Structure: []TrackNavSection{
			{ID: "intro", Label: "intro", Harmony: "Dm9", Duration: 30 * time.Second},
			{ID: "body", Label: "body", Harmony: "Gm7", Duration: 90 * time.Second},
		},
	}}
	m.trackVisible = true
	return m
}

func TestTrackDetailShowsFormMap(t *testing.T) {
	m := formMapTestModel()
	out := trackPanel(m, 118, 32, DefaultTheme())
	if !strings.Contains(out, "FORM") || !strings.Contains(out, "▰") {
		t.Fatalf("detail missing form map: %q", out)
	}
	if !strings.Contains(out, "0:30") || !strings.Contains(out, "1:30") {
		t.Fatalf("detail missing section durations: %q", out)
	}
	if !strings.Contains(out, "textures  rain -36 dB · vinyl -44 dB") {
		t.Fatalf("detail missing textures line: %q", out)
	}
	if !strings.Contains(out, "#lofi") || !strings.Contains(out, "#rhodes") {
		t.Fatalf("detail missing tag row: %q", out)
	}
	if !strings.Contains(out, "warm rhodes over rain") {
		t.Fatalf("detail missing description: %q", out)
	}
}

func TestTrackHeaderShowsCountAndFilters(t *testing.T) {
	m := formMapTestModel()
	out := trackPanel(m, 118, 32, DefaultTheme())
	if !strings.Contains(out, "1 authored track") {
		t.Fatalf("header missing count: %q", out)
	}
	if !strings.Contains(out, "▌") {
		t.Fatalf("style bar missing active marker: %q", out)
	}
	if !strings.Contains(out, "[t] close   [←→] style   [↑↓] browse   [enter] play") {
		t.Fatalf("footer changed: %q", out)
	}
}

// TestTrackPanelNeverExceedsSize sweeps every width the app can run at, at
// three terminal heights, with content chosen to overflow: a >60 char title,
// long harmony, eight sections of varied length, textures, and many tags.
// The pane is full screen now, so anything that will not fit has to be
// trimmed rather than wrapped — lipgloss's Width().Render() wraps silently,
// which the per-line assertion catches. The empty library is swept too: its
// copy is longer than a narrow pane.
func TestTrackPanelNeverExceedsSize(t *testing.T) {
	theme := DefaultTheme()
	long := TrackNavEntry{
		ID:           "lofi/an-unusually-long-authored-track-identifier-for-width-testing",
		Style:        "lofi",
		Substyle:     "dusty-rhodes-with-a-long-substyle",
		Title:        "An Unusually Long Authored Track Title For Terminal Width Testing",
		Description:  "a description that keeps going well past any reasonable pane width so it has to be trimmed",
		Key:          "Dmin",
		Tempo:        "72",
		ListenMode:   "album-side",
		SectionCount: 8,
		EventCount:   9,
		Complexity:   "arranged",
		Engine:       "acestep",
		Ensemble:     []string{"ep", "bass", "drums", "reed", "pad", "tape"},
		Textures:     []string{"rain -36 dB", "vinyl -44 dB", "tape hiss -48 dB"},
		Tags:         []string{"lofi", "rhodes", "rain", "night", "study", "mellow", "tape", "loop"},
	}
	for i := 0; i < 8; i++ {
		long.Structure = append(long.Structure, TrackNavSection{
			ID:        "section-" + string(rune('a'+i)),
			Label:     "a rather long section label " + string(rune('a'+i)),
			Harmony:   "Dm9 | G13 | Bbmaj9 | C13sus4 | Fmaj7#11 | Ebmaj9 | Am7b5 | D7alt",
			Events:    []string{"pickup", "fill", "swell"},
			RoleNames: []string{"ep", "bass", "drums"},
			Duration:  time.Duration(20+i*47) * time.Second,
		})
	}
	models := map[string]Model{
		"populated": {
			tracks:        []TrackNavEntry{long, {ID: "jazz/short", Style: "jazz", Title: "Short", Engine: "sf2"}},
			activeTrackID: long.ID,
			debug:         gen.DebugStatus{Section: "section-c"},
		},
		"empty": {},
	}
	for name, base := range models {
		for _, h := range []int{18, 24, 32} {
			for w := 40; w <= 200; w++ {
				m := base
				m.trackVisible = true
				m.width, m.height = w, h
				out := trackPanel(m, w, h, theme)
				lines := strings.Split(out, "\n")
				if len(lines) > h {
					t.Fatalf("track panel (%s) is %d rows tall at w=%d h=%d, want <= %d\n%s",
						name, len(lines), w, h, h, out)
				}
				for i, line := range lines {
					if got := lipgloss.Width(line); got > w {
						t.Fatalf("track panel (%s) line %d overflows: w=%d h=%d rendered %d cols\n%q",
							name, i, w, h, got, line)
					}
				}
			}
		}
	}
}

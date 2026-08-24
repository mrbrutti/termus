package tui

import (
	"fmt"
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
	// A crowded library: enough styles that the filter row has to window, and
	// a selection deep enough that the list pane has to scroll.
	crowded := make([]TrackNavEntry, 0, 14)
	styles := []string{"ambient", "bells", "classical", "drone", "jazz", "lofi", "phase"}
	for i := 0; i < 14; i++ {
		crowded = append(crowded, TrackNavEntry{
			ID:       fmt.Sprintf("%s/track-%02d", styles[i%len(styles)], i),
			Style:    styles[i%len(styles)],
			Substyle: "a-substyle-that-is-not-short",
			Title:    fmt.Sprintf("Crowded Library Entry Number %02d With A Long Title", i),
			Tempo:    "88",
			Ensemble: []string{"ep", "bass", "drums"},
		})
	}
	models := map[string]Model{
		"populated": {
			tracks:        []TrackNavEntry{long, {ID: "jazz/short", Style: "jazz", Title: "Short", Engine: "sf2"}},
			activeTrackID: long.ID,
			debug:         gen.DebugStatus{Section: "section-c"},
		},
		"long-list": {
			tracks:        crowded,
			trackIdx:      11,
			trackStyleIdx: 4,
			activeTrackID: crowded[3].ID,
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

// TestTrackPanelKeepsTailAndActiveFilter guards the two things a naive
// truncation drops first: the detail pane's tail ("● currently loaded" sits
// below a form map long enough to be cut) and the active style filter (which
// is not the first chip, so a row that packs from the head hides it and ←/→
// looks dead).
func TestTrackPanelKeepsTailAndActiveFilter(t *testing.T) {
	theme := DefaultTheme()

	loaded := TrackNavEntry{
		ID: "lofi/long-form", Style: "lofi", Substyle: "rainy-cafe", Title: "Long Form",
		Key: "D minor", Tempo: "84", ListenMode: "hour-stream",
		Description: "twenty sections, more than any pane can show",
		Ensemble:    []string{"rhodes", "bass"},
		Textures:    []string{"rain -36 dB"},
		Tags:        []string{"lofi", "rain"},
	}
	for i := 0; i < 20; i++ {
		loaded.Structure = append(loaded.Structure, TrackNavSection{
			ID:       fmt.Sprintf("part-%02d", i),
			Label:    fmt.Sprintf("part %02d", i),
			Harmony:  "Dm9 | G13",
			Duration: time.Duration(30+i*10) * time.Second,
		})
	}
	m := Model{width: 118, height: 32, tracks: []TrackNavEntry{loaded}, activeTrackID: loaded.ID}
	m.trackVisible = true
	out := trackPanel(m, 118, 32, theme)
	if !strings.Contains(out, "…") {
		t.Fatalf("form map should have truncated at 20 sections: %q", out)
	}
	if !strings.Contains(out, "● currently loaded") {
		t.Fatalf("truncated form map pushed the detail tail off the pane: %q", out)
	}
	if !strings.Contains(out, "textures  rain -36 dB") {
		t.Fatalf("truncated form map pushed the textures row off the pane: %q", out)
	}

	// A short terminal has no room for the form map at all. Spending two rows
	// on "FORM" plus a lone "…" conveys nothing and costs the tail its place,
	// so the whole block drops instead.
	short := m
	short.width, short.height = 120, 18
	shortOut := trackPanel(short, 120, 18, theme)
	if !strings.Contains(shortOut, "● currently loaded") {
		t.Fatalf("h=18 detail pane dropped the tail instead of the form map: %q", shortOut)
	}
	shortLines := strings.Split(shortOut, "\n")
	if len(shortLines) > 18 {
		t.Fatalf("h=18 track panel is %d rows tall, want <= 18\n%s", len(shortLines), shortOut)
	}
	for i, line := range shortLines {
		if got := lipgloss.Width(line); got > 120 {
			t.Fatalf("h=18 track panel line %d rendered %d cols, want <= 120\n%q", i, got, line)
		}
	}

	// Seven styles at the narrowest supported width: the filter row cannot
	// show them all, and the active one is well past the head.
	var wide []TrackNavEntry
	for i, style := range []string{"ambient", "bells", "classical", "drone", "jazz", "lofi", "phase"} {
		wide = append(wide, TrackNavEntry{ID: fmt.Sprintf("%s/t%d", style, i), Style: style, Title: "T"})
	}
	for _, idx := range []int{0, 3, 5, 7} {
		narrow := Model{width: 40, height: 24, tracks: wide, trackStyleIdx: idx}
		narrow.trackVisible = true
		got := trackPanel(narrow, 40, 24, theme)
		if !strings.Contains(got, "▌"+trackStyleGlyph(narrow.currentTrackStyle())) {
			t.Fatalf("active filter %q (idx %d) not visible at w=40: %q", narrow.currentTrackStyle(), idx, got)
		}
	}
}

package tui

import (
	"strings"
	"testing"
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
				{Label: "intro", Harmony: "Dm9 G13", Events: []string{"pickup"}, RoleNames: []string{"ep", "bass"}},
				{Label: "a", Harmony: "Bbmaj9 C13", Events: []string{"fill"}, RoleNames: []string{"ep", "reed", "drums"}},
			},
		},
		{ID: "jazz/dusty-swing-after-hours", Style: "jazz", Substyle: "trio-after-hours", Title: "Dusty Swing / After Hours", SectionCount: 4, EventCount: 3, Complexity: "through"},
	}, nil, true)
	panel := trackPanel(m, 90, 18, DefaultTheme())
	for _, want := range []string{"TRACK LIBRARY", "Soft Tape / Rain Bus", "TRACKS", "dusty-rhodes", "03 sections", "04 moments", "ensemble  ep · bass · drums · reed", "[t] close"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("track panel missing %q:\n%s", want, panel)
		}
	}
}

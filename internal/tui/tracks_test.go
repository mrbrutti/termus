package tui

import (
	"strings"
	"testing"
)

func TestTrackPanelShowsEntries(t *testing.T) {
	m := New(nil, &tuiCommanderStub{}, "Tracks", "Cmin", 42, 70).WithTrackBrowser([]TrackNavEntry{
		{ID: "lofi/soft-tape-rain-bus", Style: "lofi", Title: "Soft Tape / Rain Bus", Description: "Late-night ride"},
		{ID: "jazz/dusty-swing-after-hours", Style: "jazz", Title: "Dusty Swing / After Hours"},
	}, nil, true)
	panel := trackPanel(m, 90, 18, DefaultTheme())
	for _, want := range []string{"TRACK NAVIGATOR", "lofi/soft-tape-rain-bus", "Soft Tape / Rain Bus"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("track panel missing %q:\n%s", want, panel)
		}
	}
}

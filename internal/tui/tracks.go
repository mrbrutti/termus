package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func trackPanel(m Model, w, h int, theme ColorTheme) string {
	outer := lipgloss.NewStyle().
		Width(w).
		Height(h).
		Padding(1, 2).
		Foreground(theme.BarFg)
	if len(m.tracks) == 0 {
		return outer.Render(strings.Join([]string{
			"TRACKS",
			"",
			"No authored tracks found.",
			"Add .tm files under tracks/<style>/ to browse them here.",
		}, "\n"))
	}
	lines := []string{
		lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render("TRACK NAVIGATOR"),
		lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render("Open an authored track and let the playlist walk its sections."),
		"",
	}
	maxRows := maxInt(1, h-7)
	start := 0
	if m.trackIdx >= maxRows {
		start = m.trackIdx - maxRows + 1
	}
	end := minInt(len(m.tracks), start+maxRows)
	for i := start; i < end; i++ {
		entry := m.tracks[i]
		line := fmt.Sprintf("%-28s  %s", entry.ID, entry.Title)
		if entry.Description != "" {
			line += " · " + entry.Description
		}
		if i == m.trackIdx {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("#111111")).Background(theme.BarHi).Render(" " + line + " ")
		} else if entry.ID == m.activeTrackID {
			line = lipgloss.NewStyle().Foreground(theme.BarHi).Render("• " + line)
		}
		lines = append(lines, line)
	}
	lines = append(lines, "", lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render("[enter] load   [↑↓] browse   [esc] close"))
	return outer.Render(strings.Join(lines, "\n"))
}

package tui

import (
	"fmt"
	"sort"
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
			lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render("TRACK LIBRARY"),
			"",
			"No authored tracks found.",
			"Add .tm files under tracks/<style>/ to browse them here.",
		}, "\n"))
	}
	title := lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render("TRACK LIBRARY")
	subtitle := lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render("authored songs · one performer · [enter] play")
	accent := renderStartupBrailleBar(maxInt(18, minInt(w-8, 42)), 1, 1, 0.3, theme)
	header := lipgloss.JoinVertical(lipgloss.Left, title, subtitle, accent)

	styleBar := renderTrackStyleBar(m, theme, w-4)
	bodyH := maxInt(8, h-8)
	leftW := clampInt(int(float64(w)*0.38), 24, maxInt(24, w-34))
	rightW := maxInt(18, w-leftW-7)
	listPane := renderTrackListPane(m, leftW, bodyH, theme)
	detailPane := renderTrackDetailPane(m, rightW, bodyH, theme)
	divider := renderTrackDivider(bodyH, theme)
	footer := lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render("[t] close   [←→] style   [↑↓] browse   [enter] play")

	body := lipgloss.JoinHorizontal(lipgloss.Top, listPane, "  ", divider, "  ", detailPane)
	return outer.Render(lipgloss.JoinVertical(lipgloss.Left, header, "", styleBar, "", body, "", footer))
}

func (m Model) trackStyleOptions() []string {
	if len(m.tracks) == 0 {
		return []string{"all"}
	}
	seen := map[string]bool{"all": true}
	out := []string{"all"}
	for _, entry := range m.tracks {
		style := strings.TrimSpace(entry.Style)
		if style == "" || seen[style] {
			continue
		}
		seen[style] = true
		out = append(out, style)
	}
	sort.Strings(out[1:])
	return out
}

func (m Model) currentTrackStyle() string {
	styles := m.trackStyleOptions()
	if len(styles) == 0 {
		return "all"
	}
	if m.trackStyleIdx < 0 || m.trackStyleIdx >= len(styles) {
		return styles[0]
	}
	return styles[m.trackStyleIdx]
}

func (m Model) filteredTrackIndices() []int {
	style := m.currentTrackStyle()
	out := make([]int, 0, len(m.tracks))
	for i, entry := range m.tracks {
		if style == "all" || strings.EqualFold(entry.Style, style) {
			out = append(out, i)
		}
	}
	return out
}

func renderTrackStyleBar(m Model, theme ColorTheme, width int) string {
	styles := m.trackStyleOptions()
	active := m.currentTrackStyle()
	parts := make([]string, 0, len(styles))
	for _, style := range styles {
		count := 0
		for _, entry := range m.tracks {
			if style == "all" || strings.EqualFold(entry.Style, style) {
				count++
			}
		}
		text := fmt.Sprintf("%s %d", style, count)
		if strings.EqualFold(style, active) {
			parts = append(parts, lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render("["+text+"]"))
			continue
		}
		parts = append(parts, lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render(text))
	}
	return trimToWidth(strings.Join(parts, "  "), width)
}

func renderTrackListPane(m Model, w, h int, theme ColorTheme) string {
	indices := m.filteredTrackIndices()
	style := lipgloss.NewStyle().Width(w).Height(h)
	if len(indices) == 0 {
		return style.Render("No tracks in this style filter.")
	}
	lines := []string{
		lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render("TRACKS"),
		"",
	}
	maxRows := maxInt(2, (h-2)/3)
	currentPos := 0
	for i, idx := range indices {
		if idx == m.trackIdx {
			currentPos = i
			break
		}
	}
	start := 0
	if currentPos >= maxRows {
		start = currentPos - maxRows + 1
	}
	end := minInt(len(indices), start+maxRows)
	for _, idx := range indices[start:end] {
		entry := m.tracks[idx]
		title := entry.Title
		if title == "" {
			title = entry.ID
		}
		meta := entry.ID
		if entry.Description != "" {
			meta = entry.Description
		}
		prefix := "  "
		if idx == m.trackIdx {
			prefix = "› "
		} else if entry.ID == m.activeTrackID {
			prefix = "• "
		}
		block := lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Bold(idx == m.trackIdx).Render(prefix+trimToWidth(title, maxInt(8, w-2))),
			lipgloss.NewStyle().Faint(true).Render(trimToWidth("  "+meta, maxInt(8, w-2))),
		)
		if idx == m.trackIdx {
			block = lipgloss.NewStyle().Foreground(theme.BarHi).Render(block)
		} else if entry.ID == m.activeTrackID {
			block = lipgloss.NewStyle().Foreground(theme.BarFg).Render(block)
		}
		lines = append(lines, block, "")
	}
	return style.Render(strings.TrimRight(strings.Join(lines, "\n"), "\n"))
}

func renderTrackDetailPane(m Model, w, h int, theme ColorTheme) string {
	style := lipgloss.NewStyle().Width(w).Height(h)
	if len(m.tracks) == 0 || m.trackIdx < 0 || m.trackIdx >= len(m.tracks) {
		return style.Render("")
	}
	entry := m.tracks[m.trackIdx]
	title := entry.Title
	if title == "" {
		title = entry.ID
	}
	lines := []string{
		lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render(title),
	}
	meta := make([]string, 0, 4)
	if entry.Style != "" {
		meta = append(meta, entry.Style)
	}
	if entry.Key != "" {
		meta = append(meta, entry.Key)
	}
	if entry.Tempo != "" {
		meta = append(meta, entry.Tempo+" bpm")
	}
	if entry.ListenMode != "" {
		meta = append(meta, entry.ListenMode)
	}
	if len(meta) > 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render(strings.Join(meta, " · ")))
	}
	if entry.Description != "" {
		lines = append(lines, "", trimToWidth(entry.Description, w))
	}
	if len(entry.Tags) > 0 {
		lines = append(lines, "", renderTrackTags(entry.Tags, theme, w))
	}
	if len(entry.Sections) > 0 {
		lines = append(lines, "", lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render("SECTIONS"))
		maxSections := maxInt(3, h-len(lines)-2)
		for i, section := range entry.Sections {
			if i >= maxSections {
				lines = append(lines, lipgloss.NewStyle().Faint(true).Render("…"))
				break
			}
			lines = append(lines, fmt.Sprintf("%02d  %s", i+1, trimToWidth(section, maxInt(8, w-4))))
		}
	}
	if entry.ID == m.activeTrackID {
		lines = append(lines, "", lipgloss.NewStyle().Foreground(theme.BarHi).Render("currently loaded"))
	}
	return style.Render(strings.Join(lines, "\n"))
}

func renderTrackTags(tags []string, theme ColorTheme, width int) string {
	parts := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		parts = append(parts, lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render("#"+tag))
	}
	return trimToWidth(strings.Join(parts, "  "), width)
}

func renderTrackDivider(height int, theme ColorTheme) string {
	lines := make([]string, maxInt(1, height))
	for i := range lines {
		lines[i] = lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render("│")
	}
	return strings.Join(lines, "\n")
}

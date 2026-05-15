package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrbrutti/termus/internal/gen"
)

type controlTab int

const (
	controlTabMusic controlTab = iota
	controlTabCurate
	controlTabSessions
	controlTabAudio
)

type controlItem struct {
	Title    string
	Value    string
	Hint     string
	Disabled bool
	Adjust   func(*Model, int)
	Activate func(*Model) tea.Cmd
}

func (t controlTab) label() string {
	switch t {
	case controlTabMusic:
		return "music"
	case controlTabCurate:
		return "curate"
	case controlTabSessions:
		return "sessions"
	case controlTabAudio:
		return "audio"
	default:
		return "music"
	}
}

func (t controlTab) next() controlTab {
	return controlTab((int(t) + 1) % 4)
}

func (t controlTab) prev() controlTab {
	return controlTab((int(t) + 3) % 4)
}

func (m *Model) moveControlRow(delta int) {
	items := m.controlItems()
	if len(items) == 0 {
		m.controlRow = 0
		return
	}
	m.controlRow = (m.controlRow + delta + len(items)) % len(items)
}

func (m *Model) adjustControlRow(delta int) {
	items := m.controlItems()
	if len(items) == 0 {
		return
	}
	idx := clampInt(m.controlRow, 0, len(items)-1)
	if item := items[idx]; item.Adjust != nil && !item.Disabled {
		item.Adjust(m, delta)
	}
}

func (m *Model) activateControlRow() tea.Cmd {
	items := m.controlItems()
	if len(items) == 0 {
		return nil
	}
	idx := clampInt(m.controlRow, 0, len(items)-1)
	if item := items[idx]; item.Activate != nil && !item.Disabled {
		return item.Activate(m)
	}
	return nil
}

func (m Model) controlItems() []controlItem {
	switch m.controlTab {
	case controlTabMusic:
		profile := gen.DefaultControlProfile()
		if m.musicProfile != nil {
			profile = *m.musicProfile
		}
		return []controlItem{
			{
				Title: "density",
				Value: macroLabel(profile.Density, []string{"air", "lean", "steady", "lush", "full"}),
				Hint:  "left/right rebuild",
				Adjust: func(m *Model, delta int) {
					m.updateMusicProfile("density", func(profile *gen.ControlProfile) {
						profile.Density += delta
					})
				},
			},
			{
				Title: "brightness",
				Value: macroLabel(profile.Brightness, []string{"soft", "warm", "natural", "clear", "gloss"}),
				Hint:  "left/right rebuild",
				Adjust: func(m *Model, delta int) {
					m.updateMusicProfile("brightness", func(profile *gen.ControlProfile) {
						profile.Brightness += delta
					})
				},
			},
			{
				Title: "motion",
				Value: macroLabel(profile.Motion, []string{"still", "settled", "breathing", "glide", "orbit"}),
				Hint:  "left/right rebuild",
				Adjust: func(m *Model, delta int) {
					m.updateMusicProfile("motion", func(profile *gen.ControlProfile) {
						profile.Motion += delta
					})
				},
			},
			{
				Title: "reverb",
				Value: macroLabel(profile.Reverb, []string{"dry", "close", "room", "hall", "halo"}),
				Hint:  "left/right rebuild",
				Adjust: func(m *Model, delta int) {
					m.updateMusicProfile("reverb", func(profile *gen.ControlProfile) {
						profile.Reverb += delta
					})
				},
			},
			{
				Title: "swing",
				Value: macroLabel(profile.Swing, []string{"straight", "tight", "groove", "late", "loose"}),
				Hint:  "left/right rebuild",
				Adjust: func(m *Model, delta int) {
					m.updateMusicProfile("swing", func(profile *gen.ControlProfile) {
						profile.Swing += delta
					})
				},
			},
			{
				Title: "drone depth",
				Value: macroLabel(profile.DroneDepth, []string{"light", "trim", "grounded", "deep", "sub"}),
				Hint:  "left/right rebuild",
				Adjust: func(m *Model, delta int) {
					m.updateMusicProfile("drone depth", func(profile *gen.ControlProfile) {
						profile.DroneDepth += delta
					})
				},
			},
		}
	case controlTabCurate:
		rec, ok := m.currentSeedRecord()
		recent, hasRecent := m.selectedRecentRecord()
		best, hasBest := m.selectedBestRecord()
		tagName := ""
		if len(curationTags) > 0 {
			tagName = curationTags[m.curateTagIdx%len(curationTags)]
		}
		ratingValue := "0"
		favoriteValue := "off"
		tagsValue := "none"
		if ok {
			ratingValue = ratingString(rec.Rating)
			favoriteValue = onOff(rec.Favorite)
			tagsValue = currentTagsLabel(rec.Tags)
		}
		recentValue := "no history yet"
		recentHint := "browse after playing"
		if hasRecent {
			recentValue = curationLabel(recent)
			recentHint = "left/right browse · enter load"
		}
		bestValue := "no rated takes yet"
		bestHint := "favorite or rate a take"
		if hasBest {
			bestValue = curationLabel(best)
			bestHint = "left/right browse · enter load"
		}
		return []controlItem{
			{
				Title: "keep current",
				Value: fmt.Sprintf("%s/%d", m.algo, m.seed),
				Hint:  "enter save",
				Activate: func(m *Model) tea.Cmd {
					m.keepSeed()
					return nil
				},
			},
			{
				Title: "rating",
				Value: ratingValue,
				Hint:  "left/right adjust",
				Adjust: func(m *Model, delta int) {
					m.adjustCurrentRating(delta)
				},
			},
			{
				Title: "favorite",
				Value: favoriteValue,
				Hint:  "enter toggle",
				Activate: func(m *Model) tea.Cmd {
					m.toggleCurrentFavorite()
					return nil
				},
			},
			{
				Title: "tags",
				Value: tagsValue,
				Hint:  fmt.Sprintf("left/right pick · enter toggle %s", tagName),
				Adjust: func(m *Model, delta int) {
					m.cycleCurationTag(delta)
				},
				Activate: func(m *Model) tea.Cmd {
					m.toggleCurrentTag()
					return nil
				},
			},
			{
				Title:    "recent history",
				Value:    recentValue,
				Hint:     recentHint,
				Disabled: !hasRecent,
				Adjust: func(m *Model, delta int) {
					m.browseRecent(delta)
				},
				Activate: func(m *Model) tea.Cmd {
					m.loadSelectedRecent()
					return nil
				},
			},
			{
				Title:    "best takes",
				Value:    bestValue,
				Hint:     bestHint,
				Disabled: !hasBest,
				Adjust: func(m *Model, delta int) {
					m.browseBest(delta)
				},
				Activate: func(m *Model) tea.Cmd {
					m.loadSelectedBest()
					return nil
				},
			},
			{
				Title: "saved library",
				Value: fmt.Sprintf("%d items", len(m.savedSeeds)),
				Hint:  "enter open",
				Activate: func(m *Model) tea.Cmd {
					m.toggleLibrary()
					return nil
				},
			},
		}
	case controlTabSessions:
		selected, ok := m.selectedSession()
		sessionValue := "none saved"
		sessionHint := "enter save one"
		if ok {
			sessionValue = sessionLabel(selected)
			sessionHint = fmt.Sprintf("left/right browse · %s ago", formatSessionAge(time.Now(), selected.SavedAt))
		}
		return []controlItem{
			{
				Title: "save snapshot",
				Value: fmt.Sprintf("%s · %s · %s", m.algo, Visuals[m.visualIdx].Name, m.activeTheme().Name),
				Hint:  "enter save",
				Activate: func(m *Model) tea.Cmd {
					m.saveCurrentSession()
					return nil
				},
			},
			{
				Title:    "saved sessions",
				Value:    sessionValue,
				Hint:     sessionHint,
				Disabled: !ok,
				Adjust: func(m *Model, delta int) {
					m.browseSession(delta)
				},
			},
			{
				Title:    "load selected",
				Value:    "restore algo / seed / view / volume",
				Hint:     "enter load",
				Disabled: !ok,
				Activate: func(m *Model) tea.Cmd {
					m.loadSelectedSession()
					return nil
				},
			},
			{
				Title:    "remove selected",
				Value:    "delete saved snapshot",
				Hint:     "enter remove",
				Disabled: !ok,
				Activate: func(m *Model) tea.Cmd {
					m.deleteSelectedSession()
					return nil
				},
			},
		}
	default:
		return []controlItem{
			{
				Title:    "backend",
				Value:    m.currentStatusLabel(),
				Hint:     "status",
				Disabled: true,
			},
			{
				Title: "export drawer",
				Value: "wav / midi / stems",
				Hint:  "enter open",
				Activate: func(m *Model) tea.Cmd {
					m.toggleExportDrawer()
					return nil
				},
			},
			{
				Title: "recording",
				Value: recordingLabel(m.recording),
				Hint:  "enter toggle",
				Activate: func(m *Model) tea.Cmd {
					path, err := m.cmd.ToggleRecord()
					if err != nil {
						m.flashStatus("rec error: "+err.Error(), 3*time.Second)
						m.recording = false
						return nil
					}
					if path != "" {
						m.recording = true
						m.recordStartedAt = time.Now()
						m.flashStatus("rec → "+path, 3*time.Second)
						return nil
					}
					m.recording = false
					m.recordStartedAt = time.Time{}
					m.flashStatus("rec stopped", 3*time.Second)
					return nil
				},
			},
			{
				Title: "debug inspector",
				Value: onOff(m.debugVisible),
				Hint:  "enter toggle",
				Activate: func(m *Model) tea.Cmd {
					m.debugVisible = !m.debugVisible
					if m.debugVisible {
						m.flashStatus("debug: on", 2*time.Second)
					} else {
						m.flashStatus("debug: off", 2*time.Second)
					}
					return nil
				},
			},
		}
	}
}

func currentTabItems(m Model) []controlItem {
	return m.controlItems()
}

func controlsPanel(m Model, w, h int, theme ColorTheme) string {
	bodyW := maxInt(34, minInt(w-6, 88))
	bodyH := maxInt(14, minInt(h-2, 22))
	tabs := []string{
		renderControlTab(theme, m.controlTab == controlTabMusic, controlTabMusic.label()),
		renderControlTab(theme, m.controlTab == controlTabCurate, controlTabCurate.label()),
		renderControlTab(theme, m.controlTab == controlTabSessions, controlTabSessions.label()),
		renderControlTab(theme, m.controlTab == controlTabAudio, controlTabAudio.label()),
	}
	items := currentTabItems(m)
	lines := make([]string, 0, len(items))
	for i, item := range items {
		lines = append(lines, renderControlItem(theme, i == m.controlRow, item, bodyW-8))
	}
	panel := lipgloss.NewStyle().
		Width(bodyW).
		Height(bodyH).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.BarFg).
		Padding(1, 2).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render("CONTROL CENTER"),
				lipgloss.NewStyle().Faint(true).Render(strings.Join(tabs, "  ")),
				"",
				strings.Join(lines, "\n"),
				"",
				lipgloss.NewStyle().Faint(true).Render("[tab] switch  [↑↓] browse  [←→] adjust  [enter] apply  [m] close"),
			),
		)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, panel)
}

func renderControlTab(theme ColorTheme, active bool, label string) string {
	style := lipgloss.NewStyle().Faint(true)
	if active {
		style = lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true)
	}
	return style.Render(strings.ToUpper(label))
}

func renderControlItem(theme ColorTheme, active bool, item controlItem, w int) string {
	cursor := " "
	if active {
		cursor = "›"
	}
	titleStyle := lipgloss.NewStyle().Foreground(theme.BarHi)
	valueStyle := lipgloss.NewStyle().Foreground(theme.BarFg)
	hintStyle := lipgloss.NewStyle().Faint(true)
	if item.Disabled {
		titleStyle = titleStyle.Faint(true)
		valueStyle = valueStyle.Faint(true)
	}
	left := titleStyle.Render(item.Title)
	value := valueStyle.Render(item.Value)
	base := cursor + " " + left
	right := value
	if item.Hint != "" {
		right += "  " + hintStyle.Render(item.Hint)
	}
	base = trimToWidth(base, maxInt(0, w/2))
	right = trimToWidth(right, maxInt(0, w-lipgloss.Width(base)-1))
	pad := w - lipgloss.Width(base) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	return base + spaces(pad) + right
}

func (m Model) currentStatusLabel() string {
	if status := m.currentStatus(time.Now()); status != "" {
		return status
	}
	return "audio: ready"
}

func recordingLabel(recording bool) string {
	if recording {
		return "on"
	}
	return "off"
}

func macroLabel(value int, labels []string) string {
	if len(labels) == 0 {
		return ""
	}
	value = clampInt(value, 0, len(labels)-1)
	return labels[value]
}

func onOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

func clampInt(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

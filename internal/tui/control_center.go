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
	controlTabNow controlTab = iota
	controlTabLook
	controlTabMusic
	controlTabSeeds
	controlTabLibrary
	controlTabExport
	controlTabAudio
	controlTabDebug
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
	case controlTabNow:
		return "now"
	case controlTabLook:
		return "look"
	case controlTabMusic:
		return "music"
	case controlTabSeeds:
		return "seeds"
	case controlTabLibrary:
		return "library"
	case controlTabExport:
		return "export"
	case controlTabAudio:
		return "audio"
	case controlTabDebug:
		return "debug"
	default:
		return "now"
	}
}

func (t controlTab) next() controlTab {
	return controlTab((int(t) + 1) % 8)
}

func (t controlTab) prev() controlTab {
	return controlTab((int(t) + 7) % 8)
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
	case controlTabNow:
		return m.nowControlItems()
	case controlTabLook:
		return m.lookControlItems()
	case controlTabMusic:
		return m.musicControlItems()
	case controlTabSeeds:
		return m.seedControlItems()
	case controlTabLibrary:
		return m.libraryControlItems()
	case controlTabExport:
		return m.exportControlItems()
	case controlTabAudio:
		return m.audioControlItems()
	default:
		return m.debugControlItems()
	}
}

func (m Model) nowControlItems() []controlItem {
	playback := "playing"
	if m.paused {
		playback = "paused"
	}
	trackValue := fmt.Sprintf("%s · seed %d", m.algo, m.seed)
	trackHint := "current take"
	if m.playlist != nil && m.playlistIdx < len(m.playlist.Tracks) {
		trackValue = fmt.Sprintf("%s · %d/%d", m.algo, m.playlistIdx+1, len(m.playlist.Tracks))
		trackHint = shortDuration(time.Until(m.nextTrackAt)) + " to next"
	}
	modeValue := m.listeningMode
	if modeValue == "" {
		modeValue = "endless"
	}
	return []controlItem{
		{
			Title: "playback",
			Value: playback,
			Hint:  "enter toggle",
			Activate: func(m *Model) tea.Cmd {
				m.paused = !m.paused
				m.cmd.TogglePause()
				if m.paused {
					m.flashStatus("paused", 2*time.Second)
				} else {
					m.flashStatus("playing", 2*time.Second)
				}
				return nil
			},
		},
		{
			Title:    "mode",
			Value:    modeValue,
			Hint:     "session profile",
			Disabled: true,
		},
		{
			Title:    "current track",
			Value:    trackValue,
			Hint:     trackHint,
			Disabled: true,
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
			Title:    "skip playlist track",
			Value:    skipLabel(m.playlist != nil),
			Hint:     "enter skip",
			Disabled: m.playlist == nil,
			Activate: func(m *Model) tea.Cmd {
				if m.playlist != nil {
					m.advancePlaylist()
				}
				return nil
			},
		},
	}
}

func (m Model) lookControlItems() []controlItem {
	return []controlItem{
		{
			Title: "visual",
			Value: Visuals[m.visualIdx].Name,
			Hint:  "left/right cycle",
			Adjust: func(m *Model, delta int) {
				m.startVisualTransition(m.visualIdx + delta)
				m.flashStatus("visual: "+Visuals[m.visualIdx].Name, 2*time.Second)
			},
		},
		{
			Title:    "theme",
			Value:    m.activeTheme().Name,
			Hint:     "left/right cycle",
			Disabled: len(m.themes) <= 1,
			Adjust: func(m *Model, delta int) {
				if len(m.themes) <= 1 {
					return
				}
				m.themeIdx = (m.themeIdx + delta + len(m.themes)) % len(m.themes)
				m.flashStatus("theme: "+m.themes[m.themeIdx].Name, 2*time.Second)
			},
		},
		{
			Title: "chrome",
			Value: chromeModeLabel(m.reducedChrome),
			Hint:  "enter toggle",
			Activate: func(m *Model) tea.Cmd {
				m.toggleReducedChrome()
				return nil
			},
		},
		{
			Title: "help overlay",
			Value: onOff(m.helpVisible),
			Hint:  "enter toggle",
			Activate: func(m *Model) tea.Cmd {
				m.helpVisible = !m.helpVisible
				m.controlsVisible = false
				if m.helpVisible {
					m.flashStatus("help: on", 2*time.Second)
				} else {
					m.flashStatus("help: off", 2*time.Second)
				}
				return nil
			},
		},
	}
}

func (m Model) musicControlItems() []controlItem {
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
		{
			Title: "tempo",
			Value: macroLabel(profile.Tempo, []string{"slower", "laid back", "natural", "driven", "urgent"}),
			Hint:  "rhythmic genres",
			Adjust: func(m *Model, delta int) {
				m.updateMusicProfile("tempo", func(profile *gen.ControlProfile) {
					profile.Tempo += delta
				})
			},
		},
		{
			Title: "phrase length",
			Value: macroLabel(profile.Phrase, []string{"short", "trim", "natural", "long", "floating"}),
			Hint:  "ambient textures",
			Adjust: func(m *Model, delta int) {
				m.updateMusicProfile("phrase length", func(profile *gen.ControlProfile) {
					profile.Phrase += delta
				})
			},
		},
		{
			Title: "seed morph",
			Value: macroLabel(m.morphMode, []string{"cut", "quick", "blend", "wash", "drift"}),
			Hint:  "seed and algo swaps",
			Adjust: func(m *Model, delta int) {
				m.morphMode = clampInt(m.morphMode+delta, 0, 4)
				m.flashStatus("seed morph: "+macroLabel(m.morphMode, []string{"cut", "quick", "blend", "wash", "drift"}), 2*time.Second)
			},
		},
	}
}

func (m Model) seedControlItems() []controlItem {
	hasAB := m.seedA != nil || m.seedB != nil
	return []controlItem{
		{
			Title:    "algorithm",
			Value:    m.algo,
			Hint:     "left/right cycle",
			Disabled: len(m.genres) <= 1 || m.playlist != nil,
			Adjust: func(m *Model, delta int) {
				if delta < 0 {
					m.switchAlgo(-1)
				} else if delta > 0 {
					m.switchAlgo(1)
				}
			},
		},
		{
			Title:    "current seed",
			Value:    fmt.Sprintf("%d", m.seed),
			Hint:     "left/right browse",
			Disabled: m.playlist != nil,
			Adjust: func(m *Model, delta int) {
				m.browseSeed(int64(delta))
			},
		},
		{
			Title: "slot A",
			Value: slotSeedLabel(m.seedA),
			Hint:  "enter store",
			Activate: func(m *Model) tea.Cmd {
				m.storeSeed("A")
				return nil
			},
		},
		{
			Title: "slot B",
			Value: slotSeedLabel(m.seedB),
			Hint:  "enter store",
			Activate: func(m *Model) tea.Cmd {
				m.storeSeed("B")
				return nil
			},
		},
		{
			Title:    "compare A/B",
			Value:    compareSeedLabel(m.seedA, m.seedB),
			Hint:     "enter toggle",
			Disabled: !hasAB,
			Activate: func(m *Model) tea.Cmd {
				m.toggleSeedCompare()
				return nil
			},
		},
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
			Title:    "reject current",
			Value:    "advance to next seed",
			Hint:     "enter reject",
			Disabled: m.playlist != nil,
			Activate: func(m *Model) tea.Cmd {
				m.rejectSeed()
				return nil
			},
		},
	}
}

func (m Model) libraryControlItems() []controlItem {
	rec, ok := m.currentSeedRecord()
	recent, hasRecent := m.selectedRecentRecord()
	best, hasBest := m.selectedBestRecord()
	selectedSession, hasSession := m.selectedSession()
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
	sessionValue := "none saved"
	sessionHint := "enter save one"
	if hasSession {
		sessionValue = sessionLabel(selectedSession)
		sessionHint = fmt.Sprintf("left/right browse · %s ago", formatSessionAge(time.Now(), selectedSession.SavedAt))
	}
	return []controlItem{
		{
			Title: "saved library",
			Value: fmt.Sprintf("%d items", len(m.savedSeeds)),
			Hint:  "enter open",
			Activate: func(m *Model) tea.Cmd {
				m.toggleLibrary()
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
			Title: "save session",
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
			Disabled: !hasSession,
			Adjust: func(m *Model, delta int) {
				m.browseSession(delta)
			},
		},
		{
			Title:    "load session",
			Value:    "restore algo / seed / view / volume",
			Hint:     "enter load",
			Disabled: !hasSession,
			Activate: func(m *Model) tea.Cmd {
				m.loadSelectedSession()
				return nil
			},
		},
	}
}

func (m Model) exportControlItems() []controlItem {
	duration := "unavailable"
	if m.exporter != nil {
		duration = m.exporter.durationLabel()
	}
	exportDisabled := m.exporter == nil || m.exportBusy
	hint := "enter render"
	if exportDisabled && m.exportBusy {
		hint = "busy"
	} else if exportDisabled {
		hint = "unavailable"
	}
	return []controlItem{
		{
			Title:    "WAV",
			Value:    duration,
			Hint:     hint,
			Disabled: exportDisabled || m.exporter.WAV == nil,
			Activate: func(m *Model) tea.Cmd { return m.startExport("wav") },
		},
		{
			Title:    "MIDI",
			Value:    duration,
			Hint:     hint,
			Disabled: exportDisabled || m.exporter.MIDI == nil,
			Activate: func(m *Model) tea.Cmd { return m.startExport("midi") },
		},
		{
			Title:    "stems",
			Value:    duration,
			Hint:     hint,
			Disabled: exportDisabled || m.exporter.Stems == nil,
			Activate: func(m *Model) tea.Cmd { return m.startExport("stems") },
		},
		{
			Title:    "recording",
			Value:    recordingLabel(m.recording),
			Hint:     "enter toggle",
			Disabled: false,
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
	}
}

func (m Model) audioControlItems() []controlItem {
	return []controlItem{
		{
			Title:    "backend",
			Value:    m.currentStatusLabel(),
			Hint:     "status",
			Disabled: true,
		},
		{
			Title: "retry live audio",
			Value: "new startup attempt",
			Hint:  "enter retry",
			Activate: func(m *Model) tea.Cmd {
				m.retryAudio()
				return nil
			},
		},
		{
			Title: "render-only fallback",
			Value: "disable live audio expectation",
			Hint:  "enter fallback",
			Activate: func(m *Model) tea.Cmd {
				m.fallbackRenderOnly()
				return nil
			},
		},
	}
}

func (m Model) debugControlItems() []controlItem {
	status := gen.FormatDebugStatus(m.debug)
	if status == "" {
		status = "debug unavailable"
	}
	barValue := "—"
	if m.debug.Bar > 0 {
		barValue = fmt.Sprintf("%d", m.debug.Bar)
	}
	sectionValue := m.debug.Section
	if sectionValue == "" {
		sectionValue = "—"
	}
	chordValue := m.debug.Chord
	if chordValue == "" {
		chordValue = "—"
	}
	presetValue := m.debug.Preset
	if presetValue == "" {
		presetValue = "—"
	}
	return []controlItem{
		{
			Title: "debug overlay",
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
		{
			Title:    "current state",
			Value:    status,
			Hint:     "live",
			Disabled: true,
		},
		{
			Title:    "bar",
			Value:    barValue,
			Hint:     "live",
			Disabled: true,
		},
		{
			Title:    "section",
			Value:    sectionValue,
			Hint:     "live",
			Disabled: true,
		},
		{
			Title:    "chord",
			Value:    chordValue,
			Hint:     "live",
			Disabled: true,
		},
		{
			Title:    "preset",
			Value:    presetValue,
			Hint:     "live",
			Disabled: true,
		},
	}
}

func currentTabItems(m Model) []controlItem {
	return m.controlItems()
}

func controlsPanel(m Model, w, h int, theme ColorTheme) string {
	bodyW := maxInt(42, minInt(w-6, 100))
	bodyH := maxInt(16, minInt(h-2, 24))
	sidebarW := 13
	rightW := bodyW - sidebarW - 7
	sections := []controlTab{
		controlTabNow,
		controlTabLook,
		controlTabMusic,
		controlTabSeeds,
		controlTabLibrary,
		controlTabExport,
		controlTabAudio,
		controlTabDebug,
	}
	sidebarLines := make([]string, 0, len(sections))
	for _, section := range sections {
		sidebarLines = append(sidebarLines, renderControlSection(theme, m.controlTab == section, section.label()))
	}
	items := currentTabItems(m)
	lines := make([]string, 0, len(items)+4)
	lines = append(lines,
		lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render(strings.ToUpper(m.controlTab.label())),
		lipgloss.NewStyle().Faint(true).Render(controlCenterSummary(m)),
		"",
	)
	for i, item := range items {
		lines = append(lines, renderControlItem(theme, i == m.controlRow, item, rightW))
	}
	sidebar := lipgloss.NewStyle().Width(sidebarW).Render(strings.Join(sidebarLines, "\n"))
	content := lipgloss.NewStyle().Width(rightW).Render(strings.Join(lines, "\n"))
	main := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)
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
				"",
				main,
				"",
				lipgloss.NewStyle().Faint(true).Render("[tab] next section  [↑↓] browse  [←→] adjust  [enter] apply  [m] close"),
			),
		)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, panel)
}

func renderControlSection(theme ColorTheme, active bool, label string) string {
	cursor := " "
	style := lipgloss.NewStyle().Faint(true)
	if active {
		cursor = "›"
		style = lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true)
	}
	return style.Render(cursor + " " + strings.ToUpper(label))
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

func controlCenterSummary(m Model) string {
	parts := []string{m.algo, fmt.Sprintf("seed %d", m.seed)}
	if m.paused {
		parts = append(parts, "paused")
	} else {
		parts = append(parts, "playing")
	}
	if m.debug.Preset != "" {
		parts = append(parts, m.debug.Preset)
	}
	return strings.Join(parts, " · ")
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

func skipLabel(active bool) string {
	if active {
		return "available"
	}
	return "off"
}

func chromeModeLabel(reduced bool) string {
	if reduced {
		return "zen"
	}
	return "standard"
}

func compareSeedLabel(a, b *seedBookmark) string {
	switch {
	case a != nil && b != nil:
		return fmt.Sprintf("%d ↔ %d", a.Seed, b.Seed)
	case a != nil:
		return fmt.Sprintf("A %d", a.Seed)
	case b != nil:
		return fmt.Sprintf("B %d", b.Seed)
	default:
		return "empty"
	}
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

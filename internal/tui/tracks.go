package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const trackFooterHint = "[t] close   [←→] style   [↑↓] browse   [enter] play"

// trackPanel draws the full-screen track library: a header row, the style
// filter bar, the browse list and the detail pane side by side, and the key
// hints. It owns the whole terminal now, so every row is trimmed as plain
// text before styling and the assembled block is clipped vertically —
// lipgloss's Width()/Height() wrap and pad but never truncate.
func trackPanel(m Model, w, h int, theme ColorTheme) string {
	outer := lipgloss.NewStyle().
		Width(w).
		Height(h).
		Padding(1, 2).
		Foreground(theme.BarFg)
	innerW := maxInt(1, w-4)
	innerH := maxInt(1, h-2)
	header := renderTrackHeader(m, innerW, theme)
	footer := lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).
		Render(trimToWidth(trackFooterHint, innerW))
	if len(m.tracks) == 0 {
		empty := []string{
			header,
			"",
			lipgloss.NewStyle().Faint(true).Render(trimToWidth("No authored tracks found.", innerW)),
			lipgloss.NewStyle().Faint(true).Render(trimToWidth("Add .tm files under tracks/<style>/ to browse them here.", innerW)),
			"",
			footer,
		}
		return outer.Render(clipLines(strings.Join(empty, "\n"), innerH))
	}

	styleBar := renderTrackStyleBar(m, theme, innerW)
	// header + blank + filters + blank + body + blank + footer
	bodyH := maxInt(1, innerH-6)
	// leftW + "  " + divider + "  " + rightW must fit the padded width, and
	// the detail pane needs a floor of its own or the form map collapses.
	leftW := clampInt(int(float64(w)*0.40), 24, 46)
	leftW = clampInt(leftW, 1, maxInt(1, innerW-5-12))
	rightW := maxInt(1, innerW-leftW-5)
	listPane := renderTrackListPane(m, leftW, bodyH, theme)
	detailPane := renderTrackDetailPane(m, rightW, bodyH, theme)
	divider := renderTrackDivider(bodyH, theme)

	body := lipgloss.JoinHorizontal(lipgloss.Top, listPane, "  ", divider, "  ", detailPane)
	content := lipgloss.JoinVertical(lipgloss.Left, header, "", styleBar, "", clipLines(body, bodyH), "", footer)
	return outer.Render(clipLines(content, innerH))
}

// renderTrackHeader draws "TRACK LIBRARY" with the library census pushed to
// the right edge. The census sheds detail rather than overflowing a narrow
// terminal.
func renderTrackHeader(m Model, w int, theme ColorTheme) string {
	titleText := trimToWidth("TRACK LIBRARY", w)
	title := lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render(titleText)
	countNoun := "authored tracks"
	if len(m.tracks) == 1 {
		countNoun = "authored track"
	}
	available := maxInt(0, w-lipgloss.Width(titleText)-1)
	rightText := ""
	for _, candidate := range []string{
		fmt.Sprintf("%d %s · one performer", len(m.tracks), countNoun),
		fmt.Sprintf("%d %s", len(m.tracks), countNoun),
		"",
	} {
		rightText = candidate
		if lipgloss.Width(candidate) <= available {
			break
		}
	}
	if rightText == "" {
		return title
	}
	right := lipgloss.NewStyle().Faint(true).Render(rightText)
	headPad := maxInt(1, w-lipgloss.Width(titleText)-lipgloss.Width(rightText))
	return title + spaces(headPad) + right
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

// renderTrackStyleBar draws the style filter row. The active filter carries a
// "▌" bar and bold highlight; the rest stay faint. Chunks are measured as
// plain text and dropped whole once the row is full — the old code trimmed the
// already-styled join, which slices through ANSI escapes.
func renderTrackStyleBar(m Model, theme ColorTheme, width int) string {
	styles := m.trackStyleOptions()
	active := m.currentTrackStyle()
	parts := make([]string, 0, len(styles))
	used := 0
	for _, style := range styles {
		count := 0
		for _, entry := range m.tracks {
			if style == "all" || strings.EqualFold(entry.Style, style) {
				count++
			}
		}
		text := fmt.Sprintf("%s %s %d", trackStyleGlyph(style), style, count)
		isActive := strings.EqualFold(style, active)
		if isActive {
			text = "▌" + text
		}
		gap := 0
		if len(parts) > 0 {
			gap = 2
		}
		if used+gap+lipgloss.Width(text) > width {
			break
		}
		used += gap + lipgloss.Width(text)
		if isActive {
			parts = append(parts, lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render(text))
			continue
		}
		parts = append(parts, lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render(text))
	}
	return strings.Join(parts, "  ")
}

func renderTrackListPane(m Model, w, h int, theme ColorTheme) string {
	indices := m.filteredTrackIndices()
	style := lipgloss.NewStyle().Width(w).Height(h)
	if len(indices) == 0 {
		return style.Render(trimToWidth("No tracks in this style filter.", w))
	}
	lines := []string{
		lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render(trimToWidth("TRACKS", w)),
		lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render(trimToWidth(strings.ToUpper(m.currentTrackStyle())+" filter", w)),
		"",
	}
	maxRows := maxInt(1, (h-3)/2)
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
		title := firstNonEmpty(entry.Title, entry.ID)
		meta := trackCompactMeta(entry)
		prefix := "  "
		if idx == m.trackIdx {
			prefix = "▌ "
		} else if entry.ID == m.activeTrackID {
			prefix = "• "
		}
		titleGlyphs := trackStyleGlyph(entry.Style) + trackSubstyleGlyph(entry.Substyle)
		badge := renderEngineBadge(entry.Engine, theme, idx == m.trackIdx)
		head := prefix + titleGlyphs + " "
		// Reserve 6 columns for the badge slot ("[AI] " / "[SF2] ") so
		// the title block doesn't wobble when scrolling.
		titleW := maxInt(1, w-lipgloss.Width(head)-6)
		titleLine := lipgloss.NewStyle().Bold(idx == m.trackIdx).Render(head + trimToWidth(title, titleW))
		if badge != "" && lipgloss.Width(titleLine)+1+lipgloss.Width(badge) <= w {
			titleLine = titleLine + " " + badge
		}
		metaStyle := lipgloss.NewStyle().Faint(true)
		if idx == m.trackIdx {
			// Dimmed BarHi keeps the selected row's meta legible without
			// competing with its title, and stays theme-relative.
			metaStyle = lipgloss.NewStyle().Foreground(blendColor(theme.BarHi, lipgloss.Color("#000000"), 0.27))
		}
		block := lipgloss.JoinVertical(lipgloss.Left,
			titleLine,
			metaStyle.Render(trimToWidth("  "+meta, maxInt(1, w))),
		)
		if idx == m.trackIdx {
			block = lipgloss.NewStyle().Foreground(theme.BarHi).Render(block)
		} else if entry.ID == m.activeTrackID {
			block = lipgloss.NewStyle().Foreground(theme.BarFg).Render(block)
		}
		lines = append(lines, block)
	}
	return style.Render(clipLines(strings.TrimRight(strings.Join(lines, "\n"), "\n"), h))
}

// renderTrackDetailPane describes the highlighted track: what it is, how long
// it runs, the shape of its form, who plays it, and the textures underneath.
func renderTrackDetailPane(m Model, w, h int, theme ColorTheme) string {
	style := lipgloss.NewStyle().Width(w).Height(h)
	if len(m.tracks) == 0 || m.trackIdx < 0 || m.trackIdx >= len(m.tracks) {
		return style.Render("")
	}
	entry := m.tracks[m.trackIdx]
	title := firstNonEmpty(entry.Title, entry.ID)
	head := trackStyleGlyph(entry.Style) + trackSubstyleGlyph(entry.Substyle) + " "
	badge := renderEngineBadge(entry.Engine, theme, true)
	badgeSlot := 0
	if badge != "" {
		badgeSlot = lipgloss.Width(badge) + 1
	}
	// Trim the title while it is still plain text, then style it; the badge
	// arrives already styled and is only appended when it fits.
	titleLine := lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).
		Render(head + trimToWidth(title, maxInt(1, w-lipgloss.Width(head)-badgeSlot)))
	if badge != "" && lipgloss.Width(titleLine)+1+lipgloss.Width(badge) <= w {
		titleLine += " " + badge
	}
	lines := []string{titleLine}

	meta := make([]string, 0, 6)
	for _, part := range []string{entry.Style, entry.Substyle, entry.Key} {
		if part != "" {
			meta = append(meta, part)
		}
	}
	if entry.Tempo != "" {
		meta = append(meta, entry.Tempo+" bpm")
	}
	if entry.ListenMode != "" {
		meta = append(meta, entry.ListenMode)
	}
	var total time.Duration
	for _, s := range entry.Structure {
		total += s.Duration
	}
	if total > 0 {
		meta = append(meta, formMapDuration(total))
	}
	if len(meta) > 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.BarFg).Render(trimToWidth(strings.Join(meta, " · "), w)))
	}
	if desc := strings.TrimSpace(entry.Description); desc != "" {
		lines = append(lines, lipgloss.NewStyle().Faint(true).Render(trimToWidth(desc, w)))
	}

	// Reserve the rows the tail still needs so the form map yields to the
	// ensemble/textures/tags block instead of pushing it off the pane.
	tail := 0
	if len(entry.Ensemble) > 0 {
		tail++
	}
	if len(entry.Textures) > 0 {
		tail++
	}
	if len(entry.Tags) > 0 {
		tail++
	}
	if entry.ID == m.activeTrackID {
		tail += 2
	}
	if len(entry.Structure) > 0 {
		lines = append(lines, "", lipgloss.NewStyle().Faint(true).Render(trimToWidth("FORM", w)))
		budget := maxInt(1, h-len(lines)-tail)
		lines = append(lines, renderTrackFormMap(m, entry, w, budget, theme)...)
	}
	if len(entry.Ensemble) > 0 {
		lines = append(lines, lipgloss.NewStyle().Faint(true).Render(trimToWidth("ensemble  "+strings.Join(entry.Ensemble, " · "), w)))
	}
	if len(entry.Textures) > 0 {
		lines = append(lines, lipgloss.NewStyle().Faint(true).Render(trimToWidth("textures  "+strings.Join(entry.Textures, " · "), w)))
	}
	if len(entry.Tags) > 0 {
		lines = append(lines, renderTrackTags(entry.Tags, theme, w))
	}
	if entry.ID == m.activeTrackID {
		lines = append(lines, "", lipgloss.NewStyle().Foreground(theme.BarHi).Render(trimToWidth("● currently loaded", w)))
	}
	return style.Render(clipLines(strings.Join(lines, "\n"), h))
}

// renderTrackFormMap draws one row per section: label, a duration-proportional
// ▰ bar, the section length, and harmony/events meta. The section currently
// playing (only meaningful for the loaded track) is highlighted.
func renderTrackFormMap(m Model, entry TrackNavEntry, w, budget int, theme ColorTheme) []string {
	var longest time.Duration
	labelW, durW := 0, 0
	for _, s := range entry.Structure {
		if s.Duration > longest {
			longest = s.Duration
		}
		if n := lipgloss.Width(firstNonEmpty(s.Label, s.ID)); n > labelW {
			labelW = n
		}
		if s.Duration > 0 {
			if n := lipgloss.Width(formMapDuration(s.Duration)) + 1; n > durW {
				durW = n
			}
		}
	}
	labelW = clampInt(labelW, 4, 14)
	// Shrink the label column before the bar so label + gap + at least one
	// bar cell + the duration always fit the pane.
	labelW = clampInt(labelW, 1, maxInt(1, w-durW-3))
	barMax := clampInt(w-labelW-2-durW, 1, 24)
	isActive := entry.ID == m.activeTrackID
	current, hasCurrent := currentTrackStructureSection(entry, m.debug.Section)
	lines := make([]string, 0, len(entry.Structure))
	for i, s := range entry.Structure {
		if i >= budget {
			lines = append(lines, lipgloss.NewStyle().Faint(true).Render("…"))
			break
		}
		label := firstNonEmpty(s.Label, s.ID)
		cells := 1
		if longest > 0 && s.Duration > 0 {
			cells = maxInt(1, int(float64(barMax)*float64(s.Duration)/float64(longest)))
		}
		cells = minInt(cells, barMax)
		labelStyle := lipgloss.NewStyle().Foreground(theme.BarFg)
		if isActive && hasCurrent && sectionsMatch(s, current) {
			labelStyle = lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true)
		}
		dur := ""
		if s.Duration > 0 {
			dur = padLeft(formMapDuration(s.Duration), durW)
		}
		line := labelStyle.Render(padRight(trimToWidth(label, labelW), labelW)) + "  " +
			labelStyle.Render(strings.Repeat("▰", cells)) + spaces(barMax-cells) +
			lipgloss.NewStyle().Faint(true).Render(dur)
		if meta := formMapSectionMeta(s); meta != "" {
			room := maxInt(0, w-lipgloss.Width(line)-2)
			if room > 0 {
				line += "  " + lipgloss.NewStyle().Faint(true).Render(trimToWidth(meta, room))
			}
		}
		lines = append(lines, line)
	}
	return lines
}

func formMapSectionMeta(s TrackNavSection) string {
	if harmony := compactHarmony(s.Harmony); harmony != "" {
		return harmony
	}
	if len(s.Events) > 0 {
		return strings.Join(s.Events, " · ")
	}
	return strings.Join(s.RoleNames, " · ")
}

// formMapDuration renders M:SS without the zero-padded minute of
// shortDuration, matching the mock ("0:30", "6:00").
func formMapDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Round(time.Second).Seconds())
	return fmt.Sprintf("%d:%02d", total/60, total%60)
}

func padRight(s string, w int) string {
	if n := lipgloss.Width(s); n < w {
		return s + spaces(w-n)
	}
	return s
}

func padLeft(s string, w int) string {
	if n := lipgloss.Width(s); n < w {
		return spaces(w-n) + s
	}
	return s
}

func trackCompactMeta(entry TrackNavEntry) string {
	parts := make([]string, 0, 6)
	if entry.Substyle != "" {
		parts = append(parts, entry.Substyle)
	}
	if count := maxInt(entry.SectionCount, len(entry.Sections)); count > 0 {
		parts = append(parts, fmt.Sprintf("%02d sec", count))
	}
	if entry.EventCount > 0 {
		parts = append(parts, fmt.Sprintf("%02d evt", entry.EventCount))
	}
	if len(entry.Ensemble) > 0 {
		parts = append(parts, strings.Join(entry.Ensemble[:minInt(3, len(entry.Ensemble))], "/"))
	}
	if entry.Complexity != "" {
		parts = append(parts, entry.Complexity)
	}
	if entry.Tempo != "" {
		parts = append(parts, entry.Tempo+" bpm")
	}
	if len(parts) == 0 {
		return entry.ID
	}
	return strings.Join(parts, " · ")
}

// renderEngineBadge returns the small "[AI]" / "[SF2]" tag shown to the right
// of a track title, so the user can see at a glance which engine will play
// a given .tm. Highlighted (the currently-selected row) gets the BarHi color;
// the others stay faint. An empty engine string is treated as "sf2" — that
// matches the default for .tm files that omit render_engine entirely.
func renderEngineBadge(engine string, theme ColorTheme, highlighted bool) string {
	label := strings.ToUpper(strings.TrimSpace(engine))
	if label == "" || label == "SF2" {
		label = "SF2"
	}
	if label == "ACESTEP" || label == "ACE-STEP" || label == "ACE" {
		label = "AI"
	}
	text := "[" + label + "]"
	if highlighted {
		return lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render(text)
	}
	return lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render(text)
}

func trackStyleGlyph(style string) string {
	switch strings.ToLower(strings.TrimSpace(style)) {
	case "ambient":
		return "◌"
	case "bells":
		return "✶"
	case "classical":
		return "◇"
	case "drone":
		return "▤"
	case "jazz":
		return "♬"
	case "lofi":
		return "◒"
	case "lullaby":
		return "☾"
	case "phase":
		return "∿"
	default:
		return "•"
	}
}

func trackSubstyleGlyph(substyle string) string {
	lower := strings.ToLower(strings.TrimSpace(substyle))
	switch {
	case strings.Contains(lower, "rhodes"):
		return "◒"
	case strings.Contains(lower, "vibes"):
		return "✶"
	case strings.Contains(lower, "guitar"):
		return "⌁"
	case strings.Contains(lower, "organ"):
		return "▤"
	case strings.Contains(lower, "trio"):
		return "◇"
	case strings.Contains(lower, "choir"):
		return "☾"
	case strings.Contains(lower, "glass"):
		return "⋄"
	case strings.Contains(lower, "station"):
		return "◌"
	case strings.Contains(lower, "paper"):
		return "◠"
	case strings.Contains(lower, "static"):
		return "▥"
	default:
		return "·"
	}
}

func compactHarmony(harmony string) string {
	harmony = strings.TrimSpace(strings.ReplaceAll(harmony, "\n", " "))
	if harmony == "" {
		return ""
	}
	return trimToWidth(strings.ReplaceAll(harmony, " | ", " · "), 42)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// renderTrackTags lays out "#tag" chips, dropping whole chips once the row is
// full — trimming the styled join would slice through an ANSI escape.
func renderTrackTags(tags []string, theme ColorTheme, width int) string {
	parts := make([]string, 0, len(tags))
	used := 0
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		text := "#" + tag
		gap := 0
		if len(parts) > 0 {
			gap = 2
		}
		if used+gap+lipgloss.Width(text) > width {
			break
		}
		used += gap + lipgloss.Width(text)
		parts = append(parts, lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render(text))
	}
	return strings.Join(parts, "  ")
}

func renderTrackDivider(height int, theme ColorTheme) string {
	lines := make([]string, maxInt(1, height))
	for i := range lines {
		lines[i] = lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render("│")
	}
	return strings.Join(lines, "\n")
}

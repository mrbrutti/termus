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

// chunkGap is the blank run between two chips in a packed row, and chunkMarker
// is what a "there is more this way" arrow costs including its gap.
const (
	chunkGap    = 2
	chunkMarker = 1 + chunkGap
)

// packChunks picks the run of chips that fits in width, keeping the focused
// chip visible: a filter row that drops chunks off the tail would hide the
// active filter on a narrow terminal, so ←/→ would appear to do nothing. It
// returns the half-open window plus whether chips are hidden on either side.
// Chunks are measured as plain text; the caller styles the survivors.
// A negative focus packs from the head (no chip has to stay visible).
func packChunks(chunks []string, width, focus int) (start, end int, hiddenBefore, hiddenAfter bool) {
	if len(chunks) == 0 || width <= 0 {
		return 0, 0, false, false
	}
	if focus < 0 || focus >= len(chunks) {
		focus = 0
	}
	total := 0
	for i, chunk := range chunks {
		if i > 0 {
			total += chunkGap
		}
		total += lipgloss.Width(chunk)
	}
	if total <= width {
		return 0, len(chunks), false, false
	}
	// Truncation is certain, so pay for both markers up front rather than
	// re-deciding the budget every time the window grows.
	budget := width - 2*chunkMarker
	start, end = focus, focus+1
	used := lipgloss.Width(chunks[focus])
	for used <= budget {
		grew := false
		// Extend forward first so the row keeps its reading order.
		if end < len(chunks) {
			if cost := chunkGap + lipgloss.Width(chunks[end]); used+cost <= budget {
				used += cost
				end++
				grew = true
			}
		}
		if start > 0 {
			if cost := chunkGap + lipgloss.Width(chunks[start-1]); used+cost <= budget {
				used += cost
				start--
				grew = true
			}
		}
		if !grew {
			break
		}
	}
	hiddenBefore, hiddenAfter = start > 0, end < len(chunks)
	markers := 0
	if hiddenBefore {
		markers += chunkMarker
	}
	if hiddenAfter {
		markers += chunkMarker
	}
	if used+markers > width {
		// The focused chip barely fits on its own; show it bare rather than
		// spending its columns on arrows.
		return start, end, false, false
	}
	return start, end, hiddenBefore, hiddenAfter
}

// renderTrackStyleBar draws the style filter row. The active filter carries a
// "▌" bar and bold highlight; the rest stay faint. Chunks are measured as
// plain text and windowed around the active filter — the old code trimmed the
// already-styled join, which slices through ANSI escapes.
func renderTrackStyleBar(m Model, theme ColorTheme, width int) string {
	styles := m.trackStyleOptions()
	active := m.currentTrackStyle()
	chunks := make([]string, len(styles))
	activeIdx := 0
	for i, style := range styles {
		count := 0
		for _, entry := range m.tracks {
			if style == "all" || strings.EqualFold(entry.Style, style) {
				count++
			}
		}
		if style == "all" {
			// "all" carries no genre glyph in the mock — just the bare label
			// and count.
			chunks[i] = fmt.Sprintf("%s %d", style, count)
		} else {
			chunks[i] = fmt.Sprintf("%s %s %d", trackStyleGlyph(style), style, count)
		}
		if strings.EqualFold(style, active) {
			activeIdx = i
			chunks[i] = "▌" + chunks[i]
		}
	}
	start, end, before, after := packChunks(chunks, width, activeIdx)
	marker := lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true)
	parts := make([]string, 0, len(chunks)+2)
	if before {
		parts = append(parts, marker.Render("‹"))
	}
	for i := start; i < end; i++ {
		text := trimToWidth(chunks[i], width)
		if i == activeIdx {
			parts = append(parts, lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render(text))
			continue
		}
		parts = append(parts, lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render(text))
	}
	if after {
		parts = append(parts, marker.Render("›"))
	}
	return strings.Join(parts, spaces(chunkGap))
}

// titleWithBadge trims a plain title to the room the badge slot leaves, styles
// it, and appends the already-styled badge only when it still fits. slot is the
// reserved badge width: the list pane reserves a fixed slot so rows don't
// wobble between "[AI]" and "[SF2]" while scrolling.
func titleWithBadge(head, title, badge string, w, slot int, textStyle lipgloss.Style) string {
	line := textStyle.Render(head + trimToWidth(title, maxInt(1, w-lipgloss.Width(head)-slot)))
	if badge != "" && lipgloss.Width(line)+1+lipgloss.Width(badge) <= w {
		line += " " + badge
	}
	return line
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
		// Reserve 6 columns for the badge slot ("[AI] " / "[SF2] ") so
		// the title block doesn't wobble when scrolling.
		titleLine := titleWithBadge(prefix+titleGlyphs+" ", title, badge, w, 6,
			lipgloss.NewStyle().Bold(idx == m.trackIdx))
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
	badge := renderEngineBadge(entry.Engine, theme, true)
	badgeSlot := 0
	if badge != "" {
		badgeSlot = lipgloss.Width(badge) + 1
	}
	titleLine := titleWithBadge(trackStyleGlyph(entry.Style)+trackSubstyleGlyph(entry.Substyle)+" ",
		title, badge, w, badgeSlot, lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true))
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
	if total <= 0 {
		total = entry.TotalDuration // ACE-Step: no sections, authored total is the real length
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
		// Budget the map before committing to its header: the blank + "FORM"
		// rows cost two lines, and a header over a lone "…" says nothing while
		// still pushing the tail off a short pane. A truncated map needs two
		// rows to show even one real section, since "…" takes one of them.
		budget := h - len(lines) - tail - 2
		if budget >= minInt(2, len(entry.Structure)) {
			lines = append(lines, "", lipgloss.NewStyle().Faint(true).Render(trimToWidth("FORM", w)))
			lines = append(lines, renderTrackFormMap(m, entry, w, budget, theme)...)
		}
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
	// Two sections can carry the same label ("a", "a"), and the playhead is
	// only ever in one of them: highlight the first match and no other.
	highlighted := false
	// The "…" row costs a line of its own, so a truncated form shows one
	// section fewer. Spending budget+1 rows here would push the tail
	// ("● currently loaded") off the bottom of a short pane.
	shown := len(entry.Structure)
	truncated := false
	if shown > budget {
		shown = maxInt(0, budget-1)
		truncated = true
	}
	lines := make([]string, 0, minInt(budget, len(entry.Structure)+1))
	for i, s := range entry.Structure {
		if i >= shown {
			break
		}
		label := firstNonEmpty(s.Label, s.ID)
		cells := 1
		if longest > 0 && s.Duration > 0 {
			cells = maxInt(1, int(float64(barMax)*float64(s.Duration)/float64(longest)))
		}
		cells = minInt(cells, barMax)
		labelStyle := lipgloss.NewStyle().Foreground(theme.BarFg)
		if !highlighted && isActive && hasCurrent && sectionsMatch(s, current) {
			labelStyle = lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true)
			highlighted = true
		}
		// An undated section still pays for the duration column so the
		// harmony that follows stays in line with its neighbours.
		dur := spaces(durW)
		if s.Duration > 0 {
			dur = lipgloss.NewStyle().Faint(true).Render(padLeft(formMapDuration(s.Duration), durW))
		}
		line := labelStyle.Render(padRight(trimToWidth(label, labelW), labelW)) + "  " +
			labelStyle.Render(strings.Repeat("▰", cells)) + spaces(barMax-cells) + dur
		if meta := formMapSectionMeta(s); meta != "" {
			room := maxInt(0, w-lipgloss.Width(line)-2)
			if room > 0 {
				line += "  " + lipgloss.NewStyle().Faint(true).Render(trimToWidth(meta, room))
			}
		}
		lines = append(lines, line)
	}
	if truncated {
		lines = append(lines, lipgloss.NewStyle().Faint(true).Render("…"))
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
// full — trimming the styled join would slice through an ANSI escape. No tag
// has to stay visible, so the row simply packs from the head.
func renderTrackTags(tags []string, theme ColorTheme, width int) string {
	chunks := make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag = strings.TrimSpace(tag); tag != "" {
			chunks = append(chunks, "#"+tag)
		}
	}
	start, end, _, _ := packChunks(chunks, width, -1)
	parts := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		parts = append(parts, lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render(trimToWidth(chunks[i], width)))
	}
	return strings.Join(parts, spaces(chunkGap))
}

func renderTrackDivider(height int, theme ColorTheme) string {
	lines := make([]string, maxInt(1, height))
	for i := range lines {
		lines[i] = lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render("│")
	}
	return strings.Join(lines, "\n")
}

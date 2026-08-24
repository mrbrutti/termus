package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// wordmarkLetters are the 5-row block letterforms from the design mock.
var wordmarkLetters = map[rune][5]string{
	'T': {"█████", "  █  ", "  █  ", "  █  ", "  █  "},
	'E': {"████ ", "█    ", "███  ", "█    ", "████ "},
	'R': {"████ ", "█   █", "████ ", "█  █ ", "█   █"},
	'M': {"█   █", "██ ██", "█ █ █", "█   █", "█   █"},
	'U': {"█   █", "█   █", "█   █", "█   █", " ███ "},
	'S': {" ████", "█    ", " ███ ", "    █", "████ "},
}

// renderWordmark draws text in block letters, colored per row with the
// theme's gradient. The color is sampled per cell, not per row: a vertical
// gradient returns one color for the whole row (so the runs below collapse to
// a single styled span), while a horizontal theme like rainbow spans the
// wordmark left-to-right instead of rendering flat.
func renderWordmark(text string, theme ColorTheme) string {
	rows := make([]string, 5)
	plain := make([]string, 5)
	width := 0
	for r := 0; r < 5; r++ {
		var b strings.Builder
		first := true
		for _, ch := range text {
			letter, ok := wordmarkLetters[ch]
			if !ok {
				continue
			}
			if !first {
				b.WriteString("  ")
			}
			first = false
			b.WriteString(letter[r])
		}
		plain[r] = b.String()
		width = maxInt(width, lipgloss.Width(plain[r]))
	}
	for r := 0; r < 5; r++ {
		rows[r] = colorizeRow(plain[r], r, width, theme)
	}
	return strings.Join(rows, "\n")
}

// colorizeRow paints one wordmark row, coalescing neighbouring cells that
// resolve to the same color into a single styled span so a vertical-gradient
// theme costs one Render call per row.
func colorizeRow(row string, r, width int, theme ColorTheme) string {
	if theme.ColorAt == nil {
		return lipgloss.NewStyle().Foreground(theme.BarHi).Render(row)
	}
	var b strings.Builder
	var run strings.Builder
	runColor := lipgloss.Color("")
	flush := func() {
		if run.Len() > 0 {
			b.WriteString(lipgloss.NewStyle().Foreground(runColor).Render(run.String()))
			run.Reset()
		}
	}
	for x, ch := range []rune(row) {
		if ch == ' ' {
			// Blank cells carry no ink; leaving them unstyled keeps the run
			// count (and the escape sequences) down.
			flush()
			b.WriteRune(ch)
			continue
		}
		color := theme.ColorAt(x, r, width, 5)
		if run.Len() > 0 && color != runColor {
			flush()
		}
		runColor = color
		run.WriteRune(ch)
	}
	flush()
	return b.String()
}

// stationDialLines renders the splash station chooser: the selected station
// bold BarHi, the remaining stations faint, wrapped to maxW. A loaded
// playlist owns the algorithm schedule, so the dial collapses to the station
// that is actually playing — there is nothing to browse.
func stationDialLines(m Model, theme ColorTheme, maxW int) []string {
	if len(m.genres) == 0 || maxW <= 0 {
		return nil
	}
	if m.playlist != nil {
		cur := m.genres[clampInt(m.genreIdx, 0, len(m.genres)-1)]
		return []string{lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).
			Render(trimToWidth(stationGlyph(cur.Name)+" "+strings.ToUpper(cur.Label()), maxW))}
	}
	idx := clampInt(m.genreIdx, 0, len(m.genres)-1)
	cur := m.genres[idx]
	lines := []string{lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).
		Render(trimToWidth(stationGlyph(cur.Name)+" "+strings.ToUpper(cur.Label()), maxW))}
	faint := lipgloss.NewStyle().Faint(true)
	row := ""
	flush := func() {
		if row != "" {
			lines = append(lines, faint.Render(trimToWidth(row, maxW)))
			row = ""
		}
	}
	for i, g := range m.genres {
		if i == idx {
			continue
		}
		s := stationGlyph(g.Name) + " " + strings.ToLower(g.Label())
		switch {
		case row == "":
			row = s
		case lipgloss.Width(row)+3+lipgloss.Width(s) > maxW:
			flush()
			row = s
		default:
			row += "   " + s
		}
	}
	flush()
	return lines
}

// splashProgressRow is the loading status line: percent in BarHi, then the
// loader title and its current detail, faint. Trimmed while still plain text
// — trimToWidth cannot see through ANSI sequences.
func splashProgressRow(m Model, theme ColorTheme, progress float64, maxW int) string {
	pct := fmt.Sprintf("%d%%", int(progress*100))
	title := m.startupTitle
	if title == "" {
		title = "Loading termus"
	}
	rest := " · " + title
	if m.startupDetail != "" {
		rest += " · " + m.startupDetail
	}
	rest = trimToWidth(rest, maxW-lipgloss.Width(pct))
	return lipgloss.NewStyle().Foreground(theme.BarHi).Render(pct) +
		lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render(rest)
}

// splashFooter names the things the splash can do, shortened before it would
// ever wrap. The dial hint is dropped when a playlist is driving playback:
// the arrows are inert there, and advertising them would be a lie.
func splashFooter(theme ColorTheme, maxW int, dial bool) string {
	text := "[enter] begin · [t] authored tracks"
	short := "[enter] begin"
	if dial {
		text = "[←→] choose a station · " + text
		short = "[←→] station · " + short
	}
	if lipgloss.Width(text) > maxW {
		text = short
	}
	return lipgloss.NewStyle().Faint(true).Render(trimToWidth(text, maxW))
}

// splashScreen is the merged splash + startup-loading screen: wordmark,
// tagline, station dial, braille loading bar, and (while loading) the
// progress row and any composing-context block.
//
// The stack is taller than the app's minimum 40×10 terminal, so it degrades
// in a fixed order as h shrinks: blank rows first, then the braille bar (only
// when nothing is loading), then the extra station rows, then the tagline,
// then the composing context, and finally the bar collapses to a single row.
// The wordmark, the current station, the loading progress and the footer are
// the last survivors; if even those do not fit the content is clipped.
func splashScreen(m Model, w, h int, theme ColorTheme, now time.Time) string {
	if w < 1 || h < 1 {
		return ""
	}
	loading := m.startupLoading
	barW := clampInt(w-10, 26, 64)
	barW = clampInt(barW, 1, w)
	progress := 1.0
	if loading {
		progress = clamp01(m.startupPercent)
	}
	phase := float64(now.UnixNano()) / float64(time.Second)

	wordmark := renderWordmark("TERMUS", theme)
	if lipgloss.Width(wordmark) > w {
		wordmark = lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).
			Render(trimToWidth("TERMUS", w))
	}
	tagline := lipgloss.NewStyle().Faint(true).Render(trimToWidth("a terminal music instrument", w))
	dial := stationDialLines(m, theme, w)
	// The bar is drawn once, for the stage that wins: measuring uses a
	// same-height placeholder so a rejected stage costs no rendering.
	// renderStartupBrailleBar also terminates every cell row with a newline,
	// which would spend one more row than the bar actually draws.
	measureBar := func(rows int) string { return strings.Join(make([]string, rows), "\n") }
	drawBar := func(rows int) string {
		return strings.TrimRight(renderStartupBrailleBar(barW, rows, progress, phase, theme), "\n")
	}
	progressRow := ""
	ctxBlock := ""
	if loading {
		progressRow = splashProgressRow(m, theme, progress, maxInt(1, w-2))
		ctxBlock = composingContextBlock(m, barW, theme)
	}
	footer := splashFooter(theme, w, m.playlist == nil)

	// Degradation ladder: stage 0 is the full stack, each later stage sheds
	// one more row group. Stop at the first stage that fits.
	const lastStage = 9
	build := func(stage int, bar func(rows int) string) []string {
		blanksAfterTagline := 2
		blankBeforeBar := true
		blankBeforeFooter := true
		showBar := true
		barRows := 3
		showDialExtras := true
		showTagline := true
		showCtx := true
		if stage >= 1 {
			blanksAfterTagline = 1
		}
		if stage >= 2 {
			blankBeforeFooter = false
		}
		if stage >= 3 {
			blankBeforeBar = false
		}
		if stage >= 4 && !loading {
			showBar = false
		}
		if stage >= 5 {
			showDialExtras = false
		}
		if stage >= 6 {
			showTagline = false
		}
		if stage >= 7 {
			showCtx = false
		}
		if stage >= 8 {
			barRows = 1
		}
		if stage >= 9 {
			blanksAfterTagline = 0
		}

		parts := []string{wordmark}
		if showTagline {
			parts = append(parts, tagline)
		}
		for i := 0; i < blanksAfterTagline; i++ {
			parts = append(parts, "")
		}
		if len(dial) > 0 {
			parts = append(parts, dial[0])
			if showDialExtras {
				parts = append(parts, dial[1:]...)
			}
		}
		if showBar {
			if blankBeforeBar {
				parts = append(parts, "")
			}
			parts = append(parts, bar(barRows))
		}
		if progressRow != "" {
			parts = append(parts, progressRow)
		}
		if showCtx && ctxBlock != "" {
			parts = append(parts, "", ctxBlock)
		}
		if blankBeforeFooter {
			parts = append(parts, "")
		}
		return append(parts, footer)
	}

	chosen := lastStage
	for stage := 0; stage < lastStage; stage++ {
		rows := 0
		for _, p := range build(stage, measureBar) {
			rows += lipgloss.Height(p)
		}
		if rows <= h {
			chosen = stage
			break
		}
	}
	parts := build(chosen, drawBar)

	content := lipgloss.JoinVertical(lipgloss.Center, parts...)
	// Final guards: the composing context can still run long on a narrow
	// terminal, and the most degraded stack can still be taller than h.
	content = lipgloss.NewStyle().MaxWidth(w).Render(content)
	content = clipLines(content, h)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content)
}

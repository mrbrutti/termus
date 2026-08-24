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
// theme's gradient (center rows brightest, edges toward the gradient edge).
func renderWordmark(text string, theme ColorTheme) string {
	rows := make([]string, 5)
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
		color := theme.BarHi
		if theme.ColorAt != nil {
			color = theme.ColorAt(0, r, 1, 5)
		}
		rows[r] = lipgloss.NewStyle().Foreground(color).Render(b.String())
	}
	return strings.Join(rows, "\n")
}

// stationDialLines renders the splash station chooser: the selected station
// bold BarHi, the remaining stations faint, wrapped to maxW.
func stationDialLines(m Model, theme ColorTheme, maxW int) []string {
	if len(m.genres) == 0 || maxW <= 0 {
		return nil
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

// splashFooter names the two things the splash can do, shortened before it
// would ever wrap.
func splashFooter(theme ColorTheme, maxW int) string {
	text := "[←→] choose a station · [enter] begin · [t] authored tracks"
	if lipgloss.Width(text) > maxW {
		text = "[←→] station · [enter] begin"
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
	// renderStartupBrailleBar terminates every cell row with a newline, which
	// would spend one more row than the bar actually draws.
	bars := map[int]string{
		3: strings.TrimRight(renderStartupBrailleBar(barW, 3, progress, phase, theme), "\n"),
		1: strings.TrimRight(renderStartupBrailleBar(barW, 1, progress, phase, theme), "\n"),
	}
	progressRow := ""
	ctxBlock := ""
	if loading {
		progressRow = splashProgressRow(m, theme, progress, barW)
		ctxBlock = composingContextBlock(m, barW, theme)
	}
	footer := splashFooter(theme, w)

	// Degradation ladder: stage 0 is the full stack, each later stage sheds
	// one more row group. Stop at the first stage that fits.
	const lastStage = 9
	build := func(stage int) []string {
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
		if showBar && bars[barRows] != "" {
			if blankBeforeBar {
				parts = append(parts, "")
			}
			parts = append(parts, bars[barRows])
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

	parts := build(lastStage)
	for stage := 0; stage < lastStage; stage++ {
		candidate := build(stage)
		rows := 0
		for _, p := range candidate {
			rows += lipgloss.Height(p)
		}
		if rows <= h {
			parts = candidate
			break
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Center, parts...)
	// Final guards: the composing context can still run long on a narrow
	// terminal, and the most degraded stack can still be taller than h.
	content = lipgloss.NewStyle().MaxWidth(w).Render(content)
	content = clipLines(content, h)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content)
}

package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// baseAlgoName normalizes an algorithm name for presentation lookups:
// trimmed, lower-cased, with the "-synth" variant suffix removed so those
// variants share the base algorithm's glyph and character.
func baseAlgoName(algoName string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(algoName)), "-synth")
}

// stationGlyph maps an algorithm's canonical name to its style glyph.
// trackStyleGlyph returns the neutral "•" for anything unknown, including
// the empty name, so callers without a spec can ask for stationGlyph("").
func stationGlyph(algoName string) string {
	return trackStyleGlyph(baseAlgoName(algoName))
}

// stationCharacter is the short tempo/character descriptor shown in the
// station header (presentation-only vocabulary, one entry per algorithm).
func stationCharacter(algoName string) string {
	switch baseAlgoName(algoName) {
	case "ambient":
		return "slow drift"
	case "drone":
		return "held tones"
	case "bells":
		return "bright chimes"
	case "lullaby":
		return "gentle walk"
	case "classical":
		return "chamber lines"
	case "phase":
		return "shifting patterns"
	case "lofi":
		return "dusty groove"
	case "jazz":
		return "late swing"
	default:
		return ""
	}
}

// narrationParts assembles the always-on musical narration from DebugStatus,
// in listening order: movement · episode · section · bar · chord-motion.
// Missing fields are omitted so every algorithm degrades gracefully.
func narrationParts(m Model) []string {
	parts := make([]string, 0, 5)
	d := m.debug
	if d.Movement != "" {
		parts = append(parts, "movement "+d.Movement)
	}
	if d.Episode > 0 {
		parts = append(parts, fmt.Sprintf("episode %d", d.Episode))
	}
	section := d.Section
	if label := m.currentSectionLabel(); label != "" {
		section = label
	}
	if section != "" {
		parts = append(parts, "section "+section)
	}
	if d.Bar > 0 {
		parts = append(parts, fmt.Sprintf("bar %d", d.Bar))
	}
	switch {
	case d.Chord != "" && d.NextChord != "" && d.NextChord != d.Chord:
		parts = append(parts, d.Chord+" → "+d.NextChord)
	case d.Chord != "":
		parts = append(parts, d.Chord)
	}
	return parts
}

// formRailSegments picks the form-rail source: the authored track's section
// schedule when one is playing, else the procedural episode chain from
// DebugStatus. Empty when neither exists (the rail row collapses).
func formRailSegments(m Model) ([]string, int) {
	if track, ok := m.activeTrack(); ok && len(track.Sections) > 1 {
		labels := make([]string, 0, len(track.Sections))
		for i, s := range track.Sections {
			title := strings.TrimSpace(s.Title)
			if title == "" {
				title = fmt.Sprintf("section %d", i+1)
			}
			labels = append(labels, title)
		}
		return labels, clampInt(m.sectionIdx, 0, len(labels)-1)
	}
	if len(m.debug.FormChain) > 0 {
		return m.debug.FormChain, clampInt(m.debug.FormIndex, 0, len(m.debug.FormChain)-1)
	}
	return nil, 0
}

// formRailBar renders the one-line section chain with the current-section
// marker, plus time-to-next-section and the listening mode on the right.
// Returns "" when there is no form source.
func formRailBar(m Model, w int, theme ColorTheme) string {
	segments, current := formRailSegments(m)
	if len(segments) == 0 {
		return ""
	}
	mode := m.listeningMode
	if mode == "" {
		mode = "endless"
	}
	countdown := ""
	if !m.nextSectionAt.IsZero() {
		countdown = formMapDuration(time.Until(m.nextSectionAt)) + " to next section"
	}

	// Everything is measured as PLAIN text first: styled (ANSI) strings must
	// never be trimmed. Degrade in listening priority — the full chain, then
	// a compact "current · position", then the countdown, then the mode, and
	// only as a last resort trim the section label itself.
	rightPlain := func() string {
		parts := make([]string, 0, 2)
		if countdown != "" {
			parts = append(parts, countdown)
		}
		if mode != "" {
			parts = append(parts, mode)
		}
		return strings.Join(parts, " · ")
	}
	fits := func(leftWidth int) bool {
		return leftWidth+lipgloss.Width(rightPlain())+1 <= w
	}

	const connectorPlain = " ─── "
	connectorWidth := lipgloss.Width(connectorPlain)
	displays := make([]string, len(segments))
	chainWidth := 0
	for i, label := range segments {
		display := label
		if i == current {
			display = "● " + label
		}
		displays[i] = display
		if i > 0 {
			chainWidth += connectorWidth
		}
		chainWidth += lipgloss.Width(display)
	}

	useChain := fits(chainWidth)
	compactPlain := fmt.Sprintf("● %s · %d/%d", segments[current], current+1, len(segments))
	if !useChain {
		if !fits(lipgloss.Width(compactPlain)) && countdown != "" {
			countdown = ""
		}
		if !fits(lipgloss.Width(compactPlain)) && mode != "" {
			mode = ""
		}
		if !fits(lipgloss.Width(compactPlain)) {
			compactPlain = trimToWidth(compactPlain, maxInt(0, w-lipgloss.Width(rightPlain())-1))
		}
	}

	right := ""
	if text := rightPlain(); text != "" {
		right = lipgloss.NewStyle().Faint(true).Render(text)
	}
	var left string
	if useChain {
		connector := lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render(connectorPlain)
		rendered := make([]string, 0, len(displays))
		for i, display := range displays {
			switch {
			case i == current:
				rendered = append(rendered, lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render(display))
			case i < current:
				rendered = append(rendered, lipgloss.NewStyle().Foreground(theme.BarFg).Render(display))
			default:
				rendered = append(rendered, lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render(display))
			}
		}
		left = strings.Join(rendered, connector)
	} else {
		left = lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render(compactPlain)
	}
	pad := w - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	return left + spaces(pad) + right
}

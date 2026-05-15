// Package tui contains the bubbletea model and the Braille oscilloscope
// renderer.
package tui

import (
	"strings"
)

// RenderBraille returns a w×h block of Braille glyphs visualizing the given
// mono samples in [-1, 1]. Each glyph is one terminal cell with 2×4 dots, so
// the effective resolution is (2w) × (4h) dots. Uses the default theme
// (indigo→cyan vertical gradient). For other themes, use RenderBrailleThemed.
func RenderBraille(samples []float64, w, h int) string {
	return RenderBrailleThemed(samples, w, h, DefaultTheme())
}

// RenderBrailleThemed is like RenderBraille but lets the caller pick a
// ColorTheme. See themes.go for the available presets.
func RenderBrailleThemed(samples []float64, w, h int, theme ColorTheme) string {
	return RenderBrailleWithContext(samples, w, h, RenderContext{Theme: theme})
}

// RenderBrailleWithContext is the background-aware braille renderer used by
// the TUI's visual system.
func RenderBrailleWithContext(samples []float64, w, h int, ctx RenderContext) string {
	if w < 4 || h < 1 {
		return "(too small)\n"
	}
	dotsX := 2 * w
	dotsY := 4 * h
	plot := make([][]bool, dotsY)
	for i := range plot {
		plot[i] = make([]bool, dotsX)
	}
	if len(samples) > 0 {
		for px := 0; px < dotsX; px++ {
			si := px * len(samples) / dotsX
			s := samples[si]
			if s > 1 {
				s = 1
			}
			if s < -1 {
				s = -1
			}
			py := int((1.0 - (s+1)/2) * float64(dotsY-1))
			if py < 0 {
				py = 0
			}
			if py > dotsY-1 {
				py = dotsY - 1
			}
			plot[py][px] = true
		}
	}

	var b strings.Builder
	for cy := 0; cy < h; cy++ {
		for cx := 0; cx < w; cx++ {
			r := brailleCell(plot, cx, cy)
			b.WriteString(renderCell(r, cx, cy, w, h, ctx))
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// brailleCell builds the single-rune Braille glyph for the (cx, cy) cell.
// Dot mapping (per Unicode Braille):
//
//	0x1 = (0,0)   0x8 = (1,0)
//	0x2 = (0,1)   0x10 = (1,1)
//	0x4 = (0,2)   0x20 = (1,2)
//	0x40 = (0,3)  0x80 = (1,3)
func brailleCell(plot [][]bool, cx, cy int) rune {
	const base = 0x2800
	var bits rune
	mask := [4][2]rune{
		{0x1, 0x8},
		{0x2, 0x10},
		{0x4, 0x20},
		{0x40, 0x80},
	}
	for dy := 0; dy < 4; dy++ {
		py := cy*4 + dy
		if py >= len(plot) {
			continue
		}
		for dx := 0; dx < 2; dx++ {
			px := cx*2 + dx
			if px >= len(plot[py]) {
				continue
			}
			if plot[py][px] {
				bits |= mask[dy][dx]
			}
		}
	}
	return base + bits
}

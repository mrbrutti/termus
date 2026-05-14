// Package tui contains the bubbletea model and the Braille oscilloscope
// renderer.
package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderBraille returns a w×h block of Braille glyphs visualizing the given
// mono samples in [-1, 1]. Each glyph is one terminal cell with 2×4 dots, so
// the effective resolution is (2w) × (4h) dots. Coloring uses a vertical
// gradient: deep indigo near the rails, cyan near the centerline.
func RenderBraille(samples []float64, w, h int) string {
	if w < 4 || h < 1 {
		return "(too small)\n"
	}
	dotsX := 2 * w
	dotsY := 4 * h
	// Map sample[i in 0..len(samples)) → x in 0..dotsX-1.
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
			// y=0 is top of buffer. Sample 1.0 → top, -1.0 → bottom.
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

	// Build cell grid.
	var b strings.Builder
	for cy := 0; cy < h; cy++ {
		for cx := 0; cx < w; cx++ {
			r := brailleCell(plot, cx, cy)
			// Choose color from vertical position of cell.
			color := gradientColor(cy, h)
			b.WriteString(lipgloss.NewStyle().Foreground(color).Render(string(r)))
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

// gradientColor returns the foreground color for a given vertical cell index.
// Indigo at the top/bottom rails, cyan near the middle.
func gradientColor(cy, h int) lipgloss.Color {
	mid := float64(h-1) / 2
	d := float64(cy) - mid
	// Normalize 0..1, 0 at center, 1 at rail.
	norm := 1.0
	if mid > 0 {
		norm = absf(d) / mid
	}
	// indigo (#5b4bff) at rails → cyan (#5bfaff) at center.
	r := lerp(0x5b, 0x5b, norm)
	g := lerp(0xfa, 0x4b, norm)
	b := lerp(0xff, 0xff, norm)
	return lipgloss.Color(hexColor(r, g, b))
}

func absf(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func lerp(near, far int, t float64) int {
	// t=0 → near (center), t=1 → far (rail).
	return int(float64(near) + (float64(far)-float64(near))*t)
}

func hexColor(r, g, b int) string {
	return "#" + hex2(r) + hex2(g) + hex2(b)
}

func hex2(v int) string {
	if v < 0 {
		v = 0
	}
	if v > 255 {
		v = 255
	}
	const hex = "0123456789abcdef"
	return string([]byte{hex[v>>4], hex[v&0xf]})
}

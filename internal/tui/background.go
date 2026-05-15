package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// RenderContext bundles the non-audio rendering inputs shared by all visual
// styles so foreground visuals and the animated ASCII background stay in sync.
type RenderContext struct {
	Theme      ColorTheme
	Background AnimeBackground
}

// AnimeBackground renders a faint looping city-sky scene behind the active
// visualization. The scene is ASCII-only so it still works in low-feature
// terminals.
type AnimeBackground struct {
	Enabled bool
	Frame   int
}

type backgroundCell struct {
	ch    rune
	color lipgloss.Color
	faint bool
}

func (bg AnimeBackground) cell(cx, cy, w, h int, theme ColorTheme) backgroundCell {
	if !bg.Enabled || w < 8 || h < 4 {
		return backgroundCell{ch: ' '}
	}

	skyColor := theme.ColorAt(cx, maxInt(0, cy), maxInt(1, w), maxInt(1, h))
	if moon := bg.moonCell(cx, cy, w); moon != 0 {
		return backgroundCell{ch: moon, color: theme.BarHi, faint: false}
	}
	if petal := bg.petalCell(cx, cy, w, h); petal != 0 {
		return backgroundCell{ch: petal, color: theme.BarFg, faint: true}
	}
	if cloud := bg.cloudCell(cx, cy, w); cloud != 0 {
		return backgroundCell{ch: cloud, color: skyColor, faint: true}
	}
	if skyline := bg.skylineCell(cx, cy, w, h); skyline != 0 {
		return backgroundCell{ch: skyline, color: theme.BarFg, faint: true}
	}
	if star := bg.starCell(cx, cy, w, h); star != 0 {
		return backgroundCell{ch: star, color: skyColor, faint: true}
	}
	return backgroundCell{ch: ' '}
}

func (bg AnimeBackground) moonCell(cx, cy, w int) rune {
	moonX := maxInt(4, w-10-(bg.Frame/18)%5)
	switch {
	case cy == 1 && cx == moonX:
		return '('
	case cy == 1 && cx == moonX+1:
		return '_'
	case cy == 1 && cx == moonX+2:
		return ')'
	case cy == 2 && cx == moonX:
		return '('
	case cy == 2 && cx == moonX+1:
		return '_'
	case cy == 2 && cx == moonX+2:
		return ')'
	default:
		return 0
	}
}

func (bg AnimeBackground) cloudCell(cx, cy, w int) rune {
	if cy < 1 || cy > 3 {
		return 0
	}
	for band := 0; band < 2; band++ {
		y := 1 + band*2
		offset := (bg.Frame/(6+band*2) + band*11) % maxInt(1, w+16)
		start := w - offset - 8
		end := start + 8
		if cy != y || cx < start || cx >= end {
			continue
		}
		if (cx-start)%3 == 1 {
			return '~'
		}
	}
	return 0
}

func (bg AnimeBackground) petalCell(cx, cy, w, h int) rune {
	for i := 0; i < 6; i++ {
		path := hash32(cx+i*13, cy+i*7, bg.Frame/3+i*17)
		startX := int(path%uint32(maxInt(1, w+12))) - 6
		startY := int((path / 17) % uint32(maxInt(1, h/2+1)))
		x := startX + (bg.Frame/2+i*3)%maxInt(1, w+10)
		y := startY + (bg.Frame/11+i)%maxInt(1, maxInt(1, h/2))
		if cx == x && cy == y {
			if i%2 == 0 {
				return '*'
			}
			return '.'
		}
	}
	return 0
}

func (bg AnimeBackground) skylineCell(cx, cy, w, h int) rune {
	horizon := maxInt(2, h-h/4)
	if cy < horizon {
		return 0
	}
	col := cx / 3
	height := 1 + int(hash32(col, 0, 17)%uint32(maxInt(1, h/3+1)))
	top := h - height
	if cy < top {
		return 0
	}
	if cy == top {
		switch cx % 3 {
		case 0:
			return '/'
		case 1:
			return '_'
		default:
			return '\\'
		}
	}
	if (cy+cx+bg.Frame/8)%7 == 0 && cy < h-1 {
		return '.'
	}
	if cx%3 == 1 {
		return '|'
	}
	return '#'
}

func (bg AnimeBackground) starCell(cx, cy, w, h int) rune {
	if cy >= maxInt(1, h-h/3) {
		return 0
	}
	v := hash32(cx, cy, bg.Frame/24)
	if v%31 != 0 {
		return 0
	}
	if (cx+cy+bg.Frame/16)%2 == 0 {
		return '.'
	}
	if cx > 1 && cx < w-2 {
		return '+'
	}
	return '.'
}

func renderCell(ch rune, cx, cy, w, h int, ctx RenderContext) string {
	if ch == ' ' || ch == rune(0x2800) {
		if !ctx.Background.Enabled {
			return lipgloss.NewStyle().
				Foreground(ctx.Theme.ColorAt(cx, cy, w, h)).
				Render(string(ch))
		}
		bg := ctx.Background.cell(cx, cy, w, h, ctx.Theme)
		style := lipgloss.NewStyle().Foreground(bg.color)
		if bg.faint {
			style = style.Faint(true)
		}
		return style.Render(string(bg.ch))
	}
	return lipgloss.NewStyle().
		Foreground(ctx.Theme.ColorAt(cx, cy, w, h)).
		Render(string(ch))
}

func hash32(a, b, c int) uint32 {
	v := uint32(a*73856093) ^ uint32(b*19349663) ^ uint32(c*83492791)
	v ^= v >> 13
	v *= 0x5bd1e995
	v ^= v >> 15
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (bg AnimeBackground) String() string {
	return fmt.Sprintf("anime-bg frame=%d enabled=%t", bg.Frame, bg.Enabled)
}

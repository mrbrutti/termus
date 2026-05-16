package tui

import "github.com/charmbracelet/lipgloss"

// ColorTheme is a function that returns the foreground color for a Braille
// cell at column cx, row cy in a w×h grid. This lets a theme implement
// vertical gradients (the default), horizontal gradients (rainbow), or any
// other 2-D scheme.
type ColorTheme struct {
	Name    string
	ColorAt func(cx, cy, w, h int) lipgloss.Color
	// Accent colors used for the top/bottom bar text. The renderer doesn't
	// use these — Model does, so the chrome matches the scope.
	BarFg lipgloss.Color
	BarHi lipgloss.Color // highlight color (currently-active state, e.g. PAUSED)
}

// verticalGradient builds a ColorTheme whose color depends only on the row,
// interpolating between `center` (at the middle row) and `edge` (at the top
// and bottom rows).
func verticalGradient(name string, center, edge [3]int, bar, hi lipgloss.Color) ColorTheme {
	return ColorTheme{
		Name: name,
		ColorAt: func(_, cy, _, h int) lipgloss.Color {
			mid := float64(h-1) / 2
			d := float64(cy) - mid
			norm := 1.0
			if mid > 0 {
				if d < 0 {
					d = -d
				}
				norm = d / mid
			}
			r := lerpInt(center[0], edge[0], norm)
			g := lerpInt(center[1], edge[1], norm)
			b := lerpInt(center[2], edge[2], norm)
			return lipgloss.Color(hex6(r, g, b))
		},
		BarFg: bar,
		BarHi: hi,
	}
}

// rainbowTheme maps cell column to hue so the scope is a full spectrum of
// rainbow colors left-to-right, with no vertical variation.
var rainbowTheme = ColorTheme{
	Name: "rainbow",
	ColorAt: func(cx, _, w, _ int) lipgloss.Color {
		// Hue 0..360 across the width.
		h := float64(cx) / float64(w) * 360.0
		return lipgloss.Color(hex6(hsvToRGB(h, 0.85, 1.0)))
	},
	BarFg: "#e0e0ff",
	BarHi: "#ffd76b",
}

// Themes is the full preset palette, in display order. New is the default.
var Themes = []ColorTheme{
	verticalGradient("indigo",
		[3]int{0x5b, 0xfa, 0xff}, // center: cyan
		[3]int{0x5b, 0x4b, 0xff}, // edge: indigo
		"#a0a0ff", "#5bfaff"),
	verticalGradient("amber",
		[3]int{0xff, 0xe3, 0x7a}, // center: warm yellow
		[3]int{0xa8, 0x2b, 0x10}, // edge: deep ember
		"#ffd089", "#ffb14a"),
	verticalGradient("matrix",
		[3]int{0xb8, 0xff, 0x7a}, // center: bright lime
		[3]int{0x0e, 0x55, 0x1a}, // edge: dark forest
		"#7aff89", "#b8ffd0"),
	verticalGradient("magenta",
		[3]int{0xff, 0x9b, 0xff}, // center: hot pink
		[3]int{0x4a, 0x10, 0x5c}, // edge: deep purple
		"#ffb0ff", "#ffd0ff"),
	verticalGradient("mono",
		[3]int{0xff, 0xff, 0xff}, // center: white
		[3]int{0x55, 0x55, 0x55}, // edge: dim gray
		"#bbbbbb", "#ffffff"),
	rainbowTheme,
}

// DefaultTheme is the theme used at startup. Indigo for continuity with
// the prior release.
func DefaultTheme() ColorTheme { return Themes[0] }

// hsvToRGB converts (h in [0,360), s, v in [0,1]) to (r, g, b in 0..255).
func hsvToRGB(h, s, v float64) (r, g, b int) {
	if h < 0 {
		h = 0
	}
	if h >= 360 {
		h -= 360
	}
	c := v * s
	x := c * (1 - abs(modf(h/60.0)-1))
	m := v - c
	var rf, gf, bf float64
	switch {
	case h < 60:
		rf, gf, bf = c, x, 0
	case h < 120:
		rf, gf, bf = x, c, 0
	case h < 180:
		rf, gf, bf = 0, c, x
	case h < 240:
		rf, gf, bf = 0, x, c
	case h < 300:
		rf, gf, bf = x, 0, c
	default:
		rf, gf, bf = c, 0, x
	}
	return int((rf + m) * 255), int((gf + m) * 255), int((bf + m) * 255)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// modf returns x mod 2, used so x*y/60-1 stays in [-1, 1] for hsv math.
func modf(x float64) float64 {
	for x >= 2 {
		x -= 2
	}
	for x < 0 {
		x += 2
	}
	return x
}

func lerpInt(a, b int, t float64) int {
	return int(float64(a) + (float64(b)-float64(a))*t)
}

func hex6(r, g, b int) string {
	clamp := func(v int) int {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return v
	}
	r = clamp(r)
	g = clamp(g)
	b = clamp(b)
	const hex = "0123456789abcdef"
	return string([]byte{
		'#',
		hex[r>>4], hex[r&0xf],
		hex[g>>4], hex[g&0xf],
		hex[b>>4], hex[b&0xf],
	})
}

func blendColor(base, accent lipgloss.Color, mix float64) lipgloss.Color {
	if mix <= 0 {
		return base
	}
	if mix > 1 {
		mix = 1
	}
	br, bg, bb, ok := parseHexColor(base)
	if !ok {
		return base
	}
	ar, ag, ab, ok := parseHexColor(accent)
	if !ok {
		return base
	}
	return lipgloss.Color(hex6(
		lerpInt(br, ar, mix),
		lerpInt(bg, ag, mix),
		lerpInt(bb, ab, mix),
	))
}

func parseHexColor(color lipgloss.Color) (r, g, b int, ok bool) {
	s := string(color)
	if len(s) != 7 || s[0] != '#' {
		return 0, 0, 0, false
	}
	parse := func(ch byte) (int, bool) {
		switch {
		case ch >= '0' && ch <= '9':
			return int(ch - '0'), true
		case ch >= 'a' && ch <= 'f':
			return int(ch-'a') + 10, true
		case ch >= 'A' && ch <= 'F':
			return int(ch-'A') + 10, true
		default:
			return 0, false
		}
	}
	parts := []*int{&r, &g, &b}
	for i, part := range parts {
		hi, okHi := parse(s[1+i*2])
		lo, okLo := parse(s[2+i*2])
		if !okHi || !okLo {
			return 0, 0, 0, false
		}
		*part = hi<<4 | lo
	}
	return r, g, b, true
}

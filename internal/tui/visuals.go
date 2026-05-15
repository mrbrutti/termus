package tui

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/madelynnblue/go-dsp/fft"
)

// VisualStyle is one renderable visualization mode. Render produces a w×h
// string of styled cells given a mono buffer in [-1, 1] and a color theme.
type VisualStyle struct {
	Name   string
	Render func(samples []float64, w, h int, ctx RenderContext) string
}

// RenderContext bundles the non-audio styling inputs shared by all visuals.
type RenderContext struct {
	Theme ColorTheme
}

// Visuals is the ordered list of selectable visualization styles. [C] in the
// TUI cycles through them.
var Visuals = []VisualStyle{
	{Name: "scope", Render: RenderBrailleWithContext},
	{Name: "ridge", Render: RenderSpectrum},
	{Name: "ribbon", Render: RenderBars},
	{Name: "double", Render: RenderMirror},
}

// RenderSpectrum draws the frequency-domain envelope as one continuous ridge
// line rather than a bar chart, so it stays visually aligned with the default
// scope view.
func RenderSpectrum(samples []float64, w, h int, ctx RenderContext) string {
	if w < 4 || h < 1 {
		return "(too small)\n"
	}
	// Use the largest power-of-two prefix of samples (≤ 2048 to keep FFT cheap).
	n := largestPow2AtMost(len(samples), 2048)
	if n < 32 {
		return blankCells(w, h)
	}

	// Hann window to reduce spectral leakage.
	buf := make([]float64, n)
	for i := 0; i < n; i++ {
		w := 0.5 - 0.5*math.Cos(2*math.Pi*float64(i)/float64(n-1))
		buf[i] = samples[i] * w
	}
	spec := fft.FFTReal(buf)

	// Compute magnitude for the first n/2 bins (positive frequencies).
	half := n / 2
	mag := make([]float64, half)
	for i := 0; i < half; i++ {
		mag[i] = math.Hypot(real(spec[i]), imag(spec[i]))
	}

	// Log-bin into w buckets. Skip DC (bin 0).
	bars := make([]float64, w)
	binStart := 1
	for bx := 0; bx < w; bx++ {
		// Each bar covers an exponentially-wider slice of bins.
		lo := int(math.Floor(math.Pow(float64(half-1), float64(bx)/float64(w))))
		hi := int(math.Floor(math.Pow(float64(half-1), float64(bx+1)/float64(w))))
		if lo < binStart {
			lo = binStart
		}
		if hi <= lo {
			hi = lo + 1
		}
		if hi > half {
			hi = half
		}
		peak := 0.0
		for i := lo; i < hi; i++ {
			if mag[i] > peak {
				peak = mag[i]
			}
		}
		bars[bx] = peak
	}

	// Convert to dB and normalize to 0..1.
	for i, v := range bars {
		if v < 1e-9 {
			bars[i] = 0
			continue
		}
		db := 20 * math.Log10(v/float64(n))
		// Map -60dB..0dB into 0..1.
		nrm := (db + 60) / 60
		if nrm < 0 {
			nrm = 0
		}
		if nrm > 1 {
			nrm = 1
		}
		bars[i] = nrm
	}

	bars = smoothSeries(bars, 2)
	grid := newDotGrid(w, h)
	dotsX := len(grid[0])
	dotsY := len(grid)
	prevX, prevY := 0, dotsY-2
	for px := 0; px < dotsX; px++ {
		si := px * len(bars) / dotsX
		if si >= len(bars) {
			si = len(bars) - 1
		}
		y := dotsY - 2 - int(bars[si]*float64(dotsY-3))
		if y < 0 {
			y = 0
		}
		drawLine(grid, prevX, prevY, px, y)
		plotThickDot(grid, px, y, 1)
		prevX, prevY = px, y
	}
	return renderBrailleGrid(grid, w, h, ctx)
}

// RenderBars draws a symmetric amplitude ribbon around the centerline. It uses
// the same fine braille texture as the default scope, but reads as energy mass
// rather than an instantaneous waveform.
func RenderBars(samples []float64, w, h int, ctx RenderContext) string {
	if w < 4 || h < 1 || len(samples) == 0 {
		return blankCells(w, h)
	}
	bars := make([]float64, w)
	per := len(samples) / w
	if per < 1 {
		per = 1
	}
	for bx := 0; bx < w; bx++ {
		lo := bx * len(samples) / w
		hi := (bx + 1) * len(samples) / w
		if hi > len(samples) {
			hi = len(samples)
		}
		peak := 0.0
		for i := lo; i < hi; i++ {
			a := samples[i]
			if a < 0 {
				a = -a
			}
			if a > peak {
				peak = a
			}
		}
		bars[bx] = peak
	}
	bars = smoothSeries(bars, 3)
	grid := newDotGrid(w, h)
	dotsX := len(grid[0])
	dotsY := len(grid)
	center := dotsY / 2
	maxAmp := maxInt(2, dotsY/2-2)
	prevTop, prevBottom := center, center
	for px := 0; px < dotsX; px++ {
		si := px * len(bars) / dotsX
		if si >= len(bars) {
			si = len(bars) - 1
		}
		amp := 1 + int(bars[si]*float64(maxAmp))
		top := center - amp
		bottom := center + amp
		drawLine(grid, maxInt(0, px-1), prevTop, px, top)
		drawLine(grid, maxInt(0, px-1), prevBottom, px, bottom)
		if amp <= 2 || px%4 == 0 {
			drawLine(grid, px, top, px, bottom)
		}
		prevTop, prevBottom = top, bottom
	}
	return renderBrailleGrid(grid, w, h, ctx)
}

// RenderMirror draws two synchronized waveform traces, one in the upper half
// and one reflected below. It preserves the default scope's line quality while
// creating a denser, more architectural composition.
func RenderMirror(samples []float64, w, h int, ctx RenderContext) string {
	if w < 4 || h < 1 || len(samples) == 0 {
		return blankCells(w, h)
	}
	grid := newDotGrid(w, h)
	dotsX := len(grid[0])
	dotsY := len(grid)
	topCenter := dotsY / 4
	bottomCenter := (3 * dotsY) / 4
	span := maxInt(2, dotsY/4-2)
	prevTopX, prevTopY := 0, topCenter
	prevBottomX, prevBottomY := 0, bottomCenter
	for px := 0; px < dotsX; px++ {
		si := px * len(samples) / dotsX
		if si >= len(samples) {
			si = len(samples) - 1
		}
		s := samples[si]
		if s > 1 {
			s = 1
		}
		if s < -1 {
			s = -1
		}
		topY := topCenter - int(s*float64(span))
		bottomY := bottomCenter + int(s*float64(span))
		drawLine(grid, prevTopX, prevTopY, px, topY)
		drawLine(grid, prevBottomX, prevBottomY, px, bottomY)
		plotThickDot(grid, px, topY, 1)
		plotThickDot(grid, px, bottomY, 1)
		prevTopX, prevTopY = px, topY
		prevBottomX, prevBottomY = px, bottomY
	}
	return renderBrailleGrid(grid, w, h, ctx)
}

func newDotGrid(w, h int) [][]bool {
	dotsX := 2 * w
	dotsY := 4 * h
	grid := make([][]bool, dotsY)
	for i := range grid {
		grid[i] = make([]bool, dotsX)
	}
	return grid
}

func renderBrailleGrid(grid [][]bool, w, h int, ctx RenderContext) string {
	var b strings.Builder
	for cy := 0; cy < h; cy++ {
		for cx := 0; cx < w; cx++ {
			r := brailleCell(grid, cx, cy)
			b.WriteString(renderCell(r, cx, cy, w, h, ctx))
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func plotDot(grid [][]bool, x, y int) {
	if y < 0 || y >= len(grid) || len(grid) == 0 || x < 0 || x >= len(grid[0]) {
		return
	}
	grid[y][x] = true
}

func plotThickDot(grid [][]bool, x, y, radius int) {
	for dy := -radius; dy <= radius; dy++ {
		plotDot(grid, x, y+dy)
	}
}

func drawLine(grid [][]bool, x0, y0, x1, y1 int) {
	dx := absInt(x1 - x0)
	dy := -absInt(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	for {
		plotDot(grid, x0, y0)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func smoothSeries(in []float64, radius int) []float64 {
	if radius <= 0 || len(in) == 0 {
		out := make([]float64, len(in))
		copy(out, in)
		return out
	}
	out := make([]float64, len(in))
	for i := range in {
		sum := 0.0
		count := 0
		for j := i - radius; j <= i+radius; j++ {
			if j < 0 || j >= len(in) {
				continue
			}
			sum += in[j]
			count++
		}
		if count == 0 {
			out[i] = in[i]
			continue
		}
		out[i] = sum / float64(count)
	}
	return out
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func renderCell(ch rune, cx, cy, w, h int, ctx RenderContext) string {
	return lipgloss.NewStyle().
		Foreground(ctx.Theme.ColorAt(cx, cy, w, h)).
		Render(string(ch))
}

// blankCells returns an empty w×h grid as the renderer's "nothing to draw"
// fallback so the layout stays stable.
func blankCells(w, h int) string {
	row := strings.Repeat(" ", w) + "\n"
	return strings.Repeat(row, h)
}

// largestPow2AtMost returns the largest power of two ≤ min(n, cap).
func largestPow2AtMost(n, cap int) int {
	if n > cap {
		n = cap
	}
	p := 1
	for p*2 <= n {
		p *= 2
	}
	return p
}

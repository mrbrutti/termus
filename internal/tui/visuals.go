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
	{Name: "contour", Render: RenderContour},
	{Name: "vector", Render: RenderVector},
	{Name: "drift", Render: RenderDrift},
}

// RenderContour draws the frequency envelope as a single low horizon line.
// It is intentionally restrained: one smooth contour with broad empty space.
func RenderContour(samples []float64, w, h int, ctx RenderContext) string {
	if w < 4 || h < 1 {
		return "(too small)\n"
	}
	grid := newDotGrid(w, h)
	dotsX := len(grid[0])
	dotsY := len(grid)
	halfBars := spectralBars(samples, maxInt(8, dotsX/2))
	if len(halfBars) == 0 {
		return blankCells(w, h)
	}
	halfBars = smoothSeries(halfBars, 4)
	baseY := int(float64(dotsY) * 0.72)
	centerX := dotsX / 2
	maxLift := maxInt(2, dotsY/3)
	prevX, prevY := 0, baseY
	for px := 0; px < dotsX; px++ {
		dist := math.Abs(float64(px-centerX)) / float64(maxInt(1, centerX))
		mapped := math.Pow(dist, 1.7)
		si := int(mapped * float64(len(halfBars)-1))
		if si >= len(halfBars) {
			si = len(halfBars) - 1
		}
		energy := halfBars[si]
		lift := int(energy * (0.82 + 0.18*(1.0-dist)) * float64(maxLift))
		y := baseY - lift
		if y < 0 {
			y = 0
		}
		drawLine(grid, prevX, prevY, px, y)
		if px%7 == 0 || (dist < 0.28 && px%5 == 0) {
			plotDot(grid, px, y)
		}
		prevX, prevY = px, y
	}
	return renderBrailleGrid(grid, w, h, ctx)
}

// RenderVector projects the signal into a centered phase portrait using a
// short time lag. The result feels distinct from the scope while keeping the
// same thin, architectural line language.
func RenderVector(samples []float64, w, h int, ctx RenderContext) string {
	if w < 4 || h < 1 || len(samples) == 0 {
		return blankCells(w, h)
	}
	grid := newDotGrid(w, h)
	dotsX := len(grid[0])
	dotsY := len(grid)
	centerX := dotsX / 2
	centerY := dotsY / 2
	spanX, spanY, drive := vectorGeometry(samples, dotsX, dotsY)
	lag := maxInt(2, len(samples)/64)
	step := maxInt(1, (len(samples)-lag)/maxInt(40, dotsX))
	prevX, prevY := centerX, centerY
	first := true
	for i := 0; i+lag < len(samples); i += step {
		xs := vectorProject(samples[i], drive)
		ys := vectorProject(0.7*samples[i+lag]+0.3*samples[(i+lag/2)%len(samples)], drive)
		x := centerX + int(xs*float64(spanX))
		y := centerY - int(ys*float64(spanY))
		if first {
			first = false
		} else {
			drawLine(grid, prevX, prevY, x, y)
		}
		if i%(step*4) == 0 {
			plotDot(grid, x, y)
		}
		prevX, prevY = x, y
	}
	return renderBrailleGrid(grid, w, h, ctx)
}

func vectorGeometry(samples []float64, dotsX, dotsY int) (spanX, spanY int, drive float64) {
	peak, rms := sampleStats(samples)
	energy := clamp01(0.68*peak + 0.32*rms)
	zoom := 0.30 + 0.70*math.Sqrt(energy)
	spanX = int(float64(maxInt(2, dotsX/2-2)) * (0.92 + 0.07*zoom))
	spanY = int(float64(maxInt(2, dotsY/2-2)) * (0.88 + 0.10*zoom))
	drive = 3.0 + 5.5*math.Sqrt(energy)
	if drive < 3.0 {
		drive = 3.0
	}
	if drive > 8.5 {
		drive = 8.5
	}
	return spanX, spanY, drive
}

func vectorProject(v, drive float64) float64 {
	return math.Tanh(v * drive)
}

// RenderDrift draws a restrained string field. Each horizontal line is pinned
// at the sides and vibrates mostly in the middle; low and high frequency bands
// excite different strings so the shape changes with the notes.
func RenderDrift(samples []float64, w, h int, ctx RenderContext) string {
	if w < 4 || h < 1 || len(samples) == 0 {
		return blankCells(w, h)
	}
	grid := newDotGrid(w, h)
	dotsX := len(grid[0])
	dotsY := len(grid)
	bands := spectralBars(samples, 48)
	if len(bands) == 0 {
		return blankCells(w, h)
	}
	stringsN := 6
	centerY := dotsY / 2
	spacing := maxInt(4, dotsY/10)
	maxDeflect := maxInt(3, dotsY/6)
	phaseStride := maxInt(1, len(samples)/(stringsN*7))
	peak, rms := sampleStats(samples)
	motionScale := 0.95 + 1.20*math.Sqrt(clamp01(0.5*peak+0.5*rms))
	for stringIdx := 0; stringIdx < stringsN; stringIdx++ {
		fromBottom := stringsN - 1 - stringIdx
		baseY := centerY + (fromBottom-(stringsN-1)/2)*spacing
		if stringsN%2 == 0 {
			baseY += spacing / 2
		}
		energy := math.Sqrt(clamp01(bandAverage(bands, fromBottom, stringsN)))
		amp := clamp01((0.28 + 1.15*energy) * motionScale)
		deflect := maxInt(3, int(float64(maxDeflect)*(0.72+1.25*amp)))
		offsetA := (stringIdx + 1) * phaseStride
		offsetB := (stringsN - stringIdx + 1) * phaseStride / 2
		offsetC := (stringIdx + 2) * phaseStride / 3
		prevX, prevY := 0, baseY
		for px := 0; px < dotsX; px++ {
			si := px * len(samples) / dotsX
			s0 := clamp1(samples[(si+offsetA)%len(samples)])
			s1 := clamp1(samples[(si+offsetB)%len(samples)])
			s2 := clamp1(samples[(si+offsetC)%len(samples)])
			raw := 0.95*s0 + 0.45*s1 - 0.30*s2
			wave := math.Tanh(raw * (1.8 + 2.4*amp))
			anchor := math.Pow(math.Sin(math.Pi*float64(px)/float64(maxInt(1, dotsX-1))), 0.42)
			y := baseY - int(wave*anchor*float64(deflect))
			if y < 0 {
				y = 0
			}
			if y >= dotsY {
				y = dotsY - 1
			}
			drawLine(grid, prevX, prevY, px, y)
			if px%7 == 0 {
				plotDot(grid, px, y)
			}
			prevX, prevY = px, y
		}
	}
	return renderBrailleGrid(grid, w, h, ctx)
}

func spectralBars(samples []float64, nBars int) []float64 {
	if nBars < 1 {
		return nil
	}
	n := largestPow2AtMost(len(samples), 2048)
	if n < 32 {
		return nil
	}
	buf := make([]float64, n)
	for i := 0; i < n; i++ {
		win := 0.5 - 0.5*math.Cos(2*math.Pi*float64(i)/float64(n-1))
		buf[i] = samples[i] * win
	}
	spec := fft.FFTReal(buf)
	half := n / 2
	mag := make([]float64, half)
	for i := 0; i < half; i++ {
		mag[i] = math.Hypot(real(spec[i]), imag(spec[i]))
	}
	bars := make([]float64, nBars)
	for bx := 0; bx < nBars; bx++ {
		lo := int(math.Floor(math.Pow(float64(half-1), float64(bx)/float64(nBars))))
		hi := int(math.Floor(math.Pow(float64(half-1), float64(bx+1)/float64(nBars))))
		if lo < 1 {
			lo = 1
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
		if peak < 1e-9 {
			bars[bx] = 0
			continue
		}
		db := 20 * math.Log10(peak/float64(n))
		nrm := (db + 60) / 60
		bars[bx] = clamp01(nrm)
	}
	return bars
}

func bandAverage(bars []float64, idx, total int) float64 {
	if len(bars) == 0 || total < 1 {
		return 0
	}
	lo := idx * len(bars) / total
	hi := (idx + 1) * len(bars) / total
	if hi <= lo {
		hi = lo + 1
	}
	if hi > len(bars) {
		hi = len(bars)
	}
	sum := 0.0
	for _, v := range bars[lo:hi] {
		sum += v
	}
	return sum / float64(hi-lo)
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

func clamp1(v float64) float64 {
	if v > 1 {
		return 1
	}
	if v < -1 {
		return -1
	}
	return v
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func sampleStats(samples []float64) (peak, rms float64) {
	if len(samples) == 0 {
		return 0, 0
	}
	sumSq := 0.0
	for _, s := range samples {
		a := s
		if a < 0 {
			a = -a
		}
		if a > peak {
			peak = a
		}
		sumSq += s * s
	}
	rms = math.Sqrt(sumSq / float64(len(samples)))
	return peak, rms
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

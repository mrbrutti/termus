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
	{Name: "spectrum", Render: RenderSpectrum},
	{Name: "bars", Render: RenderBars},
	{Name: "mirror", Render: RenderMirror},
}

// RenderSpectrum draws a frequency-domain magnitude plot using block-character
// vertical bars. Computes a real FFT over the most-recent power-of-two prefix
// of `samples`, log-binned into w bars; each bar's height is rendered with
// block-element characters (▁▂▃…█).
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

	return renderBarColumns(bars, h, ctx)
}

// RenderBars draws a stylized waveform: each column is a vertical bar whose
// height is the peak |sample| in that column's slice of the buffer. Quieter
// passages → shorter bars. Cheaper than the spectrum view and a nice
// rest-state when you want less visual stimulation.
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
	return renderBarColumns(bars, h, ctx)
}

// RenderMirror draws the waveform symmetrically around the middle row, so the
// top half plots the positive envelope of the signal and the bottom half
// mirrors it downward. Visually emphasizes amplitude dynamics.
func RenderMirror(samples []float64, w, h int, ctx RenderContext) string {
	if w < 4 || h < 1 || len(samples) == 0 {
		return blankCells(w, h)
	}
	bars := make([]float64, w)
	for bx := 0; bx < w; bx++ {
		lo := bx * len(samples) / w
		hi := (bx + 1) * len(samples) / w
		if hi > len(samples) {
			hi = len(samples)
		}
		var sumSq float64
		count := 0
		for i := lo; i < hi; i++ {
			sumSq += samples[i] * samples[i]
			count++
		}
		if count == 0 {
			continue
		}
		bars[bx] = math.Sqrt(sumSq / float64(count))
	}
	return renderMirrorColumns(bars, h, ctx)
}

// renderBarColumns turns a w-long slice of [0,1] bar heights into a grid of
// bottom-anchored bars using the 8-step block-element characters.
func renderBarColumns(bars []float64, h int, ctx RenderContext) string {
	w := len(bars)
	const eighths = " ▁▂▃▄▅▆▇█"
	runes := []rune(eighths)

	// Compute, for each column, the height in eighths (0..8h).
	hEighths := make([]int, w)
	for i, v := range bars {
		hEighths[i] = int(math.Round(v * float64(h*8)))
		if hEighths[i] < 0 {
			hEighths[i] = 0
		}
		if hEighths[i] > h*8 {
			hEighths[i] = h * 8
		}
	}

	var b strings.Builder
	for cy := 0; cy < h; cy++ {
		// Row 0 is the top — we want the bar to grow from the bottom.
		rowFromBottom := h - 1 - cy
		rowStart := rowFromBottom * 8
		for cx := 0; cx < w; cx++ {
			level := hEighths[cx] - rowStart
			if level < 0 {
				level = 0
			}
			if level > 8 {
				level = 8
			}
			b.WriteString(renderCell(runes[level], cx, cy, w, h, ctx))
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// renderMirrorColumns draws each column as a bar that grows up from the
// midline AND down from it — a symmetric envelope shape. Bottom half uses
// graduated 8th-block characters (smooth); top half uses full/empty cells
// (block elements don't have a clean top-anchored 8-step variant).
func renderMirrorColumns(bars []float64, h int, ctx RenderContext) string {
	w := len(bars)
	const eighths = " ▁▂▃▄▅▆▇█"
	runesBot := []rune(eighths)
	half := h / 2

	hCells := make([]int, w) // bar height in whole cells (top half uses this)
	hEighths := make([]int, w)
	for i, v := range bars {
		eight := int(math.Round(v * float64(half*8)))
		if eight < 0 {
			eight = 0
		}
		if eight > half*8 {
			eight = half * 8
		}
		hEighths[i] = eight
		hCells[i] = eight / 8 // round-down whole cells for top
	}

	var b strings.Builder
	for cy := 0; cy < h; cy++ {
		for cx := 0; cx < w; cx++ {
			var ch rune
			switch {
			case cy < half:
				// Top half: full block if covered, space otherwise.
				rowFromCenter := half - cy
				if rowFromCenter <= hCells[cx] {
					ch = '█'
				} else {
					ch = ' '
				}
			default:
				// Bottom half: bar growing from midline downward.
				rowFromCenter := cy - half
				rowStart := rowFromCenter * 8
				level := hEighths[cx] - rowStart
				if level < 0 {
					level = 0
				}
				if level > 8 {
					level = 8
				}
				ch = runesBot[level]
			}
			b.WriteString(renderCell(ch, cx, cy, w, h, ctx))
		}
		b.WriteByte('\n')
	}
	return b.String()
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

package tui

import (
	"math"
	"strings"
	"testing"

	"github.com/muesli/termenv"
)

func testSamples(n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = math.Sin(2*math.Pi*float64(i)/64) * 0.5
	}
	return out
}

func TestEachVisualRendersWithoutPanic(t *testing.T) {
	const w, h = 60, 10
	samples := testSamples(2048)
	ctx := RenderContext{Theme: DefaultTheme()}
	for _, v := range Visuals {
		out := v.Render(samples, w, h, ctx)
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		if len(lines) != h {
			t.Errorf("%s: got %d lines, want %d", v.Name, len(lines), h)
		}
	}
}

func TestSpectrumQuietSignalIsMostlyEmpty(t *testing.T) {
	const w, h = 40, 8
	samples := make([]float64, 2048) // all zeros
	out := RenderSpectrum(samples, w, h, RenderContext{Theme: DefaultTheme()})
	// All-zero signal should leave the bar area without any active columns.
	for _, glyph := range []rune{'█', '▇', '▆', '▅', '▄', '▃', '▂', '▁'} {
		if strings.ContainsRune(out, glyph) {
			t.Errorf("expected no active bars for zero input, got:\n%s", out)
			break
		}
	}
}

func TestDetectAdaptiveUIForAscii(t *testing.T) {
	ui := detectAdaptiveUIWith(termenv.Ascii, true)
	if len(ui.Themes) != 1 || ui.Themes[0].Name != "mono" {
		t.Fatalf("ascii profile should force mono theme, got %+v", ui.Themes)
	}
}

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

func TestContourQuietSignalIsMostlyEmpty(t *testing.T) {
	const w, h = 40, 8
	samples := make([]float64, 2048) // all zeros
	out := RenderContour(samples, w, h, RenderContext{Theme: DefaultTheme()})
	// All-zero signal should leave the bar area without any active columns.
	for _, glyph := range []rune{'█', '▇', '▆', '▅', '▄', '▃', '▂', '▁'} {
		if strings.ContainsRune(out, glyph) {
			t.Errorf("expected no active bars for zero input, got:\n%s", out)
			break
		}
	}
}

func TestVisualNamesMatchNewMinimalSet(t *testing.T) {
	want := []string{"scope", "contour", "vector", "drift"}
	if len(Visuals) != len(want) {
		t.Fatalf("visual count = %d, want %d", len(Visuals), len(want))
	}
	for i, v := range Visuals {
		if v.Name != want[i] {
			t.Fatalf("visual %d = %q, want %q", i, v.Name, want[i])
		}
	}
}

func TestAlternateVisualsUseBrailleTexture(t *testing.T) {
	samples := testSamples(2048)
	ctx := RenderContext{Theme: DefaultTheme()}
	for _, v := range Visuals[1:] {
		out := v.Render(samples, 60, 10, ctx)
		hasBraille := false
		for _, r := range out {
			if r >= 0x2801 && r <= 0x28FF {
				hasBraille = true
				break
			}
		}
		if !hasBraille {
			t.Fatalf("%s: expected braille-style texture, got:\n%s", v.Name, out)
		}
		for _, glyph := range []rune{'█', '▇', '▆', '▅', '▄', '▃', '▂', '▁'} {
			if strings.ContainsRune(out, glyph) {
				t.Fatalf("%s: should avoid block-bar glyph %q, got:\n%s", v.Name, string(glyph), out)
			}
		}
	}
}

func TestVectorGeometryExpandsForLouderSignal(t *testing.T) {
	soft := make([]float64, 2048)
	loud := make([]float64, 2048)
	for i := range soft {
		phase := 2 * math.Pi * float64(i) / 64
		soft[i] = math.Sin(phase) * 0.12
		loud[i] = math.Sin(phase) * 0.9
	}
	softX, softY, softDrive := vectorGeometry(soft, 120, 40)
	loudX, loudY, loudDrive := vectorGeometry(loud, 120, 40)
	if loudX <= softX || loudY <= softY || loudDrive <= softDrive {
		t.Fatalf("vector geometry should expand with louder signal: soft=(%d,%d,%.2f) loud=(%d,%d,%.2f)",
			softX, softY, softDrive, loudX, loudY, loudDrive)
	}
}

func TestDetectAdaptiveUIForAscii(t *testing.T) {
	ui := detectAdaptiveUIWith(termenv.Ascii, true)
	if len(ui.Themes) != 1 || ui.Themes[0].Name != "mono" {
		t.Fatalf("ascii profile should force mono theme, got %+v", ui.Themes)
	}
}

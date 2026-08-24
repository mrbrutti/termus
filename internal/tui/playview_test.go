package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
)

func stationTestModel() Model {
	spec, _ := gen.Resolve("ambient")
	m := Model{width: 118, height: 32, keyName: "Cmin", seed: 71001, algo: spec.Label()}
	m.genres = []gen.AlgoSpec{spec}
	m.genreIdx = 0
	m.ring = scope.NewRing(1024)
	m.cmd = &tuiCommanderStub{}
	return m
}

func TestStationHeaderShowsIdentity(t *testing.T) {
	m := stationTestModel()
	out := topBar(m, 118, DefaultTheme(), false)
	if !strings.Contains(out, "NIGHT DRIFT") {
		t.Fatalf("station header missing uppercase station name: %q", out)
	}
	if !strings.Contains(out, "ambient") || !strings.Contains(out, "Cmin") {
		t.Fatalf("station header missing algo/key line: %q", out)
	}
	if !strings.Contains(out, "seed 71001") {
		t.Fatalf("station header missing seed: %q", out)
	}
	if !strings.Contains(out, "◌") {
		t.Fatalf("station header missing style glyph: %q", out)
	}
}

func TestStationHeaderFitsNarrowWidth(t *testing.T) {
	m := stationTestModel()
	m.recording = true
	m.seedA = &seedBookmark{Seed: 70992}
	m.seedB = &seedBookmark{Seed: 71001}
	m.kept = map[string]seedBookmark{"ambient/71001": {Seed: 71001}}

	wide := topBar(m, 72, DefaultTheme(), false)
	if got := lipgloss.Width(wide); got > 72 {
		t.Fatalf("station header overflows w=72: width %d, %q", got, wide)
	}
	if !strings.Contains(wide, "NIGHT DRIFT") {
		t.Fatalf("station header should keep the station name at w=72: %q", wide)
	}
	narrow := topBar(m, 40, DefaultTheme(), false)
	if got := lipgloss.Width(narrow); got > 40 {
		t.Fatalf("station header overflows w=40: width %d, %q", got, narrow)
	}
}

func TestFooterFitsNarrowWidth(t *testing.T) {
	m := stationTestModel()
	out := bottomBar(m, 40, DefaultTheme(), true)
	if got := lipgloss.Width(out); got > 40 {
		t.Fatalf("footer overflows w=40: width %d, %q", got, out)
	}
	if !strings.Contains(out, "[?] help") {
		t.Fatalf("narrow footer should still point at help: %q", out)
	}
}

// TestPlayChromeNeverExceedsWidth sweeps every width the app can run at.
// View's chromeH math assumes one row per bar, so a bar that overflows wraps
// and silently shifts the whole layout. This caught an off-by-one in the
// footer at w=59 that the fixed-width cases above miss.
func TestPlayChromeNeverExceedsWidth(t *testing.T) {
	theme := DefaultTheme()
	for _, recording := range []bool{false, true} {
		for _, slots := range []bool{false, true} {
			for _, compact := range []bool{false, true} {
				for w := 40; w <= 200; w++ {
					m := stationTestModel()
					m.recording = recording
					m.stickyStatus = "audio: no default device; use --out file.wav"
					if slots {
						// Worst case: max-width int64 seeds in both slots.
						m.seedA = &seedBookmark{Seed: 9223372036854775807}
						m.seedB = &seedBookmark{Seed: 9223372036854775806}
						m.kept = map[string]seedBookmark{"a": {}, "b": {}}
					}
					m.debug.FormChain = []string{"intro", "statement", "development", "recapitulation", "coda"}
					m.debug.FormIndex = 2

					if got := lipgloss.Width(topBar(m, w, theme, compact)); got > w {
						t.Fatalf("topBar w=%d compact=%v recording=%v slots=%v rendered %d cols",
							w, compact, recording, slots, got)
					}
					if got := lipgloss.Width(bottomBar(m, w, theme, compact)); got > w {
						t.Fatalf("bottomBar w=%d compact=%v rendered %d cols", w, compact, got)
					}
					if got := lipgloss.Width(formRailBar(m, w, theme)); got > w {
						t.Fatalf("formRailBar w=%d rendered %d cols", w, got)
					}
				}
			}
		}
	}
}

func TestNarrationRowPromotesDebugStatus(t *testing.T) {
	m := stationTestModel()
	m.debug = gen.DebugStatus{
		Movement: "develop", Episode: 3, Section: "drift",
		Bar: 129, Chord: "Dm9", NextChord: "Gm7",
	}
	out := playbackBar(m, 118, DefaultTheme(), make([]float64, 64), false)
	for _, want := range []string{"movement develop", "episode 3", "section drift", "bar 129", "Dm9 → Gm7", "lvl"} {
		if !strings.Contains(out, want) {
			t.Fatalf("narration row missing %q: %q", want, out)
		}
	}
}

func TestNarrationRowOmitsMissingParts(t *testing.T) {
	m := stationTestModel()
	m.debug = gen.DebugStatus{Section: "A", Bar: 4, Chord: "Cmaj7"}
	out := playbackBar(m, 118, DefaultTheme(), make([]float64, 64), false)
	if strings.Contains(out, "movement") || strings.Contains(out, "episode") {
		t.Fatalf("narration row should omit movement/episode when absent: %q", out)
	}
	if !strings.Contains(out, "Cmaj7") || strings.Contains(out, "→") {
		t.Fatalf("chord without next-chord should render bare: %q", out)
	}
}

func TestFormRailMarksCurrentSection(t *testing.T) {
	m := stationTestModel()
	m.debug = gen.DebugStatus{FormChain: []string{"intro", "A", "B", "outro"}, FormIndex: 1}
	out := formRailBar(m, 118, DefaultTheme())
	if !strings.Contains(out, "● A") {
		t.Fatalf("form rail missing current-section marker: %q", out)
	}
	if !strings.Contains(out, "intro") || !strings.Contains(out, "outro") {
		t.Fatalf("form rail missing chain sections: %q", out)
	}
	if !strings.Contains(out, "endless") {
		t.Fatalf("form rail missing listening mode: %q", out)
	}
}

func TestFormRailEmptyWithoutSource(t *testing.T) {
	m := stationTestModel()
	if out := formRailBar(m, 118, DefaultTheme()); out != "" {
		t.Fatalf("form rail should collapse without a source, got %q", out)
	}
}

func TestFooterAdvertisesCoreKeys(t *testing.T) {
	m := stationTestModel()
	out := bottomBar(m, 118, DefaultTheme(), false)
	for _, want := range []string{"[space] play", "[m] control", "[t] tracks", "[?] help", "[z] zen"} {
		if !strings.Contains(out, want) {
			t.Fatalf("footer missing %q: %q", want, out)
		}
	}
}

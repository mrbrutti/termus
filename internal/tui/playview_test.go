package tui

import (
	"strings"
	"testing"
	"time"

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
	out := bottomBar(m, 40, DefaultTheme())
	if got := lipgloss.Width(out); got > 40 {
		t.Fatalf("footer overflows w=40: width %d, %q", got, out)
	}
	if !strings.Contains(out, "[?] help") {
		t.Fatalf("narrow footer should still point at help: %q", out)
	}
}

// TestPlayChromeNeverExceedsWidth sweeps every width the app can run at.
// View's chromeH math assumes one row per bar, so a bar that overflows wraps
// and silently shifts the whole layout. Fixed-width cases miss these: this
// sweep caught an off-by-one in the footer at w=59, an unbounded form-rail
// right side, and a narration budget that ignored the meter's "clip" state.
func TestPlayChromeNeverExceedsWidth(t *testing.T) {
	theme := DefaultTheme()

	longTitle := "the long slow return of the opening figure, heard now in the parallel minor"
	if len(longTitle) < 70 {
		t.Fatalf("fixture should exercise an over-wide section label, got %d chars", len(longTitle))
	}
	quiet := []float64{0.1, 0.3, -0.4}
	// >= 0.985 drives meterSummary into the clipped state, whose "clip"
	// label is two columns wider than "ok".
	clipping := []float64{0.1, 0.99, -0.4}

	scenarios := []struct {
		name    string
		samples []float64
		mutate  func(m *Model)
	}{
		{name: "bare", samples: quiet, mutate: func(m *Model) {}},
		{
			name:    "full narration + clipping",
			samples: clipping,
			mutate: func(m *Model) {
				m.recording = true
				m.recordStartedAt = time.Now().Add(-17 * time.Second)
				m.debug = gen.DebugStatus{
					Movement: "recapitulation", Episode: 12, Section: "development",
					Bar: 129, Chord: "Dm9", NextChord: "Gm7",
					FormChain: []string{"intro", "statement", "development", "recapitulation", "coda"},
					FormIndex: 2,
				}
				m.aceRenderActive = true
				m.aceRenderDetail = "generating next track"
			},
		},
		{
			name:    "long rail right side",
			samples: clipping,
			mutate: func(m *Model) {
				m.listeningMode = "hour stream"
				m.nextSectionAt = time.Now().Add(3*time.Minute + 42*time.Second)
				m.debug.FormChain = []string{"intro", "statement", "development", "coda"}
				m.debug.FormIndex = 1
			},
		},
		{
			name:    "over-wide section label",
			samples: quiet,
			mutate: func(m *Model) {
				m.listeningMode = "hour stream"
				m.nextSectionAt = time.Now().Add(90 * time.Second)
				m.debug.FormChain = []string{"intro", longTitle, "coda"}
				m.debug.FormIndex = 1
			},
		},
		{
			name:    "max seeds + recording",
			samples: clipping,
			mutate: func(m *Model) {
				m.recording = true
				m.seedA = &seedBookmark{Seed: 9223372036854775807}
				m.seedB = &seedBookmark{Seed: 9223372036854775806}
				m.kept = map[string]seedBookmark{"a": {}, "b": {}}
				m.debug.FormChain = []string{"intro", "A", "B"}
				m.debug.FormIndex = 0
			},
		},
	}

	for _, sc := range scenarios {
		for _, compact := range []bool{false, true} {
			for w := 40; w <= 200; w++ {
				m := stationTestModel()
				m.stickyStatus = "audio: no default device; use --out file.wav"
				sc.mutate(&m)

				bars := map[string]string{
					"topBar":      topBar(m, w, theme, compact),
					"playbackBar": playbackBar(m, w, theme, sc.samples, compact),
					"formRailBar": formRailBar(m, w, theme),
					"debugBar":    debugBar(m, w, theme),
					"bottomBar":   bottomBar(m, w, theme),
				}
				for name, row := range bars {
					if got := lipgloss.Width(row); got > w {
						t.Fatalf("%s overflows in scenario %q: w=%d compact=%v rendered %d cols\n%q",
							name, sc.name, w, compact, got, row)
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
	out := bottomBar(m, 118, DefaultTheme())
	for _, want := range []string{"[space] play", "[m] control", "[t] tracks", "[?] help", "[z] zen"} {
		if !strings.Contains(out, want) {
			t.Fatalf("footer missing %q: %q", want, out)
		}
	}
}

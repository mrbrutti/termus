package tui

import (
	"strings"
	"testing"

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

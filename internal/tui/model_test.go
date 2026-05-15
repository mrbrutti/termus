package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/gen"
)

type tuiCommanderStub struct {
	swaps []gen.Algorithm
}

func (s *tuiCommanderStub) SetVolume(int)                    {}
func (s *tuiCommanderStub) DebugStatus() gen.DebugStatus     { return gen.DebugStatus{} }
func (s *tuiCommanderStub) TogglePause()                     {}
func (s *tuiCommanderStub) ToggleRecord() (string, error)    { return "", nil }
func (s *tuiCommanderStub) SwapAlgorithm(algo gen.Algorithm) { s.swaps = append(s.swaps, algo) }
func (s *tuiCommanderStub) SwapAlgorithmFade(algo gen.Algorithm, fadeFrames int) {
	s.swaps = append(s.swaps, algo)
}

type tuiAlgoStub struct{ name string }

func (a *tuiAlgoStub) Name() string        { return a.name }
func (a *tuiAlgoStub) Seed(int64)          {}
func (a *tuiAlgoStub) Next(l, r []float64) {}

func TestModelAudioStateLifecycle(t *testing.T) {
	m := Model{}
	m.applyAudioState(audio.BackendState{Kind: audio.BackendStateStarting})
	if got := m.currentStatus(time.Now()); got != "audio: starting..." {
		t.Fatalf("starting status = %q", got)
	}

	m.applyAudioState(audio.BackendState{Kind: audio.BackendStateReady})
	if got := m.currentStatus(time.Now()); got != "audio: ready" {
		t.Fatalf("ready status = %q", got)
	}
	if got := m.currentStatus(time.Now().Add(3 * time.Second)); got != "" {
		t.Fatalf("ready flash should clear, got %q", got)
	}

	m.applyAudioState(audio.BackendState{Kind: audio.BackendStateNoDefaultDevice})
	if got := m.currentStatus(time.Now().Add(3 * time.Second)); got != "audio: no default device; use --out file.wav" {
		t.Fatalf("no-device status = %q", got)
	}
}

func TestBottomBarLeavesRoomForStatus(t *testing.T) {
	m := Model{
		volume:       70,
		stickyStatus: "audio: starting...",
		themes:       []ColorTheme{DefaultTheme()},
	}
	bar := bottomBar(m, 80, DefaultTheme())
	if !strings.Contains(bar, "audio: starting...") {
		t.Fatalf("bottom bar missing status: %q", bar)
	}
}

func TestTopBarShowsTitle(t *testing.T) {
	m := Model{
		algo:   "Jazz",
		seed:   42,
		debug:  gen.DebugStatus{Bar: 5, Section: "A'", Chord: "G7", Preset: "tyros4"},
		themes: []ColorTheme{DefaultTheme()},
	}
	bar := topBar(m, 120, DefaultTheme())
	if !strings.Contains(bar, "termus · Jazz") || !strings.Contains(bar, "seed=42") {
		t.Fatalf("top bar missing title info: %q", bar)
	}
}

func TestDebugBarShowsDedicatedInspector(t *testing.T) {
	m := Model{
		debugVisible: true,
		debug:        gen.DebugStatus{Bar: 3, Section: "cadence", Chord: "Dm7", Preset: "sgm"},
		themes:       []ColorTheme{DefaultTheme()},
	}
	bar := debugBar(m, 100, DefaultTheme())
	if !strings.Contains(bar, "DEBUG") || !strings.Contains(bar, "bar 3") || !strings.Contains(bar, "Dm7") {
		t.Fatalf("debug bar missing inspector fields: %q", bar)
	}
}

func TestSeedBrowserStoresAndTogglesAB(t *testing.T) {
	cmd := &tuiCommanderStub{}
	specs := []gen.AlgoSpec{{Name: "ambient", Display: "Ambient"}}
	build := func(spec gen.AlgoSpec, seed int64) gen.Algorithm {
		return &tuiAlgoStub{name: spec.Name}
	}
	m := Model{
		cmd:     cmd,
		genres:  specs,
		buildFn: build,
		algo:    "Ambient",
		seed:    42,
	}

	m.storeSeed("A")
	if m.seedA == nil || m.seedA.Seed != 42 {
		t.Fatalf("seedA = %+v, want seed 42", m.seedA)
	}
	m.seed = 43
	m.storeSeed("B")
	m.toggleSeedCompare()
	if m.seed != 42 {
		t.Fatalf("toggle from B should recall A, got seed %d", m.seed)
	}
	m.toggleSeedCompare()
	if m.seed != 43 {
		t.Fatalf("toggle from A should recall B, got seed %d", m.seed)
	}
	if len(cmd.swaps) != 2 {
		t.Fatalf("swap count = %d, want 2", len(cmd.swaps))
	}
}

package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

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
	bar := bottomBar(m, 120, DefaultTheme())
	if !strings.Contains(bar, "audio: starting...") {
		t.Fatalf("bottom bar missing status: %q", bar)
	}
	if !strings.Contains(bar, "[?] help") {
		t.Fatalf("bottom bar should expose help entry point: %q", bar)
	}
	if !strings.Contains(bar, "[l] library") {
		t.Fatalf("bottom bar should expose saved-seed library: %q", bar)
	}
	if strings.Contains(bar, "[[/]] seed") {
		t.Fatalf("bottom bar should stay compact, got: %q", bar)
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

func TestPlaybackBarShowsTimingAndMeter(t *testing.T) {
	now := time.Now()
	m := Model{
		recording:       true,
		startedAt:       now.Add(-95 * time.Second),
		recordStartedAt: now.Add(-17 * time.Second),
		playlist: &gen.Playlist{Tracks: []gen.Track{
			{Duration: 5 * time.Minute},
		}},
		playlistIdx:    0,
		trackStartedAt: now.Add(-32 * time.Second),
		nextTrackAt:    now.Add(4*time.Minute + 28*time.Second),
		playlistFade:   88200,
		themes:         []ColorTheme{DefaultTheme()},
	}
	samples := []float64{0.1, 0.3, 0.85, -0.4}
	bar := playbackBar(m, 120, DefaultTheme(), samples)
	for _, want := range []string{"live 01:35", "track 00:32/05:00", "next 04:28", "fade 00:02", "rec 00:17", "lvl"} {
		if !strings.Contains(bar, want) {
			t.Fatalf("playback bar missing %q: %q", want, bar)
		}
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

func TestHelpPanelShowsCoreControls(t *testing.T) {
	m := Model{
		helpVisible: true,
		genres:      []gen.AlgoSpec{{Name: "ambient", Display: "Ambient"}, {Name: "jazz", Display: "Jazz"}},
		playlist:    &gen.Playlist{Name: "mix", Tracks: []gen.Track{{Duration: time.Second}}},
		themes:      []ColorTheme{DefaultTheme()},
	}
	panel := helpPanel(m, 90, 18, DefaultTheme())
	for _, want := range []string{"TERMUS HELP", "Playback", "Seeds", "[l] library", "Tracks", "[?] close this overlay"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("help panel missing %q:\n%s", want, panel)
		}
	}
}

func TestHelpBlocksNonHelpKeys(t *testing.T) {
	cmd := &tuiCommanderStub{}
	m := Model{
		cmd:         cmd,
		helpVisible: true,
		volume:      60,
	}
	next, _ := m.Update(keyMsg("up"))
	got := next.(Model)
	if got.volume != 60 {
		t.Fatalf("volume changed while help overlay visible: %d", got.volume)
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

func TestLibraryPanelShowsSavedSeeds(t *testing.T) {
	m := Model{
		libraryVisible: true,
		libraryIdx:     0,
		savedSeeds: []savedSeedRecord{
			{Algo: "ambient", Display: "Ambient", Seed: 42, SavedAt: time.Now().Add(-2 * time.Minute)},
		},
		themes: []ColorTheme{DefaultTheme()},
	}
	panel := libraryPanel(m, 90, 18, DefaultTheme())
	for _, want := range []string{"SAVED SEEDS", "Ambient", "42", "[enter] load"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("library panel missing %q:\n%s", want, panel)
		}
	}
}

func TestMeterSummaryDetectsClip(t *testing.T) {
	peak, clipped := meterSummary([]float64{0.2, -0.99, 0.3})
	if peak < 0.99 || !clipped {
		t.Fatalf("meterSummary = (%v, %v), want clipped peak", peak, clipped)
	}
}

func keyMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

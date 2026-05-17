package tui

import (
	"fmt"
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
		algo:         "Ambient",
		volume:       70,
		stickyStatus: "audio: starting...",
		themes:       []ColorTheme{DefaultTheme()},
	}
	bar := bottomBar(m, 120, DefaultTheme(), false)
	if !strings.Contains(bar, "audio: starting...") {
		t.Fatalf("bottom bar missing status: %q", bar)
	}
	if !strings.Contains(bar, "Ambient") {
		t.Fatalf("bottom bar should show current music type: %q", bar)
	}
	if !strings.Contains(bar, "?  m") {
		t.Fatalf("bottom bar should expose help entry point: %q", bar)
	}
	if strings.Contains(bar, "[l] library") || strings.Contains(bar, "[i] inspect") || strings.Contains(bar, "[space]") {
		t.Fatalf("bottom bar should stay minimal, got: %q", bar)
	}
}

func TestTopBarShowsTitle(t *testing.T) {
	m := Model{
		algo:   "Jazz",
		seed:   42,
		debug:  gen.DebugStatus{Bar: 5, Section: "A'", Chord: "G7", Preset: "tyros4"},
		themes: []ColorTheme{DefaultTheme()},
	}
	bar := topBar(m, 120, DefaultTheme(), false)
	if !strings.Contains(bar, "termus · Jazz") || !strings.Contains(bar, "seed=42") {
		t.Fatalf("top bar missing title info: %q", bar)
	}
}

func TestTopBarShowsStationAndAlgoNameWhenSpecAvailable(t *testing.T) {
	m := Model{
		algo:     "Night Drift",
		seed:     42,
		keyName:  "Cmin",
		genreIdx: 0,
		genres:   []gen.AlgoSpec{{Name: "ambient", Display: "Ambient", Station: "Night Drift"}},
		themes:   []ColorTheme{DefaultTheme()},
	}
	bar := topBar(m, 140, DefaultTheme(), false)
	if !strings.Contains(bar, "Night Drift · ambient") {
		t.Fatalf("top bar should surface both station and canonical algo name: %q", bar)
	}
}

func TestPlaybackBarShowsTimingAndMeter(t *testing.T) {
	now := time.Now()
	m := Model{
		recording:       true,
		listeningMode:   "hour stream",
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
	bar := playbackBar(m, 120, DefaultTheme(), samples, false)
	for _, want := range []string{"live 01:35", "hour stream", "track 00:32/05:00", "next 04:28", "fade 00:02", "rec 00:17", "lvl"} {
		if !strings.Contains(bar, want) {
			t.Fatalf("playback bar missing %q: %q", want, bar)
		}
	}
}

func TestStartVisualTransitionTracksPreviousVisual(t *testing.T) {
	m := Model{visualIdx: 1, visualPrevIdx: -1}
	m.startVisualTransition(3)
	if m.visualIdx != 3 || m.visualPrevIdx != 1 {
		t.Fatalf("transition state = (%d,%d), want current=3 previous=1", m.visualIdx, m.visualPrevIdx)
	}
	if !m.visualTransitionActive(time.Now()) {
		t.Fatal("expected active visual transition")
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
	for _, want := range []string{"TERMUS HELP", "Global", "[m] control center", "Inside Control Center", "Sections", "Now   Look   Music", "[?] close help"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("help panel missing %q:\n%s", want, panel)
		}
	}
}

func TestControlsPanelShowsTabbedOverlay(t *testing.T) {
	m := Model{
		controlsVisible: true,
		controlTab:      controlTabMusic,
		algo:            "Ambient",
		seed:            42,
		volume:          70,
		themes:          []ColorTheme{DefaultTheme()},
	}
	panel := controlsPanel(m, 100, 22, DefaultTheme())
	for _, want := range []string{"CONTROL CENTER", "NOW", "LOOK", "MUSIC", "SEEDS", "LIBRARY", "EXPORT", "AUDIO", "DEBUG", "density", "brightness", "reverb", "[tab] next section"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("controls panel missing %q:\n%s", want, panel)
		}
	}
}

func TestControlsPanelShowsAudioRecoveryActions(t *testing.T) {
	m := Model{
		controlsVisible: true,
		controlTab:      controlTabAudio,
		algo:            "Ambient",
		seed:            42,
		volume:          70,
		themes:          []ColorTheme{DefaultTheme()},
	}
	panel := controlsPanel(m, 100, 22, DefaultTheme())
	for _, want := range []string{"CONTROL CENTER", "retry live audio", "render-only fallback", "backend"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("audio controls panel missing %q:\n%s", want, panel)
		}
	}
}

func TestControlsPanelShowsTrackStructureInspector(t *testing.T) {
	m := Model{
		controlsVisible: true,
		controlTab:      controlTabDebug,
		algo:            "Lofi",
		seed:            42,
		debug:           gen.DebugStatus{Section: "intro", Chord: "Dm9"},
		activeTrackID:   "lofi/demo",
		tracks: []TrackNavEntry{{
			ID:           "lofi/demo",
			Style:        "lofi",
			Substyle:     "dusty-rhodes",
			Title:        "Demo Track",
			SectionCount: 3,
			EventCount:   4,
			Complexity:   "arranged",
			Ensemble:     []string{"ep", "bass", "drums", "reed"},
			Structure: []TrackNavSection{
				{ID: "intro", Label: "Intro", Harmony: "Dm9 G13", Events: []string{"pickup"}, RoleNames: []string{"ep", "bass"}},
				{ID: "head", Label: "Head", Harmony: "Bbmaj9 C13", Events: []string{"fill"}, RoleNames: []string{"ep", "reed", "drums"}},
			},
		}},
		themes: []ColorTheme{DefaultTheme()},
	}
	panel := controlsPanel(m, 100, 24, DefaultTheme())
	for _, want := range []string{"TRACK FORM", "Demo Track", "live  Intro", "pickup", "ep · bass"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("track structure inspector missing %q:\n%s", want, panel)
		}
	}
}

func TestSplashPanelShowsOnboarding(t *testing.T) {
	m := Model{
		splashVisible: true,
		themes:        []ColorTheme{DefaultTheme()},
	}
	panel := splashPanel(m, 90, 18, DefaultTheme())
	for _, want := range []string{"TERMUS", "Play", "Open", "[m] control center", "Press any key"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("splash panel missing %q:\n%s", want, panel)
		}
	}
}

func TestSplashPanelShowsStartupLoading(t *testing.T) {
	m := Model{
		splashVisible:  true,
		startupLoading: true,
		startupTitle:   "Loading MAX palette · Dusty Swing · jazz",
		startupDetail:  "ready 1/2 · last sgm",
		startupPercent: 0.5,
		themes:         []ColorTheme{DefaultTheme()},
	}
	panel := splashPanel(m, 90, 18, DefaultTheme())
	for _, want := range []string{"TERMUS", "Play", "Open"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("onboarding splash missing %q:\n%s", want, panel)
		}
	}
}

func TestLoadSelectedTrackUsesStartupLoaderAndSwapsOnResult(t *testing.T) {
	cmdr := &tuiCommanderStub{}
	m := New(nil, cmdr, "Tracks", "Cmin", 42, 70).
		WithSwitcher([]gen.AlgoSpec{{Name: "lofi", Display: "Lofi"}}, 0, func(spec gen.AlgoSpec, seed int64) gen.Algorithm {
			return &tuiAlgoStub{name: spec.Name}
		}).
		WithTrackBrowser([]TrackNavEntry{{ID: "lofi/demo", Style: "lofi", Title: "Demo Track"}}, func(id string) (*gen.Playlist, string, error) {
			return &gen.Playlist{
				Name: "Demo",
				Tracks: []gen.Track{{
					Spec:     gen.AlgoSpec{Name: "lofi", Display: "Lofi"},
					Seed:     88,
					Duration: 4 * time.Second,
					Title:    "Demo Section",
				}},
			}, "album-side", nil
		}, true)

	loadCmd := m.loadSelectedTrack()
	if loadCmd == nil {
		t.Fatal("expected track load command")
	}
	if !m.startupLoading || m.startupTitle != "Demo Track" {
		t.Fatalf("track load should raise startup loader, got loading=%v title=%q", m.startupLoading, m.startupTitle)
	}
	msg := loadCmd()
	gotModel, _ := m.Update(msg)
	got := gotModel.(Model)
	if got.trackVisible {
		t.Fatal("track browser should close after successful load")
	}
	if got.activeTrackID != "lofi/demo" || got.listeningMode != "album-side" {
		t.Fatalf("unexpected loaded track state: id=%q mode=%q", got.activeTrackID, got.listeningMode)
	}
	if len(cmdr.swaps) != 1 {
		t.Fatalf("expected one algorithm swap, got %d", len(cmdr.swaps))
	}
}

func TestStartupLoadingViewShowsBrailleStyleProgress(t *testing.T) {
	m := Model{
		width:          90,
		height:         18,
		startupLoading: true,
		startupTitle:   "Loading MAX palette · Dusty Swing · jazz",
		startupDetail:  "ready 1/2 · last sgm",
		startupPercent: 0.5,
		themes:         []ColorTheme{DefaultTheme()},
	}
	view := startupLoadingView(m, 90, 18, DefaultTheme(), time.Unix(0, 0))
	for _, want := range []string{"Loading MAX palette", "50%", "ready 1/2"} {
		if !strings.Contains(view, want) {
			t.Fatalf("startup loading view missing %q:\n%s", want, view)
		}
	}
	if !strings.ContainsAny(view, "⠄⡀⠤⠶⠒⠂⠦") {
		t.Fatalf("startup loading view should use braille texture:\n%s", view)
	}
}

func TestStartupLoadingBlocksDismissal(t *testing.T) {
	cmd := &tuiCommanderStub{}
	m := Model{
		cmd:            cmd,
		splashVisible:  true,
		startupLoading: true,
		themes:         []ColorTheme{DefaultTheme()},
	}
	next, _ := m.Update(keyMsg("c"))
	got := next.(Model)
	if !got.splashVisible || !got.startupLoading {
		t.Fatal("startup loading should keep splash visible until loading completes")
	}
}

func TestStartupLoadingViewBypassesChrome(t *testing.T) {
	m := Model{
		width:          90,
		height:         18,
		startupLoading: true,
		startupTitle:   "Loading MAX palette · Dusty Swing · jazz",
		startupDetail:  "loading sgm, tyros4",
		startupPercent: 0.2,
		themes:         []ColorTheme{DefaultTheme()},
	}
	view := m.View()
	if strings.Contains(view, "?  m") || strings.Contains(view, "termus ·") {
		t.Fatalf("startup loading should bypass normal chrome:\n%s", view)
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

func TestHiddenGlobalShortcutsStillWork(t *testing.T) {
	cmd := &tuiCommanderStub{}
	m := Model{
		cmd:      cmd,
		themeIdx: 0,
		themes:   []ColorTheme{DefaultTheme(), Themes[1]},
		seed:     42,
	}
	next, _ := m.Update(keyMsg("c"))
	got := next.(Model)
	if got.themeIdx != 1 {
		t.Fatalf("theme shortcut should still work, got themeIdx=%d", got.themeIdx)
	}
	next, _ = got.Update(keyMsg("z"))
	got = next.(Model)
	if !got.reducedChrome {
		t.Fatal("zen shortcut should still toggle reduced chrome")
	}
	next, _ = got.Update(keyMsg("l"))
	got = next.(Model)
	if !got.libraryVisible {
		t.Fatal("library shortcut should still open saved-seed library")
	}
	next, _ = got.Update(keyMsg("l"))
	got = next.(Model)
	if got.libraryVisible {
		t.Fatal("library shortcut should still close saved-seed library")
	}
}

func TestVisualShortcutCyclesWithoutControlCenter(t *testing.T) {
	m := Model{
		visualIdx: 0,
		themes:    []ColorTheme{DefaultTheme()},
	}
	next, _ := m.Update(keyMsg("C"))
	got := next.(Model)
	if got.visualIdx != 1 {
		t.Fatalf("visual shortcut should advance visual, got %d", got.visualIdx)
	}
	if !got.visualTransitionActive(time.Now()) {
		t.Fatal("visual shortcut should trigger transition")
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
	for _, want := range []string{"SAVED SEEDS", "Night Drift · ambient", "42", "[enter] load"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("library panel missing %q:\n%s", want, panel)
		}
	}
}

func TestInspectorPanelShowsTrackState(t *testing.T) {
	m := Model{
		algo:             "Jazz",
		keyName:          "Cmin",
		seed:             42,
		inspectorVisible: true,
		seedA:            &seedBookmark{Spec: gen.AlgoSpec{Name: "ambient", Display: "Ambient"}, Seed: 11},
		seedB:            &seedBookmark{Spec: gen.AlgoSpec{Name: "jazz", Display: "Jazz"}, Seed: 12},
		kept:             map[string]seedBookmark{"jazz:42": {Spec: gen.AlgoSpec{Name: "jazz", Display: "Jazz"}, Seed: 42}},
		debug:            gen.DebugStatus{Bar: 3, Section: "A", Chord: "Dm7", Preset: "general"},
		themes:           []ColorTheme{DefaultTheme()},
	}
	panel := inspectorPanel(m, 90, 18, DefaultTheme())
	for _, want := range []string{"TRACK INSPECTOR", "Jazz · Cmin", "42", "Ambient/11", "Jazz/12", "bar 3", "[e] export drawer"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("inspector panel missing %q:\n%s", want, panel)
		}
	}
}

func TestExportPanelShowsArtifactActions(t *testing.T) {
	m := Model{
		algo:          "Ambient",
		seed:          42,
		exportVisible: true,
		exporter:      &ExportController{Seconds: 60},
		themes:        []ColorTheme{DefaultTheme()},
	}
	panel := exportPanel(m, 90, 16, DefaultTheme())
	for _, want := range []string{"EXPORT", "[w] WAV 60s", "[m] MIDI 60s", "[t] stems 60s"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("export panel missing %q:\n%s", want, panel)
		}
	}
}

func TestStartExportRunsCallback(t *testing.T) {
	specs := []gen.AlgoSpec{{Name: "ambient", Display: "Ambient"}}
	m := Model{
		genres:   specs,
		genreIdx: 0,
		seed:     42,
		exporter: &ExportController{
			WAV: func(spec gen.AlgoSpec, seed int64) (string, error) {
				return fmt.Sprintf("%s-%d.wav", spec.Name, seed), nil
			},
		},
	}
	cmd := m.startExport("wav")
	if cmd == nil {
		t.Fatal("startExport returned nil cmd")
	}
	msg := cmd().(exportResultMsg)
	if msg.path != "ambient-42.wav" || msg.err != nil {
		t.Fatalf("export result = %+v", msg)
	}
}

func TestMeterSummaryDetectsClip(t *testing.T) {
	peak, clipped := meterSummary([]float64{0.2, -0.99, 0.3})
	if peak < 0.99 || !clipped {
		t.Fatalf("meterSummary = (%v, %v), want clipped peak", peak, clipped)
	}
}

func TestCompactBottomBarUsesMinimalHints(t *testing.T) {
	m := Model{
		algo:   "Ambient",
		volume: 70,
		themes: []ColorTheme{DefaultTheme()},
	}
	bar := bottomBar(m, 64, DefaultTheme(), true)
	if !strings.Contains(bar, "?  m") || !strings.Contains(bar, "Ambient") {
		t.Fatalf("compact bottom bar missing minimal hints: %q", bar)
	}
	if strings.Contains(bar, "[l] library") || strings.Contains(bar, "[i] inspect") || strings.Contains(bar, "[q]") {
		t.Fatalf("compact bottom bar should omit extended chrome: %q", bar)
	}
}

func TestReducedChromeBottomBarShowsReturnHint(t *testing.T) {
	m := Model{
		algo:          "Ambient",
		volume:        70,
		reducedChrome: true,
		themes:        []ColorTheme{DefaultTheme()},
	}
	bar := bottomBar(m, 90, DefaultTheme(), false)
	if !strings.Contains(bar, "Ambient") || !strings.Contains(bar, "?") {
		t.Fatalf("reduced chrome bar missing minimal chrome: %q", bar)
	}
	if strings.Contains(bar, "[q]") || strings.Contains(bar, "[z]") || strings.Contains(bar, "70%") {
		t.Fatalf("reduced chrome bar should stay minimal: %q", bar)
	}
}

func TestRenderVolumeLineShowsCenteredFeedback(t *testing.T) {
	m := Model{volume: 70}
	line := renderVolumeLine(m, 40, DefaultTheme())
	if strings.Contains(line, "%") {
		t.Fatalf("volume line should not show numeric label: %q", line)
	}
	if !strings.Contains(line, "─") {
		t.Fatalf("volume line should render as a line: %q", line)
	}
}

func keyMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	bar := bottomBar(m, 120, DefaultTheme())
	if !strings.Contains(bar, "audio: starting...") {
		t.Fatalf("bottom bar missing status: %q", bar)
	}
	if !strings.Contains(bar, "[space] play") || !strings.Contains(bar, "[?] help") {
		t.Fatalf("bottom bar should name the core keys: %q", bar)
	}
	if !strings.Contains(bar, "[z] zen") {
		t.Fatalf("bottom bar should expose the zen toggle on the right: %q", bar)
	}
	if strings.Contains(bar, "[l] library") || strings.Contains(bar, "[i] inspect") {
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
	if !strings.Contains(bar, "JAZZ") || !strings.Contains(bar, "seed 42") {
		t.Fatalf("top bar missing station identity: %q", bar)
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
	if !strings.Contains(bar, "NIGHT DRIFT") || !strings.Contains(bar, "ambient") {
		t.Fatalf("top bar should surface both station and canonical algo name: %q", bar)
	}
	if !strings.Contains(bar, "Cmin") {
		t.Fatalf("top bar should surface the musical key: %q", bar)
	}
}

func TestPlaybackBarShowsNarrationAndMeter(t *testing.T) {
	now := time.Now()
	m := Model{
		recording:       true,
		listeningMode:   "hour stream",
		startedAt:       now.Add(-95 * time.Second),
		recordStartedAt: now.Add(-17 * time.Second),
		debug: gen.DebugStatus{
			Movement: "develop", Episode: 2, Section: "A'",
			Bar: 17, Chord: "G7", NextChord: "Cmaj7",
		},
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
	for _, want := range []string{"movement develop", "episode 2", "section A'", "bar 17", "G7 → Cmaj7", "rec 00:17", "lvl"} {
		if !strings.Contains(bar, want) {
			t.Fatalf("playback bar missing %q: %q", want, bar)
		}
	}
	// Playlist timing clutter moved off the default chrome.
	for _, unwanted := range []string{"live ", "track 00:32", "fade "} {
		if strings.Contains(bar, unwanted) {
			t.Fatalf("playback bar should drop timing clutter %q: %q", unwanted, bar)
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
	m := Model{width: 118, height: 32}
	out := helpPanel(m, 118, 32, DefaultTheme())
	for _, want := range []string{
		"TERMUS HELP", "PLAYBACK", "VIEW", "OPEN", "SEEDS", "INSIDE PANELS", "GLOBAL",
		"n / p", "[ ]", "a / b", "k / x", "zen — scope only",
		"every key still works everywhere",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("help panel missing %q", want)
		}
	}
}

// TestHelpPanelNeverExceedsSize sweeps every runnable width at four heights,
// asserting the bordered help overlay never overflows its allotted body
// area: every line <= w, total rows <= h. w=40 forces the single-column
// fallback (innerW = bodyW-6 < 64), so this also exercises that path.
func TestHelpPanelNeverExceedsSize(t *testing.T) {
	theme := DefaultTheme()
	for _, h := range []int{10, 14, 18, 32} {
		for w := 40; w <= 200; w++ {
			m := Model{width: w, height: h}
			out := helpPanel(m, w, h, theme)
			lines := strings.Split(out, "\n")
			if len(lines) > h {
				t.Fatalf("help panel is %d rows tall at w=%d h=%d, want <= %d\n%s",
					len(lines), w, h, h, out)
			}
			for i, line := range lines {
				if got := lipgloss.Width(line); got > w {
					t.Fatalf("help panel line %d overflows at w=%d h=%d: rendered %d cols\n%q",
						i, w, h, got, line)
				}
			}
		}
	}
}

// TestHelpPanelKeepsAllGroupsOnSmallTerminals guards against clipLines
// eating the tail of the group list — and, worse, the quit/close keys —
// on small-but-common terminal sizes. helpPanel's h argument is the
// play-view body height, not the raw window height: View() subtracts
// ~4 rows of chrome (station header, footer, etc.) before calling
// helpPanel, so an 80x24 stock terminal actually renders the panel at
// h=20, and a 40x10 terminal at h=8. Using the raw window heights here
// would understate how tight real terminals get and let this regress
// again.
//
// 80 wide clears the two-column threshold (innerW = bodyW-6 = 74 >=
// twoColMinInnerW), so all six group headers must render even with the
// h=20 chrome-adjusted height. 40x8 is small enough to force the
// single-column fallback and a tight clipLines budget that clips the
// body's own GLOBAL group; "quit" must still be reachable because
// helpTitleRow's tiered hint pins at least "[q] quit" into the first
// line of inner content, which clipLines never trims.
func TestHelpPanelKeepsAllGroupsOnSmallTerminals(t *testing.T) {
	theme := DefaultTheme()

	m80 := Model{width: 80, height: 24}
	out80 := helpPanel(m80, 80, 20, theme)
	for _, want := range []string{"PLAYBACK", "VIEW", "OPEN", "SEEDS", "INSIDE PANELS", "GLOBAL", "quit", "close"} {
		if !strings.Contains(out80, want) {
			t.Fatalf("help panel at 80x20 (80x24 terminal minus chrome) missing %q:\n%s", want, out80)
		}
	}

	m40 := Model{width: 40, height: 10}
	out40 := helpPanel(m40, 40, 8, theme)
	if !strings.Contains(out40, "quit") {
		t.Fatalf("help panel at 40x8 (40x10 terminal minus chrome) missing %q:\n%s", "quit", out40)
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
	panel := controlsPanel(m, 118, 32, DefaultTheme())
	for _, want := range []string{
		"CONTROL CENTER", "MUSIC",
		"▌ music", "  now", "  look", "  seeds", "  library", "  export", "  audio", "  debug",
		"density", "brightness", "reverb", "[tab] section", "3 of 8",
	} {
		if !strings.Contains(panel, want) {
			t.Fatalf("controls panel missing %q:\n%s", want, panel)
		}
	}
	if strings.ContainsAny(panel, "╭╰│") {
		t.Fatalf("full-screen control center should not draw a border:\n%s", panel)
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
	panel := controlsPanel(m, 118, 32, DefaultTheme())
	for _, want := range []string{"CONTROL CENTER", "AUDIO", "backend health and recovery", "retry live audio", "render-only fallback", "backend", "7 of 8"} {
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
	panel := controlsPanel(m, 118, 32, DefaultTheme())
	for _, want := range []string{"DEBUG", "TRACK FORM", "Demo Track", "live  Intro", "pickup", "ep · bass", "8 of 8"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("track structure inspector missing %q:\n%s", want, panel)
		}
	}
}

func TestSplashPanelShowsOnboarding(t *testing.T) {
	m := Model{
		width:         90,
		height:        18,
		splashVisible: true,
		themes:        []ColorTheme{DefaultTheme()},
	}
	panel := splashScreen(m, 90, 18, DefaultTheme(), time.Unix(0, 0))
	for _, want := range []string{"█", "a terminal music instrument", "[enter] begin", "[t] authored tracks"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("splash panel missing %q:\n%s", want, panel)
		}
	}
}

// TestLoadSelectedTrackRoutesAITrackThroughEngineSwitcher verifies that
// picking an AI track from the browser no longer flashes the legacy
// "relaunch with --engine acestep" status. Instead, the model invokes the
// wired-up EngineSwitcher and returns its tea.Cmd to bubbletea for execution.
func TestLoadSelectedTrackRoutesAITrackThroughEngineSwitcher(t *testing.T) {
	cmdr := &tuiCommanderStub{}
	var switcherCalled bool
	var capturedEngine string
	m := New(nil, cmdr, "Tracks", "Cmin", 42, 70).
		WithSwitcher([]gen.AlgoSpec{{Name: "ambient", Display: "Ambient"}}, 0, func(spec gen.AlgoSpec, seed int64) gen.Algorithm {
			return &tuiAlgoStub{name: spec.Name}
		}).
		WithTrackBrowser([]TrackNavEntry{{ID: "ai/demo", Style: "ambient", Title: "AI Demo", Engine: "acestep"}}, func(id string) (*gen.Playlist, string, error) {
			t.Fatalf("trackLoader should NOT be called for an AI track")
			return nil, "", nil
		}, true).
		WithEngineSwitcher(func(id, title, engine string) tea.Cmd {
			switcherCalled = true
			capturedEngine = engine
			return func() tea.Msg {
				return TrackEngineSwitchMsg{EntryID: id, EntryTitle: title, Engine: engine}
			}
		})

	loadCmd := m.loadSelectedTrack()
	if loadCmd == nil {
		t.Fatal("expected engine-switch command for AI track")
	}
	if !switcherCalled {
		t.Fatal("EngineSwitcher should be invoked for AI tracks")
	}
	if capturedEngine != "acestep" {
		t.Fatalf("EngineSwitcher engine arg = %q, want acestep", capturedEngine)
	}
	if !m.startupLoading {
		t.Fatalf("track selection should raise the startup loader")
	}
	if m.activeTrackID != "ai/demo" {
		t.Fatalf("activeTrackID = %q, want ai/demo", m.activeTrackID)
	}
	if m.trackVisible {
		t.Fatal("track browser should close after dispatching engine switch")
	}
}

// TestLoadSelectedTrackAITrackFallbackWhenNoSwitcher verifies the legacy
// SP25 behaviour is preserved when no EngineSwitcher is wired: the model
// flashes the "relaunch" status rather than crashing.
func TestLoadSelectedTrackAITrackFallbackWhenNoSwitcher(t *testing.T) {
	cmdr := &tuiCommanderStub{}
	m := New(nil, cmdr, "Tracks", "Cmin", 42, 70).
		WithSwitcher([]gen.AlgoSpec{{Name: "ambient", Display: "Ambient"}}, 0, func(spec gen.AlgoSpec, seed int64) gen.Algorithm {
			return &tuiAlgoStub{name: spec.Name}
		}).
		WithTrackBrowser([]TrackNavEntry{{ID: "ai/demo", Style: "ambient", Title: "AI Demo", Engine: "acestep"}}, func(id string) (*gen.Playlist, string, error) {
			return nil, "", nil
		}, true)

	if cmd := m.loadSelectedTrack(); cmd != nil {
		t.Fatalf("expected nil cmd when no EngineSwitcher is wired (legacy fallback)")
	}
	if m.status == "" {
		t.Fatalf("expected legacy status flash for AI track without switcher")
	}
}

// TestTrackEngineSwitchMsgDismissesLoaderOnSF2Success verifies the model's
// TrackEngineSwitchMsg handler clears the startup loader once an SF2 hot-
// switch reports success. AI tracks keep the loader open until
// ACEStepReadyMsg arrives.
func TestTrackEngineSwitchMsgDismissesLoaderOnSF2Success(t *testing.T) {
	cmdr := &tuiCommanderStub{}
	m := Model{
		cmd:            cmdr,
		startupLoading: true,
		splashVisible:  true,
	}
	updated, _ := m.Update(TrackEngineSwitchMsg{EntryID: "lofi/demo", EntryTitle: "Demo", Engine: "sf2"})
	got := updated.(Model)
	if got.startupLoading {
		t.Fatalf("SF2 engine switch should dismiss the loader")
	}
	if got.activeTrackID != "lofi/demo" {
		t.Fatalf("activeTrackID = %q, want lofi/demo", got.activeTrackID)
	}
}

// TestTrackEngineSwitchMsgKeepsLoaderOpenForACEStep verifies the loader stays
// up after an ACE-Step engine-switch success message; the streamer's
// ACEStepReadyMsg dismisses it.
func TestTrackEngineSwitchMsgKeepsLoaderOpenForACEStep(t *testing.T) {
	cmdr := &tuiCommanderStub{}
	m := Model{
		cmd:            cmdr,
		startupLoading: true,
		splashVisible:  true,
	}
	updated, _ := m.Update(TrackEngineSwitchMsg{EntryID: "ai/demo", EntryTitle: "Demo", Engine: "acestep"})
	got := updated.(Model)
	if !got.startupLoading {
		t.Fatalf("ACE-Step engine switch should keep the loader open until ACEStepReadyMsg")
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

// TestStartupLoadingProgressOnSplashScreen covers the loading half of the
// merged splash: the wordmark stays, and the braille bar, percent, loader
// title and current detail all report progress.
func TestStartupLoadingProgressOnSplashScreen(t *testing.T) {
	m := Model{
		width:          90,
		height:         18,
		splashVisible:  true,
		startupLoading: true,
		startupTitle:   "Loading MAX palette · Dusty Swing · jazz",
		startupDetail:  "ready 1/2 · last sgm",
		startupPercent: 0.5,
		themes:         []ColorTheme{DefaultTheme()},
	}
	view := splashScreen(m, 90, 18, DefaultTheme(), time.Unix(0, 0))
	for _, want := range []string{"█", "Loading MAX palette", "50%", "ready 1/2"} {
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
	if strings.Contains(view, "[space] play") || strings.Contains(view, "seed ") {
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
	// Below w=60 the full hint row cannot fit alongside the status gutter,
	// so the footer degrades to the minimal pair.
	bar := bottomBar(m, 58, DefaultTheme())
	if !strings.Contains(bar, "[?] help") || !strings.Contains(bar, "[z]") {
		t.Fatalf("narrow bottom bar should keep the minimal hints: %q", bar)
	}
	if strings.Contains(bar, "[space] play") || strings.Contains(bar, "[m] control") {
		t.Fatalf("narrow bottom bar should shed the full hint row: %q", bar)
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
	bar := bottomBar(m, 90, DefaultTheme())
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

func TestACEStepInstallProgressMsgUpdatesLoader(t *testing.T) {
	m := Model{themes: []ColorTheme{DefaultTheme()}}
	m.applyACEStepInstall(ACEStepInstallProgressMsg{
		Phase:   "install:model",
		Title:   "Setting up AI engine",
		Detail:  "downloading model",
		Percent: 0.55,
	})
	if !m.startupLoading {
		t.Fatal("startup loader should be active after install progress")
	}
	if m.startupTitle != "Setting up AI engine" {
		t.Fatalf("title = %q", m.startupTitle)
	}
	if m.startupDetail != "downloading model" {
		t.Fatalf("detail = %q", m.startupDetail)
	}
	if m.startupPercent < 0.54 || m.startupPercent > 0.56 {
		t.Fatalf("percent = %f", m.startupPercent)
	}
}

func TestACEStepStatusMsgUpdatesLoader(t *testing.T) {
	m := Model{themes: []ColorTheme{DefaultTheme()}}
	m.applyACEStepStatus(ACEStepStatusMsg{
		Phase:   "loading-model",
		Title:   "Starting AI engine",
		Detail:  "warming up MLX",
		Percent: 0.78,
	})
	if !m.startupLoading {
		t.Fatal("status msg should keep startup loader active")
	}
	if m.startupTitle != "Starting AI engine" {
		t.Fatalf("title = %q", m.startupTitle)
	}
	if m.startupDetail != "warming up MLX" {
		t.Fatalf("detail = %q", m.startupDetail)
	}
}

func TestACEStepReadyMsgDismissesLoader(t *testing.T) {
	m := Model{startupLoading: true, splashVisible: true, themes: []ColorTheme{DefaultTheme()}}
	m.applyACEStepReady(ACEStepReadyMsg{Detail: "engine ready"})
	if m.startupLoading {
		t.Fatal("startupLoading should be cleared after ready")
	}
	if m.splashVisible {
		t.Fatal("splashVisible should be cleared after ready")
	}
}

func TestACEStepRenderingMsgTogglesCornerIndicator(t *testing.T) {
	m := Model{themes: []ColorTheme{DefaultTheme()}}
	m.applyACEStepRendering(ACEStepRenderingMsg{Seq: 2, Detail: "generating track 3"})
	if !m.aceRenderActive {
		t.Fatal("rendering should be marked active")
	}
	if m.aceRenderSeq != 2 || m.aceRenderDetail != "generating track 3" {
		t.Fatalf("seq=%d detail=%q", m.aceRenderSeq, m.aceRenderDetail)
	}
	m.applyACEStepRendering(ACEStepRenderingMsg{Seq: 2, Done: true})
	if m.aceRenderActive {
		t.Fatal("rendering should be cleared on Done")
	}
}

func TestACEStepRenderingMsgShowsInPlaybackBar(t *testing.T) {
	m := Model{
		algo:            "ACE-Step",
		volume:          70,
		aceRenderActive: true,
		aceRenderSeq:    1,
		aceRenderDetail: "generating track 2",
		startedAt:       time.Now().Add(-30 * time.Second),
		themes:          []ColorTheme{DefaultTheme()},
	}
	bar := playbackBar(m, 100, DefaultTheme(), make([]float64, 200), false)
	if !strings.Contains(bar, "generating track 2") {
		t.Fatalf("playback bar should surface rendering detail: %q", bar)
	}
}

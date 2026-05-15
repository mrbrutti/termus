package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
)

// BuildAlgoFn constructs a fresh Algorithm for the given spec. main.go closes
// over the loaded SoundFont (or nil) and any per-build wiring like IR setup.
type BuildAlgoFn func(spec gen.AlgoSpec, seed int64) gen.Algorithm

type seedBookmark struct {
	Spec gen.AlgoSpec
	Seed int64
}

// Model is the bubbletea model for termus.
type Model struct {
	width, height int

	ring    *scope.Ring
	cmd     audio.Commander
	algo    string
	debug   gen.DebugStatus
	keyName string
	seed    int64

	volume           int
	paused           bool
	recording        bool
	debugVisible     bool
	helpVisible      bool
	libraryVisible   bool
	inspectorVisible bool
	exportVisible    bool
	exportBusy       bool
	status           string
	statusTTL        time.Time
	stickyStatus     string

	themeIdx  int // index into Themes
	visualIdx int // index into Visuals
	themes    []ColorTheme
	ui        AdaptiveUI

	// Algorithm switching ([n]/[p]).
	genres     []gen.AlgoSpec // ordered list of switchable algorithms
	genreIdx   int            // current index into genres
	buildFn    BuildAlgoFn    // closure used to construct a new algorithm
	seedA      *seedBookmark
	seedB      *seedBookmark
	kept       map[string]seedBookmark
	savedSeeds []savedSeedRecord
	libraryIdx int
	exporter   *ExportController

	// Playlist auto-advance.
	playlist        *gen.Playlist
	playlistIdx     int // index of currently-playing track
	trackStartedAt  time.Time
	nextTrackAt     time.Time // when to advance to the next track
	playlistFade    int       // crossfade length in audio frames (44.1 kHz)
	startedAt       time.Time
	recordStartedAt time.Time
}

// New constructs a Model. keyName is e.g. "Cmin".
func New(ring *scope.Ring, cmd audio.Commander, algo, keyName string, seed int64, initialVol int) Model {
	ui := DetectAdaptiveUI()
	savedSeeds, _ := loadSavedSeedRecords()
	return Model{
		ring:       ring,
		cmd:        cmd,
		algo:       algo,
		debug:      cmd.DebugStatus(),
		keyName:    keyName,
		seed:       seed,
		volume:     initialVol,
		ui:         ui,
		themes:     append([]ColorTheme(nil), ui.Themes...),
		themeIdx:   ui.DefaultThemeIdx,
		kept:       recordsToBookmarks(savedSeeds),
		savedSeeds: savedSeeds,
		startedAt:  time.Now(),
	}
}

// WithSwitcher enables in-app algorithm switching. genres is the ordered list
// the user cycles through; startIdx is the index of the algorithm currently
// playing; buildFn constructs a fresh Algorithm for a chosen spec.
func (m Model) WithSwitcher(genres []gen.AlgoSpec, startIdx int, buildFn BuildAlgoFn) Model {
	m.genres = genres
	m.genreIdx = startIdx
	m.buildFn = buildFn
	return m
}

// WithDebug controls whether the dedicated debug inspector starts visible.
func (m Model) WithDebug(visible bool) Model {
	m.debugVisible = visible
	return m
}

func (m Model) WithExportController(exporter *ExportController) Model {
	m.exporter = exporter
	return m
}

// WithPlaylist enables playlist auto-advance. The model walks through the
// playlist's tracks, swapping the algorithm at each track's Duration boundary
// with a crossfade of fadeFrames samples. buildFn must be set via
// WithSwitcher first so the model knows how to construct algorithms.
func (m Model) WithPlaylist(p *gen.Playlist, startIdx int, fadeFrames int) Model {
	m.playlist = p
	m.playlistIdx = startIdx
	m.playlistFade = fadeFrames
	if p != nil && startIdx < len(p.Tracks) {
		m.trackStartedAt = time.Now()
		m.nextTrackAt = time.Now().Add(p.Tracks[startIdx].Duration)
	}
	return m
}

// advancePlaylist moves to the next track in the playlist (wrapping) and
// crossfades into it. Re-arms nextTrackAt for the new track's duration.
func (m *Model) advancePlaylist() {
	if m.playlist == nil || m.buildFn == nil || len(m.playlist.Tracks) == 0 {
		return
	}
	m.playlistIdx = (m.playlistIdx + 1) % len(m.playlist.Tracks)
	track := m.playlist.Tracks[m.playlistIdx]
	algo := m.buildFn(track.Spec, track.Seed)
	m.cmd.SwapAlgorithmFade(algo, m.playlistFade)
	m.algo = track.Spec.Display
	m.seed = track.Seed
	m.trackStartedAt = time.Now()
	m.nextTrackAt = time.Now().Add(track.Duration)
	m.flashStatus(fmt.Sprintf("▶ %d/%d %s",
		m.playlistIdx+1, len(m.playlist.Tracks), track.Spec.Display), 3*time.Second)

	// Keep the genre cycle index in sync if this track matches a genre.
	for i, g := range m.genres {
		if g.Name == track.Spec.Name {
			m.genreIdx = i
			break
		}
	}
}

// switchAlgo cycles the current algorithm by step (+1 or -1) and asks the
// audio thread to swap in a freshly-built instance.
func (m *Model) switchAlgo(step int) {
	if len(m.genres) == 0 || m.buildFn == nil {
		return
	}
	m.genreIdx = (m.genreIdx + step + len(m.genres)) % len(m.genres)
	spec := m.genres[m.genreIdx]
	algo := m.buildFn(spec, m.seed)
	m.cmd.SwapAlgorithm(algo)
	m.algo = spec.Display
	m.flashStatus("switched: "+spec.Display, 2*time.Second)
}

func (m Model) currentSpec() (gen.AlgoSpec, bool) {
	if m.genreIdx >= 0 && m.genreIdx < len(m.genres) {
		return m.genres[m.genreIdx], true
	}
	return gen.AlgoSpec{}, false
}

func (m *Model) swapToSeed(spec gen.AlgoSpec, seed int64, status string) {
	if m.playlist != nil || m.buildFn == nil {
		return
	}
	if seed < 0 {
		seed = 0
	}
	algo := m.buildFn(spec, seed)
	m.cmd.SwapAlgorithm(algo)
	m.algo = spec.Display
	m.seed = seed
	for i, g := range m.genres {
		if g.Name == spec.Name {
			m.genreIdx = i
			break
		}
	}
	m.flashStatus(status, 2*time.Second)
}

func (m *Model) browseSeed(delta int64) {
	if m.playlist != nil {
		return
	}
	spec, ok := m.currentSpec()
	if !ok {
		return
	}
	next := m.seed + delta
	if next < 0 {
		next = 0
	}
	m.swapToSeed(spec, next, fmt.Sprintf("seed: %d", next))
}

func (m *Model) storeSeed(slot string) {
	if m.playlist != nil {
		return
	}
	spec, ok := m.currentSpec()
	if !ok {
		return
	}
	bookmark := &seedBookmark{Spec: spec, Seed: m.seed}
	if slot == "A" {
		m.seedA = bookmark
	} else {
		m.seedB = bookmark
	}
	m.flashStatus(fmt.Sprintf("%s ← %s/%d", slot, spec.Display, m.seed), 2*time.Second)
}

func (m *Model) toggleSeedCompare() {
	if m.playlist != nil {
		return
	}
	switch {
	case m.seedA != nil && seedMatches(m.seedA, m) && m.seedB != nil:
		m.swapToSeed(m.seedB.Spec, m.seedB.Seed, fmt.Sprintf("B → %d", m.seedB.Seed))
	case m.seedB != nil && seedMatches(m.seedB, m) && m.seedA != nil:
		m.swapToSeed(m.seedA.Spec, m.seedA.Seed, fmt.Sprintf("A → %d", m.seedA.Seed))
	case m.seedA != nil:
		m.swapToSeed(m.seedA.Spec, m.seedA.Seed, fmt.Sprintf("A → %d", m.seedA.Seed))
	case m.seedB != nil:
		m.swapToSeed(m.seedB.Spec, m.seedB.Seed, fmt.Sprintf("B → %d", m.seedB.Seed))
	}
}

func (m *Model) keepSeed() {
	if m.playlist != nil {
		return
	}
	spec, ok := m.currentSpec()
	if !ok {
		return
	}
	if m.kept == nil {
		m.kept = make(map[string]seedBookmark)
	}
	key := bookmarkKey(spec, m.seed)
	m.kept[key] = seedBookmark{Spec: spec, Seed: m.seed}
	rec := savedSeedRecord{
		Algo:    spec.Name,
		Display: spec.Display,
		Seed:    m.seed,
		SavedAt: time.Now(),
	}
	m.savedSeeds = append([]savedSeedRecord{rec}, removeSavedSeedRecord(m.savedSeeds, spec.Name, m.seed)...)
	if err := saveSavedSeedRecords(m.savedSeeds); err != nil {
		m.flashStatus("keep saved locally failed", 3*time.Second)
		return
	}
	m.flashStatus(fmt.Sprintf("kept %s/%d (%d)", spec.Display, m.seed, len(m.kept)), 2*time.Second)
}

func (m *Model) toggleLibrary() {
	m.libraryVisible = !m.libraryVisible
	if m.libraryVisible {
		m.helpVisible = false
		m.inspectorVisible = false
		m.exportVisible = false
		if m.libraryIdx >= len(m.savedSeeds) {
			m.libraryIdx = maxInt(0, len(m.savedSeeds)-1)
		}
		m.flashStatus("library: on", 2*time.Second)
		return
	}
	m.flashStatus("library: off", 2*time.Second)
}

func (m *Model) toggleInspector() {
	m.inspectorVisible = !m.inspectorVisible
	if m.inspectorVisible {
		m.helpVisible = false
		m.libraryVisible = false
		m.exportVisible = false
		m.flashStatus("inspector: on", 2*time.Second)
		return
	}
	m.flashStatus("inspector: off", 2*time.Second)
}

func (m *Model) toggleExportDrawer() {
	if m.exporter == nil {
		m.flashStatus("export: unavailable", 2*time.Second)
		return
	}
	if m.exportBusy {
		return
	}
	m.exportVisible = !m.exportVisible
	if m.exportVisible {
		m.helpVisible = false
		m.libraryVisible = false
		m.inspectorVisible = false
		m.flashStatus("export: on", 2*time.Second)
		return
	}
	m.flashStatus("export: off", 2*time.Second)
}

func (m Model) currentExportTarget() (gen.AlgoSpec, int64, bool) {
	spec, ok := m.currentSpec()
	if !ok {
		return gen.AlgoSpec{}, 0, false
	}
	return spec, m.seed, true
}

func (m *Model) startExport(kind string) tea.Cmd {
	if m.exporter == nil || m.exportBusy {
		return nil
	}
	spec, seed, ok := m.currentExportTarget()
	if !ok {
		m.flashStatus("export: no active track", 2*time.Second)
		return nil
	}
	var fn func(gen.AlgoSpec, int64) (string, error)
	switch kind {
	case "wav":
		fn = m.exporter.WAV
	case "midi":
		fn = m.exporter.MIDI
	case "stems":
		fn = m.exporter.Stems
	}
	if fn == nil {
		m.flashStatus("export: unsupported", 2*time.Second)
		return nil
	}
	m.exportBusy = true
	m.flashStatus("exporting "+kind+"...", 3*time.Second)
	return runExport(kind, func() (string, error) {
		return fn(spec, seed)
	})
}

func (m *Model) moveLibrary(delta int) {
	if len(m.savedSeeds) == 0 {
		m.libraryIdx = 0
		return
	}
	m.libraryIdx = (m.libraryIdx + delta + len(m.savedSeeds)) % len(m.savedSeeds)
}

func (m *Model) recallLibrarySeed() {
	if len(m.savedSeeds) == 0 {
		return
	}
	rec := m.savedSeeds[m.libraryIdx]
	bookmark, label, ok := resolveSavedSeedRecord(rec)
	if !ok {
		m.flashStatus("saved algo unavailable: "+label, 3*time.Second)
		return
	}
	m.libraryVisible = false
	m.swapToSeed(bookmark.Spec, bookmark.Seed, fmt.Sprintf("saved → %s/%d", label, bookmark.Seed))
}

func (m *Model) deleteLibrarySeed() {
	if len(m.savedSeeds) == 0 {
		return
	}
	rec := m.savedSeeds[m.libraryIdx]
	m.savedSeeds = append([]savedSeedRecord(nil), removeSavedSeedRecord(m.savedSeeds, rec.Algo, rec.Seed)...)
	if spec, ok := gen.Resolve(rec.Algo); ok && m.kept != nil {
		delete(m.kept, bookmarkKey(spec, rec.Seed))
	}
	if m.libraryIdx >= len(m.savedSeeds) {
		m.libraryIdx = maxInt(0, len(m.savedSeeds)-1)
	}
	if err := saveSavedSeedRecords(m.savedSeeds); err != nil {
		m.flashStatus("library save failed", 3*time.Second)
		return
	}
	if len(m.savedSeeds) == 0 {
		m.flashStatus("library cleared", 2*time.Second)
		return
	}
	m.flashStatus("removed saved seed", 2*time.Second)
}

func (m *Model) rejectSeed() {
	if m.playlist != nil {
		return
	}
	spec, ok := m.currentSpec()
	if !ok {
		return
	}
	next := m.seed + 1
	m.swapToSeed(spec, next, fmt.Sprintf("reject → %d", next))
}

func seedMatches(mark *seedBookmark, m *Model) bool {
	return mark != nil && mark.Seed == m.seed && mark.Spec.Name == m.algoSpecName()
}

func (m Model) algoSpecName() string {
	if spec, ok := m.currentSpec(); ok {
		return spec.Name
	}
	return ""
}

func bookmarkKey(spec gen.AlgoSpec, seed int64) string {
	return fmt.Sprintf("%s:%d", spec.Name, seed)
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second/30, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) Init() tea.Cmd { return tick() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case audio.BackendState:
		m.applyAudioState(msg)
		return m, nil
	case exportResultMsg:
		m.exportBusy = false
		if msg.err != nil {
			m.flashStatus(msg.kind+" export failed: "+msg.err.Error(), 4*time.Second)
		} else {
			m.flashStatus(msg.kind+" → "+msg.path, 4*time.Second)
			m.exportVisible = false
		}
		return m, nil
	case tea.KeyMsg:
		if m.libraryVisible {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "l", "esc":
				m.toggleLibrary()
			case "up":
				m.moveLibrary(-1)
			case "down":
				m.moveLibrary(1)
			case "enter":
				m.recallLibrarySeed()
			case "backspace", "delete", "x":
				m.deleteLibrarySeed()
			}
			return m, nil
		}
		if m.exportVisible {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "e", "esc":
				if !m.exportBusy {
					m.toggleExportDrawer()
				}
			case "w":
				return m, m.startExport("wav")
			case "m":
				return m, m.startExport("midi")
			case "t":
				return m, m.startExport("stems")
			case "r":
				path, err := m.cmd.ToggleRecord()
				if err != nil {
					m.flashStatus("rec error: "+err.Error(), 3*time.Second)
					m.recording = false
				} else if path != "" {
					m.recording = true
					m.recordStartedAt = time.Now()
					m.flashStatus("rec → "+path, 3*time.Second)
				} else {
					m.recording = false
					m.recordStartedAt = time.Time{}
					m.flashStatus("rec stopped", 3*time.Second)
				}
			}
			return m, nil
		}
		action := matchKey(msg)
		if m.helpVisible && action != actionHelp && action != actionQuit {
			return m, nil
		}
		if m.inspectorVisible && action != actionInspector && action != actionQuit && action != actionExport {
			return m, nil
		}
		switch action {
		case actionQuit:
			return m, tea.Quit
		case actionPause:
			m.paused = !m.paused
			m.cmd.TogglePause()
		case actionVolUp:
			m.volume += 5
			if m.volume > 100 {
				m.volume = 100
			}
			m.cmd.SetVolume(m.volume)
		case actionVolDown:
			m.volume -= 5
			if m.volume < 0 {
				m.volume = 0
			}
			m.cmd.SetVolume(m.volume)
		case actionRecord:
			path, err := m.cmd.ToggleRecord()
			if err != nil {
				m.flashStatus("rec error: "+err.Error(), 3*time.Second)
				m.recording = false
			} else if path != "" {
				m.recording = true
				m.recordStartedAt = time.Now()
				m.flashStatus("rec → "+path, 3*time.Second)
			} else {
				m.recording = false
				m.recordStartedAt = time.Time{}
				m.flashStatus("rec stopped", 3*time.Second)
			}
		case actionTheme:
			if len(m.themes) > 1 {
				m.themeIdx = (m.themeIdx + 1) % len(m.themes)
				m.flashStatus("theme: "+m.themes[m.themeIdx].Name, 2*time.Second)
			}
		case actionVisual:
			m.visualIdx = (m.visualIdx + 1) % len(Visuals)
			m.flashStatus("visual: "+Visuals[m.visualIdx].Name, 2*time.Second)
		case actionDebug:
			m.debugVisible = !m.debugVisible
			if m.debugVisible {
				m.flashStatus("debug: on", 2*time.Second)
			} else {
				m.flashStatus("debug: off", 2*time.Second)
			}
		case actionHelp:
			m.helpVisible = !m.helpVisible
			if m.helpVisible {
				m.libraryVisible = false
				m.inspectorVisible = false
				m.exportVisible = false
				m.flashStatus("help: on", 2*time.Second)
			} else {
				m.flashStatus("help: off", 2*time.Second)
			}
		case actionLibrary:
			m.toggleLibrary()
		case actionInspector:
			m.toggleInspector()
		case actionExport:
			m.toggleExportDrawer()
		case actionNextAlgo:
			m.switchAlgo(1)
		case actionPrevAlgo:
			m.switchAlgo(-1)
		case actionNextTrack:
			if m.playlist != nil {
				m.advancePlaylist()
			}
		case actionPrevSeed:
			m.browseSeed(-1)
		case actionNextSeed:
			m.browseSeed(1)
		case actionStoreA:
			m.storeSeed("A")
		case actionStoreB:
			m.storeSeed("B")
		case actionToggleAB:
			m.toggleSeedCompare()
		case actionKeepSeed:
			m.keepSeed()
		case actionRejectSeed:
			m.rejectSeed()
		}
		return m, nil
	case tickMsg:
		m.debug = m.cmd.DebugStatus()
		if m.playlist != nil && !m.paused && time.Now().After(m.nextTrackAt) {
			m.advancePlaylist()
		}
		return m, tick()
	}
	return m, nil
}

func (m Model) View() string {
	if m.width < 40 || m.height < 10 {
		return centerBox(m.width, m.height, "terminal too small — resize to ≥ 40 × 10")
	}
	chromeH := 3 // top + now-playing + bottom bars
	if m.debugVisible {
		chromeH++
	}
	innerH := m.height - chromeH
	innerW := m.width

	// Snapshot scope and render with the active visual + theme.
	samples := make([]float64, innerW*2)
	m.ring.Snapshot(samples)
	theme := m.activeTheme()
	visual := Visuals[m.visualIdx]
	scopeStr := visual.Render(samples, innerW, innerH, RenderContext{
		Theme: theme,
	})

	top := topBar(m, innerW, theme)
	playback := playbackBar(m, innerW, theme, samples)
	bottom := bottomBar(m, innerW, theme)
	body := scopeStr
	if m.helpVisible {
		body = helpPanel(m, innerW, innerH, theme)
	} else if m.libraryVisible {
		body = libraryPanel(m, innerW, innerH, theme)
	} else if m.inspectorVisible {
		body = inspectorPanel(m, innerW, innerH, theme)
	} else if m.exportVisible {
		body = exportPanel(m, innerW, innerH, theme)
	}
	if m.debugVisible {
		debug := debugBar(m, innerW, theme)
		return lipgloss.JoinVertical(lipgloss.Left, top, playback, debug, body, bottom)
	}
	return lipgloss.JoinVertical(lipgloss.Left, top, playback, body, bottom)
}

func (m Model) activeTheme() ColorTheme {
	if len(m.themes) == 0 {
		return DefaultTheme()
	}
	if m.themeIdx < 0 || m.themeIdx >= len(m.themes) {
		return m.themes[0]
	}
	return m.themes[m.themeIdx]
}

func (m *Model) flashStatus(text string, ttl time.Duration) {
	m.status = text
	m.statusTTL = time.Now().Add(ttl)
}

func (m *Model) setStickyStatus(text string) {
	m.stickyStatus = text
}

func (m Model) currentStatus(now time.Time) string {
	if now.Before(m.statusTTL) {
		return m.status
	}
	return m.stickyStatus
}

func (m *Model) applyAudioState(state audio.BackendState) {
	switch state.Kind {
	case audio.BackendStateStarting:
		m.setStickyStatus(state.StatusText())
	case audio.BackendStateReady:
		m.setStickyStatus("")
		m.flashStatus(state.StatusText(), 2*time.Second)
	case audio.BackendStateNoDefaultDevice, audio.BackendStateHung,
		audio.BackendStateRenderOnly, audio.BackendStateInitFailed:
		m.setStickyStatus(state.StatusText())
	}
}

func topBar(m Model, w int, theme ColorTheme) string {
	var label string
	if m.playlist != nil {
		label = fmt.Sprintf("termus · %s · %d/%d %s · seed=%d",
			m.playlist.Name, m.playlistIdx+1, len(m.playlist.Tracks),
			m.algo, m.seed)
	} else {
		label = fmt.Sprintf("termus · %s · %s · seed=%d",
			m.algo, m.keyName, m.seed)
	}
	right := ""
	if m.recording {
		rec := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5b5b")).Render("● REC")
		if right == "" {
			right = rec
		} else {
			right += "  " + rec
		}
	}
	if seeds := m.seedSlotsLabel(); seeds != "" {
		seeds = lipgloss.NewStyle().Faint(true).Render(seeds)
		if right == "" {
			right = seeds
		} else {
			right = seeds + "  " + right
		}
	}
	if right != "" {
		label = trimToWidth(label, maxInt(0, w-lipgloss.Width(right)-1))
	}
	left := lipgloss.NewStyle().Foreground(theme.BarFg).Render(label)
	pad := w - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	return left + spaces(pad) + right
}

func (m Model) seedSlotsLabel() string {
	parts := make([]string, 0, 3)
	if m.seedA != nil {
		parts = append(parts, fmt.Sprintf("A=%d", m.seedA.Seed))
	}
	if m.seedB != nil {
		parts = append(parts, fmt.Sprintf("B=%d", m.seedB.Seed))
	}
	if len(m.kept) > 0 {
		parts = append(parts, fmt.Sprintf("keep=%d", len(m.kept)))
	}
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	}
	out := parts[0]
	for _, part := range parts[1:] {
		out += " · " + part
	}
	return out
}

func playbackBar(m Model, w int, theme ColorTheme, samples []float64) string {
	leftParts := []string{formatElapsed("live", time.Since(m.startedAt))}
	if m.playlist != nil && m.playlistIdx < len(m.playlist.Tracks) {
		track := m.playlist.Tracks[m.playlistIdx]
		leftParts = append(leftParts,
			fmt.Sprintf("track %s/%s", shortDuration(time.Since(m.trackStartedAt)), shortDuration(track.Duration)),
			fmt.Sprintf("next %s", shortDuration(time.Until(m.nextTrackAt))),
			fmt.Sprintf("fade %s", shortDuration(time.Duration(m.playlistFade)*time.Second/44100)),
		)
		if len(m.playlist.Tracks) > 0 {
			leftParts = append(leftParts, fmt.Sprintf("%d/%d", m.playlistIdx+1, len(m.playlist.Tracks)))
		}
	}
	if m.recording && !m.recordStartedAt.IsZero() {
		leftParts = append(leftParts, formatElapsed("rec", time.Since(m.recordStartedAt)))
	}
	leftText := trimToWidth(strings.Join(leftParts, " · "), maxInt(0, w-22))
	meter, clipped := meterSummary(samples)
	right := renderCompactMeter(theme, meter, clipped, 14)
	left := lipgloss.NewStyle().Faint(true).Render(leftText)
	pad := w - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	return left + spaces(pad) + right
}

func debugBar(m Model, w int, theme ColorTheme) string {
	status := gen.FormatDebugStatus(m.debug)
	if status == "" {
		status = "debug unavailable"
	}
	left := lipgloss.NewStyle().
		Foreground(theme.BarHi).
		Render("DEBUG")
	right := lipgloss.NewStyle().
		Faint(true).
		Render(trimToWidth(status, maxInt(0, w-lipgloss.Width(left)-3)))
	pad := w - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	return left + spaces(pad) + right
}

func meterSummary(samples []float64) (float64, bool) {
	peak := 0.0
	for _, s := range samples {
		if s < 0 {
			s = -s
		}
		if s > peak {
			peak = s
		}
	}
	return peak, peak >= 0.985
}

func renderCompactMeter(theme ColorTheme, peak float64, clipped bool, width int) string {
	if width < 4 {
		width = 4
	}
	if peak < 0 {
		peak = 0
	}
	if peak > 1 {
		peak = 1
	}
	filled := int(peak * float64(width))
	if peak > 0 && filled == 0 {
		filled = 1
	}
	if filled > width {
		filled = width
	}
	active := lipgloss.NewStyle().Foreground(theme.BarHi).Render(strings.Repeat("─", filled))
	idle := lipgloss.NewStyle().Faint(true).Render(strings.Repeat("─", width-filled))
	label := lipgloss.NewStyle().Foreground(theme.BarFg).Render("lvl")
	clip := lipgloss.NewStyle().Faint(true).Render("ok")
	if clipped {
		clip = lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render("clip")
	}
	return label + " " + active + idle + " " + clip
}

func slotSeedLabel(mark *seedBookmark) string {
	if mark == nil {
		return "—"
	}
	return fmt.Sprintf("%s/%d", mark.Spec.Display, mark.Seed)
}

func inspectorDebugLabel(status gen.DebugStatus) string {
	text := gen.FormatDebugStatus(status)
	if text == "" {
		return "debug unavailable"
	}
	return text
}

func bottomBar(m Model, w int, theme ColorTheme) string {
	state := "play"
	if m.paused {
		state = "PAUSED"
	}
	hintParts := []string{
		fmt.Sprintf("[space] %s", state),
		fmt.Sprintf("[↑↓] %d%%", m.volume),
		fmt.Sprintf("[C] %s", Visuals[m.visualIdx].Name),
		"[l] library",
		"[i] inspect",
		"[e] export",
		"[?] help",
	}
	if m.recording {
		hintParts = append(hintParts, "[r] stop rec")
	} else {
		hintParts = append(hintParts, "[r] rec")
	}
	if len(m.themes) > 1 {
		hintParts = append(hintParts, fmt.Sprintf("[c] %s", theme.Name))
	}
	if len(m.genres) > 1 {
		hintParts = append(hintParts, "[n/p] algo")
	}
	if m.debugVisible {
		hintParts = append(hintParts, "[d] debug on")
	}
	if m.playlist != nil {
		hintParts = append(hintParts, "[s] skip")
	}
	hintParts = append(hintParts, "[q] quit")
	if m.helpVisible {
		hintParts = []string{"[?] close help", "[q] quit"}
	} else if m.libraryVisible {
		hintParts = []string{"[↑↓] browse", "[enter] load", "[delete] remove", "[l] close", "[q] quit"}
	} else if m.inspectorVisible {
		hintParts = []string{"[i] close", "[e] export", "[r] record", "[q] quit"}
	} else if m.exportVisible {
		hintParts = []string{"[w] wav", "[m] midi", "[t] stems", "[r] record", "[e] close", "[q] quit"}
	}
	hint := strings.Join(hintParts, "   ")

	status := m.currentStatus(time.Now())
	if status != "" {
		status = trimToWidth(status, maxInt(0, w/2))
		hint = trimToWidth(hint, maxInt(0, w-lipgloss.Width(status)-1))
	}
	left := lipgloss.NewStyle().Faint(true).Render(hint)
	right := ""
	if status != "" {
		right = lipgloss.NewStyle().Foreground(theme.BarHi).Render(status)
	}
	pad := w - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	return left + spaces(pad) + right
}

func helpPanel(m Model, w, h int, theme ColorTheme) string {
	bodyW := maxInt(24, minInt(w-6, 76))
	bodyH := maxInt(10, minInt(h-2, 18))
	lines := []string{
		styleHelpLine(theme, false, "Playback", "[space] pause/resume   [↑↓] volume   [r] record"),
		styleHelpLine(theme, false, "Look", "[C] visual   [c] theme   [d] debug   [i] inspect"),
		styleHelpLine(theme, false, "Seeds", "[[/]] browse   [a/b] store   [tab] compare   [k/x] keep/reject   [l] library"),
		styleHelpLine(theme, false, "Export", "[e] drawer   [w] wav   [m] midi   [t] stems"),
		styleHelpLine(theme, false, "Tracks", "[n/p] algorithm   [s] skip playlist track"),
		styleHelpLine(theme, false, "Close", "[?] close this overlay   [q] quit"),
	}
	lines = filterHelpLines(lines, m)
	content := strings.Join(lines, "\n")
	panel := lipgloss.NewStyle().
		Width(bodyW).
		Height(bodyH).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.BarFg).
		Padding(1, 2).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render("TERMUS HELP"),
				"",
				content,
			),
		)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, panel)
}

func styleHelpLine(theme ColorTheme, dim bool, title, text string) string {
	label := lipgloss.NewStyle().Foreground(theme.BarHi).Render(title)
	valueStyle := lipgloss.NewStyle()
	if dim {
		valueStyle = valueStyle.Faint(true)
	}
	return label + "  " + valueStyle.Render(text)
}

func filterHelpLines(lines []string, m Model) []string {
	out := make([]string, 0, len(lines))
	for idx, line := range lines {
		switch idx {
		case 4:
			if len(m.genres) <= 1 && m.playlist == nil {
				continue
			}
			if m.playlist == nil {
				line = styleHelpLine(m.activeTheme(), false, "Tracks", "[n/p] algorithm")
			} else if len(m.genres) <= 1 {
				line = styleHelpLine(m.activeTheme(), false, "Tracks", "[s] skip playlist track")
			}
		}
		out = append(out, line)
	}
	return out
}

func inspectorPanel(m Model, w, h int, theme ColorTheme) string {
	bodyW := maxInt(30, minInt(w-6, 84))
	bodyH := maxInt(12, minInt(h-2, 18))
	details := []string{
		styleHelpLine(theme, false, "Track", fmt.Sprintf("%s · %s", m.algo, m.keyName)),
		styleHelpLine(theme, false, "Seed", fmt.Sprintf("%d", m.seed)),
		styleHelpLine(theme, false, "Slots", fmt.Sprintf("A %s   B %s   kept %d", slotSeedLabel(m.seedA), slotSeedLabel(m.seedB), len(m.kept))),
		styleHelpLine(theme, false, "State", inspectorDebugLabel(m.debug)),
		styleHelpLine(theme, false, "Export", "[e] export drawer   [r] record   --out/--stems/--midi available"),
	}
	if m.playlist != nil && m.playlistIdx < len(m.playlist.Tracks) {
		details = append(details, styleHelpLine(theme, false, "Playlist",
			fmt.Sprintf("%s · %d/%d · next %s", m.playlist.Name, m.playlistIdx+1, len(m.playlist.Tracks), shortDuration(time.Until(m.nextTrackAt)))))
	}
	panel := lipgloss.NewStyle().
		Width(bodyW).
		Height(bodyH).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.BarFg).
		Padding(1, 2).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render("TRACK INSPECTOR"),
				"",
				strings.Join(details, "\n"),
				"",
				lipgloss.NewStyle().Faint(true).Render("[i] close   [e] export   [q] quit"),
			),
		)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, panel)
}

func exportPanel(m Model, w, h int, theme ColorTheme) string {
	bodyW := maxInt(30, minInt(w-6, 78))
	bodyH := maxInt(12, minInt(h-2, 16))
	duration := "60s"
	if m.exporter != nil {
		duration = m.exporter.durationLabel()
	}
	lines := []string{
		lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render("EXPORT"),
		"",
		styleHelpLine(theme, false, "Track", fmt.Sprintf("%s · seed %d", m.algo, m.seed)),
		styleHelpLine(theme, false, "Artifacts", fmt.Sprintf("[w] WAV %s   [m] MIDI %s   [t] stems %s", duration, duration, duration)),
		styleHelpLine(theme, false, "Live", "[r] toggle recording"),
		styleHelpLine(theme, false, "Status", "exports write to ./exports with the current theme and mix settings"),
		"",
	}
	if m.exportBusy {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.BarHi).Render("rendering in background..."))
	} else {
		lines = append(lines, lipgloss.NewStyle().Faint(true).Render("[e] close   [q] quit"))
	}
	panel := lipgloss.NewStyle().
		Width(bodyW).
		Height(bodyH).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.BarFg).
		Padding(1, 2).
		Render(strings.Join(lines, "\n"))
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, panel)
}

func libraryPanel(m Model, w, h int, theme ColorTheme) string {
	bodyW := maxInt(28, minInt(w-6, 82))
	bodyH := maxInt(10, minInt(h-2, 18))
	lines := []string{
		lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render("SAVED SEEDS"),
		"",
	}
	if len(m.savedSeeds) == 0 {
		lines = append(lines,
			"No saved seeds yet.",
			"",
			lipgloss.NewStyle().Faint(true).Render("Press [k] while browsing seeds to keep one here."),
		)
	} else {
		now := time.Now()
		maxRows := maxInt(1, bodyH-5)
		start := 0
		if m.libraryIdx >= maxRows {
			start = m.libraryIdx - maxRows + 1
		}
		end := minInt(len(m.savedSeeds), start+maxRows)
		for i := start; i < end; i++ {
			rec := m.savedSeeds[i]
			_, label, ok := resolveSavedSeedRecord(rec)
			entry := fmt.Sprintf("%s · %d · %s", label, rec.Seed, formatSavedSeedAge(now, rec.SavedAt))
			if !ok {
				entry += " · unavailable"
			}
			if i == m.libraryIdx {
				entry = lipgloss.NewStyle().Foreground(theme.BarHi).Render("› " + entry)
			} else {
				entry = "  " + entry
			}
			lines = append(lines, entry)
		}
	}
	lines = append(lines, "", lipgloss.NewStyle().Faint(true).Render("[↑↓] browse   [enter] load   [delete] remove   [l] close"))
	panel := lipgloss.NewStyle().
		Width(bodyW).
		Height(bodyH).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.BarFg).
		Padding(1, 2).
		Render(strings.Join(lines, "\n"))
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, panel)
}

func spaces(n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = ' '
	}
	return string(out)
}

func trimToWidth(text string, max int) string {
	if max <= 0 {
		return ""
	}
	if lipgloss.Width(text) <= max {
		return text
	}
	if max <= 3 {
		runes := []rune(text)
		if len(runes) > max {
			runes = runes[:max]
		}
		return string(runes)
	}
	runes := []rune(text)
	limit := max - 3
	if len(runes) > limit {
		runes = runes[:limit]
	}
	return string(runes) + "..."
}

func formatElapsed(label string, d time.Duration) string {
	return label + " " + shortDuration(d)
}

func shortDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Round(time.Second).Seconds())
	mins := total / 60
	secs := total % 60
	return fmt.Sprintf("%02d:%02d", mins, secs)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func centerBox(w, h int, text string) string {
	if w < 1 || h < 1 {
		return text
	}
	lines := make([]string, h)
	mid := h / 2
	for i := range lines {
		if i == mid {
			pad := (w - lipgloss.Width(text)) / 2
			if pad < 0 {
				pad = 0
			}
			lines[i] = spaces(pad) + trimToWidth(text, maxInt(0, w-pad))
		} else {
			lines[i] = ""
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

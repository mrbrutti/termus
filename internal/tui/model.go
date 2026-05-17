package tui

import (
	"fmt"
	"math"
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

type TrackNavEntry struct {
	ID          string
	Style       string
	Title       string
	Description string
	Tags        []string
	Key         string
	Tempo       string
	ListenMode  string
	Sections    []string
}

type TrackLoader func(id string) (*gen.Playlist, string, error)

type AudioControl struct {
	Retry      func()
	RenderOnly func()
}

type StartupLoadMsg struct {
	Title   string
	Detail  string
	Percent float64
	Done    bool
}

type TrackLoadResultMsg struct {
	EntryID    string
	EntryTitle string
	Playlist   *gen.Playlist
	ModeLabel  string
	Spec       gen.AlgoSpec
	Seed       int64
	Algo       gen.Algorithm
	Err        error
}

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

	volume             int
	paused             bool
	recording          bool
	debugVisible       bool
	helpVisible        bool
	trackVisible       bool
	libraryVisible     bool
	inspectorVisible   bool
	exportVisible      bool
	controlsVisible    bool
	exportBusy         bool
	reducedChrome      bool
	splashVisible      bool
	startupLoading     bool
	startupTitle       string
	startupDetail      string
	startupPercent     float64
	status             string
	statusTTL          time.Time
	stickyStatus       string
	volumeOverlayUntil time.Time

	themeIdx          int // index into Themes
	visualIdx         int // index into Visuals
	visualPrevIdx     int
	visualSwitchUntil time.Time
	themes            []ColorTheme
	ui                AdaptiveUI
	musicProfile      *gen.ControlProfile
	morphMode         int
	audioControl      *AudioControl
	trackLoader       TrackLoader
	tracks            []TrackNavEntry
	trackIdx          int
	trackStyleIdx     int
	activeTrackID     string

	// Algorithm switching ([n]/[p]).
	genres          []gen.AlgoSpec // ordered list of switchable algorithms
	genreIdx        int            // current index into genres
	buildFn         BuildAlgoFn    // closure used to construct a new algorithm
	seedA           *seedBookmark
	seedB           *seedBookmark
	kept            map[string]seedBookmark
	savedSeeds      []savedSeedRecord
	savedSessions   []savedSessionRecord
	curation        map[string]seedCurationRecord
	libraryIdx      int
	exporter        *ExportController
	controlTab      controlTab
	controlRow      int
	sessionIdx      int
	curateTagIdx    int
	curateRecentIdx int
	curateBestIdx   int

	// Playlist auto-advance.
	playlist        *gen.Playlist
	playlistIdx     int // index of currently-playing track
	trackStartedAt  time.Time
	nextTrackAt     time.Time // when to advance to the next track
	playlistFade    int       // crossfade length in audio frames (44.1 kHz)
	startedAt       time.Time
	splashUntil     time.Time
	recordStartedAt time.Time
	listeningMode   string
}

// New constructs a Model. keyName is e.g. "Cmin".
func New(ring *scope.Ring, cmd audio.Commander, algo, keyName string, seed int64, initialVol int) Model {
	ui := DetectAdaptiveUI()
	savedSeeds, _ := loadSavedSeedRecords()
	savedSessions, _ := loadSavedSessionRecords()
	curation, _ := loadSeedCuration()
	m := Model{
		ring:          ring,
		cmd:           cmd,
		algo:          algo,
		debug:         cmd.DebugStatus(),
		keyName:       keyName,
		seed:          seed,
		volume:        initialVol,
		ui:            ui,
		themes:        append([]ColorTheme(nil), ui.Themes...),
		themeIdx:      ui.DefaultThemeIdx,
		morphMode:     2,
		kept:          recordsToBookmarks(savedSeeds),
		savedSeeds:    savedSeeds,
		savedSessions: savedSessions,
		curation:      curation,
		startedAt:     time.Now(),
		splashUntil:   time.Now().Add(5 * time.Second),
		splashVisible: true,
		visualPrevIdx: -1,
	}
	m.touchCurrentSeed()
	return m
}

// WithSwitcher enables in-app algorithm switching. genres is the ordered list
// the user cycles through; startIdx is the index of the algorithm currently
// playing; buildFn constructs a fresh Algorithm for a chosen spec.
func (m Model) WithSwitcher(genres []gen.AlgoSpec, startIdx int, buildFn BuildAlgoFn) Model {
	m.genres = genres
	m.genreIdx = startIdx
	m.buildFn = buildFn
	m.touchCurrentSeed()
	return m
}

// WithDebug controls whether the dedicated debug inspector starts visible.
func (m Model) WithDebug(visible bool) Model {
	m.debugVisible = visible
	return m
}

func (m Model) WithListeningMode(label string) Model {
	m.listeningMode = label
	return m
}

func (m Model) WithExportController(exporter *ExportController) Model {
	m.exporter = exporter
	return m
}

func (m Model) WithControlProfile(profile *gen.ControlProfile) Model {
	m.musicProfile = profile
	return m
}

func (m Model) WithAudioControl(control *AudioControl) Model {
	m.audioControl = control
	return m
}

func (m Model) WithStartupLoading(title, detail string, percent float64) Model {
	m.startupLoading = true
	m.startupTitle = title
	m.startupDetail = detail
	m.startupPercent = clamp01(percent)
	m.splashVisible = true
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

func (m Model) WithTrackBrowser(entries []TrackNavEntry, loader TrackLoader, visible bool) Model {
	m.tracks = append([]TrackNavEntry(nil), entries...)
	m.trackLoader = loader
	m.trackVisible = visible
	if m.trackIdx >= len(m.tracks) {
		m.trackIdx = maxInt(0, len(m.tracks)-1)
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
	m.algo = track.Spec.Label()
	m.seed = track.Seed
	m.touchCurrentSeed()
	m.trackStartedAt = time.Now()
	m.nextTrackAt = time.Now().Add(track.Duration)
	label := track.Spec.Label()
	if strings.TrimSpace(track.Title) != "" {
		label = track.Title + " · " + label
	}
	m.flashStatus(fmt.Sprintf("▶ %d/%d %s",
		m.playlistIdx+1, len(m.playlist.Tracks), label), 3*time.Second)

	// Keep the genre cycle index in sync if this track matches a genre.
	for i, g := range m.genres {
		if g.Name == track.Spec.Name {
			m.genreIdx = i
			break
		}
	}
}

func (m Model) morphFadeFrames() int {
	seconds := []float64{0.0, 0.18, 0.40, 0.90, 1.60}
	idx := clampInt(m.morphMode, 0, len(seconds)-1)
	return int(seconds[idx] * 44100.0)
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
	m.cmd.SwapAlgorithmFade(algo, m.morphFadeFrames())
	m.algo = spec.Label()
	m.touchCurrentSeed()
	m.flashStatus("switched: "+spec.Label(), 2*time.Second)
}

func (m Model) currentSpec() (gen.AlgoSpec, bool) {
	if m.genreIdx >= 0 && m.genreIdx < len(m.genres) {
		return m.genres[m.genreIdx], true
	}
	return gen.AlgoSpec{}, false
}

func algoIdentity(spec gen.AlgoSpec) string {
	label := spec.Label()
	if spec.Name == "" {
		return label
	}
	if label == "" || strings.EqualFold(label, spec.Name) {
		return spec.Name
	}
	return fmt.Sprintf("%s · %s", label, spec.Name)
}

func (m Model) currentAlgoIdentity() string {
	if spec, ok := m.currentSpec(); ok {
		return algoIdentity(spec)
	}
	if m.playlist != nil && m.playlistIdx >= 0 && m.playlistIdx < len(m.playlist.Tracks) {
		return algoIdentity(m.playlist.Tracks[m.playlistIdx].Spec)
	}
	return m.algo
}

func (m *Model) swapToSeed(spec gen.AlgoSpec, seed int64, status string) {
	if m.playlist != nil || m.buildFn == nil {
		return
	}
	if seed < 0 {
		seed = 0
	}
	algo := m.buildFn(spec, seed)
	m.cmd.SwapAlgorithmFade(algo, m.morphFadeFrames())
	m.algo = spec.Label()
	m.seed = seed
	for i, g := range m.genres {
		if g.Name == spec.Name {
			m.genreIdx = i
			break
		}
	}
	m.touchCurrentSeed()
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
	m.flashStatus(fmt.Sprintf("%s ← %s/%d", slot, spec.Label(), m.seed), 2*time.Second)
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
		Display: algoIdentity(spec),
		Seed:    m.seed,
		SavedAt: time.Now(),
	}
	m.savedSeeds = append([]savedSeedRecord{rec}, removeSavedSeedRecord(m.savedSeeds, spec.Name, m.seed)...)
	m.markCurrentSeedKept()
	if err := saveSavedSeedRecords(m.savedSeeds); err != nil {
		m.flashStatus("keep saved locally failed", 3*time.Second)
		return
	}
	m.flashStatus(fmt.Sprintf("kept %s/%d (%d)", spec.Label(), m.seed, len(m.kept)), 2*time.Second)
}

func (m *Model) toggleTracks() {
	if len(m.tracks) == 0 {
		m.flashStatus("tracks: none found", 2*time.Second)
		return
	}
	m.trackVisible = !m.trackVisible
	if m.trackVisible {
		m.helpVisible = false
		m.libraryVisible = false
		m.inspectorVisible = false
		m.exportVisible = false
		m.controlsVisible = false
		m.alignTrackSelection()
		m.flashStatus("tracks: on", 2*time.Second)
		return
	}
	m.flashStatus("tracks: off", 2*time.Second)
}

func (m *Model) moveTrack(delta int) {
	visible := m.filteredTrackIndices()
	if len(visible) == 0 {
		m.trackIdx = 0
		return
	}
	current := 0
	for i, idx := range visible {
		if idx == m.trackIdx {
			current = i
			break
		}
	}
	current = (current + delta + len(visible)) % len(visible)
	m.trackIdx = visible[current]
}

func (m *Model) cycleTrackStyle(delta int) {
	styles := m.trackStyleOptions()
	if len(styles) == 0 {
		m.trackStyleIdx = 0
		return
	}
	m.trackStyleIdx = (m.trackStyleIdx + delta + len(styles)) % len(styles)
	m.alignTrackSelection()
}

func (m *Model) alignTrackSelection() {
	visible := m.filteredTrackIndices()
	if len(visible) == 0 {
		m.trackIdx = 0
		return
	}
	for _, idx := range visible {
		if idx == m.trackIdx {
			return
		}
	}
	m.trackIdx = visible[0]
}

func (m *Model) loadSelectedTrack() tea.Cmd {
	if m.trackLoader == nil || len(m.tracks) == 0 || m.buildFn == nil {
		return nil
	}
	entry := m.tracks[m.trackIdx]
	title := strings.TrimSpace(entry.Title)
	if title == "" {
		title = entry.ID
	}
	m.startupLoading = true
	m.startupTitle = title
	m.startupDetail = "compiling authored arrangement"
	m.startupPercent = 0.18
	m.splashVisible = true
	return func() tea.Msg {
		pl, modeLabel, err := m.trackLoader(entry.ID)
		if err != nil {
			return TrackLoadResultMsg{EntryID: entry.ID, EntryTitle: title, Err: err}
		}
		if pl == nil || len(pl.Tracks) == 0 {
			return TrackLoadResultMsg{EntryID: entry.ID, EntryTitle: title, Err: fmt.Errorf("track empty")}
		}
		first := pl.Tracks[0]
		algo := m.buildFn(first.Spec, first.Seed)
		return TrackLoadResultMsg{
			EntryID:    entry.ID,
			EntryTitle: title,
			Playlist:   pl,
			ModeLabel:  modeLabel,
			Spec:       first.Spec,
			Seed:       first.Seed,
			Algo:       algo,
		}
	}
}

func (m *Model) toggleLibrary() {
	m.libraryVisible = !m.libraryVisible
	if m.libraryVisible {
		m.helpVisible = false
		m.trackVisible = false
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
		m.trackVisible = false
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
		m.trackVisible = false
		m.libraryVisible = false
		m.inspectorVisible = false
		m.flashStatus("export: on", 2*time.Second)
		return
	}
	m.flashStatus("export: off", 2*time.Second)
}

func (m *Model) toggleControls() {
	m.controlsVisible = !m.controlsVisible
	if m.controlsVisible {
		m.helpVisible = false
		m.trackVisible = false
		m.libraryVisible = false
		m.inspectorVisible = false
		m.exportVisible = false
		m.splashVisible = false
		m.controlRow = 0
		m.flashStatus("controls: on", 2*time.Second)
		return
	}
	m.flashStatus("controls: off", 2*time.Second)
}

func (m *Model) toggleRecording() {
	path, err := m.cmd.ToggleRecord()
	if err != nil {
		m.flashStatus("rec error: "+err.Error(), 3*time.Second)
		m.recording = false
		return
	}
	if path != "" {
		m.recording = true
		m.recordStartedAt = time.Now()
		m.flashStatus("rec → "+path, 3*time.Second)
		return
	}
	m.recording = false
	m.recordStartedAt = time.Time{}
	m.flashStatus("rec stopped", 3*time.Second)
}

func (m *Model) toggleReducedChrome() {
	m.reducedChrome = !m.reducedChrome
	if m.reducedChrome {
		m.flashStatus("zen: on", 2*time.Second)
		return
	}
	m.flashStatus("zen: off", 2*time.Second)
}

func (m *Model) activeSpec() (gen.AlgoSpec, bool) {
	if spec, ok := m.currentSpec(); ok {
		return spec, true
	}
	if m.playlist != nil && m.playlistIdx >= 0 && m.playlistIdx < len(m.playlist.Tracks) {
		return m.playlist.Tracks[m.playlistIdx].Spec, true
	}
	return gen.AlgoSpec{}, false
}

func (m *Model) ensureMusicProfile() *gen.ControlProfile {
	if m.musicProfile == nil {
		profile := gen.DefaultControlProfile()
		m.musicProfile = &profile
	}
	return m.musicProfile
}

func (m *Model) updateMusicProfile(label string, mutate func(*gen.ControlProfile)) {
	profile := m.ensureMusicProfile()
	mutate(profile)
	profile.Density = clampInt(profile.Density, 0, 4)
	profile.Brightness = clampInt(profile.Brightness, 0, 4)
	profile.Motion = clampInt(profile.Motion, 0, 4)
	profile.Reverb = clampInt(profile.Reverb, 0, 4)
	profile.Swing = clampInt(profile.Swing, 0, 4)
	profile.DroneDepth = clampInt(profile.DroneDepth, 0, 4)
	profile.Tempo = clampInt(profile.Tempo, 0, 4)
	profile.Phrase = clampInt(profile.Phrase, 0, 4)
	m.refreshCurrentTake(label)
}

func (m *Model) refreshCurrentTake(label string) {
	spec, ok := m.activeSpec()
	if !ok || m.buildFn == nil {
		return
	}
	algo := m.buildFn(spec, m.seed)
	fade := m.morphFadeFrames()
	if fade == 0 {
		fade = 15435
	}
	m.cmd.SwapAlgorithmFade(algo, fade)
	m.algo = spec.Label()
	m.flashStatus(label, 2*time.Second)
}

func (m *Model) retryAudio() {
	if m.audioControl == nil || m.audioControl.Retry == nil {
		m.flashStatus("audio retry unavailable", 2*time.Second)
		return
	}
	m.audioControl.Retry()
	m.flashStatus("audio retrying...", 2*time.Second)
}

func (m *Model) fallbackRenderOnly() {
	if m.audioControl == nil || m.audioControl.RenderOnly == nil {
		m.flashStatus("render-only fallback unavailable", 2*time.Second)
		return
	}
	m.audioControl.RenderOnly()
	m.flashStatus("audio: render-only", 2*time.Second)
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
	if spec, ok := gen.Resolve(rec.Algo); ok {
		if m.kept != nil {
			delete(m.kept, bookmarkKey(spec, rec.Seed))
		}
		if m.curation != nil {
			key := bookmarkKey(spec, rec.Seed)
			if cur, found := m.curation[key]; found {
				cur.Kept = false
				m.curation[key] = cur
				_ = saveSeedCuration(m.curation)
			}
		}
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
	case StartupLoadMsg:
		m.applyStartupLoad(msg)
		return m, nil
	case TrackLoadResultMsg:
		m.startupLoading = false
		m.splashVisible = false
		if msg.Err != nil {
			m.flashStatus("track load failed", 3*time.Second)
			return m, nil
		}
		if msg.Playlist == nil || len(msg.Playlist.Tracks) == 0 || msg.Algo == nil {
			m.flashStatus("track empty", 3*time.Second)
			return m, nil
		}
		first := msg.Playlist.Tracks[0]
		m.cmd.SwapAlgorithmFade(msg.Algo, m.morphFadeFrames())
		m.playlist = msg.Playlist
		m.playlistIdx = 0
		m.playlistFade = 88200
		m.trackStartedAt = time.Now()
		m.nextTrackAt = time.Now().Add(first.Duration)
		m.algo = msg.Spec.Label()
		m.seed = msg.Seed
		m.activeTrackID = msg.EntryID
		m.trackVisible = false
		m.touchCurrentSeed()
		if msg.ModeLabel != "" {
			m.listeningMode = msg.ModeLabel
		}
		for i, g := range m.genres {
			if g.Name == msg.Spec.Name {
				m.genreIdx = i
				break
			}
		}
		m.flashStatus("track: "+msg.EntryTitle, 3*time.Second)
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
		if m.startupLoading {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}
		if m.splashVisible {
			m.splashVisible = false
		}
		if m.trackVisible {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "t", "esc":
				m.toggleTracks()
			case "left":
				m.cycleTrackStyle(-1)
			case "right":
				m.cycleTrackStyle(1)
			case "up":
				m.moveTrack(-1)
			case "down":
				m.moveTrack(1)
			case "enter":
				return m, m.loadSelectedTrack()
			}
			return m, nil
		}
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
				m.toggleRecording()
			}
			return m, nil
		}
		if m.controlsVisible {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "m", "esc":
				m.toggleControls()
			case "tab":
				m.controlTab = m.controlTab.next()
				m.controlRow = 0
			case "shift+tab":
				m.controlTab = m.controlTab.prev()
				m.controlRow = 0
			case "up":
				m.moveControlRow(-1)
			case "down":
				m.moveControlRow(1)
			case "left":
				m.adjustControlRow(-1)
			case "right":
				m.adjustControlRow(1)
			case "enter", " ":
				return m, m.activateControlRow()
			}
			return m, nil
		}
		action := matchKey(msg)
		if m.helpVisible && action != actionHelp && action != actionQuit && action != actionControls {
			return m, nil
		}
		if m.inspectorVisible && action != actionInspector && action != actionQuit && action != actionExport && action != actionControls {
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
			m.showVolumeOverlay()
		case actionVolDown:
			m.volume -= 5
			if m.volume < 0 {
				m.volume = 0
			}
			m.cmd.SetVolume(m.volume)
			m.showVolumeOverlay()
		case actionRecord:
			m.toggleRecording()
		case actionTracks:
			m.toggleTracks()
		case actionTheme:
			if len(m.themes) > 1 {
				m.themeIdx = (m.themeIdx + 1) % len(m.themes)
				m.flashStatus("theme: "+m.themes[m.themeIdx].Name, 2*time.Second)
			}
		case actionNextAlgo:
			m.switchAlgo(1)
		case actionPrevAlgo:
			m.switchAlgo(-1)
		case actionNextTrack:
			if m.playlist != nil {
				m.advancePlaylist()
			}
		case actionVisual:
			m.startVisualTransition(m.visualIdx + 1)
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
				m.trackVisible = false
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
		case actionControls:
			m.toggleControls()
		case actionZen:
			m.toggleReducedChrome()
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
		if m.visualPrevIdx >= 0 && time.Now().After(m.visualSwitchUntil) {
			m.visualPrevIdx = -1
		}
		if m.splashVisible && time.Now().After(m.splashUntil) {
			m.splashVisible = false
		}
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
	now := time.Now()
	theme := m.activeTheme()
	if m.startupLoading {
		return startupLoadingView(m, m.width, m.height, theme, now)
	}
	compact := useCompactLayout(m.width, m.height)
	chromeH := 3 // top + now-playing + bottom bars
	if m.reducedChrome {
		chromeH = 1
	} else if m.debugVisible {
		chromeH++
	}
	if m.volumeOverlayVisible(now) {
		chromeH++
	}
	innerH := m.height - chromeH
	innerW := m.width

	// Snapshot scope and render with the active visual + theme.
	samples := make([]float64, innerW*2)
	m.ring.Snapshot(samples)
	peak, rms := sampleStats(samples)
	pulse := clamp01(0.55*peak + 0.45*rms)
	phase := float64(now.UnixNano()) / float64(time.Second)
	ctx := RenderContext{
		Theme: theme,
		Pulse: pulse,
		Phase: phase + 0.35*math.Sin(phase*0.35),
	}
	visual := Visuals[m.visualIdx]
	scopeStr := visual.Render(samples, innerW, innerH, ctx)
	if m.visualTransitionActive(now) {
		prev := Visuals[m.visualPrevIdx].Render(samples, innerW, innerH, ctx)
		progress := 1 - float64(time.Until(m.visualSwitchUntil))/float64(visualSwitchDuration)
		scopeStr = blendVisualFrames(prev, scopeStr, clamp01(progress))
	}

	top := topBar(m, innerW, theme, compact)
	playback := playbackBar(m, innerW, theme, samples, compact)
	bottom := bottomBar(m, innerW, theme, compact)
	volumeLine := ""
	if m.volumeOverlayVisible(now) {
		volumeLine = renderVolumeLine(m, innerW, theme)
	}
	body := scopeStr
	if m.helpVisible {
		body = helpPanel(m, innerW, innerH, theme)
	} else if m.trackVisible {
		body = trackPanel(m, innerW, innerH, theme)
	} else if m.libraryVisible {
		body = libraryPanel(m, innerW, innerH, theme)
	} else if m.inspectorVisible {
		body = inspectorPanel(m, innerW, innerH, theme)
	} else if m.exportVisible {
		body = exportPanel(m, innerW, innerH, theme)
	} else if m.controlsVisible {
		body = controlsPanel(m, innerW, innerH, theme)
	} else if m.splashVisible {
		body = splashPanel(m, innerW, innerH, theme)
	}
	if m.reducedChrome {
		if volumeLine != "" {
			return lipgloss.JoinVertical(lipgloss.Left, body, volumeLine, bottom)
		}
		return lipgloss.JoinVertical(lipgloss.Left, body, bottom)
	}
	if m.debugVisible {
		debug := debugBar(m, innerW, theme)
		if volumeLine != "" {
			return lipgloss.JoinVertical(lipgloss.Left, top, playback, volumeLine, debug, body, bottom)
		}
		return lipgloss.JoinVertical(lipgloss.Left, top, playback, debug, body, bottom)
	}
	if volumeLine != "" {
		return lipgloss.JoinVertical(lipgloss.Left, top, playback, volumeLine, body, bottom)
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

const visualSwitchDuration = 340 * time.Millisecond

func (m *Model) startVisualTransition(next int) {
	if len(Visuals) == 0 {
		return
	}
	next = (next + len(Visuals)) % len(Visuals)
	if next == m.visualIdx {
		return
	}
	m.visualPrevIdx = m.visualIdx
	m.visualIdx = next
	m.visualSwitchUntil = time.Now().Add(visualSwitchDuration)
}

func (m Model) visualTransitionActive(now time.Time) bool {
	return m.visualPrevIdx >= 0 &&
		m.visualPrevIdx < len(Visuals) &&
		m.visualPrevIdx != m.visualIdx &&
		now.Before(m.visualSwitchUntil)
}

func (m *Model) flashStatus(text string, ttl time.Duration) {
	m.status = text
	m.statusTTL = time.Now().Add(ttl)
}

func (m *Model) showVolumeOverlay() {
	m.volumeOverlayUntil = time.Now().Add(1200 * time.Millisecond)
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
		if m.startupLoading {
			m.startupDetail = state.StatusText()
		}
		m.setStickyStatus(state.StatusText())
	case audio.BackendStateReady:
		if m.startupLoading {
			m.startupLoading = false
			m.splashVisible = false
		}
		m.setStickyStatus("")
		m.flashStatus(state.StatusText(), 2*time.Second)
	case audio.BackendStateNoDefaultDevice, audio.BackendStateHung,
		audio.BackendStateRenderOnly, audio.BackendStateInitFailed:
		if m.startupLoading {
			m.startupLoading = false
			m.splashVisible = false
		}
		m.setStickyStatus(state.StatusText())
	}
}

func (m *Model) applyStartupLoad(msg StartupLoadMsg) {
	m.startupLoading = !msg.Done
	if msg.Title != "" {
		m.startupTitle = msg.Title
	}
	if msg.Detail != "" {
		m.startupDetail = msg.Detail
	}
	m.startupPercent = clamp01(msg.Percent)
	m.splashVisible = true
}

func topBar(m Model, w int, theme ColorTheme, compact bool) string {
	currentAlgo := m.currentAlgoIdentity()
	var label string
	if m.playlist != nil {
		if compact {
			label = fmt.Sprintf("termus · %s · %d/%d · %d", currentAlgo, m.playlistIdx+1, len(m.playlist.Tracks), m.seed)
		} else {
			label = fmt.Sprintf("termus · %s · %d/%d %s · seed=%d",
				m.playlist.Name, m.playlistIdx+1, len(m.playlist.Tracks),
				currentAlgo, m.seed)
		}
	} else {
		if compact {
			label = fmt.Sprintf("termus · %s · %d", currentAlgo, m.seed)
		} else {
			label = fmt.Sprintf("termus · %s · %s · seed=%d",
				currentAlgo, m.keyName, m.seed)
		}
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
	if !compact {
		if seeds := m.seedSlotsLabel(); seeds != "" {
			seeds = lipgloss.NewStyle().Faint(true).Render(seeds)
			if right == "" {
				right = seeds
			} else {
				right = seeds + "  " + right
			}
		}
	}
	if compact && len(m.kept) > 0 {
		kept := lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("keep=%d", len(m.kept)))
		if right == "" {
			right = kept
		} else {
			right = kept + "  " + right
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

func playbackBar(m Model, w int, theme ColorTheme, samples []float64, compact bool) string {
	leftParts := []string{formatElapsed("live", time.Since(m.startedAt))}
	if m.listeningMode != "" {
		leftParts = append(leftParts, m.listeningMode)
	}
	if m.playlist != nil && m.playlistIdx < len(m.playlist.Tracks) {
		track := m.playlist.Tracks[m.playlistIdx]
		if compact {
			leftParts = append(leftParts,
				fmt.Sprintf("%s/%s", shortDuration(time.Since(m.trackStartedAt)), shortDuration(track.Duration)),
				fmt.Sprintf("next %s", shortDuration(time.Until(m.nextTrackAt))),
			)
		} else {
			leftParts = append(leftParts,
				fmt.Sprintf("track %s/%s", shortDuration(time.Since(m.trackStartedAt)), shortDuration(track.Duration)),
				fmt.Sprintf("next %s", shortDuration(time.Until(m.nextTrackAt))),
				fmt.Sprintf("fade %s", shortDuration(time.Duration(m.playlistFade)*time.Second/44100)),
			)
		}
		if len(m.playlist.Tracks) > 0 {
			leftParts = append(leftParts, fmt.Sprintf("%d/%d", m.playlistIdx+1, len(m.playlist.Tracks)))
		}
	}
	if m.recording && !m.recordStartedAt.IsZero() {
		leftParts = append(leftParts, formatElapsed("rec", time.Since(m.recordStartedAt)))
	}
	leftText := trimToWidth(strings.Join(leftParts, " · "), maxInt(0, w-22))
	meter, clipped := meterSummary(samples)
	meterWidth := 14
	if compact {
		meterWidth = 8
	}
	right := renderCompactMeter(theme, meter, clipped, meterWidth)
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

func (m Model) volumeOverlayVisible(now time.Time) bool {
	return now.Before(m.volumeOverlayUntil)
}

func renderVolumeLine(m Model, w int, theme ColorTheme) string {
	if w < 2 {
		return ""
	}
	side := w / 2
	activeSide := int(float64(side) * float64(m.volume) / 100.0)
	if m.volume > 0 && activeSide == 0 {
		activeSide = 1
	}
	if activeSide > side {
		activeSide = side
	}
	idleSide := side - activeSide
	left := lipgloss.NewStyle().Faint(true).Render(strings.Repeat("─", idleSide)) +
		lipgloss.NewStyle().Foreground(theme.BarHi).Render(strings.Repeat("─", activeSide))
	center := ""
	if w%2 != 0 {
		center = lipgloss.NewStyle().Foreground(theme.BarHi).Render("─")
	}
	right := lipgloss.NewStyle().Foreground(theme.BarHi).Render(strings.Repeat("─", activeSide)) +
		lipgloss.NewStyle().Faint(true).Render(strings.Repeat("─", idleSide))
	return left + center + right
}

func slotSeedLabel(mark *seedBookmark) string {
	if mark == nil {
		return "—"
	}
	return fmt.Sprintf("%s/%d", mark.Spec.Label(), mark.Seed)
}

func inspectorDebugLabel(status gen.DebugStatus) string {
	text := gen.FormatDebugStatus(status)
	if text == "" {
		return "debug unavailable"
	}
	return text
}

func bottomBar(m Model, w int, theme ColorTheme, compact bool) string {
	if m.reducedChrome {
		left := lipgloss.NewStyle().Foreground(theme.BarFg).Render(m.algo)
		right := lipgloss.NewStyle().Faint(true).Render("?")
		pad := w - lipgloss.Width(left) - lipgloss.Width(right)
		if pad < 1 {
			pad = 1
		}
		return left + spaces(pad) + right
	}
	leftText := lipgloss.NewStyle().Foreground(theme.BarFg).Render(m.algo)
	rightText := lipgloss.NewStyle().Faint(true).Render("?  m")
	status := m.currentStatus(time.Now())
	if status == "" && compact {
		status = " "
	}
	statusStyle := lipgloss.NewStyle().Foreground(theme.BarHi)
	if status == "" {
		statusStyle = lipgloss.NewStyle().Faint(true)
		status = " "
	}
	centerText := statusStyle.Render(trimToWidth(status, maxInt(0, w/3)))
	available := w - lipgloss.Width(leftText) - lipgloss.Width(rightText) - 2
	if available < 1 {
		available = 1
	}
	centerWidth := lipgloss.Width(centerText)
	if centerWidth > available {
		centerText = statusStyle.Render(trimToWidth(status, available))
		centerWidth = lipgloss.Width(centerText)
	}
	leftPad := (available - centerWidth) / 2
	rightPad := available - centerWidth - leftPad
	return leftText + spaces(leftPad+1) + centerText + spaces(rightPad+1) + rightText
}

func helpPanel(m Model, w, h int, theme ColorTheme) string {
	bodyW := maxInt(24, minInt(w-6, 76))
	bodyH := maxInt(10, minInt(h-2, 18))
	lines := []string{
		styleHelpLine(theme, false, "Global", "[space] play / pause   [↑↓] volume   [m] control center   [t] tracks"),
		styleHelpLine(theme, false, "View", "The main screen stays minimal. Use the control center for everything deeper."),
		styleHelpLine(theme, false, "Track Library", "[t] open library   [←→] style   [↑↓] browse   [enter] play"),
		styleHelpLine(theme, false, "Inside Control Center", "[↑↓] browse   [←→] adjust   [enter] apply / open   [tab] next section"),
		styleHelpLine(theme, false, "Sections", "Now   Look   Music   Seeds   Library   Export   Audio   Debug"),
		styleHelpLine(theme, false, "Close", "[?] close help   [q] quit"),
	}
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
	return append([]string(nil), lines...)
}

func splashPanel(m Model, w, h int, theme ColorTheme) string {
	bodyW := maxInt(30, minInt(w-6, 72))
	bodyH := maxInt(12, minInt(h-2, 16))
	lines := []string{
		lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render("TERMUS"),
		"",
		styleHelpLine(theme, false, "Play", "[space] pause / resume   [↑↓] volume"),
		styleHelpLine(theme, false, "Open", "[m] control center   [t] tracks   [?] help"),
		styleHelpLine(theme, false, "Global", "[q] quit   [z] zen"),
		"",
		lipgloss.NewStyle().Faint(true).Render("Press any key to start exploring."),
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

func startupLoadingView(m Model, w, h int, theme ColorTheme, now time.Time) string {
	title := m.startupTitle
	if title == "" {
		title = "Loading termus"
	}
	barW := maxInt(26, minInt(w-10, 72))
	barH := 3
	phase := float64(now.UnixNano()) / float64(time.Second)
	bar := renderStartupBrailleBar(barW, barH, clamp01(m.startupPercent), phase, theme)
	pct := lipgloss.NewStyle().Foreground(theme.BarHi).Render(fmt.Sprintf("%3d%%", int(clamp01(m.startupPercent)*100)))
	titleLine := lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render(title)
	detail := ""
	if m.startupDetail != "" {
		detail = lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render(m.startupDetail)
	}
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleLine,
		"",
		bar,
		pct,
		"",
		detail,
	)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content)
}

func renderStartupBrailleBar(wCells, hCells int, progress, phase float64, theme ColorTheme) string {
	grid := newDotGrid(wCells, hCells)
	if len(grid) == 0 || len(grid[0]) == 0 {
		return ""
	}
	dotsX := len(grid[0])
	dotsY := len(grid)
	centerY := dotsY / 2
	for px := 0; px < dotsX; px++ {
		grid[centerY][px] = true
	}
	activeDots := int(clamp01(progress) * float64(maxInt(1, dotsX-1)))
	prevX, prevY := 0, centerY
	for px := 0; px <= activeDots; px++ {
		t := float64(px) / float64(maxInt(1, dotsX-1))
		env := math.Sin(math.Pi * t)
		wobble := math.Sin(phase*4.5+t*9.0)*0.55 + math.Sin(phase*2.1+t*4.0)*0.25
		head := math.Exp(-math.Pow(float64(px-activeDots)/8.0, 2))
		y := centerY - int((wobble*env+head*0.6)*float64(maxInt(1, dotsY/3)))
		y = clampInt(y, 0, dotsY-1)
		if px > 0 {
			drawLine(grid, prevX, prevY, px, y)
		}
		plotDot(grid, px, y)
		prevX, prevY = px, y
	}
	var b strings.Builder
	activeCell := activeDots / 2
	for cy := 0; cy < hCells; cy++ {
		for cx := 0; cx < wCells; cx++ {
			r := brailleCell(grid, cx, cy)
			switch {
			case r == '\u2800':
				b.WriteRune(' ')
			case cx <= activeCell:
				b.WriteString(renderCell(r, cx, cy, wCells, hCells, RenderContext{
					Theme: theme,
					Pulse: 0.35 + 0.65*progress,
					Phase: phase,
				}))
			default:
				b.WriteString(lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render(string(r)))
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func inspectorPanel(m Model, w, h int, theme ColorTheme) string {
	bodyW := maxInt(30, minInt(w-6, 84))
	bodyH := maxInt(12, minInt(h-2, 18))
	details := []string{
		styleHelpLine(theme, false, "Track", fmt.Sprintf("%s · %s", m.currentAlgoIdentity(), m.keyName)),
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
		styleHelpLine(theme, false, "Track", fmt.Sprintf("%s · seed %d", m.currentAlgoIdentity(), m.seed)),
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
			bookmark, label, ok := resolveSavedSeedRecord(rec)
			entry := fmt.Sprintf("%s · %d · %s", label, rec.Seed, formatSavedSeedAge(now, rec.SavedAt))
			if ok {
				if cur, found := m.curation[bookmarkKey(bookmark.Spec, bookmark.Seed)]; found {
					badges := make([]string, 0, 3)
					if cur.Favorite {
						badges = append(badges, "★")
					}
					if cur.Rating > 0 {
						badges = append(badges, ratingString(cur.Rating))
					}
					if len(cur.Tags) > 0 {
						badges = append(badges, strings.Join(cur.Tags, ","))
					}
					if len(badges) > 0 {
						entry += " · " + strings.Join(badges, " ")
					}
				}
			}
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

func useCompactLayout(w, h int) bool {
	return w < 72 || h < 18
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

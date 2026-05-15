package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
)

// BuildAlgoFn constructs a fresh Algorithm for the given spec. main.go closes
// over the loaded SoundFont (or nil) and any per-build wiring like IR setup.
type BuildAlgoFn func(spec gen.AlgoSpec) gen.Algorithm

// Model is the bubbletea model for termus.
type Model struct {
	width, height int

	ring    *scope.Ring
	cmd     audio.Commander
	algo    string
	debug   gen.DebugStatus
	keyName string
	seed    int64

	volume       int
	paused       bool
	recording    bool
	status       string
	statusTTL    time.Time
	stickyStatus string

	themeIdx  int // index into Themes
	visualIdx int // index into Visuals
	themes    []ColorTheme
	ui        AdaptiveUI

	// Algorithm switching ([n]/[p]).
	genres   []gen.AlgoSpec // ordered list of switchable algorithms
	genreIdx int            // current index into genres
	buildFn  BuildAlgoFn    // closure used to construct a new algorithm

	// Playlist auto-advance.
	playlist     *gen.Playlist
	playlistIdx  int       // index of currently-playing track
	nextTrackAt  time.Time // when to advance to the next track
	playlistFade int       // crossfade length in audio frames (44.1 kHz)
}

// New constructs a Model. keyName is e.g. "Cmin".
func New(ring *scope.Ring, cmd audio.Commander, algo, keyName string, seed int64, initialVol int) Model {
	ui := DetectAdaptiveUI()
	return Model{
		ring:     ring,
		cmd:      cmd,
		algo:     algo,
		debug:    cmd.DebugStatus(),
		keyName:  keyName,
		seed:     seed,
		volume:   initialVol,
		ui:       ui,
		themes:   append([]ColorTheme(nil), ui.Themes...),
		themeIdx: ui.DefaultThemeIdx,
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

// WithPlaylist enables playlist auto-advance. The model walks through the
// playlist's tracks, swapping the algorithm at each track's Duration boundary
// with a crossfade of fadeFrames samples. buildFn must be set via
// WithSwitcher first so the model knows how to construct algorithms.
func (m Model) WithPlaylist(p *gen.Playlist, startIdx int, fadeFrames int) Model {
	m.playlist = p
	m.playlistIdx = startIdx
	m.playlistFade = fadeFrames
	if p != nil && startIdx < len(p.Tracks) {
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
	algo := m.buildFn(track.Spec)
	algo.Seed(track.Seed)
	m.cmd.SwapAlgorithmFade(algo, m.playlistFade)
	m.algo = track.Spec.Display
	m.seed = track.Seed
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
	algo := m.buildFn(spec)
	m.cmd.SwapAlgorithm(algo)
	m.algo = spec.Display
	m.flashStatus("switched: "+spec.Display, 2*time.Second)
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
	case tea.KeyMsg:
		switch matchKey(msg) {
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
				m.flashStatus("rec → "+path, 3*time.Second)
			} else {
				m.recording = false
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
		case actionNextAlgo:
			m.switchAlgo(1)
		case actionPrevAlgo:
			m.switchAlgo(-1)
		case actionNextTrack:
			if m.playlist != nil {
				m.advancePlaylist()
			}
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
	innerH := m.height - 2 // minus top + bottom bars
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
	bottom := bottomBar(m, innerW, theme)
	return lipgloss.JoinVertical(lipgloss.Left, top, scopeStr, bottom)
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
	debug := trimToWidth(gen.FormatDebugStatus(m.debug), maxInt(0, w/2))
	right := ""
	if debug != "" {
		right = lipgloss.NewStyle().Foreground(theme.BarHi).Render(debug)
	}
	if m.recording {
		rec := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5b5b")).Render("● REC")
		if right == "" {
			right = rec
		} else {
			right += "  " + rec
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

func bottomBar(m Model, w int, theme ColorTheme) string {
	state := "play"
	if m.paused {
		state = "PAUSED"
	}
	hint := fmt.Sprintf("[space] %s   [↑↓] vol %d%%   [r] rec",
		state, m.volume)
	if len(m.themes) > 1 {
		hint += fmt.Sprintf("   [c] %s", theme.Name)
	} else {
		hint += fmt.Sprintf("   [%s]", theme.Name)
	}
	hint += fmt.Sprintf("   [C] %s", Visuals[m.visualIdx].Name)
	if len(m.genres) > 1 {
		hint += "   [n/p] algo"
	}
	if m.playlist != nil {
		hint += "   [s] skip"
	}
	hint += "   [q] quit"

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

func maxInt(a, b int) int {
	if a > b {
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

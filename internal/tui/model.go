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
	keyName string
	seed    int64

	volume    int
	paused    bool
	recording bool
	status    string
	statusTTL time.Time

	themeIdx int // index into Themes

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
	return Model{
		ring:    ring,
		cmd:     cmd,
		algo:    algo,
		keyName: keyName,
		seed:    seed,
		volume:  initialVol,
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
	m.status = fmt.Sprintf("▶ %d/%d %s",
		m.playlistIdx+1, len(m.playlist.Tracks), track.Spec.Display)
	m.statusTTL = time.Now().Add(3 * time.Second)

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
	m.status = "switched: " + spec.Display
	m.statusTTL = time.Now().Add(2 * time.Second)
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
				m.status = "rec error: " + err.Error()
				m.statusTTL = time.Now().Add(3 * time.Second)
				m.recording = false
			} else if path != "" {
				m.recording = true
				m.status = "rec → " + path
				m.statusTTL = time.Now().Add(3 * time.Second)
			} else {
				m.recording = false
				m.status = "rec stopped"
				m.statusTTL = time.Now().Add(3 * time.Second)
			}
		case actionTheme:
			m.themeIdx = (m.themeIdx + 1) % len(Themes)
			m.status = "theme: " + Themes[m.themeIdx].Name
			m.statusTTL = time.Now().Add(2 * time.Second)
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

	// Snapshot scope and render with the active theme.
	samples := make([]float64, innerW*2)
	m.ring.Snapshot(samples)
	theme := Themes[m.themeIdx]
	scopeStr := RenderBrailleThemed(samples, innerW, innerH, theme)

	top := topBar(m, innerW, theme)
	bottom := bottomBar(m, innerW, theme)
	return lipgloss.JoinVertical(lipgloss.Left, top, scopeStr, bottom)
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
	left := lipgloss.NewStyle().Foreground(theme.BarFg).Render(label)
	right := ""
	if m.recording {
		right = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5b5b")).Render("● REC")
	}
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
	hint := fmt.Sprintf("[space] %s   [↑↓] vol %d%%   [r] rec   [c] %s",
		state, m.volume, theme.Name)
	if len(m.genres) > 1 {
		hint += "   [n/p] algo"
	}
	if m.playlist != nil {
		hint += "   [s] skip"
	}
	hint += "   [q] quit"
	left := lipgloss.NewStyle().Faint(true).Render(hint)
	right := ""
	if time.Now().Before(m.statusTTL) {
		right = lipgloss.NewStyle().Foreground(theme.BarHi).Render(m.status)
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
			lines[i] = spaces(pad) + text
		} else {
			lines[i] = ""
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/scope"
)

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
		}
		return m, nil
	case tickMsg:
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

	// Snapshot scope and render.
	samples := make([]float64, innerW*2)
	m.ring.Snapshot(samples)
	scopeStr := RenderBraille(samples, innerW, innerH)

	top := topBar(m, innerW)
	bottom := bottomBar(m, innerW)
	return lipgloss.JoinVertical(lipgloss.Left, top, scopeStr, bottom)
}

func topBar(m Model, w int) string {
	left := lipgloss.NewStyle().Foreground(lipgloss.Color("#a0a0ff")).Render(
		fmt.Sprintf("termus · %s · %s · seed=%d", m.algo, m.keyName, m.seed),
	)
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

func bottomBar(m Model, w int) string {
	state := "play"
	if m.paused {
		state = "PAUSED"
	}
	left := lipgloss.NewStyle().Faint(true).Render(
		fmt.Sprintf("[space] %s   [↑↓] vol %d%%   [r] rec   [q] quit", state, m.volume),
	)
	right := ""
	if time.Now().Before(m.statusTTL) {
		right = lipgloss.NewStyle().Foreground(lipgloss.Color("#5bfaff")).Render(m.status)
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

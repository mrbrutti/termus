package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mrbrutti/termus/internal/gen"
)

type ExportController struct {
	Seconds float64
	WAV     func(spec gen.AlgoSpec, seed int64) (string, error)
	MIDI    func(spec gen.AlgoSpec, seed int64) (string, error)
	Stems   func(spec gen.AlgoSpec, seed int64) (string, error)
}

type exportResultMsg struct {
	kind string
	path string
	err  error
}

func runExport(kind string, fn func() (string, error)) tea.Cmd {
	return func() tea.Msg {
		path, err := fn()
		return exportResultMsg{kind: kind, path: path, err: err}
	}
}

func (e ExportController) durationLabel() string {
	if e.Seconds <= 0 {
		return "60s"
	}
	return fmt.Sprintf("%.0fs", e.Seconds)
}

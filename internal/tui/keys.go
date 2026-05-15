package tui

import tea "github.com/charmbracelet/bubbletea"

type keyAction int

const (
	actionNone keyAction = iota
	actionQuit
	actionPause
	actionVolUp
	actionVolDown
	actionRecord
	actionTheme
	actionNextAlgo
	actionPrevAlgo
	actionNextTrack
	actionVisual
)

func matchKey(msg tea.KeyMsg) keyAction {
	switch msg.String() {
	case "q", "ctrl+c":
		return actionQuit
	case " ":
		return actionPause
	case "up", "+":
		return actionVolUp
	case "down", "-":
		return actionVolDown
	case "r":
		return actionRecord
	case "c":
		return actionTheme
	case "C":
		return actionVisual
	case "n", "right":
		return actionNextAlgo
	case "p", "left":
		return actionPrevAlgo
	case "s":
		return actionNextTrack
	}
	return actionNone
}

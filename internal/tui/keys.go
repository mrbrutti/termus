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
	actionDebug
	actionHelp
	actionLibrary
	actionInspector
	actionExport
	actionZen
	actionPrevSeed
	actionNextSeed
	actionStoreA
	actionStoreB
	actionToggleAB
	actionKeepSeed
	actionRejectSeed
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
	case "d":
		return actionDebug
	case "?", "esc":
		return actionHelp
	case "l":
		return actionLibrary
	case "i":
		return actionInspector
	case "e":
		return actionExport
	case "z":
		return actionZen
	case "[":
		return actionPrevSeed
	case "]":
		return actionNextSeed
	case "a":
		return actionStoreA
	case "b":
		return actionStoreB
	case "tab":
		return actionToggleAB
	case "k":
		return actionKeepSeed
	case "x":
		return actionRejectSeed
	}
	return actionNone
}

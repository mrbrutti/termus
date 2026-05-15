package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/mrbrutti/termus/internal/audio"
)

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
		volume:       70,
		stickyStatus: "audio: starting...",
		themes:       []ColorTheme{DefaultTheme()},
	}
	bar := bottomBar(m, 80, DefaultTheme())
	if !strings.Contains(bar, "audio: starting...") {
		t.Fatalf("bottom bar missing status: %q", bar)
	}
}

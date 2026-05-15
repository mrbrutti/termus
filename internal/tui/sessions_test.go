package tui

import (
	"os"
	"testing"
	"time"

	"github.com/mrbrutti/termus/internal/gen"
)

func TestSavedSessionRecordsRoundTrip(t *testing.T) {
	t.Setenv("TERMUS_CONFIG_DIR", t.TempDir())
	want := []savedSessionRecord{
		{Label: "Jazz / 77", Algo: "jazz", Display: "Jazz", Seed: 77, Visual: "scope", Theme: "Default", Volume: 68, SavedAt: time.Now()},
		{Label: "Ambient / 42", Algo: "ambient", Display: "Ambient", Seed: 42, Visual: "vector", Theme: "Default", Volume: 70, SavedAt: time.Now().Add(-time.Minute)},
	}
	if err := saveSavedSessionRecords(want); err != nil {
		t.Fatalf("saveSavedSessionRecords: %v", err)
	}
	got, err := loadSavedSessionRecords()
	if err != nil {
		t.Fatalf("loadSavedSessionRecords: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("session count = %d, want %d", len(got), len(want))
	}
	if got[0].Algo != "jazz" || got[1].Algo != "ambient" {
		t.Fatalf("sessions not sorted newest-first: %+v", got)
	}
	path, err := savedSessionsPath()
	if err != nil {
		t.Fatalf("savedSessionsPath: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected saved session file at %s: %v", path, err)
	}
}

func TestSaveAndLoadCurrentSession(t *testing.T) {
	t.Setenv("TERMUS_CONFIG_DIR", t.TempDir())
	cmd := &tuiCommanderStub{}
	specs := []gen.AlgoSpec{{Name: "ambient", Display: "Ambient"}}
	build := func(spec gen.AlgoSpec, seed int64) gen.Algorithm {
		return &tuiAlgoStub{name: spec.Name}
	}
	m := New(nil, cmd, "Ambient", "Cmin", 42, 70).WithSwitcher(specs, 0, build)
	m.visualIdx = 2
	m.saveCurrentSession()
	if len(m.savedSessions) != 1 {
		t.Fatalf("savedSessions = %d, want 1", len(m.savedSessions))
	}

	m.seed = 91
	m.volume = 25
	m.visualIdx = 0
	m.loadSelectedSession()
	if m.seed != 42 {
		t.Fatalf("seed = %d, want 42", m.seed)
	}
	if m.volume != 70 {
		t.Fatalf("volume = %d, want 70", m.volume)
	}
	if m.visualIdx != 2 {
		t.Fatalf("visualIdx = %d, want 2", m.visualIdx)
	}
	if len(cmd.swaps) != 1 {
		t.Fatalf("swap count = %d, want 1", len(cmd.swaps))
	}
}

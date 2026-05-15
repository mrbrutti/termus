package tui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mrbrutti/termus/internal/gen"
)

func TestSavedSeedRecordsRoundTrip(t *testing.T) {
	t.Setenv("TERMUS_CONFIG_DIR", t.TempDir())
	want := []savedSeedRecord{
		{Algo: "ambient", Display: "Ambient", Seed: 42, SavedAt: time.Now().Add(-time.Minute)},
		{Algo: "jazz", Display: "Jazz", Seed: 77, SavedAt: time.Now()},
	}
	if err := saveSavedSeedRecords(want); err != nil {
		t.Fatalf("saveSavedSeedRecords: %v", err)
	}
	got, err := loadSavedSeedRecords()
	if err != nil {
		t.Fatalf("loadSavedSeedRecords: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("record count = %d, want %d", len(got), len(want))
	}
	if got[0].Algo != "jazz" || got[1].Algo != "ambient" {
		t.Fatalf("records not sorted newest-first: %+v", got)
	}
	path, err := savedSeedsPath()
	if err != nil {
		t.Fatalf("savedSeedsPath: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected saved seed file at %s: %v", path, err)
	}
}

func TestNewLoadsSavedSeeds(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TERMUS_CONFIG_DIR", tmp)
	if err := saveSavedSeedRecords([]savedSeedRecord{{Algo: "ambient", Display: "Ambient", Seed: 42, SavedAt: time.Now()}}); err != nil {
		t.Fatalf("saveSavedSeedRecords: %v", err)
	}
	m := New(nil, &tuiCommanderStub{}, "Ambient", "Cmin", 42, 70)
	if len(m.savedSeeds) != 1 {
		t.Fatalf("savedSeeds = %d, want 1", len(m.savedSeeds))
	}
	if len(m.kept) != 1 {
		t.Fatalf("kept = %d, want 1", len(m.kept))
	}
	path := filepath.Join(tmp, "saved_seeds.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected persisted file at %s: %v", path, err)
	}
}

func TestKeepSeedPersistsLibraryAndRecall(t *testing.T) {
	t.Setenv("TERMUS_CONFIG_DIR", t.TempDir())
	cmd := &tuiCommanderStub{}
	specs := []gen.AlgoSpec{{Name: "ambient", Display: "Ambient"}}
	build := func(spec gen.AlgoSpec, seed int64) gen.Algorithm {
		return &tuiAlgoStub{name: spec.Name}
	}
	m := New(nil, cmd, "Ambient", "Cmin", 42, 70).WithSwitcher(specs, 0, build)
	m.keepSeed()
	if len(m.savedSeeds) != 1 {
		t.Fatalf("savedSeeds = %d, want 1", len(m.savedSeeds))
	}
	m.seed = 99
	m.libraryVisible = true
	m.recallLibrarySeed()
	if m.seed != 42 {
		t.Fatalf("recallLibrarySeed restored seed %d, want 42", m.seed)
	}
	if len(cmd.swaps) != 1 {
		t.Fatalf("swap count = %d, want 1", len(cmd.swaps))
	}
}

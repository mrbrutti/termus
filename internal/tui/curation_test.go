package tui

import (
	"os"
	"testing"
	"time"

	"github.com/mrbrutti/termus/internal/gen"
)

func TestSeedCurationRoundTrip(t *testing.T) {
	t.Setenv("TERMUS_CONFIG_DIR", t.TempDir())
	records := map[string]seedCurationRecord{
		"ambient:42": {
			Algo:         "ambient",
			Display:      "Ambient",
			Seed:         42,
			Favorite:     true,
			Rating:       4,
			Tags:         []string{"calm", "warm"},
			PlayCount:    3,
			LastPlayedAt: time.Now(),
			UpdatedAt:    time.Now(),
		},
	}
	if err := saveSeedCuration(records); err != nil {
		t.Fatalf("saveSeedCuration: %v", err)
	}
	got, err := loadSeedCuration()
	if err != nil {
		t.Fatalf("loadSeedCuration: %v", err)
	}
	rec, ok := got["ambient:42"]
	if !ok {
		t.Fatalf("expected ambient:42 in curation map: %+v", got)
	}
	if !rec.Favorite || rec.Rating != 4 || len(rec.Tags) != 2 {
		t.Fatalf("unexpected curation record: %+v", rec)
	}
	path, err := seedCurationPath()
	if err != nil {
		t.Fatalf("seedCurationPath: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected curation file at %s: %v", path, err)
	}
}

func TestCurrentSeedCurationControls(t *testing.T) {
	t.Setenv("TERMUS_CONFIG_DIR", t.TempDir())
	cmd := &tuiCommanderStub{}
	specs := []gen.AlgoSpec{{Name: "ambient", Display: "Ambient"}}
	build := func(spec gen.AlgoSpec, seed int64) gen.Algorithm {
		return &tuiAlgoStub{name: spec.Name}
	}
	m := New(nil, cmd, "Ambient", "Cmin", 42, 70).WithSwitcher(specs, 0, build)
	m.adjustCurrentRating(3)
	m.toggleCurrentFavorite()
	m.curateTagIdx = 0
	m.toggleCurrentTag()
	rec, ok := m.currentSeedRecord()
	if !ok {
		t.Fatal("expected current seed record")
	}
	if rec.Rating != 3 || !rec.Favorite || !hasTag(rec.Tags, curationTags[0]) {
		t.Fatalf("unexpected current seed curation: %+v", rec)
	}
	if len(m.currentBestRecords()) == 0 {
		t.Fatal("expected rated/favorited record to appear in best takes")
	}
	if len(m.currentRecentRecords()) == 0 {
		t.Fatal("expected touched record to appear in recent history")
	}
}

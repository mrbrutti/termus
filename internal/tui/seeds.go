package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/mrbrutti/termus/internal/gen"
)

type savedSeedRecord struct {
	Algo    string    `json:"algo"`
	Display string    `json:"display,omitempty"`
	Seed    int64     `json:"seed"`
	SavedAt time.Time `json:"saved_at"`
}

func savedSeedsPath() (string, error) {
	if root := os.Getenv("TERMUS_CONFIG_DIR"); root != "" {
		return filepath.Join(root, "saved_seeds.json"), nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "termus", "saved_seeds.json"), nil
}

func loadSavedSeedRecords() ([]savedSeedRecord, error) {
	path, err := savedSeedsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var records []savedSeedRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].SavedAt.After(records[j].SavedAt)
	})
	return records, nil
}

func saveSavedSeedRecords(records []savedSeedRecord) error {
	path, err := savedSeedsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func recordsToBookmarks(records []savedSeedRecord) map[string]seedBookmark {
	if len(records) == 0 {
		return nil
	}
	out := make(map[string]seedBookmark, len(records))
	for _, rec := range records {
		spec, ok := gen.Resolve(rec.Algo)
		if !ok {
			continue
		}
		out[bookmarkKey(spec, rec.Seed)] = seedBookmark{Spec: spec, Seed: rec.Seed}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func removeSavedSeedRecord(records []savedSeedRecord, algo string, seed int64) []savedSeedRecord {
	out := records[:0]
	for _, rec := range records {
		if rec.Algo == algo && rec.Seed == seed {
			continue
		}
		out = append(out, rec)
	}
	return out
}

func resolveSavedSeedRecord(rec savedSeedRecord) (seedBookmark, string, bool) {
	spec, ok := gen.Resolve(rec.Algo)
	if !ok {
		label := rec.Display
		if label == "" {
			label = rec.Algo
		}
		return seedBookmark{}, label, false
	}
	return seedBookmark{Spec: spec, Seed: rec.Seed}, spec.Label(), true
}

func formatSavedSeedAge(now, savedAt time.Time) string {
	if savedAt.IsZero() {
		return "just now"
	}
	age := now.Sub(savedAt)
	switch {
	case age < time.Minute:
		return "just now"
	case age < time.Hour:
		return pluralDuration(int(age.Minutes()), "m")
	case age < 24*time.Hour:
		return pluralDuration(int(age.Hours()), "h")
	default:
		return pluralDuration(int(age.Hours()/24), "d")
	}
}

func pluralDuration(n int, suffix string) string {
	if n < 1 {
		n = 1
	}
	return strconv.Itoa(n) + suffix
}

package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mrbrutti/termus/internal/gen"
)

var curationTags = []string{
	"calm",
	"warm",
	"dark",
	"bright",
	"dreamy",
	"driving",
	"glass",
	"tape",
}

type seedCurationRecord struct {
	Algo         string    `json:"algo"`
	Display      string    `json:"display,omitempty"`
	Seed         int64     `json:"seed"`
	Favorite     bool      `json:"favorite,omitempty"`
	Kept         bool      `json:"kept,omitempty"`
	Rating       int       `json:"rating,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
	PlayCount    int       `json:"play_count,omitempty"`
	LastPlayedAt time.Time `json:"last_played_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

func seedCurationPath() (string, error) {
	if root := os.Getenv("TERMUS_CONFIG_DIR"); root != "" {
		return filepath.Join(root, "curation.json"), nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "termus", "curation.json"), nil
}

func loadSeedCuration() (map[string]seedCurationRecord, error) {
	path, err := seedCurationPath()
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
	var records []seedCurationRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	out := make(map[string]seedCurationRecord, len(records))
	for _, rec := range records {
		spec, ok := gen.Resolve(rec.Algo)
		if ok {
			out[bookmarkKey(spec, rec.Seed)] = rec
			continue
		}
		out[rec.Algo+":"+formatSeed(rec.Seed)] = rec
	}
	return out, nil
}

func saveSeedCuration(records map[string]seedCurationRecord) error {
	path, err := seedCurationPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	flat := make([]seedCurationRecord, 0, len(records))
	for _, rec := range records {
		flat = append(flat, rec)
	}
	sort.SliceStable(flat, func(i, j int) bool {
		return flat[i].UpdatedAt.After(flat[j].UpdatedAt)
	})
	data, err := json.MarshalIndent(flat, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func formatSeed(seed int64) string {
	return strconv.FormatInt(seed, 10)
}

func (m *Model) currentCurationKey() (string, gen.AlgoSpec, bool) {
	spec, ok := m.currentSpec()
	if !ok {
		return "", gen.AlgoSpec{}, false
	}
	return bookmarkKey(spec, m.seed), spec, true
}

func (m *Model) currentSeedRecord() (seedCurationRecord, bool) {
	key, spec, ok := m.currentCurationKey()
	if !ok {
		return seedCurationRecord{}, false
	}
	if m.curation == nil {
		m.curation = map[string]seedCurationRecord{}
	}
	rec, found := m.curation[key]
	if !found {
		rec = seedCurationRecord{
			Algo:    spec.Name,
			Display: spec.Display,
			Seed:    m.seed,
		}
	}
	return rec, true
}

func (m *Model) storeCurrentSeedRecord(rec seedCurationRecord) {
	key, _, ok := m.currentCurationKey()
	if !ok {
		return
	}
	if m.curation == nil {
		m.curation = map[string]seedCurationRecord{}
	}
	rec.UpdatedAt = time.Now()
	m.curation[key] = rec
	_ = saveSeedCuration(m.curation)
}

func (m *Model) touchCurrentSeed() {
	rec, ok := m.currentSeedRecord()
	if !ok {
		return
	}
	rec.PlayCount++
	rec.LastPlayedAt = time.Now()
	m.storeCurrentSeedRecord(rec)
}

func (m *Model) markCurrentSeedKept() {
	rec, ok := m.currentSeedRecord()
	if !ok {
		return
	}
	rec.Kept = true
	m.storeCurrentSeedRecord(rec)
}

func (m *Model) toggleCurrentFavorite() {
	rec, ok := m.currentSeedRecord()
	if !ok {
		return
	}
	rec.Favorite = !rec.Favorite
	m.storeCurrentSeedRecord(rec)
	if rec.Favorite {
		m.flashStatus("favorite: on", 2*time.Second)
	} else {
		m.flashStatus("favorite: off", 2*time.Second)
	}
}

func (m *Model) adjustCurrentRating(delta int) {
	rec, ok := m.currentSeedRecord()
	if !ok {
		return
	}
	rec.Rating += delta
	if rec.Rating < 0 {
		rec.Rating = 0
	}
	if rec.Rating > 5 {
		rec.Rating = 5
	}
	m.storeCurrentSeedRecord(rec)
	m.flashStatus("rating: "+ratingString(rec.Rating), 2*time.Second)
}

func (m *Model) cycleCurationTag(delta int) {
	if len(curationTags) == 0 {
		return
	}
	m.curateTagIdx = (m.curateTagIdx + delta + len(curationTags)) % len(curationTags)
}

func (m *Model) toggleCurrentTag() {
	rec, ok := m.currentSeedRecord()
	if !ok || len(curationTags) == 0 {
		return
	}
	tag := curationTags[m.curateTagIdx]
	if hasTag(rec.Tags, tag) {
		rec.Tags = removeTag(rec.Tags, tag)
		m.flashStatus("tag removed: "+tag, 2*time.Second)
	} else {
		rec.Tags = append(rec.Tags, tag)
		sort.Strings(rec.Tags)
		m.flashStatus("tag added: "+tag, 2*time.Second)
	}
	m.storeCurrentSeedRecord(rec)
}

func (m Model) currentRecentRecords() []seedCurationRecord {
	if len(m.curation) == 0 {
		return nil
	}
	out := make([]seedCurationRecord, 0, len(m.curation))
	for _, rec := range m.curation {
		if rec.PlayCount > 0 || !rec.LastPlayedAt.IsZero() {
			out = append(out, rec)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].LastPlayedAt.After(out[j].LastPlayedAt)
	})
	return out
}

func (m Model) currentBestRecords() []seedCurationRecord {
	if len(m.curation) == 0 {
		return nil
	}
	out := make([]seedCurationRecord, 0, len(m.curation))
	for _, rec := range m.curation {
		if rec.Favorite || rec.Kept || rec.Rating > 0 {
			out = append(out, rec)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Favorite != out[j].Favorite {
			return out[i].Favorite
		}
		if out[i].Rating != out[j].Rating {
			return out[i].Rating > out[j].Rating
		}
		if out[i].Kept != out[j].Kept {
			return out[i].Kept
		}
		if out[i].PlayCount != out[j].PlayCount {
			return out[i].PlayCount > out[j].PlayCount
		}
		return out[i].LastPlayedAt.After(out[j].LastPlayedAt)
	})
	return out
}

func (m *Model) browseRecent(delta int) {
	records := m.currentRecentRecords()
	if len(records) == 0 {
		m.curateRecentIdx = 0
		return
	}
	m.curateRecentIdx = (m.curateRecentIdx + delta + len(records)) % len(records)
}

func (m *Model) browseBest(delta int) {
	records := m.currentBestRecords()
	if len(records) == 0 {
		m.curateBestIdx = 0
		return
	}
	m.curateBestIdx = (m.curateBestIdx + delta + len(records)) % len(records)
}

func (m Model) selectedRecentRecord() (seedCurationRecord, bool) {
	records := m.currentRecentRecords()
	if len(records) == 0 {
		return seedCurationRecord{}, false
	}
	return records[clampInt(m.curateRecentIdx, 0, len(records)-1)], true
}

func (m Model) selectedBestRecord() (seedCurationRecord, bool) {
	records := m.currentBestRecords()
	if len(records) == 0 {
		return seedCurationRecord{}, false
	}
	return records[clampInt(m.curateBestIdx, 0, len(records)-1)], true
}

func (m *Model) recallCurated(rec seedCurationRecord, prefix string) {
	spec, ok := gen.Resolve(rec.Algo)
	if !ok {
		label := rec.Display
		if label == "" {
			label = rec.Algo
		}
		m.flashStatus("saved algo unavailable: "+label, 3*time.Second)
		return
	}
	m.swapToSeed(spec, rec.Seed, prefix+" → "+curationLabel(rec))
}

func (m *Model) loadSelectedRecent() {
	rec, ok := m.selectedRecentRecord()
	if !ok {
		m.flashStatus("no recent takes", 2*time.Second)
		return
	}
	m.recallCurated(rec, "recent")
}

func (m *Model) loadSelectedBest() {
	rec, ok := m.selectedBestRecord()
	if !ok {
		m.flashStatus("no best takes", 2*time.Second)
		return
	}
	m.recallCurated(rec, "best")
}

func curationLabel(rec seedCurationRecord) string {
	label := rec.Display
	if label == "" {
		label = rec.Algo
	}
	return label + "/" + formatSeed(rec.Seed)
}

func currentTagsLabel(tags []string) string {
	if len(tags) == 0 {
		return "none"
	}
	return strings.Join(tags, ", ")
}

func ratingString(rating int) string {
	if rating <= 0 {
		return "0"
	}
	return strings.Repeat("★", rating)
}

func hasTag(tags []string, target string) bool {
	for _, tag := range tags {
		if strings.EqualFold(tag, target) {
			return true
		}
	}
	return false
}

func removeTag(tags []string, target string) []string {
	out := tags[:0]
	for _, tag := range tags {
		if strings.EqualFold(tag, target) {
			continue
		}
		out = append(out, tag)
	}
	return out
}

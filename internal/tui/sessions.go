package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mrbrutti/termus/internal/gen"
)

type savedSessionRecord struct {
	Label    string             `json:"label"`
	Algo     string             `json:"algo"`
	Display  string             `json:"display,omitempty"`
	Seed     int64              `json:"seed"`
	Visual   string             `json:"visual"`
	Theme    string             `json:"theme"`
	Volume   int                `json:"volume"`
	Controls gen.ControlProfile `json:"controls"`
	Morph    int                `json:"morph"`
	Playlist string             `json:"playlist,omitempty"`
	Track    int                `json:"track,omitempty"`
	SavedAt  time.Time          `json:"saved_at"`
}

func savedSessionsPath() (string, error) {
	if root := os.Getenv("TERMUS_CONFIG_DIR"); root != "" {
		return filepath.Join(root, "sessions.json"), nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "termus", "sessions.json"), nil
}

func loadSavedSessionRecords() ([]savedSessionRecord, error) {
	path, err := savedSessionsPath()
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
	var records []savedSessionRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].SavedAt.After(records[j].SavedAt)
	})
	return records, nil
}

func saveSavedSessionRecords(records []savedSessionRecord) error {
	path, err := savedSessionsPath()
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

func removeSavedSessionRecord(records []savedSessionRecord, savedAt time.Time) []savedSessionRecord {
	out := records[:0]
	for _, rec := range records {
		if rec.SavedAt.Equal(savedAt) {
			continue
		}
		out = append(out, rec)
	}
	return out
}

func sessionLabel(rec savedSessionRecord) string {
	if rec.Label != "" {
		return rec.Label
	}
	label := rec.Display
	if label == "" {
		label = rec.Algo
	}
	if label == "" {
		label = "session"
	}
	return fmt.Sprintf("%s / %d", label, rec.Seed)
}

func formatSessionAge(now, savedAt time.Time) string {
	return formatSavedSeedAge(now, savedAt)
}

func (m *Model) saveCurrentSession() {
	spec, ok := m.currentSpec()
	if !ok {
		if resolved, found := gen.Resolve(strings.ToLower(m.algo)); found {
			spec = resolved
			ok = true
		}
	}
	if !ok {
		m.flashStatus("session save unavailable", 3*time.Second)
		return
	}
	rec := savedSessionRecord{
		Label:    fmt.Sprintf("%s / %d", spec.Display, m.seed),
		Algo:     spec.Name,
		Display:  spec.Display,
		Seed:     m.seed,
		Visual:   Visuals[m.visualIdx].Name,
		Theme:    m.activeTheme().Name,
		Volume:   m.volume,
		Controls: gen.DefaultControlProfile(),
		Morph:    m.morphMode,
		SavedAt:  time.Now(),
	}
	if m.musicProfile != nil {
		rec.Controls = *m.musicProfile
	}
	if m.playlist != nil {
		rec.Playlist = m.playlist.Name
		rec.Track = m.playlistIdx + 1
	}
	m.savedSessions = append([]savedSessionRecord{rec}, m.savedSessions...)
	if err := saveSavedSessionRecords(m.savedSessions); err != nil {
		m.flashStatus("session save failed", 3*time.Second)
		return
	}
	m.sessionIdx = 0
	m.flashStatus("session saved", 2*time.Second)
}

func (m *Model) browseSession(delta int) {
	if len(m.savedSessions) == 0 {
		m.sessionIdx = 0
		return
	}
	m.sessionIdx = (m.sessionIdx + delta + len(m.savedSessions)) % len(m.savedSessions)
}

func (m *Model) selectedSession() (savedSessionRecord, bool) {
	if len(m.savedSessions) == 0 {
		return savedSessionRecord{}, false
	}
	idx := clampInt(m.sessionIdx, 0, len(m.savedSessions)-1)
	return m.savedSessions[idx], true
}

func (m *Model) loadSelectedSession() {
	rec, ok := m.selectedSession()
	if !ok {
		m.flashStatus("no saved sessions", 2*time.Second)
		return
	}
	if idx := themeIndexByName(m.themes, rec.Theme); idx >= 0 {
		m.themeIdx = idx
	}
	if idx := visualIndexByName(rec.Visual); idx >= 0 {
		m.visualIdx = idx
	}
	if rec.Volume >= 0 && rec.Volume <= 100 {
		m.volume = rec.Volume
		m.cmd.SetVolume(rec.Volume)
	}
	if rec.Controls == (gen.ControlProfile{}) {
		rec.Controls = gen.DefaultControlProfile()
	}
	*m.ensureMusicProfile() = rec.Controls
	m.morphMode = clampInt(rec.Morph, 0, 4)
	spec, resolved := gen.Resolve(rec.Algo)
	if !resolved {
		label := rec.Display
		if label == "" {
			label = rec.Algo
		}
		m.flashStatus("saved algo unavailable: "+label, 3*time.Second)
		return
	}
	if m.playlist != nil {
		m.flashStatus("session loaded: view only during playlist", 3*time.Second)
		return
	}
	m.swapToSeed(spec, rec.Seed, "session → "+sessionLabel(rec))
}

func (m *Model) deleteSelectedSession() {
	rec, ok := m.selectedSession()
	if !ok {
		m.flashStatus("no saved sessions", 2*time.Second)
		return
	}
	m.savedSessions = append([]savedSessionRecord(nil), removeSavedSessionRecord(m.savedSessions, rec.SavedAt)...)
	if m.sessionIdx >= len(m.savedSessions) {
		m.sessionIdx = maxInt(0, len(m.savedSessions)-1)
	}
	if err := saveSavedSessionRecords(m.savedSessions); err != nil {
		m.flashStatus("session delete failed", 3*time.Second)
		return
	}
	m.flashStatus("session removed", 2*time.Second)
}

func themeIndexByName(themes []ColorTheme, name string) int {
	for i, theme := range themes {
		if strings.EqualFold(theme.Name, name) {
			return i
		}
	}
	return -1
}

func visualIndexByName(name string) int {
	for i, visual := range Visuals {
		if strings.EqualFold(visual.Name, name) {
			return i
		}
	}
	return -1
}

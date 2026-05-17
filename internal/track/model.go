package track

import (
	"fmt"

	"github.com/mrbrutti/termus/internal/gen"
)

type File struct {
	Title       string          `yaml:"title"`
	Description string          `yaml:"description,omitempty"`
	Style       string          `yaml:"style"`
	ListenMode  string          `yaml:"listen_mode,omitempty"`
	Seed        int64           `yaml:"seed,omitempty"`
	Tags        []string        `yaml:"tags,omitempty"`
	Key         string          `yaml:"key,omitempty"`
	Tempo       string          `yaml:"tempo,omitempty"`
	Roles       map[string]Role `yaml:"roles,omitempty"`
	Sections    []Section       `yaml:"sections"`
	Globals     Profile         `yaml:"globals,omitempty"`
	Lint        LintControl     `yaml:"lint,omitempty"`
}

type Section struct {
	ID         string          `yaml:"id,omitempty"`
	Title      string          `yaml:"title,omitempty"`
	Derive     string          `yaml:"derive,omitempty"`
	Transforms []string        `yaml:"transforms,omitempty"`
	Duration   string          `yaml:"duration"`
	Seed       *int64          `yaml:"seed,omitempty"`
	SeedOffset *int64          `yaml:"seed_offset,omitempty"`
	Key        string          `yaml:"key,omitempty"`
	Tempo      string          `yaml:"tempo,omitempty"`
	Harmony    string          `yaml:"harmony,omitempty"`
	Scene      string          `yaml:"scene,omitempty"`
	Variation  string          `yaml:"variation,omitempty"`
	Profile    Profile         `yaml:"profile,omitempty"`
	Roles      map[string]Role `yaml:"roles,omitempty"`
	Events     []Event         `yaml:"events,omitempty"`
}

type Role struct {
	Family       string   `yaml:"family,omitempty"`
	Tone         []string `yaml:"tone,omitempty"`
	Articulation string   `yaml:"articulation,omitempty"`
	Register     string   `yaml:"register,omitempty"`
	Prominence   string   `yaml:"prominence,omitempty"`
	Pattern      string   `yaml:"pattern,omitempty"`
	Motif        string   `yaml:"motif,omitempty"`
	Harmony      string   `yaml:"harmony,omitempty"`
	Active       *bool    `yaml:"active,omitempty"`
}

type Event struct {
	Kind    string   `yaml:"kind"`
	Bar     int      `yaml:"bar,omitempty"`
	Bars    int      `yaml:"bars,omitempty"`
	Slot    int      `yaml:"slot,omitempty"`
	Roles   []string `yaml:"roles,omitempty"`
	Pattern string   `yaml:"pattern,omitempty"`
	Motif   string   `yaml:"motif,omitempty"`
}

type Profile struct {
	Density    MacroValue `yaml:"density,omitempty"`
	Brightness MacroValue `yaml:"brightness,omitempty"`
	Motion     MacroValue `yaml:"motion,omitempty"`
	Reverb     MacroValue `yaml:"reverb,omitempty"`
	Swing      MacroValue `yaml:"swing,omitempty"`
	DroneDepth MacroValue `yaml:"drone_depth,omitempty"`
	Tempo      MacroValue `yaml:"tempo,omitempty"`
	Phrase     MacroValue `yaml:"phrase,omitempty"`
}

type MacroValue struct {
	set bool
	raw string
}

type LintControl struct {
	RequireContrast bool `yaml:"require_contrast,omitempty"`
}

type Warning struct {
	Path    string
	Message string
}

type Compiled struct {
	Playlist gen.Playlist
	Profiles map[string]gen.ControlProfile
	Plans    map[string]gen.AuthoredTrackPlan
	Warnings []Warning
}

type Entry struct {
	ID          string
	Path        string
	Style       string
	Title       string
	Description string
	Tags        []string
	Key         string
	Tempo       string
	ListenMode  string
	Sections    []string
}

func playlistKey(spec gen.AlgoSpec, seed int64) string {
	return fmt.Sprintf("%s:%d", spec.Name, seed)
}

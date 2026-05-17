package tm

import (
	"fmt"

	"github.com/mrbrutti/termus/internal/gen"
)

type Score struct {
	Title       string      `yaml:"title"`
	Description string      `yaml:"description,omitempty"`
	ListenMode  string      `yaml:"listen_mode,omitempty"`
	Seed        int64       `yaml:"seed,omitempty"`
	Tags        []string    `yaml:"tags,omitempty"`
	Sections    []Section   `yaml:"sections"`
	Globals     TMProfile   `yaml:"globals,omitempty"`
	Lint        LintControl `yaml:"lint,omitempty"`
}

type Section struct {
	Title      string    `yaml:"title"`
	Algo       string    `yaml:"algo"`
	Duration   string    `yaml:"duration"`
	Seed       *int64    `yaml:"seed,omitempty"`
	SeedOffset *int64    `yaml:"seed_offset,omitempty"`
	Profile    TMProfile `yaml:"profile,omitempty"`
	Audit      Audit     `yaml:"audit,omitempty"`
}

type TMProfile struct {
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

type Audit struct {
	Form      string `yaml:"form,omitempty"`
	Harmony   string `yaml:"harmony,omitempty"`
	Lead      string `yaml:"lead,omitempty"`
	Comp      string `yaml:"comp,omitempty"`
	Drums     string `yaml:"drums,omitempty"`
	Arrange   string `yaml:"arrange,omitempty"`
	Variation string `yaml:"variation,omitempty"`
}

type LintControl struct {
	RequireContrast bool `yaml:"require_contrast,omitempty"`
}

type Warning struct {
	Path    string
	Message string
}

type Compiled struct {
	Playlist   gen.Playlist
	Overrides  map[string]gen.ControlProfile
	Blueprints map[string]gen.ScoreBlueprint
	Warnings   []Warning
}

func overrideKey(spec gen.AlgoSpec, seed int64) string {
	return fmt.Sprintf("%s:%d", spec.Name, seed)
}

package track

import (
	"fmt"

	"github.com/mrbrutti/termus/internal/gen"
)

type File struct {
	Title           string          `yaml:"title"`
	Description     string          `yaml:"description,omitempty"`
	Style           string          `yaml:"style"`
	Substyle        string          `yaml:"substyle,omitempty"`
	ListenMode      string          `yaml:"listen_mode,omitempty"`
	Seed            int64           `yaml:"seed,omitempty"`
	Tags            []string        `yaml:"tags,omitempty"`
	Key             string          `yaml:"key,omitempty"`
	Tempo           string          `yaml:"tempo,omitempty"`
	// MixBus is an optional top-level mix-bus profile selector (SP6).
	// One of: lofi, jazz, chill, ambient. Resolved via gen.MixBusByName.
	// If absent, no profile is applied (behavior unchanged).
	MixBus          string          `yaml:"mix_bus,omitempty"`
	Roles           map[string]Role `yaml:"roles,omitempty"`
	Sections        []Section       `yaml:"sections"`
	Globals         Profile         `yaml:"globals,omitempty"`
	VariationBudget VariationBudget `yaml:"variation_budget,omitempty"`
	Lint            LintControl     `yaml:"lint,omitempty"`
}

type Section struct {
	ID            string          `yaml:"id,omitempty"`
	Title         string          `yaml:"title,omitempty"`
	Derive        string          `yaml:"derive,omitempty"`
	Transforms    []string        `yaml:"transforms,omitempty"`
	Duration      string          `yaml:"duration"`
	Seed          *int64          `yaml:"seed,omitempty"`
	SeedOffset    *int64          `yaml:"seed_offset,omitempty"`
	Key           string          `yaml:"key,omitempty"`
	Tempo         string          `yaml:"tempo,omitempty"`
	Harmony       string          `yaml:"harmony,omitempty"`
	Scene         string          `yaml:"scene,omitempty"`
	Variation     string          `yaml:"variation,omitempty"`
	// Groove is an optional named groove template (SP6). References a
	// GrooveTemplate by name via gen.GrooveByName. If absent, no template
	// is applied (existing behaviour).
	Groove        string          `yaml:"groove,omitempty"`
	// HarmonyChords holds voice-leading-aware chord specs (SP6).
	// Accepts both plain string and map form per entry; see ChordSpec.
	// When present it augments the plain Harmony string with voicing hints.
	HarmonyChords []ChordSpec     `yaml:"harmony_chords,omitempty"`
	Profile       Profile         `yaml:"profile,omitempty"`
	Roles         map[string]Role `yaml:"roles,omitempty"`
	Orchestration Orchestration   `yaml:"orchestration,omitempty"`
	Arrangement   Arrangement     `yaml:"arrangement,omitempty"`
	Events        []Event         `yaml:"events,omitempty"`
}

type Arrangement struct {
	Events []Event `yaml:"events,omitempty"`
}

type Orchestration struct {
	Roles map[string]OrchestrationRole `yaml:"roles,omitempty"`
}

type OrchestrationRole struct {
	Family       string   `yaml:"family,omitempty"`
	Tone         []string `yaml:"tone,omitempty"`
	Articulation string   `yaml:"articulation,omitempty"`
	Register     string   `yaml:"register,omitempty"`
	Prominence   string   `yaml:"prominence,omitempty"`
	Active       *bool    `yaml:"active,omitempty"`
}

// WowOverride is an optional per-role wow modulator override (SP6).
// Zero values mean "not set" — the mix-bus or algorithm default applies.
type WowOverride struct {
	DepthCents float64 `yaml:"depth_cents,omitempty"`
	RateHz     float64 `yaml:"rate_hz,omitempty"`
}

type Role struct {
	Family       string                 `yaml:"family,omitempty"`
	Tone         []string               `yaml:"tone,omitempty"`
	Articulation string                 `yaml:"articulation,omitempty"`
	Register     string                 `yaml:"register,omitempty"`
	Prominence   string                 `yaml:"prominence,omitempty"`
	Pattern      string                 `yaml:"pattern,omitempty"`
	Motif        string                 `yaml:"motif,omitempty"`
	Harmony      string                 `yaml:"harmony,omitempty"`
	Phrases      map[string]PhraseBlock `yaml:"phrases,omitempty"`
	Active       *bool                  `yaml:"active,omitempty"`
	// Character knobs (SP6) — all optional; zero = use default.
	// Personality selects a synth.PersonalityPreset by name.
	Personality string      `yaml:"personality,omitempty"`
	// Room selects a synth.IRPreset by name for the per-role reverb.
	Room        string      `yaml:"room,omitempty"`
	// ReverbSendDB is the wet level into the reverb bus in dBFS (e.g. -12).
	ReverbSendDB *float64   `yaml:"reverb_send_db,omitempty"`
	// Wow is an optional per-role wow modulator override.
	Wow         *WowOverride `yaml:"wow,omitempty"`
	// VelocityCurve is a string identifier for a velocity mapping preset.
	VelocityCurve string    `yaml:"velocity_curve,omitempty"`
}

type PhraseBlock struct {
	Pattern string `yaml:"pattern,omitempty"`
	Motif   string `yaml:"motif,omitempty"`
	Harmony string `yaml:"harmony,omitempty"`
	Active  *bool  `yaml:"active,omitempty"`
}

// ChordSpec is the voice-leading-aware chord specification (SP6).
// It can be authored in YAML as either a plain string (chord symbol only)
// or as a mapping with optional voice-leading directives.
// Use the harmony_chords field on Section for this richer form.
//
//	harmony_chords:
//	  - Cmaj9                              # string form → ChordSpec{Symbol:"Cmaj9"}
//	  - {chord: Am7, voicing: rootless_A}  # map form
type ChordSpec struct {
	Symbol  string `yaml:"chord"`   // e.g. "Cmaj9"
	Voicing string `yaml:"voicing"` // e.g. "drop2", "rootless_A", "" = default
	Top     string `yaml:"top"`     // e.g. "9", "3" — top note of voicing
	Smooth  bool   `yaml:"smooth"`  // true = pick inversion minimising voice motion
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

type VariationBudget struct {
	MaxHarmonyRepeat       int  `yaml:"max_harmony_repeat,omitempty"`
	MaxSceneRepeat         int  `yaml:"max_scene_repeat,omitempty"`
	MaxMotifRepeat         int  `yaml:"max_motif_repeat,omitempty"`
	RequireReturnTransform bool `yaml:"require_return_transform,omitempty"`
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
	ID           string
	Path         string
	Style        string
	Substyle     string
	Title        string
	Description  string
	Tags         []string
	Key          string
	Tempo        string
	ListenMode   string
	SectionCount int
	Sections     []string
	Ensemble     []string
	EventCount   int
	Complexity   string
	Structure    []EntrySection
}

type EntrySection struct {
	ID        string
	Label     string
	Harmony   string
	RoleNames []string
	Events    []string
}

func playlistKey(spec gen.AlgoSpec, seed int64) string {
	return fmt.Sprintf("%s:%d", spec.Name, seed)
}

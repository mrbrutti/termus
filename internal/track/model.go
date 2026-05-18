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
	// Motifs is an optional library of named motifs (SP7).
	// Each entry may reference others via based_on and apply textual
	// transforms (transpose, retrograde, invert, augment, diminish).
	Motifs          []MotifEntry    `yaml:"motifs,omitempty"`
	// ChordMarkov is an optional file-level Markov table for chord progressions
	// (SP7). Weights per state should sum to ~1.0; a warning is emitted if not.
	ChordMarkov     *ChordMarkov    `yaml:"chord_markov,omitempty"`
	Roles           map[string]Role `yaml:"roles,omitempty"`
	Sections        []Section       `yaml:"sections"`
	Globals         Profile         `yaml:"globals,omitempty"`
	VariationBudget VariationBudget `yaml:"variation_budget,omitempty"`
	Lint            LintControl     `yaml:"lint,omitempty"`
}

// MotifEntry defines a named motif with optional transforms (SP7).
// Transforms are applied textually on the pattern string.
type MotifEntry struct {
	Name       string  `yaml:"name"`
	Pattern    string  `yaml:"pattern"`   // e.g. "5 . . 7 | 9 . 7 5"
	BasedOn    string  `yaml:"based_on"`  // optional: name of parent motif
	Transpose  int     `yaml:"transpose"` // semitones — shifts scale-degree digits
	Retrograde bool    `yaml:"retrograde"`
	Invert     int     `yaml:"invert"`   // 0 = no, else pivot scale degree
	Augment    float64 `yaml:"augment"`  // duration multiplier, e.g. 1.5
	Diminish   float64 `yaml:"diminish"` // duration multiplier, e.g. 0.5
}

// ChordMarkov holds a Markov successor table for chord progressions (SP7).
// Weights per state should sum to ~1.0; a warning is emitted by ValidateChordMarkov.
type ChordMarkov struct {
	Transitions map[string]map[string]float64 `yaml:"transitions"`
}

// NotePool is a weighted-random note pool for a role (SP7).
// Keys are scale degrees (e.g. "1", "3", "5") and values are relative weights.
// Weights should sum to ~1.0; a warning is emitted if they do not.
type NotePool struct {
	Choices map[string]float64 `yaml:"choices"`
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
	// Automation holds per-section breakpoint curves for parameters like
	// cutoff, pan, expression (SP7). Inert at render time; consumed by
	// future compile-time rewriters.
	Automation    []AutomationLane `yaml:"automation,omitempty"`
	// Substitutions holds harmonic substitution directives for this section
	// (SP7). Applied deterministically via a seed when compiling.
	Substitutions []SubstitutionRule `yaml:"substitutions,omitempty"`
	Profile       Profile         `yaml:"profile,omitempty"`
	Roles         map[string]Role `yaml:"roles,omitempty"`
	Orchestration Orchestration   `yaml:"orchestration,omitempty"`
	Arrangement   Arrangement     `yaml:"arrangement,omitempty"`
	Events        []Event         `yaml:"events,omitempty"`
	// RoleEvents (SP14) is a per-section explicit event list keyed by role
	// name. When set, completely replaces both the role's default Events and
	// the algorithm's pattern logic for that role in this section.
	// Resolution precedence: section.RoleEvents[role] > role.Events > algorithm fallback.
	RoleEvents map[string][]NoteEvent `yaml:"role_events,omitempty"`

	// LoopBars (SP15) is an optional explicit loop length in bars (4/4 = 4 beats per bar).
	// When set, event lists for this section are repeated every LoopBars bars across
	// the section's full duration. 0 = auto-detect from the max event beat.
	// Applies to both role-level and section-level RoleEvents.
	LoopBars int `yaml:"loop_bars,omitempty"`

	// RoleLoopBars (SP15) is an optional per-role override of LoopBars, keyed by
	// role name. Use this when one role's pattern is a different loop length than
	// the rest of the section (e.g. a 4-bar bass line over 2-bar drums).
	RoleLoopBars map[string]int `yaml:"role_loop_bars,omitempty"`
}

// NoteEvent is one explicit note in a role's event list (SP14).
// All fields are optional except Beat. When present, the role's per-section
// pattern/motif/algorithm logic is bypassed for that role; the engine plays
// these events verbatim.
type NoteEvent struct {
	// Beat is the position within the section, in beats. 1.0 = first beat.
	// Sub-beat resolution: 1.5 = "1 and", 1.25 = first 16th of beat 1, etc.
	Beat float64 `yaml:"beat"`

	// Pitch can be:
	//   - a MIDI note name like "C4", "F#3", "Bb5"
	//   - a scale degree relative to the section's key: "1", "b3", "#5", "7"
	//     (octave shifts: ">" raises an octave, "<" lowers an octave)
	//   - a chord-relative degree: "R" (root), "3", "5", "7", "9", "11", "13"
	//   - empty string for drums (the role's family determines the hit kind)
	Pitch string `yaml:"pitch"`

	// Dur is duration in beats. 1.0 = quarter note in 4/4. Defaults to 0.5
	// when omitted (an eighth-note-ish duration; the engine gates this to
	// leave breath between consecutive notes).
	Dur float64 `yaml:"dur"`

	// Vel is MIDI velocity 1..127. Defaults to 80.
	Vel int `yaml:"vel"`

	// Art is the articulation hint:
	//   - "ghost"    — very low velocity, abbreviated duration
	//   - "accent"   — bump velocity by +15
	//   - "staccato" — gate duration to 25%
	//   - "tenuto"   — gate duration to 100% (full hold)
	//   - "legato"   — overlap into next note slightly
	// Empty = normal.
	Art string `yaml:"art"`
}

// AutomationLane describes a per-section breakpoint curve for a named
// parameter (SP7). Param is one of: "cutoff", "pan", "expression".
type AutomationLane struct {
	Param       string `yaml:"param"`
	Breakpoints []Bkpt `yaml:"breakpoints"`
}

// Bkpt is a single breakpoint in an AutomationLane.
// AtPercent is 0..100 of the section duration; Value is the parameter value.
type Bkpt struct {
	AtPercent float64 `yaml:"at"`
	Value     float64 `yaml:"value"`
}

// SubstitutionRule specifies a harmonic substitution directive (SP7).
// The renderer does not apply these at runtime; they are consumed by the
// ApplySubstitutions compile-time rewriter in internal/gen.
type SubstitutionRule struct {
	// Rule is one of: tritone_sub, ii_V_chain, secondary_dominant, deceptive.
	Rule        string  `yaml:"rule"`
	// ApplyTo constrains which chord role triggers the rule (e.g. "V", "I").
	ApplyTo     string  `yaml:"apply_to"`
	// Before is an optional anchor chord for ii_V_chain insertion.
	Before      string  `yaml:"before"`
	// Of is the target chord for secondary_dominant (e.g. "ii").
	Of          string  `yaml:"of"`
	// Probability is 0..1; when < 1 the rule is applied probabilistically.
	Probability float64 `yaml:"probability"`
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
	// Harmony was the v1 per-role harmony string. Removed in SP8; it was never
	// consumed by the render pipeline. Use Section.Harmony or Section.HarmonyChords.
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
	// Notes is an optional weighted-random note pool (SP7).
	// When present the role's generator draws from these scale degrees
	// according to their relative weights.
	Notes         *NotePool `yaml:"notes,omitempty"`
	// Events (SP14) is the role's default explicit event list. When non-empty
	// it overrides the role's pattern/motif logic for any section that does
	// not have its own per-section RoleEvents entry for this role. The engine
	// repeats the event sequence across the section.
	Events        []NoteEvent `yaml:"events,omitempty"`
	// LoopBars (SP15) is the role's explicit loop length in bars (4 beats each).
	// When set, the Events list is repeated every LoopBars bars across the
	// section. 0 = auto-detect from the max event beat (rounded up to nearest bar).
	// Section.LoopBars (or Section.RoleLoopBars[name]) overrides this.
	LoopBars int `yaml:"loop_bars,omitempty"`
}

type PhraseBlock struct {
	Pattern string `yaml:"pattern,omitempty"`
	Motif   string `yaml:"motif,omitempty"`
	// Harmony was the v1 per-phrase harmony override. Removed in SP8; it was
	// dead weight — never consumed by the render pipeline.
	Active *bool `yaml:"active,omitempty"`
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

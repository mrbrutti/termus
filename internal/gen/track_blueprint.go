package gen

import "strings"

// RoleBlueprint describes one musical role in an authored track section.
// The language stays generic; style-specific generators decide how to render
// these requests into their own channel/program layouts.
type RoleBlueprint struct {
	Name         string
	Family       string
	Tone         []string
	Articulation string
	Register     string
	Prominence   string
	Pattern      string
	Motif        string
	Harmony      string
	Active       bool
}

// TrackBlueprint is the generic section-level IR compiled from a .tm track.
// It is intentionally style-neutral: one schema can drive lofi, jazz,
// ambient, or future styles by changing the renderer, not the language.
type TrackBlueprint struct {
	Style     string
	Section   string
	Tempo     string
	Key       string
	Harmony   string
	Scene     string
	Variation string
	Roles     map[string]RoleBlueprint
}

func (b TrackBlueprint) Empty() bool {
	if strings.TrimSpace(b.Style) != "" ||
		strings.TrimSpace(b.Section) != "" ||
		strings.TrimSpace(b.Tempo) != "" ||
		strings.TrimSpace(b.Key) != "" ||
		strings.TrimSpace(b.Harmony) != "" ||
		strings.TrimSpace(b.Scene) != "" ||
		strings.TrimSpace(b.Variation) != "" {
		return false
	}
	for _, role := range b.Roles {
		if strings.TrimSpace(role.Family) != "" ||
			len(role.Tone) > 0 ||
			strings.TrimSpace(role.Articulation) != "" ||
			strings.TrimSpace(role.Register) != "" ||
			strings.TrimSpace(role.Prominence) != "" ||
			strings.TrimSpace(role.Pattern) != "" ||
			strings.TrimSpace(role.Motif) != "" ||
			strings.TrimSpace(role.Harmony) != "" ||
			role.Active {
			return false
		}
	}
	return true
}

type TrackBlueprintAware interface {
	ApplyTrackBlueprint(TrackBlueprint)
}

func ApplyTrackBlueprint(algo Algorithm, blueprint TrackBlueprint) Algorithm {
	if blueprint.Empty() {
		return algo
	}
	if aware, ok := algo.(TrackBlueprintAware); ok {
		aware.ApplyTrackBlueprint(blueprint)
	}
	return algo
}

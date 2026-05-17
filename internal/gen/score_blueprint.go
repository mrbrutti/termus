package gen

import "strings"

// ScoreBlueprint is a compact authored-section description compiled from .tm
// files. Algorithms can opt into it to make scored pieces influence actual
// harmony, phrasing, groove, and arrangement instead of only seed/profile.
type ScoreBlueprint struct {
	Form      string
	Harmony   string
	Lead      string
	Comp      string
	Drums     string
	Arrange   string
	Variation string
}

func (b ScoreBlueprint) Empty() bool {
	return strings.TrimSpace(b.Form) == "" &&
		strings.TrimSpace(b.Harmony) == "" &&
		strings.TrimSpace(b.Lead) == "" &&
		strings.TrimSpace(b.Comp) == "" &&
		strings.TrimSpace(b.Drums) == "" &&
		strings.TrimSpace(b.Arrange) == "" &&
		strings.TrimSpace(b.Variation) == ""
}

type ScoreBlueprintAware interface {
	ApplyScoreBlueprint(ScoreBlueprint)
}

func ApplyScoreBlueprint(algo Algorithm, blueprint ScoreBlueprint) Algorithm {
	if blueprint.Empty() {
		return algo
	}
	if aware, ok := algo.(ScoreBlueprintAware); ok {
		aware.ApplyScoreBlueprint(blueprint)
	}
	return algo
}

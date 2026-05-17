package tm

import (
	"fmt"
	"strings"
	"time"

	"github.com/mrbrutti/termus/internal/gen"
)

func Compile(score *Score, defaultSeed int64, defaultListenMode gen.ListeningMode) (*Compiled, error) {
	if score == nil {
		return nil, fmt.Errorf("score is nil")
	}
	if strings.TrimSpace(score.Title) == "" {
		return nil, fmt.Errorf("title is required")
	}
	if len(score.Sections) == 0 {
		return nil, fmt.Errorf("at least one section is required")
	}
	listenMode := defaultListenMode
	if listenMode == "" {
		listenMode = gen.ListeningModeEndless
	}
	if score.ListenMode != "" {
		mode, ok := gen.ResolveListeningMode(score.ListenMode)
		if !ok {
			return nil, fmt.Errorf("unknown listen mode %q", score.ListenMode)
		}
		listenMode = mode.Name
	}
	baseSeed := defaultSeed
	if score.Seed != 0 {
		baseSeed = score.Seed
	}
	compiled := &Compiled{
		Playlist: gen.Playlist{
			Name:       score.Title,
			Mode:       gen.PlaylistScore,
			ListenMode: listenMode,
			Tracks:     make([]gen.Track, 0, len(score.Sections)),
		},
		Overrides:  make(map[string]gen.ControlProfile, len(score.Sections)),
		Blueprints: make(map[string]gen.ScoreBlueprint, len(score.Sections)),
	}
	globalProfile, err := score.Globals.resolve(gen.DefaultControlProfile())
	if err != nil {
		return nil, fmt.Errorf("globals: %w", err)
	}
	for i, section := range score.Sections {
		spec, ok := gen.Resolve(section.Algo)
		if !ok {
			return nil, fmt.Errorf("sections[%d].algo: unknown algorithm %q", i, section.Algo)
		}
		dur, err := time.ParseDuration(section.Duration)
		if err != nil || dur <= 0 {
			return nil, fmt.Errorf("sections[%d].duration: invalid duration %q", i, section.Duration)
		}
		if err := validateSectionAudit(section); err != nil {
			return nil, fmt.Errorf("sections[%d]: %w", i, err)
		}
		seed := baseSeed + int64(i)*1009
		if section.Seed != nil {
			seed = *section.Seed + int64(i)*1009
		} else if section.SeedOffset != nil {
			seed = baseSeed + *section.SeedOffset + int64(i)*1009
		}
		profile, err := section.Profile.resolve(globalProfile)
		if err != nil {
			return nil, fmt.Errorf("sections[%d].profile: %w", i, err)
		}
		title := section.Title
		if strings.TrimSpace(title) == "" {
			title = spec.Label()
		}
		compiled.Playlist.Tracks = append(compiled.Playlist.Tracks, gen.Track{
			Spec:     spec,
			Seed:     seed,
			Duration: dur,
			Title:    title,
		})
		key := overrideKey(spec, seed)
		compiled.Overrides[key] = profile
		compiled.Blueprints[key] = gen.ScoreBlueprint{
			Form:      section.Audit.Form,
			Harmony:   section.Audit.Harmony,
			Lead:      section.Audit.Lead,
			Comp:      section.Audit.Comp,
			Drums:     section.Audit.Drums,
			Arrange:   section.Audit.Arrange,
			Variation: section.Audit.Variation,
		}
	}
	compiled.Warnings = lintScore(score, compiled.Playlist.Tracks)
	return compiled, nil
}

func (p TMProfile) resolve(base gen.ControlProfile) (gen.ControlProfile, error) {
	out := base
	fields := []struct {
		in  MacroValue
		set func(int)
	}{
		{p.Density, func(v int) { out.Density = v }},
		{p.Brightness, func(v int) { out.Brightness = v }},
		{p.Motion, func(v int) { out.Motion = v }},
		{p.Reverb, func(v int) { out.Reverb = v }},
		{p.Swing, func(v int) { out.Swing = v }},
		{p.DroneDepth, func(v int) { out.DroneDepth = v }},
		{p.Tempo, func(v int) { out.Tempo = v }},
		{p.Phrase, func(v int) { out.Phrase = v }},
	}
	for _, field := range fields {
		v, ok, err := field.in.Resolve()
		if err != nil {
			return gen.ControlProfile{}, err
		}
		if ok {
			field.set(v)
		}
	}
	return out, nil
}

func validateSectionAudit(section Section) error {
	if err := validatePattern(section.Audit.Harmony, "harmony"); err != nil {
		return fmt.Errorf("audit.harmony: %w", err)
	}
	if err := validatePattern(section.Audit.Lead, "melody"); err != nil {
		return fmt.Errorf("audit.lead: %w", err)
	}
	if err := validatePattern(section.Audit.Comp, "rhythm"); err != nil {
		return fmt.Errorf("audit.comp: %w", err)
	}
	if err := validatePattern(section.Audit.Drums, "rhythm"); err != nil {
		return fmt.Errorf("audit.drums: %w", err)
	}
	if err := validatePattern(section.Audit.Arrange, "arrange"); err != nil {
		return fmt.Errorf("audit.arrange: %w", err)
	}
	return nil
}

func lintScore(score *Score, tracks []gen.Track) []Warning {
	var warnings []Warning
	if len(tracks) < 2 {
		warnings = append(warnings, Warning{Path: "sections", Message: "single-section score may still feel static; add contrast sections"})
		return warnings
	}
	sameAlgo := true
	sameTitle := true
	firstAlgo := tracks[0].Spec.Name
	firstTitle := tracks[0].Title
	hasLead := false
	hasDrums := false
	for i, section := range score.Sections {
		if section.Audit.Lead != "" {
			hasLead = true
		}
		if section.Audit.Drums != "" {
			hasDrums = true
		}
		if tracks[i].Spec.Name != firstAlgo {
			sameAlgo = false
		}
		if tracks[i].Title != firstTitle {
			sameTitle = false
		}
	}
	if sameAlgo && sameTitle {
		warnings = append(warnings, Warning{Path: "sections", Message: "all sections share the same algorithm and title; consider stronger sectional contrast"})
	}
	if !hasLead {
		warnings = append(warnings, Warning{Path: "audit.lead", Message: "no lead motif audit found; melodic identity may be weak"})
	}
	if !hasDrums {
		warnings = append(warnings, Warning{Path: "audit.drums", Message: "no drum audit found; rhythmic genres may feel underspecified"})
	}
	return warnings
}

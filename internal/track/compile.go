package track

import (
	"fmt"
	"strings"
	"time"

	"github.com/mrbrutti/termus/internal/gen"
)

func Compile(file *File, defaultSeed int64, defaultListenMode gen.ListeningMode) (*Compiled, error) {
	if file == nil {
		return nil, fmt.Errorf("track is nil")
	}
	if strings.TrimSpace(file.Title) == "" {
		return nil, fmt.Errorf("title is required")
	}
	if strings.TrimSpace(file.Style) == "" {
		return nil, fmt.Errorf("style is required")
	}
	if len(file.Sections) == 0 {
		return nil, fmt.Errorf("at least one section is required")
	}
	sections, err := resolveSections(file)
	if err != nil {
		return nil, err
	}
	spec, ok := gen.Resolve(file.Style)
	if !ok {
		return nil, fmt.Errorf("unknown style %q", file.Style)
	}
	listenMode := defaultListenMode
	if listenMode == "" {
		listenMode = gen.ListeningModeEndless
	}
	if file.ListenMode != "" {
		mode, ok := gen.ResolveListeningMode(file.ListenMode)
		if !ok {
			return nil, fmt.Errorf("unknown listen mode %q", file.ListenMode)
		}
		listenMode = mode.Name
	}
	baseSeed := defaultSeed
	if file.Seed != 0 {
		baseSeed = file.Seed
	}
	globalProfile, err := file.Globals.resolve(gen.DefaultControlProfile())
	if err != nil {
		return nil, fmt.Errorf("globals: %w", err)
	}
	compiled := &Compiled{
		Playlist: gen.Playlist{
			Name:       file.Title,
			Mode:       gen.PlaylistScore,
			ListenMode: listenMode,
			Tracks:     make([]gen.Track, 0, len(sections)),
		},
		Profiles: make(map[string]gen.ControlProfile, len(sections)),
		Plans:    make(map[string]gen.AuthoredTrackPlan, len(sections)),
	}
	for name, role := range file.Roles {
		if err := validateRole(name, role); err != nil {
			return nil, err
		}
	}
	for i, section := range sections {
		dur, err := time.ParseDuration(section.Duration)
		if err != nil || dur <= 0 {
			return nil, fmt.Errorf("sections[%d].duration: invalid duration %q", i, section.Duration)
		}
		if err := validatePattern(section.Harmony, "harmony"); err != nil {
			return nil, fmt.Errorf("sections[%d].harmony: %w", i, err)
		}
		if err := validatePattern(section.Scene, "scene"); err != nil {
			return nil, fmt.Errorf("sections[%d].scene: %w", i, err)
		}
		mergedRoles := mergeRoles(file.Roles, section.Roles)
		mergedRoles = applyRoleTransforms(mergedRoles, section.Transforms)
		for name, role := range mergedRoles {
			if err := validateRole(name, role); err != nil {
				return nil, fmt.Errorf("sections[%d]: %w", i, err)
			}
		}
		for eventIdx, event := range sectionEvents(section) {
			if err := validateEvent(i, eventIdx, event); err != nil {
				return nil, err
			}
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
		title := strings.TrimSpace(section.Title)
		if title == "" {
			title = strings.TrimSpace(section.ID)
		}
		if title == "" {
			title = spec.Label()
		}
		compiled.Playlist.Tracks = append(compiled.Playlist.Tracks, gen.Track{
			Spec:     spec,
			Seed:     seed,
			Duration: dur,
			Title:    title,
		})
		key := playlistKey(spec, seed)
		compiled.Profiles[key] = profile
		plan, err := buildAuthoredPlan(spec, file, section, mergedRoles, dur, profile, seed)
		if err != nil {
			return nil, fmt.Errorf("sections[%d].plan: %w", i, err)
		}
		compiled.Plans[key] = plan
	}
	resolvedFile := *file
	resolvedFile.Sections = sections
	compiled.Warnings = lintFile(&resolvedFile, compiled.Playlist.Tracks)
	return compiled, nil
}

func mergeRoles(base map[string]Role, override map[string]Role) map[string]Role {
	out := make(map[string]Role, len(base)+len(override))
	for name, role := range base {
		out[name] = role
	}
	for name, role := range override {
		current := out[name]
		if strings.TrimSpace(role.Family) != "" {
			current.Family = role.Family
		}
		if len(role.Tone) > 0 {
			current.Tone = append([]string(nil), role.Tone...)
		}
		if strings.TrimSpace(role.Articulation) != "" {
			current.Articulation = role.Articulation
		}
		if strings.TrimSpace(role.Register) != "" {
			current.Register = role.Register
		}
		if strings.TrimSpace(role.Prominence) != "" {
			current.Prominence = role.Prominence
		}
		if strings.TrimSpace(role.Pattern) != "" {
			current.Pattern = role.Pattern
		}
		if strings.TrimSpace(role.Motif) != "" {
			current.Motif = role.Motif
		}
		if strings.TrimSpace(role.Harmony) != "" {
			current.Harmony = role.Harmony
		}
		if len(role.Phrases) > 0 {
			if current.Phrases == nil {
				current.Phrases = map[string]PhraseBlock{}
			}
			for phrase, block := range role.Phrases {
				current.Phrases[phrase] = block
			}
		}
		if role.Active != nil {
			current.Active = role.Active
		}
		out[name] = current
	}
	return out
}

func (p Profile) resolve(base gen.ControlProfile) (gen.ControlProfile, error) {
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

func lintFile(file *File, tracks []gen.Track) []Warning {
	var warnings []Warning
	if len(file.Roles) == 0 {
		warnings = append(warnings, Warning{Path: "roles", Message: "no top-level roles defined; authored arrangement may stay thin"})
	}
	if len(tracks) < 2 {
		warnings = append(warnings, Warning{Path: "sections", Message: "single-section track may still feel static; add contrast sections"})
		return warnings
	}
	uniqueHarmony := map[string]bool{}
	sceneCount := 0
	for _, section := range file.Sections {
		if strings.TrimSpace(section.Harmony) != "" {
			uniqueHarmony[strings.TrimSpace(section.Harmony)] = true
		}
		if strings.TrimSpace(section.Scene) != "" {
			sceneCount++
		}
	}
	if len(uniqueHarmony) < 2 {
		warnings = append(warnings, Warning{Path: "sections.harmony", Message: "all sections share similar harmony; consider stronger sectional contrast"})
	}
	if sceneCount == 0 {
		warnings = append(warnings, Warning{Path: "sections.scene", Message: "no section scenes defined; role contrast may be weak"})
	}
	return warnings
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

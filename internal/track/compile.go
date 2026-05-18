package track

import (
	"fmt"
	"sort"
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
	// SP18: if Form is set and the authored file has no explicit sections,
	// expand the form template into a default section list. Explicit sections
	// always win.
	if len(file.Sections) == 0 && strings.TrimSpace(file.Form) != "" {
		template, ok := ResolveForm(file.Form)
		if !ok {
			return nil, fmt.Errorf("unknown form %q", file.Form)
		}
		bpm := resolveBPMHint(file.Tempo, template.DefaultBPM)
		file.Sections = expandFormTemplate(template, bpm)
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
	pack := resolveStylePack(file.Style, file.Substyle, file.Title, file.Tags)
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
			Tracks:     make([]gen.Track, 0, 1),
		},
		Profiles: make(map[string]gen.ControlProfile, len(sections)),
		Plans:    make(map[string]gen.AuthoredTrackPlan, len(sections)),
	}
	for name, role := range file.Roles {
		if err := validateRole(name, role); err != nil {
			return nil, err
		}
	}
	// SP17: build a single seamless Track holding all sections as an internal
	// schedule. Each section still gets its own plan keyed by spec+seed; the
	// playback engine swaps plans at section boundaries without crossfading.
	sectionStops := make([]gen.SectionStop, 0, len(sections))
	var cursor time.Duration
	var firstSectionSeed int64
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
		mergedRoles = applyOrchestration(mergedRoles, section.Orchestration)
		section, mergedRoles = applyStyleLibrary(pack, section, mergedRoles)
		for name, role := range mergedRoles {
			if err := validateRole(name, role); err != nil {
				return nil, fmt.Errorf("sections[%d]: %w", i, err)
			}
		}
		for name, directive := range section.Orchestration.Roles {
			if err := validateOrchestrationRole(name, directive); err != nil {
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
		key := playlistKey(spec, seed)
		compiled.Profiles[key] = profile
		plan, err := buildAuthoredPlan(spec, file, section, mergedRoles, dur, profile, seed)
		if err != nil {
			return nil, fmt.Errorf("sections[%d].plan: %w", i, err)
		}
		compiled.Plans[key] = plan
		sectionStops = append(sectionStops, gen.SectionStop{
			Title:     title,
			Seed:      seed,
			Duration:  dur,
			StartTime: cursor,
			PlanKey:   key,
		})
		if i == 0 {
			firstSectionSeed = seed
		}
		cursor += dur
	}
	// SP17: assemble the single Track. The Track's outer Seed/Title come from
	// the first section so legacy code that reads Track.Spec / Track.Seed (e.g.
	// the build closures that key plan lookups) still sees a valid lookup.
	loopEvolving := listenMode == gen.ListeningModeHourStream || listenMode == gen.ListeningModeAlbumSide
	trackTitle := strings.TrimSpace(file.Title)
	if trackTitle == "" {
		trackTitle = spec.Label()
	}
	compiled.Playlist.Tracks = append(compiled.Playlist.Tracks, gen.Track{
		Spec:                spec,
		Seed:                firstSectionSeed,
		Duration:            cursor,
		Title:               trackTitle,
		Sections:            sectionStops,
		LoopForeverEvolving: loopEvolving,
	})
	resolvedFile := *file
	resolvedFile.Sections = sections
	compiled.Warnings = lintFile(&resolvedFile, compiled.Playlist.Tracks, len(sections))
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
		// Role.Harmony removed in SP8 (v1 dead field).
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

// lintFile inspects the authored file plus the compiled track schedule and
// returns warnings. sectionCount is the number of resolved sections in the
// compiled output (post-SP17 a single Track may carry many sections, so the
// legacy "len(tracks)" check was no longer meaningful).
func lintFile(file *File, tracks []gen.Track, sectionCount int) []Warning {
	var warnings []Warning
	if len(file.Roles) == 0 {
		warnings = append(warnings, Warning{Path: "roles", Message: "no top-level roles defined; authored arrangement may stay thin"})
	}
	if sectionCount < 2 {
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
	cadenceFound := false
	for idx, section := range file.Sections {
		roles := resolvedSectionRoles(file, section)
		if sectionRoleDensity(roles) > 80 {
			warnings = append(warnings, Warning{Path: fmt.Sprintf("sections[%d].roles", idx), Message: "section writing is dense across too many active roles; consider thinning or alternating ownership"})
		}
		if brightAttackCount(roles) > 6 {
			warnings = append(warnings, Warning{Path: fmt.Sprintf("sections[%d].roles", idx), Message: "too many simultaneous bright attack roles; soften or stagger the ensemble"})
		}
		if strings.Contains(strings.ToLower(section.Scene+" "+section.Variation), "cadence") || strings.Contains(strings.ToLower(section.Scene+" "+section.Variation), "outro") {
			cadenceFound = true
		}
		for _, event := range sectionEvents(section) {
			switch strings.ToLower(strings.TrimSpace(event.Kind)) {
			case "ending", "tag", "stop":
				cadenceFound = true
			}
		}
	}
	if !cadenceFound {
		warnings = append(warnings, Warning{Path: "sections", Message: "track has no clear cadence or ending shape; add a cadence/outro scene or ending event"})
	}
	for i := 1; i < len(file.Sections); i++ {
		if sectionSimilarity(file, file.Sections[i-1], file.Sections[i]) >= 0.80 {
			warnings = append(warnings, Warning{Path: fmt.Sprintf("sections[%d]", i), Message: "section is too similar to its neighbor; change harmony, phrase blocks, orchestration, or arrangement"})
		}
	}
	budget := file.VariationBudget
	if budget.MaxHarmonyRepeat > 0 {
		for harmony, count := range countSectionField(file.Sections, func(section Section) string {
			return strings.TrimSpace(section.Harmony)
		}) {
			if harmony != "" && count > budget.MaxHarmonyRepeat {
				warnings = append(warnings, Warning{Path: "variation_budget.max_harmony_repeat", Message: fmt.Sprintf("harmony %q repeats %d times (budget %d)", harmony, count, budget.MaxHarmonyRepeat)})
			}
		}
	}
	if budget.MaxSceneRepeat > 0 {
		for scene, count := range countSectionField(file.Sections, func(section Section) string {
			return strings.TrimSpace(section.Scene)
		}) {
			if scene != "" && count > budget.MaxSceneRepeat {
				warnings = append(warnings, Warning{Path: "variation_budget.max_scene_repeat", Message: fmt.Sprintf("scene %q repeats %d times (budget %d)", scene, count, budget.MaxSceneRepeat)})
			}
		}
	}
	if budget.MaxMotifRepeat > 0 {
		for motif, count := range countSectionField(file.Sections, func(section Section) string {
			merged := applyOrchestration(applyRoleTransforms(mergeRoles(file.Roles, section.Roles), section.Transforms), section.Orchestration)
			values := make([]string, 0, len(merged))
			for name, role := range merged {
				if authoredRoleKind(name, role) != "melody" {
					continue
				}
				values = append(values, strings.TrimSpace(roleValue(role.Motif, role.Pattern)))
			}
			sort.Strings(values)
			return strings.Join(values, " || ")
		}) {
			if motif != "" && count > budget.MaxMotifRepeat {
				warnings = append(warnings, Warning{Path: "variation_budget.max_motif_repeat", Message: fmt.Sprintf("melodic phrase pack %q repeats %d times (budget %d)", motif, count, budget.MaxMotifRepeat)})
			}
		}
	}
	if budget.RequireReturnTransform {
		for idx, section := range file.Sections {
			if strings.TrimSpace(section.Derive) == "" {
				continue
			}
			if len(section.Transforms) == 0 {
				warnings = append(warnings, Warning{Path: fmt.Sprintf("sections[%d].transforms", idx), Message: "derived section has no transforms; return may sound like a copy"})
			}
		}
	}
	return warnings
}

func countSectionField(sections []Section, keyFn func(Section) string) map[string]int {
	counts := map[string]int{}
	for _, section := range sections {
		key := keyFn(section)
		if key == "" {
			continue
		}
		counts[key]++
	}
	return counts
}

func resolvedSectionRoles(file *File, section Section) map[string]Role {
	return applyOrchestration(applyRoleTransforms(mergeRoles(file.Roles, section.Roles), section.Transforms), section.Orchestration)
}

func sectionRoleDensity(roles map[string]Role) int {
	total := 0
	for name, role := range roles {
		if role.Active != nil && !*role.Active {
			continue
		}
		pattern := strings.TrimSpace(roleValue(role.Pattern, role.Motif))
		if pattern == "" {
			continue
		}
		total += rolePatternWeight(name, pattern)
	}
	return total
}

func rolePatternWeight(name, pattern string) int {
	switch authoredRoleKind(name, Role{}) {
	case "drum":
		return strings.Count(pattern, "x") / 2
	case "pad":
		return maxInt(1, strings.Count(pattern, "x"))
	default:
		return maxInt(1, strings.Count(pattern, "x"))
	}
}

func brightAttackCount(roles map[string]Role) int {
	count := 0
	for name, role := range roles {
		if role.Active != nil && !*role.Active {
			continue
		}
		if !roleHasBrightAttack(name, role) {
			continue
		}
		count++
	}
	return count
}

func roleHasBrightAttack(name string, role Role) bool {
	family := strings.ToLower(strings.TrimSpace(role.Family))
	if family == "drums" || family == "bass" || family == "synth_bass" {
		return false
	}
	articulation := strings.ToLower(strings.TrimSpace(role.Articulation))
	prominence := strings.ToLower(strings.TrimSpace(role.Prominence))
	if strings.Contains(articulation, "sustain") && !strings.Contains(prominence, "lead") {
		return false
	}
	if strings.Contains(articulation, "bloom") || strings.Contains(articulation, "echo") || strings.Contains(articulation, "pulse") || strings.Contains(articulation, "stab") {
		return true
	}
	for _, tone := range role.Tone {
		lower := strings.ToLower(strings.TrimSpace(tone))
		switch lower {
		case "glass", "sparkle", "bright", "metallic", "luminous", "icy":
			return true
		}
	}
	switch family {
	case "bells", "music_box", "mallet", "brass":
		return true
	}
	return false
}

func sectionSimilarity(file *File, a, b Section) float64 {
	score := 0.0
	total := 4.0
	if strings.TrimSpace(a.Harmony) == strings.TrimSpace(b.Harmony) {
		score++
	}
	if strings.TrimSpace(a.Scene) == strings.TrimSpace(b.Scene) {
		score++
	}
	if strings.TrimSpace(a.Variation) == strings.TrimSpace(b.Variation) {
		score++
	}
	if roleSignature(resolvedSectionRoles(file, a)) == roleSignature(resolvedSectionRoles(file, b)) {
		score++
	}
	return score / total
}

func roleSignature(roles map[string]Role) string {
	if len(roles) == 0 {
		return ""
	}
	parts := make([]string, 0, len(roles))
	for name, role := range roles {
		active := role.Active == nil || *role.Active
		if !active {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s:%s:%s:%s", name, role.Family, roleValue(role.Pattern, role.Motif), role.Register))
	}
	sort.Strings(parts)
	return strings.Join(parts, " | ")
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

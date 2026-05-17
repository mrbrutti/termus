package track

import (
	"fmt"
	"strconv"
	"strings"
)

func resolveSections(file *File) ([]Section, error) {
	if file == nil {
		return nil, fmt.Errorf("track is nil")
	}
	resolved := make([]Section, 0, len(file.Sections))
	index := map[string]Section{}
	for i, section := range file.Sections {
		current := section
		if key := strings.TrimSpace(section.Derive); key != "" {
			base, ok := index[strings.ToLower(key)]
			if !ok {
				return nil, fmt.Errorf("sections[%d].derive: unknown base section %q", i, section.Derive)
			}
			current = mergeSection(base, section)
			if err := applySectionTransforms(&current, file.Key); err != nil {
				return nil, fmt.Errorf("sections[%d].transforms: %w", i, err)
			}
		}
		resolved = append(resolved, current)
		for _, key := range sectionLookupKeys(current) {
			index[key] = current
		}
	}
	return resolved, nil
}

func mergeSection(base, override Section) Section {
	out := base
	if strings.TrimSpace(override.ID) != "" {
		out.ID = override.ID
	}
	if strings.TrimSpace(override.Title) != "" {
		out.Title = override.Title
	}
	out.Derive = override.Derive
	if len(override.Transforms) > 0 {
		out.Transforms = append([]string(nil), override.Transforms...)
	}
	if strings.TrimSpace(override.Duration) != "" {
		out.Duration = override.Duration
	}
	if override.Seed != nil {
		out.Seed = override.Seed
	}
	if override.SeedOffset != nil {
		out.SeedOffset = override.SeedOffset
	}
	if strings.TrimSpace(override.Key) != "" {
		out.Key = override.Key
	}
	if strings.TrimSpace(override.Tempo) != "" {
		out.Tempo = override.Tempo
	}
	if strings.TrimSpace(override.Harmony) != "" {
		out.Harmony = override.Harmony
	}
	if strings.TrimSpace(override.Scene) != "" {
		out.Scene = override.Scene
	}
	if strings.TrimSpace(override.Variation) != "" {
		out.Variation = override.Variation
	}
	out.Profile = mergeProfile(base.Profile, override.Profile)
	out.Roles = mergeRoles(base.Roles, override.Roles)
	if len(override.Events) > 0 {
		out.Events = append(append([]Event(nil), base.Events...), override.Events...)
	}
	return out
}

func mergeProfile(base, override Profile) Profile {
	out := base
	if override.Density.set {
		out.Density = override.Density
	}
	if override.Brightness.set {
		out.Brightness = override.Brightness
	}
	if override.Motion.set {
		out.Motion = override.Motion
	}
	if override.Reverb.set {
		out.Reverb = override.Reverb
	}
	if override.Swing.set {
		out.Swing = override.Swing
	}
	if override.DroneDepth.set {
		out.DroneDepth = override.DroneDepth
	}
	if override.Tempo.set {
		out.Tempo = override.Tempo
	}
	if override.Phrase.set {
		out.Phrase = override.Phrase
	}
	return out
}

func applySectionTransforms(section *Section, fallbackKey string) error {
	if section == nil {
		return nil
	}
	for _, raw := range section.Transforms {
		transform := strings.ToLower(strings.TrimSpace(raw))
		switch transform {
		case "", "none":
			continue
		case "sequence":
			section.Scene = appendDescriptorToken(section.Scene, "sequence-up")
			section.Variation = appendDescriptorToken(section.Variation, "sequence-up")
		case "invert":
			section.Scene = appendDescriptorToken(section.Scene, "mirror")
			section.Variation = appendDescriptorToken(section.Variation, "invert")
		case "thin":
			section.Scene = appendDescriptorToken(section.Scene, "thin")
			section.Variation = appendDescriptorToken(section.Variation, "thin")
			section.Profile.Density = newMacroValue("sparse")
		case "lift-register":
			section.Scene = appendDescriptorToken(section.Scene, "lift-register")
			section.Variation = appendDescriptorToken(section.Variation, "lift-register")
		case "cadence-rewrite":
			section.Scene = appendDescriptorToken(section.Scene, "cadence")
			section.Variation = appendDescriptorToken(section.Variation, "cadence")
			section.Harmony = rewriteCadenceHarmony(section.Harmony, firstNonBlank(section.Key, fallbackKey))
		default:
			return fmt.Errorf("unsupported transform %q", raw)
		}
	}
	return nil
}

func applyRoleTransforms(roles map[string]Role, transforms []string) map[string]Role {
	if len(roles) == 0 || len(transforms) == 0 {
		return roles
	}
	out := cloneRoles(roles)
	for _, raw := range transforms {
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "", "none":
			continue
		case "sequence":
			out = transformRoleMotifs(out, func(pattern string) string {
				return shiftMelodyPattern(pattern, 1, false)
			})
		case "invert":
			out = transformRoleMotifs(out, invertMelodyPattern)
		case "lift-register":
			out = liftRoleRegisters(out)
		case "cadence-rewrite":
			out = transformRoleMotifs(out, rewriteCadenceMotif)
		}
	}
	return out
}

func cloneRoles(roles map[string]Role) map[string]Role {
	if len(roles) == 0 {
		return roles
	}
	out := make(map[string]Role, len(roles))
	for name, role := range roles {
		out[name] = role
	}
	return out
}

func transformRoleMotifs(roles map[string]Role, fn func(string) string) map[string]Role {
	if len(roles) == 0 || fn == nil {
		return roles
	}
	out := make(map[string]Role, len(roles))
	for name, role := range roles {
		if strings.TrimSpace(role.Motif) != "" {
			role.Motif = fn(role.Motif)
		}
		out[name] = role
	}
	return out
}

func liftRoleRegisters(roles map[string]Role) map[string]Role {
	if len(roles) == 0 {
		return roles
	}
	out := make(map[string]Role, len(roles))
	for name, role := range roles {
		lowerFamily := strings.ToLower(strings.TrimSpace(role.Family))
		if lowerFamily != "bass" && lowerFamily != "synth_bass" && lowerFamily != "drums" {
			role.Register = liftRegister(role.Register)
		}
		out[name] = role
	}
	return out
}

func liftRegister(register string) string {
	switch strings.TrimSpace(register) {
	case "sub":
		return "low"
	case "low":
		return "mid"
	case "mid":
		return "mid-high"
	case "mid-high":
		return "high"
	case "high":
		return "air"
	default:
		return register
	}
}

func shiftMelodyPattern(pattern string, degreeDelta int, octaveUp bool) string {
	parts := strings.Fields(strings.ReplaceAll(pattern, "|", " | "))
	for i, token := range parts {
		if token == "|" {
			continue
		}
		parts[i] = shiftMelodyToken(token, degreeDelta, octaveUp)
	}
	return strings.Join(parts, " ")
}

func invertMelodyPattern(pattern string) string {
	parts := strings.Fields(strings.ReplaceAll(pattern, "|", " | "))
	var pivot int
	found := false
	for _, token := range parts {
		if token == "|" || token == "." || token == "-" || token == "r" {
			continue
		}
		value, ok := melodyTokenValue(token)
		if ok {
			pivot = value
			found = true
			break
		}
	}
	if !found {
		return pattern
	}
	for i, token := range parts {
		if token == "|" || token == "." || token == "-" || token == "r" {
			continue
		}
		value, ok := melodyTokenValue(token)
		if !ok {
			continue
		}
		parts[i] = melodyTokenFromValue(pivot - (value - pivot))
	}
	return strings.Join(parts, " ")
}

func rewriteCadenceMotif(pattern string) string {
	parts := strings.Fields(strings.ReplaceAll(pattern, "|", " | "))
	line := []string{"3", ".", "2", ".", "1", ".", ".", "."}
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == "|" {
			continue
		}
		start := maxInt(0, i-len(line)+1)
		for j := 0; j < len(line) && start+j < len(parts); j++ {
			if parts[start+j] == "|" {
				continue
			}
			parts[start+j] = line[j]
		}
		break
	}
	return strings.Join(parts, " ")
}

func rewriteCadenceHarmony(harmony, key string) string {
	harmony = strings.TrimSpace(harmony)
	if harmony == "" {
		return harmony
	}
	parts := strings.Split(harmony, "|")
	if len(parts) == 0 {
		return harmony
	}
	tonic := tonicChordForKey(key)
	last := strings.Fields(strings.TrimSpace(parts[len(parts)-1]))
	switch len(last) {
	case 0:
		parts[len(parts)-1] = " " + tonic
	case 1:
		last[0] = tonic
		parts[len(parts)-1] = " " + strings.Join(last, " ")
	default:
		last[len(last)-1] = tonic
		parts[len(parts)-1] = " " + strings.Join(last, " ")
	}
	return strings.Join(parts, " | ")
}

func tonicChordForKey(key string) string {
	root, _, ok := parseRootToken(strings.TrimSpace(key))
	if !ok {
		return "Cmaj7"
	}
	name := pitchClassName(root)
	lower := strings.ToLower(strings.TrimSpace(key))
	switch {
	case strings.Contains(lower, "min"):
		return name + "m9"
	default:
		return name + "maj9"
	}
}

func pitchClassName(pc int) string {
	names := []string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}
	return names[wrapPitchClass(pc)]
}

func melodyTokenValue(token string) (int, bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		return 0, false
	}
	octave := 0
	if strings.HasPrefix(token, ">") {
		octave = 7
		token = strings.TrimPrefix(token, ">")
	} else if strings.HasPrefix(token, "^") {
		octave = 7
		token = strings.TrimPrefix(token, "^")
	}
	sign := 0
	if strings.HasPrefix(token, "b") {
		sign = -1
		token = strings.TrimPrefix(token, "b")
	} else if strings.HasPrefix(token, "#") {
		sign = 1
		token = strings.TrimPrefix(token, "#")
	}
	value, err := strconv.Atoi(token)
	if err != nil {
		return 0, false
	}
	return value + sign + octave, true
}

func melodyTokenFromValue(value int) string {
	prefix := ""
	for value > 13 {
		prefix += ">"
		value -= 7
	}
	for value < 1 {
		value += 7
	}
	return prefix + fmt.Sprintf("%d", value)
}

func sectionLookupKeys(section Section) []string {
	keys := []string{}
	if id := strings.ToLower(strings.TrimSpace(section.ID)); id != "" {
		keys = append(keys, id)
	}
	if title := strings.ToLower(strings.TrimSpace(section.Title)); title != "" {
		keys = append(keys, title)
	}
	return keys
}

func appendDescriptorToken(base, token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return strings.TrimSpace(base)
	}
	parts := strings.Fields(strings.TrimSpace(base))
	for _, part := range parts {
		if strings.EqualFold(part, token) {
			return strings.Join(parts, " ")
		}
	}
	parts = append(parts, token)
	return strings.Join(parts, " ")
}

func newMacroValue(raw string) MacroValue {
	return MacroValue{set: true, raw: strings.TrimSpace(raw)}
}

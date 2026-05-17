package gen

import (
	_ "embed"
	"encoding/json"
	"sort"
	"strings"
)

//go:embed sf2_inventory.json
var sf2InventoryJSON []byte

type sf2PresetProfile struct {
	Name          string              `json:"name"`
	Styles        []string            `json:"styles"`
	Families      []string            `json:"families"`
	Tones         []string            `json:"tones"`
	Realism       string              `json:"realism,omitempty"`
	Blend         []string            `json:"blend,omitempty"`
	Articulations []string            `json:"articulations,omitempty"`
	Programs      []sf2ProgramProfile `json:"programs,omitempty"`
}

type sf2ProgramProfile struct {
	Family        string   `json:"family"`
	Program       int32    `json:"program"`
	Roles         []string `json:"roles,omitempty"`
	Tones         []string `json:"tones,omitempty"`
	Articulations []string `json:"articulations,omitempty"`
	Registers     []string `json:"registers,omitempty"`
	Blend         []string `json:"blend,omitempty"`
	Realism       string   `json:"realism,omitempty"`
}

type SF2RoleIntent struct {
	Channel      int32
	Role         string
	Family       string
	Tone         []string
	Articulation string
	Register     string
	Prominence   string
	Active       bool
}

type SF2Selection struct {
	Primary  string
	Routes   map[int32]string
	Programs map[int32]int32
	Presets  []string
}

var sf2Inventory = loadSF2Inventory()

func loadSF2Inventory() map[string]sf2PresetProfile {
	var list []sf2PresetProfile
	_ = json.Unmarshal(sf2InventoryJSON, &list)
	out := make(map[string]sf2PresetProfile, len(list))
	for _, item := range list {
		out[item.Name] = item
	}
	return out
}

func InventoryPresetProfile(name string) (sf2PresetProfile, bool) {
	profile, ok := sf2Inventory[strings.TrimSpace(name)]
	return profile, ok
}

func InventoryProgramProfiles(name string) []sf2ProgramProfile {
	profile, ok := InventoryPresetProfile(name)
	if !ok {
		return nil
	}
	return append([]sf2ProgramProfile(nil), profile.Programs...)
}

func ResolveSF2Selection(spec AlgoSpec, strategy, fallback string) SF2Selection {
	if !spec.RequiresSF2 {
		return SF2Selection{}
	}
	if strategy == "single" {
		return SF2Selection{
			Primary:  fallback,
			Programs: defaultProgramsForIntents(roleIntentsForSpec(spec)),
			Presets:  dedupePresets([]string{fallback}),
		}
	}
	intents := roleIntentsForSpec(spec)
	if len(intents) == 0 {
		return SF2Selection{Primary: fallback, Presets: dedupePresets([]string{fallback})}
	}
	return resolveSelection(spec, intents, strategy, fallback)
}

func ResolveSF2SelectionForPlan(spec AlgoSpec, plan *AuthoredTrackPlan, strategy, fallback string) SF2Selection {
	if !spec.RequiresSF2 {
		return SF2Selection{}
	}
	if plan == nil {
		return ResolveSF2Selection(spec, strategy, fallback)
	}
	if strategy == "single" {
		return SF2Selection{
			Primary:  fallback,
			Programs: defaultProgramsForPlan(plan),
			Presets:  dedupePresets([]string{fallback}),
		}
	}
	intents := intentsFromPlan(plan)
	if len(intents) == 0 {
		return ResolveSF2Selection(spec, strategy, fallback)
	}
	return resolveSelection(spec, intents, strategy, fallback)
}

func resolveSelection(spec AlgoSpec, intents []SF2RoleIntent, strategy, fallback string) SF2Selection {
	primary := resolvePrimaryPreset(spec, intents, fallback)
	if strategy == "pro" {
		programs := resolveProgramsForPreset(spec.Name, intents, primary)
		return SF2Selection{
			Primary:  primary,
			Programs: programs,
			Presets:  dedupePresets([]string{primary}),
		}
	}
	routes, programs := resolveRouteAssignments(spec.Name, intents, primary, fallback)
	presets := make([]string, 0, len(routes)+1)
	if primary != "" {
		presets = append(presets, primary)
	}
	for _, preset := range routes {
		presets = append(presets, preset)
	}
	return SF2Selection{
		Primary:  primary,
		Routes:   routes,
		Programs: programs,
		Presets:  dedupePresets(presets),
	}
}

func resolvePrimaryPreset(spec AlgoSpec, intents []SF2RoleIntent, fallback string) string {
	if preferred := strings.TrimSpace(spec.PreferredSF2); preferred != "" {
		if _, ok := sf2Inventory[preferred]; ok {
			// Prefer it, but still let the scorer compare nearby options when a
			// track asks for something very different.
		}
	}
	best := fallback
	bestScore := -1 << 30
	for name, preset := range sf2Inventory {
		score := presetAggregateScore(spec.Name, preset, intents, nil)
		if strings.EqualFold(name, spec.PreferredSF2) {
			score += 10
		}
		if strings.EqualFold(name, fallback) {
			score += 2
		}
		if score > bestScore {
			best = name
			bestScore = score
		}
	}
	if best == "" {
		best = fallback
	}
	return best
}

func resolveProgramsForPreset(style string, intents []SF2RoleIntent, presetName string) map[int32]int32 {
	programs := make(map[int32]int32, len(intents))
	preset, ok := sf2Inventory[presetName]
	if !ok {
		return defaultProgramsForIntents(intents)
	}
	for _, intent := range intents {
		program, _, ok := bestProgramChoice(style, preset, intent, false)
		if ok && presetSupportsIntent(preset, intent) {
			programs[intent.Channel] = program
			continue
		}
		programs[intent.Channel] = defaultProgramForIntent(intent)
	}
	return programs
}

func resolveRouteAssignments(style string, intents []SF2RoleIntent, primary, fallback string) (map[int32]string, map[int32]int32) {
	routes := make(map[int32]string, len(intents))
	programs := make(map[int32]int32, len(intents))
	for _, intent := range intents {
		best := primary
		bestProgram := defaultProgramForIntent(intent)
		bestScore := -1 << 30
		for name, preset := range sf2Inventory {
			program, programScore, ok := bestProgramChoice(style, preset, intent, true)
			score := presetAggregateScore(style, preset, []SF2RoleIntent{intent}, stringSet(primary))
			if ok {
				score += programScore
			}
			if strings.EqualFold(name, primary) {
				score += 4
			}
			if score > bestScore {
				best = name
				bestScore = score
				if ok {
					bestProgram = program
				}
			}
		}
		if best == "" {
			best = fallback
		}
		routes[intent.Channel] = best
		programs[intent.Channel] = bestProgram
	}
	return routes, programs
}

func presetAggregateScore(style string, preset sf2PresetProfile, intents []SF2RoleIntent, cohesion map[string]bool) int {
	score := 0
	if containsFold(preset.Styles, style) {
		score += 8
	}
	switch strings.ToLower(strings.TrimSpace(preset.Realism)) {
	case "utility":
		score -= 14
	case "organic":
		score += 4
	case "storybook":
		if style == "bells" || style == "lullaby" || style == "ambient" {
			score += 6
		}
	case "polished":
		if style == "jazz" || style == "lofi" {
			score += 4
		}
	}
	if containsFold(preset.Tones, "generalist") {
		score -= 6
	}
	if containsFold(preset.Tones, "safe") {
		score -= 4
	}
	for _, intent := range intents {
		if !intent.Active {
			continue
		}
		score += bestProgramMatchScore(style, preset, intent, false)
		if containsFold(preset.Families, intent.Family) {
			score += 10
		}
	}
	if cohesion != nil && cohesion[preset.Name] {
		score += 3
	}
	return score
}

func bestProgramChoice(style string, preset sf2PresetProfile, intent SF2RoleIntent, strict bool) (int32, int, bool) {
	bestProgram := int32(0)
	bestScore := -1 << 30
	found := false
	for _, program := range preset.Programs {
		score := programMatchScore(style, preset, program, intent, strict)
		if score > bestScore {
			bestProgram = program.Program
			bestScore = score
			found = true
		}
	}
	return bestProgram, bestScore, found
}

func presetSupportsIntent(preset sf2PresetProfile, intent SF2RoleIntent) bool {
	for _, program := range preset.Programs {
		if strings.EqualFold(program.Family, intent.Family) {
			return true
		}
		if containsFold(program.Roles, intent.Role) || containsFold(program.Roles, roleBaseName(intent.Role)) {
			return true
		}
	}
	return false
}

func bestProgramMatchScore(style string, preset sf2PresetProfile, intent SF2RoleIntent, strict bool) int {
	_, score, ok := bestProgramChoice(style, preset, intent, strict)
	if !ok {
		return 0
	}
	return score
}

func programMatchScore(style string, preset sf2PresetProfile, program sf2ProgramProfile, intent SF2RoleIntent, strict bool) int {
	score := 0
	if strings.EqualFold(program.Family, intent.Family) {
		score += 24
	} else if strict {
		score -= 18
	}
	if containsFold(program.Roles, intent.Role) {
		score += 14
	}
	if containsFold(program.Roles, roleBaseName(intent.Role)) {
		score += 8
	}
	for _, tone := range intent.Tone {
		if containsFold(program.Tones, tone) {
			score += 5
		}
		if containsFold(preset.Tones, tone) {
			score += 2
		}
	}
	if intent.Articulation != "" {
		if containsFold(program.Articulations, intent.Articulation) {
			score += 8
		} else if containsFold(preset.Articulations, intent.Articulation) {
			score += 3
		}
	}
	if intent.Register != "" {
		if containsFold(program.Registers, intent.Register) {
			score += 6
		} else if strict {
			score -= 2
		}
	}
	switch strings.ToLower(intent.Prominence) {
	case "lead", "front":
		if containsFold(program.Blend, "front") || containsFold(preset.Blend, "front") {
			score += 6
		}
		if containsFold(program.Tones, "present") || containsFold(program.Tones, "clear") {
			score += 4
		}
	case "support":
		if containsFold(program.Blend, "core") || containsFold(program.Blend, "support") || containsFold(program.Blend, "glue") {
			score += 6
		}
		if containsFold(program.Tones, "soft") || containsFold(program.Tones, "warm") {
			score += 3
		}
	case "air":
		if containsFold(program.Blend, "air") || containsFold(program.Blend, "halo") || containsFold(program.Blend, "sparkle") {
			score += 6
		}
	case "anchor":
		if containsFold(program.Blend, "glue") || containsFold(program.Blend, "core") {
			score += 5
		}
		if containsFold(program.Registers, "low") || containsFold(program.Registers, "sub") {
			score += 3
		}
	}
	if containsFold(preset.Styles, style) {
		score += 4
	}
	switch strings.ToLower(strings.TrimSpace(program.Realism)) {
	case "utility":
		score -= 6
	case "organic":
		score += 2
	}
	return score
}

func defaultProgramsForPlan(plan *AuthoredTrackPlan) map[int32]int32 {
	if plan == nil {
		return nil
	}
	out := make(map[int32]int32, len(plan.Tracks))
	for _, track := range plan.Tracks {
		out[track.Channel] = track.Program
	}
	return out
}

func defaultProgramsForIntents(intents []SF2RoleIntent) map[int32]int32 {
	if len(intents) == 0 {
		return nil
	}
	out := make(map[int32]int32, len(intents))
	for _, intent := range intents {
		out[intent.Channel] = defaultProgramForIntent(intent)
	}
	return out
}

func defaultProgramForIntent(intent SF2RoleIntent) int32 {
	switch strings.ToLower(strings.TrimSpace(intent.Family)) {
	case "acoustic_piano":
		return 0
	case "electric_piano":
		return 5
	case "bass":
		return 32
	case "synth_bass":
		return 39
	case "reed_lead":
		return 66
	case "woodwind":
		return 73
	case "brass":
		return 56
	case "guitar":
		return 24
	case "mallet":
		return 11
	case "bells":
		return 14
	case "music_box":
		return 10
	case "pad":
		return 89
	case "choir":
		return 52
	case "strings":
		return 48
	case "organ":
		return 19
	case "drums":
		return 0
	default:
		return 0
	}
}

func roleIntentsForSpec(spec AlgoSpec) []SF2RoleIntent {
	return defaultRolePlan(spec.Name)
}

func intentsFromPlan(plan *AuthoredTrackPlan) []SF2RoleIntent {
	if plan == nil {
		return nil
	}
	seen := map[int32]int{}
	intents := make([]SF2RoleIntent, 0, len(plan.Tracks))
	for _, track := range plan.Tracks {
		family := strings.TrimSpace(track.Family)
		if family == "" {
			family = inferFamilyFromRole(track.Name)
		}
		intent := SF2RoleIntent{
			Channel:      track.Channel,
			Role:         track.Name,
			Family:       family,
			Tone:         append([]string(nil), track.Tone...),
			Articulation: track.Articulation,
			Register:     track.Register,
			Prominence:   track.Prominence,
			Active:       true,
		}
		if strings.TrimSpace(intent.Prominence) == "" {
			intent.Prominence = inferredProminence(track.Name, family)
		}
		if idx, ok := seen[intent.Channel]; ok {
			if intent.Family != "" && intents[idx].Family == "" {
				intents[idx].Family = intent.Family
			}
			intents[idx].Tone = dedupeFold(append(intents[idx].Tone, intent.Tone...))
			if intents[idx].Prominence == "" {
				intents[idx].Prominence = intent.Prominence
			}
			if intents[idx].Articulation == "" {
				intents[idx].Articulation = intent.Articulation
			}
			if intents[idx].Register == "" {
				intents[idx].Register = intent.Register
			}
			continue
		}
		seen[intent.Channel] = len(intents)
		intents = append(intents, intent)
	}
	return intents
}

func roleIntentChannel(style, name, family string) int32 {
	lowerName := strings.ToLower(strings.TrimSpace(name))
	family = strings.ToLower(strings.TrimSpace(family))
	switch lowerName {
	case "kick", "snare", "hat", "hihat", "ride", "crash", "openhat", "clap", "rim", "tom", "tom-high", "tom-low", "perc", "drums":
		return 9
	}
	switch style {
	case "lofi":
		switch lowerName {
		case "keys", "rhodes", "ep", "chords":
			return 0
		case "bass", "sub":
			return 1
		case "texture", "vibes", "vibraphone", "mallet":
			return 2
		case "lead", "sax", "hook", "counter", "flute":
			return 3
		case "guitar", "pluck":
			return 4
		case "pad", "choir":
			return 5
		}
	case "jazz":
		switch lowerName {
		case "keys", "piano", "comp":
			return 0
		case "bass", "walk":
			return 1
		case "lead", "sax", "horn", "alto", "tenor", "clarinet", "trumpet":
			return 2
		case "guitar", "vibes", "vibraphone":
			return 3
		case "organ":
			return 4
		}
	case "bells":
		switch lowerName {
		case "bells":
			return 0
		case "celesta":
			return 1
		case "glock":
			return 2
		case "box", "music_box":
			return 3
		case "pad":
			return 4
		case "choir":
			return 5
		case "strings":
			return 6
		case "bass":
			return 7
		case "shimmer":
			return 8
		}
	case "ambient":
		switch lowerName {
		case "pad":
			return 0
		case "choir":
			return 1
		case "texture", "bells", "sparkle":
			return 2
		case "lead", "flute", "woodwind":
			return 3
		case "bass":
			return 4
		case "strings":
			return 5
		case "shimmer":
			return 6
		}
	case "drone":
		switch lowerName {
		case "bed":
			return 0
		case "strings":
			return 1
		case "choir":
			return 2
		case "shimmer":
			return 3
		case "bass":
			return 4
		case "lead":
			return 5
		}
	case "classical":
		switch lowerName {
		case "piano":
			return 0
		case "strings":
			return 1
		case "winds":
			return 2
		case "brass":
			return 3
		case "harp":
			return 4
		case "choir":
			return 5
		}
	case "phase":
		switch lowerName {
		case "mallet-a", "mallet_a":
			return 0
		case "mallet-b", "mallet_b":
			return 1
		case "pad":
			return 2
		case "bass":
			return 3
		case "shimmer":
			return 4
		case "choir":
			return 5
		}
	case "lullaby":
		switch lowerName {
		case "lead":
			return 0
		case "harp":
			return 1
		case "choir":
			return 2
		case "box":
			return 3
		case "pad":
			return 4
		}
	}
	switch family {
	case "acoustic_piano", "electric_piano":
		return 0
	case "bass", "synth_bass":
		return 1
	case "reed_lead", "woodwind", "brass":
		return 2
	case "guitar", "mallet", "music_box":
		return 3
	case "pad":
		return 4
	case "choir":
		return 5
	case "strings":
		return 6
	default:
		return 0
	}
}

func inferFamilyFromRole(name string) string {
	lowerName := strings.ToLower(strings.TrimSpace(name))
	switch lowerName {
	case "kick", "snare", "hat", "hihat", "ride", "crash", "openhat", "clap", "rim", "tom", "tom-high", "tom-low", "perc", "drums":
		return "drums"
	case "keys", "rhodes", "ep", "chords":
		return "electric_piano"
	case "piano", "comp":
		return "acoustic_piano"
	case "bass", "sub", "walk":
		return "bass"
	case "texture", "vibes", "vibraphone", "celesta":
		return "mallet"
	case "glock", "bells":
		return "bells"
	case "box", "music_box":
		return "music_box"
	case "guitar", "pluck":
		return "guitar"
	case "lead", "sax", "alto", "tenor", "clarinet", "hook", "counter":
		return "reed_lead"
	case "trumpet", "horn", "brass":
		return "brass"
	case "flute", "winds", "woodwind":
		return "woodwind"
	case "strings":
		return "strings"
	case "choir":
		return "choir"
	case "pad", "bed":
		return "pad"
	case "shimmer":
		return "lead"
	case "harp":
		return "strings"
	default:
		return ""
	}
}

func inferredProminence(name, family string) string {
	lowerName := strings.ToLower(strings.TrimSpace(name))
	family = strings.ToLower(strings.TrimSpace(family))
	switch lowerName {
	case "lead", "bells", "sax", "alto", "tenor", "trumpet", "flute":
		return "lead"
	case "texture", "choir", "glock", "box", "music_box", "shimmer":
		return "air"
	case "bass", "sub", "walk", "kick", "snare", "hat", "ride", "drums":
		return "anchor"
	case "keys", "piano", "guitar", "strings", "winds", "harp", "organ", "vibes":
		return "support"
	}
	switch family {
	case "bass", "synth_bass", "drums":
		return "anchor"
	case "choir", "pad", "strings", "mallet":
		return "support"
	case "reed_lead", "woodwind", "brass", "bells":
		return "lead"
	default:
		return "support"
	}
}

func roleBaseName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	if idx := strings.Index(name, "-"); idx > 0 {
		return name[:idx]
	}
	return name
}

func dedupeFold(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		key := strings.ToLower(strings.TrimSpace(value))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	return out
}

func defaultRolePlan(style string) []SF2RoleIntent {
	switch style {
	case "ambient":
		return []SF2RoleIntent{
			{Channel: 0, Role: "pad", Family: "pad", Tone: []string{"dreamy", "wide"}, Prominence: "support", Active: true},
			{Channel: 1, Role: "choir", Family: "choir", Tone: []string{"airy", "soft"}, Prominence: "air", Active: true},
			{Channel: 2, Role: "texture", Family: "bells", Tone: []string{"glass", "sparkle"}, Prominence: "air", Active: true},
			{Channel: 3, Role: "lead", Family: "woodwind", Tone: []string{"soft"}, Prominence: "lead", Active: true},
			{Channel: 4, Role: "bass", Family: "synth_bass", Tone: []string{"warm"}, Prominence: "anchor", Active: true},
			{Channel: 5, Role: "shimmer", Family: "lead", Tone: []string{"shimmer"}, Prominence: "air", Active: true},
		}
	case "drone":
		return []SF2RoleIntent{
			{Channel: 0, Role: "bed", Family: "pad", Tone: []string{"wide", "soft"}, Prominence: "support", Active: true},
			{Channel: 1, Role: "strings", Family: "strings", Tone: []string{"soft", "floating"}, Prominence: "support", Active: true},
			{Channel: 2, Role: "choir", Family: "choir", Tone: []string{"airy"}, Prominence: "air", Active: true},
			{Channel: 3, Role: "shimmer", Family: "lead", Tone: []string{"icy", "shimmer"}, Prominence: "air", Active: true},
			{Channel: 4, Role: "bass", Family: "synth_bass", Tone: []string{"warm"}, Prominence: "anchor", Active: true},
		}
	case "bells":
		return []SF2RoleIntent{
			{Channel: 0, Role: "bells", Family: "bells", Tone: []string{"glass", "luminous"}, Prominence: "lead", Active: true},
			{Channel: 1, Role: "celesta", Family: "mallet", Tone: []string{"sparkle", "delicate"}, Prominence: "air", Active: true},
			{Channel: 2, Role: "glock", Family: "bells", Tone: []string{"glass", "sparkle"}, Prominence: "air", Active: true},
			{Channel: 3, Role: "music_box", Family: "music_box", Tone: []string{"delicate"}, Prominence: "air", Active: true},
			{Channel: 4, Role: "pad", Family: "pad", Tone: []string{"soft", "wide"}, Prominence: "support", Active: true},
			{Channel: 5, Role: "choir", Family: "choir", Tone: []string{"airy"}, Prominence: "support", Active: true},
			{Channel: 6, Role: "strings", Family: "strings", Tone: []string{"soft"}, Prominence: "support", Active: true},
			{Channel: 7, Role: "bass", Family: "bass", Tone: []string{"soft"}, Prominence: "anchor", Active: true},
		}
	case "lullaby":
		return []SF2RoleIntent{
			{Channel: 0, Role: "lead", Family: "mallet", Tone: []string{"delicate"}, Prominence: "lead", Active: true},
			{Channel: 1, Role: "harp", Family: "strings", Tone: []string{"soft"}, Prominence: "support", Active: true},
			{Channel: 2, Role: "keys", Family: "mallet", Tone: []string{"sparkle"}, Prominence: "support", Active: true},
			{Channel: 3, Role: "box", Family: "music_box", Tone: []string{"delicate"}, Prominence: "air", Active: true},
			{Channel: 4, Role: "choir", Family: "choir", Tone: []string{"airy"}, Prominence: "support", Active: true},
		}
	case "classical":
		return []SF2RoleIntent{
			{Channel: 0, Role: "piano", Family: "acoustic_piano", Tone: []string{"clear"}, Prominence: "lead", Active: true},
			{Channel: 1, Role: "strings", Family: "strings", Tone: []string{"lush"}, Prominence: "support", Active: true},
			{Channel: 2, Role: "winds", Family: "woodwind", Tone: []string{"soft"}, Prominence: "support", Active: true},
			{Channel: 3, Role: "brass", Family: "brass", Tone: []string{"rich"}, Prominence: "support", Active: true},
			{Channel: 4, Role: "choir", Family: "choir", Tone: []string{"airy"}, Prominence: "air", Active: true},
		}
	case "phase":
		return []SF2RoleIntent{
			{Channel: 0, Role: "mallet_a", Family: "mallet", Tone: []string{"glass", "metallic"}, Prominence: "lead", Active: true},
			{Channel: 1, Role: "mallet_b", Family: "mallet", Tone: []string{"glass", "metallic"}, Prominence: "lead", Active: true},
			{Channel: 2, Role: "pad", Family: "pad", Tone: []string{"soft"}, Prominence: "support", Active: true},
			{Channel: 3, Role: "bass", Family: "synth_bass", Tone: []string{"warm"}, Prominence: "anchor", Active: true},
			{Channel: 4, Role: "shimmer", Family: "bells", Tone: []string{"sparkle"}, Prominence: "air", Active: true},
			{Channel: 5, Role: "choir", Family: "choir", Tone: []string{"airy"}, Prominence: "air", Active: true},
		}
	case "lofi":
		return []SF2RoleIntent{
			{Channel: 0, Role: "keys", Family: "electric_piano", Tone: []string{"warm", "dusty", "soft"}, Prominence: "support", Active: true},
			{Channel: 1, Role: "bass", Family: "bass", Tone: []string{"woody", "round"}, Prominence: "anchor", Active: true},
			{Channel: 2, Role: "texture", Family: "mallet", Tone: []string{"glass", "soft"}, Prominence: "air", Active: true},
			{Channel: 3, Role: "lead", Family: "reed_lead", Tone: []string{"breathy", "intimate"}, Prominence: "lead", Active: true},
			{Channel: 4, Role: "guitar", Family: "guitar", Tone: []string{"soft", "warm"}, Prominence: "support", Active: true},
			{Channel: 9, Role: "drums", Family: "drums", Tone: []string{"tight", "dusty"}, Prominence: "anchor", Active: true},
		}
	case "jazz":
		return []SF2RoleIntent{
			{Channel: 0, Role: "keys", Family: "acoustic_piano", Tone: []string{"clear", "present"}, Prominence: "support", Active: true},
			{Channel: 1, Role: "bass", Family: "bass", Tone: []string{"woody", "round"}, Prominence: "anchor", Active: true},
			{Channel: 2, Role: "lead", Family: "reed_lead", Tone: []string{"present", "live"}, Prominence: "lead", Active: true},
			{Channel: 9, Role: "drums", Family: "drums", Tone: []string{"live", "soft"}, Prominence: "anchor", Active: true},
		}
	default:
		return nil
	}
}

func dedupePresets(names []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(names))
	for _, name := range names {
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func containsFold(values []string, want string) bool {
	for _, value := range values {
		if strings.EqualFold(value, want) {
			return true
		}
	}
	return false
}

func stringSet(values ...string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, value := range values {
		if value != "" {
			out[value] = true
		}
	}
	return out
}

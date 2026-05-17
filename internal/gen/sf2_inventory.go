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
	Name     string   `json:"name"`
	Styles   []string `json:"styles"`
	Families []string `json:"families"`
	Tones    []string `json:"tones"`
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
	Primary string
	Routes  map[int32]string
	Presets []string
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

func ResolveSF2Selection(spec AlgoSpec, blueprint *TrackBlueprint, strategy, fallback string) SF2Selection {
	if !spec.RequiresSF2 {
		return SF2Selection{}
	}
	if strategy == "single" {
		return SF2Selection{
			Primary: fallback,
			Presets: dedupePresets([]string{fallback}),
		}
	}
	intents := roleIntentsForSpec(spec, blueprint)
	if len(intents) == 0 {
		return SF2Selection{Primary: fallback, Presets: dedupePresets([]string{fallback})}
	}
	primary := resolvePrimaryPreset(spec, intents, fallback)
	if strategy == "pro" {
		return SF2Selection{
			Primary: primary,
			Presets: dedupePresets([]string{primary}),
		}
	}
	routes := resolveRoutePresets(spec.Name, intents, primary, fallback)
	presets := make([]string, 0, len(routes)+1)
	if primary != "" {
		presets = append(presets, primary)
	}
	for _, preset := range routes {
		presets = append(presets, preset)
	}
	return SF2Selection{
		Primary: primary,
		Routes:  routes,
		Presets: dedupePresets(presets),
	}
}

func resolvePrimaryPreset(spec AlgoSpec, intents []SF2RoleIntent, fallback string) string {
	if preferred := strings.TrimSpace(spec.PreferredSF2); preferred != "" {
		if _, ok := sf2Inventory[preferred]; ok {
			return preferred
		}
	}
	best := fallback
	bestScore := -1 << 30
	for name, preset := range sf2Inventory {
		score := presetScore(spec.Name, preset, intents, nil)
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

func resolveRoutePresets(style string, intents []SF2RoleIntent, primary, fallback string) map[int32]string {
	routes := make(map[int32]string, len(intents))
	for _, intent := range intents {
		best := primary
		bestScore := -1 << 30
		for name, preset := range sf2Inventory {
			score := presetScore(style, preset, []SF2RoleIntent{intent}, stringSet(primary))
			if score > bestScore {
				best = name
				bestScore = score
			}
		}
		if best == "" {
			best = fallback
		}
		routes[intent.Channel] = best
	}
	return routes
}

func presetScore(style string, preset sf2PresetProfile, intents []SF2RoleIntent, cohesion map[string]bool) int {
	score := 0
	if containsFold(preset.Styles, style) {
		score += 8
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
		if containsFold(preset.Families, intent.Family) {
			score += 12
		}
		for _, tone := range intent.Tone {
			if containsFold(preset.Tones, tone) {
				score += 4
			}
		}
		switch strings.ToLower(intent.Prominence) {
		case "front", "lead":
			if containsFold(preset.Tones, "present") || containsFold(preset.Tones, "clear") {
				score += 2
			}
		case "support", "air":
			if containsFold(preset.Tones, "soft") || containsFold(preset.Tones, "wide") {
				score += 2
			}
		}
	}
	if cohesion != nil && cohesion[preset.Name] {
		score += 3
	}
	return score
}

func roleIntentsForSpec(spec AlgoSpec, blueprint *TrackBlueprint) []SF2RoleIntent {
	base := defaultRolePlan(spec.Name)
	if blueprint == nil {
		return base
	}
	overrideRolePlan(spec.Name, base, blueprint.Roles)
	return base
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

func overrideRolePlan(style string, intents []SF2RoleIntent, roles map[string]RoleBlueprint) {
	for i := range intents {
		if role, ok := findRoleBlueprint(roles, intents[i].Role); ok {
			applyRoleOverride(&intents[i], role)
			continue
		}
		switch style {
		case "lofi":
			switch intents[i].Role {
			case "keys":
				if role, ok := findRoleBlueprint(roles, "comp"); ok {
					applyRoleOverride(&intents[i], role)
				}
			case "texture":
				if role, ok := findRoleBlueprint(roles, "texture", "vibes"); ok {
					applyRoleOverride(&intents[i], role)
				}
			case "guitar":
				if role, ok := findRoleBlueprint(roles, "counter"); ok {
					applyRoleOverride(&intents[i], role)
				}
			}
		}
	}
}

func applyRoleOverride(intent *SF2RoleIntent, role RoleBlueprint) {
	if strings.TrimSpace(role.Family) != "" {
		intent.Family = role.Family
	}
	if len(role.Tone) > 0 {
		intent.Tone = append([]string(nil), role.Tone...)
	}
	if strings.TrimSpace(role.Articulation) != "" {
		intent.Articulation = role.Articulation
	}
	if strings.TrimSpace(role.Register) != "" {
		intent.Register = role.Register
	}
	if strings.TrimSpace(role.Prominence) != "" {
		intent.Prominence = role.Prominence
	}
	intent.Active = role.Active
}

func findRoleBlueprint(roles map[string]RoleBlueprint, names ...string) (RoleBlueprint, bool) {
	for _, name := range names {
		for key, role := range roles {
			if strings.EqualFold(key, name) {
				return role, true
			}
		}
	}
	return RoleBlueprint{}, false
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

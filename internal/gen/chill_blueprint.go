package gen

import "strings"

type chillDrumBlueprint struct {
	hatDensity int
	kickBusy   bool
	ghosty     bool
	fillHeavy  bool
	openHat    bool
}

type chillBlueprint struct {
	hasTonic     bool
	tonicPC      int
	progression  []chillChord
	saxPhrase    []int
	vibePhrase   []int
	guitarPhrase []int
	roles        map[string]bool
	drums        chillDrumBlueprint
}

func (a *Chill) ApplyTrackBlueprint(blueprint TrackBlueprint) {
	a.authored = parseChillBlueprint(blueprint)
}

func parseChillBlueprint(blueprint TrackBlueprint) chillBlueprint {
	leadRole := roleFrom(blueprint, "lead")
	keysRole := roleFrom(blueprint, "keys", "comp")
	textureRole := roleFrom(blueprint, "texture", "vibes")
	guitarRole := roleFrom(blueprint, "guitar", "counter")
	drumRole := roleFrom(blueprint, "drums")
	out := chillBlueprint{
		progression: parseChillHarmony(blueprint.Harmony),
		saxPhrase:   parseChillMelody(roleValue(leadRole.Motif, leadRole.Pattern)),
		roles:       parseRoleActivity(blueprint.Roles),
		drums:       parseChillDrums(drumRole.Pattern),
	}
	if len(out.progression) > 0 {
		base := 0
		if fields := scorePatternTokens(blueprint.Harmony); len(fields) > 0 {
			if root, _, ok := parsePitchClassToken(fields[0]); ok {
				base = root
			}
		}
		out.hasTonic = true
		out.tonicPC = wrapPitchClass(base)
	}
	out.vibePhrase = parseChillComp(roleValue(textureRole.Pattern, keysRole.Pattern), []int{chillPlanNinth, chillPlanEleventh, chillPlanThirteenth, chillPlanResolveThird})
	out.guitarPhrase = parseChillComp(roleValue(guitarRole.Pattern, keysRole.Pattern), []int{chillPlanNinth, chillPlanSuspendFourth, chillPlanResolveThird, chillPlanThirteenth})
	return out
}

func parseChillHarmony(src string) []chillChord {
	fields := scorePatternTokens(src)
	if len(fields) == 0 {
		return nil
	}
	firstRoot, _, ok := parseChordToken(fields[0])
	if !ok {
		return nil
	}
	out := make([]chillChord, 0, len(fields))
	for _, token := range fields {
		root, tones, ok := parseChordToken(token)
		if !ok {
			continue
		}
		offset := wrapPitchClass(root - firstRoot)
		for i := range tones {
			tones[i] += offset
		}
		out = append(out, chillChord{tones: tones, label: token})
	}
	return out
}

func parseChordToken(token string) (int, []int, bool) {
	token = strings.TrimSpace(token)
	if token == "" || token == "|" {
		return 0, nil, false
	}
	root, rest, ok := parsePitchClassToken(token)
	if !ok {
		return 0, nil, false
	}
	lower := strings.ToLower(rest)
	switch {
	case strings.Contains(lower, "maj"):
		return root, []int{0, 4, 7, 11}, true
	case strings.Contains(lower, "m") && !strings.Contains(lower, "maj"):
		return root, []int{0, 3, 7, 10}, true
	default:
		return root, []int{0, 4, 7, 10}, true
	}
}

func parsePitchClassToken(token string) (int, string, bool) {
	if token == "" {
		return 0, "", false
	}
	baseMap := map[byte]int{
		'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11,
	}
	base, ok := baseMap[token[0]]
	if !ok {
		return 0, "", false
	}
	rest := token[1:]
	if len(rest) > 0 {
		switch rest[0] {
		case 'b':
			base--
			rest = rest[1:]
		case '#':
			base++
			rest = rest[1:]
		}
	}
	return wrapPitchClass(base), rest, true
}

func parseChillMelody(src string) []int {
	fields := scorePatternTokens(src)
	if len(fields) == 0 {
		return nil
	}
	out := make([]int, 0, len(fields))
	last := chillPlanRest
	for _, token := range fields {
		code := chillPlanToken(token, last)
		out = append(out, code)
		if code != chillPlanRest {
			last = code
		}
	}
	return out
}

func parseChillComp(src string, noteCycle []int) []int {
	if strings.TrimSpace(src) == "" {
		return nil
	}
	bars := strings.Split(src, "|")
	out := make([]int, 0, len(bars)*chillSupportSlotsPerBar)
	step := 0
	for _, bar := range bars {
		cells := strings.Fields(strings.TrimSpace(bar))
		if len(cells) == 0 {
			continue
		}
		slot0 := chillPlanRest
		slot1 := chillPlanRest
		for i, cell := range cells {
			if !strings.Contains(cell, "x") {
				continue
			}
			code := noteCycle[step%len(noteCycle)]
			if i < 2 {
				slot0 = code
			} else {
				slot1 = code
			}
			step++
		}
		out = append(out, slot0, slot1)
	}
	return out
}

func chillPlanToken(token string, last int) int {
	token = strings.TrimSpace(token)
	switch token {
	case "", "|", ".", "r":
		return chillPlanRest
	case "-":
		if last == chillPlanRest {
			return chillPlanRoot
		}
		return last
	}
	base := token
	if strings.HasPrefix(base, ">") || strings.HasPrefix(base, "^") {
		base = base[1:]
	}
	lower := strings.ToLower(base)
	switch {
	case strings.HasPrefix(lower, "b"):
		return chillPlanPickupBelow
	case strings.HasPrefix(lower, "#"):
		return chillPlanPickupAbove
	case strings.Contains(lower, "13") || lower == "6":
		return chillPlanThirteenth
	case strings.Contains(lower, "11") || lower == "4":
		return chillPlanSuspendFourth
	case strings.Contains(lower, "9") || lower == "2":
		return chillPlanNinth
	case lower == "7":
		return chillPlanSeventh
	case lower == "5":
		return chillPlanFifth
	case lower == "3":
		return chillPlanThird
	case lower == "1":
		return chillPlanRoot
	default:
		return chillPlanRest
	}
}

func parseRoleActivity(roles map[string]RoleBlueprint) map[string]bool {
	if len(roles) == 0 {
		return nil
	}
	out := make(map[string]bool, len(roles))
	for name, role := range roles {
		out[strings.ToLower(name)] = role.Active
	}
	return out
}

func roleFrom(blueprint TrackBlueprint, names ...string) RoleBlueprint {
	for _, name := range names {
		if role, ok := blueprint.Roles[name]; ok {
			return role
		}
		if role, ok := blueprint.Roles[strings.ToLower(name)]; ok {
			return role
		}
	}
	return RoleBlueprint{}
}

func roleValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func parseChillDrums(src string) chillDrumBlueprint {
	var out chillDrumBlueprint
	lower := strings.ToLower(src)
	hhCount := strings.Count(lower, "hh:") * 0
	for _, field := range strings.Fields(lower) {
		if strings.Contains(field, "x") && (strings.Contains(field, "x.x") || strings.Count(field, "x") >= 3) {
			hhCount += strings.Count(field, "x")
		}
	}
	switch {
	case hhCount >= 8:
		out.hatDensity = 2
	case hhCount >= 4:
		out.hatDensity = 1
	}
	out.kickBusy = strings.Count(lower, "bd:") > 0 && strings.Count(lower, "x") >= 8
	out.ghosty = strings.Contains(lower, "ghost") || strings.Count(lower, "sd:") > 0 && strings.Count(lower, "..x") >= 2
	out.fillHeavy = strings.Contains(lower, "fill") || strings.Count(lower, "sn:") > 0 && strings.Count(lower, "x") >= 10
	out.openHat = out.hatDensity > 0 || strings.Contains(lower, "oh:")
	return out
}

func scorePatternTokens(src string) []string {
	raw := strings.Fields(strings.ReplaceAll(src, "|", " | "))
	out := raw[:0]
	for _, token := range raw {
		if token != "|" && strings.TrimSpace(token) != "" {
			out = append(out, token)
		}
	}
	return out
}

func (a *Chill) authoredMotifBundle(numBars int) (chillMotifBundle, bool) {
	if len(a.authored.saxPhrase) == 0 && len(a.authored.vibePhrase) == 0 && len(a.authored.guitarPhrase) == 0 {
		return chillMotifBundle{}, false
	}
	supportLen := maxInt(chillSupportMotifSlots, chillSupportSlotsPerBar*numBars)
	leadLen := maxInt(chillLeadMotifBars, numBars)
	return chillMotifBundle{
		vibe:   a.authoredSupportMotifs(a.authored.vibePhrase, supportLen, chillPlanNinth),
		guitar: a.authoredSupportMotifs(a.authored.guitarPhrase, supportLen, chillPlanNinth),
		sax:    a.authoredLeadMotifs(a.authored.saxPhrase, leadLen),
	}, true
}

func (a *Chill) authoredSupportMotifs(base []int, length, fill int) MotifMemory {
	base = trimOrRepeatPhrase(base, length, fill)
	aprime := a.transformChillPhrase(base, fill)
	b := rotatePhrase(base, 2)
	cadence := trimOrRepeatPhrase(append(copyPhrase(base), chillPlanResolveThird, chillPlanNinth, chillPlanResolveThird, chillPlanRoot), length, fill)
	outro := trimOrRepeatPhrase([]int{fill, chillPlanResolveThird, chillPlanRoot, chillPlanRest}, length, chillPlanRest)
	return MotifMemory{A: base, Aprime: aprime, B: b, Cadence: cadence, Outro: outro}
}

func (a *Chill) authoredLeadMotifs(base []int, length int) MotifMemory {
	base = trimOrRepeatPhrase(base, length, chillPlanRest)
	aprime := a.transformChillPhrase(base, chillPlanRest)
	b := rotatePhrase(base, 4)
	cadence := trimOrRepeatPhrase(append(copyPhrase(base), chillPlanResolveThird, chillPlanPickupBelow, chillPlanResolveThird, chillPlanRoot), length, chillPlanRest)
	outro := trimOrRepeatPhrase([]int{chillPlanRest, chillPlanResolveThird, chillPlanRest, chillPlanRoot}, length, chillPlanRest)
	return MotifMemory{A: base, Aprime: aprime, B: b, Cadence: cadence, Outro: outro}
}

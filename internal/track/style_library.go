package track

import "strings"

type roleBuckets struct {
	melody []string
	comp   []string
	bass   []string
	pad    []string
	drums  []string
}

func applyStyleLibrary(pack stylePack, section Section, roles map[string]Role) (Section, map[string]Role) {
	roles = cloneRoles(roles)
	buckets := bucketRoles(roles)
	for name, role := range roles {
		roles[name] = applyStylePhraseLibrary(pack, section, name, role)
	}
	section = applyStyleArrangementLibrary(pack, section, buckets)
	return section, roles
}

func bucketRoles(roles map[string]Role) roleBuckets {
	var out roleBuckets
	for name, role := range roles {
		switch authoredRoleKind(name, role) {
		case "melody":
			out.melody = append(out.melody, name)
		case "bass":
			out.bass = append(out.bass, name)
		case "pad":
			out.pad = append(out.pad, name)
		case "drum":
			out.drums = append(out.drums, name)
		default:
			out.comp = append(out.comp, name)
		}
	}
	return out
}

func applyStylePhraseLibrary(pack stylePack, section Section, name string, role Role) Role {
	if strings.TrimSpace(role.Family) == "" {
		return role
	}
	if role.Phrases == nil {
		role.Phrases = map[string]PhraseBlock{}
	}
	kind := authoredRoleKind(name, role)
	switch kind {
	case "melody":
		base := firstNonBlank(role.Motif, pack.defaultMelody(name))
		role.Phrases = ensurePhraseBlock(role.Phrases, "statement", PhraseBlock{Motif: base})
		role.Phrases = ensurePhraseBlock(role.Phrases, "answer", PhraseBlock{Motif: phraseAnswerMotif(pack, base)})
		role.Phrases = ensurePhraseBlock(role.Phrases, "sequence", PhraseBlock{Motif: phraseSequenceMotif(pack, base)})
		role.Phrases = ensurePhraseBlock(role.Phrases, "release", PhraseBlock{Motif: phraseReleaseMotif(pack, base)})
		role.Phrases = ensurePhraseBlock(role.Phrases, "cadence", PhraseBlock{Motif: phraseCadenceMotif(pack, base)})
	case "comp":
		base := firstNonBlank(role.Pattern, pack.defaultRhythm(name))
		role.Phrases = ensurePhraseBlock(role.Phrases, "statement", PhraseBlock{Pattern: base})
		role.Phrases = ensurePhraseBlock(role.Phrases, "answer", PhraseBlock{Pattern: phraseAnswerRhythm(pack, kind, base)})
		role.Phrases = ensurePhraseBlock(role.Phrases, "sequence", PhraseBlock{Pattern: phraseSequenceRhythm(pack, kind, base)})
		role.Phrases = ensurePhraseBlock(role.Phrases, "release", PhraseBlock{Pattern: phraseReleaseRhythm(pack, kind, base)})
		role.Phrases = ensurePhraseBlock(role.Phrases, "cadence", PhraseBlock{Pattern: phraseCadenceRhythm(pack, kind, base)})
	case "bass":
		base := firstNonBlank(role.Pattern, pack.defaultRhythm(name))
		role.Phrases = ensurePhraseBlock(role.Phrases, "statement", PhraseBlock{Pattern: base})
		role.Phrases = ensurePhraseBlock(role.Phrases, "answer", PhraseBlock{Pattern: phraseAnswerRhythm(pack, kind, base)})
		role.Phrases = ensurePhraseBlock(role.Phrases, "sequence", PhraseBlock{Pattern: phraseSequenceRhythm(pack, kind, base)})
		role.Phrases = ensurePhraseBlock(role.Phrases, "release", PhraseBlock{Pattern: phraseReleaseRhythm(pack, kind, base)})
		role.Phrases = ensurePhraseBlock(role.Phrases, "cadence", PhraseBlock{Pattern: phraseCadenceRhythm(pack, kind, base)})
	case "drum":
		base := firstNonBlank(role.Pattern, pack.defaultRhythm(name))
		role.Phrases = ensurePhraseBlock(role.Phrases, "statement", PhraseBlock{Pattern: base})
		role.Phrases = ensurePhraseBlock(role.Phrases, "answer", PhraseBlock{Pattern: phraseAnswerRhythm(pack, kind, base)})
		role.Phrases = ensurePhraseBlock(role.Phrases, "sequence", PhraseBlock{Pattern: phraseSequenceRhythm(pack, kind, base)})
		role.Phrases = ensurePhraseBlock(role.Phrases, "release", PhraseBlock{Pattern: phraseReleaseRhythm(pack, kind, base)})
		role.Phrases = ensurePhraseBlock(role.Phrases, "cadence", PhraseBlock{Pattern: phraseCadenceRhythm(pack, kind, base)})
	case "pad":
		base := firstNonBlank(role.Pattern, pack.defaultRhythm(name))
		role.Phrases = ensurePhraseBlock(role.Phrases, "statement", PhraseBlock{Pattern: base})
		role.Phrases = ensurePhraseBlock(role.Phrases, "answer", PhraseBlock{Pattern: phraseAnswerRhythm(pack, kind, base)})
		role.Phrases = ensurePhraseBlock(role.Phrases, "release", PhraseBlock{Pattern: phraseReleaseRhythm(pack, kind, base)})
		role.Phrases = ensurePhraseBlock(role.Phrases, "cadence", PhraseBlock{Pattern: phraseCadenceRhythm(pack, kind, base)})
	}
	return role
}

func ensurePhraseBlock(phrases map[string]PhraseBlock, label string, block PhraseBlock) map[string]PhraseBlock {
	key := strings.ToLower(strings.TrimSpace(label))
	if key == "" {
		return phrases
	}
	current, ok := phrases[key]
	if !ok {
		phrases[key] = block
		return phrases
	}
	if strings.TrimSpace(current.Pattern) == "" && strings.TrimSpace(block.Pattern) != "" {
		current.Pattern = block.Pattern
	}
	if strings.TrimSpace(current.Motif) == "" && strings.TrimSpace(block.Motif) != "" {
		current.Motif = block.Motif
	}
	// PhraseBlock.Harmony removed in SP8 (v1 dead field).
	phrases[key] = current
	return phrases
}

func applyStyleArrangementLibrary(pack stylePack, section Section, buckets roleBuckets) Section {
	desc := strings.ToLower(strings.TrimSpace(strings.Join([]string{section.ID, section.Title, section.Scene, section.Variation}, " ")))
	switch pack.Name {
	case "jazz":
		if hasText(desc, "head", "statement", "intro") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "pickup", Bar: 0, Roles: buckets.melody, Motif: "5 6 7 9"})
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "fill", Bar: 0, Roles: buckets.drums})
		}
		if hasText(desc, "bridge", "reharm", "suspended") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "stab", Bar: 1, Roles: buckets.comp, Pattern: "x... ...."})
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "stop", Bar: 3, Roles: appendNames(buckets.melody, buckets.comp, buckets.bass)})
		}
		if hasText(desc, "shout") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "double", Bar: 1, Roles: appendNames(buckets.melody, buckets.comp)})
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "fill", Bar: 0, Roles: buckets.drums})
		}
		if hasText(desc, "release", "answer") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "drop", Bar: 2, Roles: buckets.drums})
		}
		if hasText(desc, "cadence", "outro", "tag") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "tag", Bar: 0, Roles: appendNames(buckets.melody, buckets.comp)})
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "ending", Bar: 0, Roles: appendNames(buckets.melody, buckets.comp, buckets.bass)})
		}
	case "lofi":
		if hasText(desc, "intro", "establish", "hush") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "breath", Bar: 1, Roles: appendNames(buckets.melody, buckets.comp)})
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "drop", Bar: 1, Roles: buckets.drums})
		}
		if hasText(desc, "bridge", "lift", "open-register") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "pickup", Bar: 0, Roles: buckets.melody, Motif: "5 6 7 9"})
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "swell", Bar: 1, Roles: appendNames(buckets.comp, buckets.pad)})
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "fill", Bar: 0, Roles: buckets.drums})
		}
		if hasText(desc, "breakdown", "subtract", "thin") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "drop", Bar: 2, Roles: appendNames(buckets.bass, buckets.drums)})
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "hold", Bar: 0, Roles: appendNames(buckets.comp, buckets.pad)})
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "breath", Bar: 0, Roles: buckets.melody})
		}
		if hasText(desc, "cadence", "outro", "home") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "pedal", Bar: 0, Roles: appendNames(buckets.bass, buckets.pad)})
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "ending", Bar: 0, Roles: appendNames(buckets.melody, buckets.comp, buckets.bass)})
		}
	case "bells":
		if hasText(desc, "vespers", "chapel", "cloister", "devotional", "intro") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "pickup", Bar: 0, Roles: buckets.melody, Motif: "5 7 9 7"})
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "swell", Bar: 1, Roles: appendNames(buckets.comp, buckets.pad)})
		}
		if hasText(desc, "answer", "release") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "hold", Bar: 0, Roles: appendNames(buckets.comp, buckets.pad)})
		}
		if hasText(desc, "cadence", "outro") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "tag", Bar: 0, Roles: buckets.melody})
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "ending", Bar: 0, Roles: appendNames(buckets.melody, buckets.pad)})
		}
	case "classical":
		if hasText(desc, "nocturne", "chamber", "intro", "statement") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "pickup", Bar: 0, Roles: buckets.melody, Motif: "3 5 6 7"})
		}
		if hasText(desc, "bridge", "development") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "breath", Bar: 2, Roles: buckets.comp})
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "swell", Bar: 1, Roles: appendNames(buckets.comp, buckets.pad)})
		}
		if hasText(desc, "cadence", "outro", "release") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "hold", Bar: 0, Roles: appendNames(buckets.comp, buckets.pad, buckets.melody)})
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "ending", Bar: 0, Roles: appendNames(buckets.melody, buckets.comp)})
		}
	case "ambient", "drone":
		if hasText(desc, "intro", "establish", "field", "haze") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "swell", Bar: 1, Roles: appendNames(buckets.comp, buckets.pad)})
		}
		if hasText(desc, "bridge", "lift", "glide", "drift") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "pedal", Bar: 0, Roles: appendNames(buckets.bass, buckets.pad)})
		}
		if hasText(desc, "cadence", "outro", "release") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "hold", Bar: 0, Roles: appendNames(buckets.comp, buckets.pad, buckets.bass)})
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "breath", Bar: 0, Roles: buckets.melody})
		}
	case "phase":
		if hasText(desc, "mirror", "steps", "interlock", "intro") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "pickup", Bar: 0, Roles: buckets.melody, Motif: "3 5 6 7"})
		}
		if hasText(desc, "break", "thin", "answer") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "drop", Bar: 2, Roles: appendNames(buckets.melody, buckets.comp)})
		}
		if hasText(desc, "cadence", "outro") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "hold", Bar: 0, Roles: appendNames(buckets.comp, buckets.pad)})
		}
	case "lullaby":
		if hasText(desc, "intro", "paper", "staircase", "verse") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "pickup", Bar: 0, Roles: buckets.melody, Motif: "3 5 6 5"})
		}
		if hasText(desc, "cadence", "outro", "sleep") {
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "breath", Bar: 0, Roles: buckets.melody})
			section.Events = appendLibraryEvent(section.Events, Event{Kind: "hold", Bar: 0, Roles: appendNames(buckets.comp, buckets.pad)})
		}
	}
	return section
}

func appendLibraryEvent(events []Event, event Event) []Event {
	if len(event.Roles) == 0 && event.Kind != "swell" && event.Kind != "ending" {
		return events
	}
	for _, existing := range events {
		if strings.EqualFold(strings.TrimSpace(existing.Kind), strings.TrimSpace(event.Kind)) &&
			existing.Bar == event.Bar &&
			sameRoleSet(existing.Roles, event.Roles) {
			return events
		}
	}
	return append(events, event)
}

func phraseAnswerMotif(pack stylePack, base string) string {
	if strings.TrimSpace(base) == "" {
		return base
	}
	switch pack.Name {
	case "jazz":
		return shiftMelodyPattern(base, 1, false)
	case "bells", "ambient", "drone", "lullaby":
		return simplifyMelodyPattern(base)
	default:
		return shiftMelodyPattern(base, 1, false)
	}
}

func phraseSequenceMotif(pack stylePack, base string) string {
	if strings.TrimSpace(base) == "" {
		return base
	}
	switch pack.Name {
	case "jazz", "classical":
		return shiftMelodyPattern(base, 1, false)
	case "lofi":
		return shiftMelodyPattern(base, 2, false)
	default:
		return shiftMelodyPattern(base, 1, false)
	}
}

func phraseReleaseMotif(pack stylePack, base string) string {
	if strings.TrimSpace(base) == "" {
		return base
	}
	switch pack.Name {
	case "bells", "ambient", "drone", "lullaby":
		return simplifyMelodyPattern(base)
	default:
		return simplifyMelodyPattern(base)
	}
}

func phraseCadenceMotif(pack stylePack, base string) string {
	if strings.TrimSpace(base) == "" {
		return base
	}
	return rewriteCadenceMotif(base)
}

func phraseAnswerRhythm(pack stylePack, kind, base string) string {
	if strings.TrimSpace(base) == "" {
		return base
	}
	switch kind {
	case "drum":
		if pack.Name == "jazz" {
			return "x.x. x.xx"
		}
		return thinRhythmPattern(base)
	case "bass":
		if pack.Name == "jazz" {
			return "x..xx..x"
		}
		return thinRhythmPattern(base)
	case "pad":
		return holdRhythmPattern(base)
	default:
		return thinRhythmPattern(base)
	}
}

func phraseSequenceRhythm(pack stylePack, kind, base string) string {
	if strings.TrimSpace(base) == "" {
		return base
	}
	switch kind {
	case "drum":
		if pack.Name == "jazz" {
			return "xxxxxxxx"
		}
		return densifyRhythmPattern(base)
	case "bass":
		if pack.Name == "jazz" {
			return "xxxxxxxx"
		}
		return densifyRhythmPattern(base)
	case "comp":
		if pack.Name == "jazz" {
			return "x... x..x"
		}
		return densifyRhythmPattern(base)
	default:
		return densifyRhythmPattern(base)
	}
}

func phraseReleaseRhythm(pack stylePack, kind, base string) string {
	if strings.TrimSpace(base) == "" {
		return base
	}
	switch kind {
	case "pad":
		return holdRhythmPattern(base)
	case "bass":
		if pack.Name == "ambient" || pack.Name == "drone" {
			return "x......."
		}
		return thinRhythmPattern(base)
	case "drum":
		return thinRhythmPattern(base)
	default:
		return holdRhythmPattern(base)
	}
}

func phraseCadenceRhythm(pack stylePack, kind, base string) string {
	switch kind {
	case "drum":
		if pack.Name == "jazz" {
			return "x..xx.xx"
		}
		return "x..xx..x"
	case "bass":
		if pack.Name == "jazz" {
			return "x.xxxxxx"
		}
		return "x...x..x"
	case "pad":
		return "x......."
	default:
		if pack.Name == "lofi" {
			return "x... ...."
		}
		return holdRhythmPattern(base)
	}
}

func simplifyMelodyPattern(pattern string) string {
	tokens := strings.Fields(strings.ReplaceAll(pattern, "|", " | "))
	noteIdx := 0
	for i, token := range tokens {
		if token == "|" || token == "." || token == "-" || token == "r" {
			continue
		}
		if noteIdx%2 == 1 {
			tokens[i] = "."
		}
		noteIdx++
	}
	return compactPatternTokens(tokens)
}

func thinRhythmPattern(pattern string) string {
	return rewriteRhythmPattern(pattern, func(idx int, hit bool) bool {
		if !hit {
			return false
		}
		return idx%2 == 0
	})
}

func densifyRhythmPattern(pattern string) string {
	return rewriteRhythmPattern(pattern, func(idx int, hit bool) bool {
		if hit {
			return true
		}
		return idx%4 == 3
	})
}

func holdRhythmPattern(pattern string) string {
	bars := maxInt(1, len(strings.Split(strings.TrimSpace(pattern), "|")))
	parts := make([]string, bars)
	for i := range parts {
		parts[i] = "x......."
	}
	return strings.Join(parts, " | ")
}

func rewriteRhythmPattern(pattern string, keep func(idx int, hit bool) bool) string {
	fields := strings.Split(strings.TrimSpace(pattern), "|")
	if len(fields) == 0 {
		return pattern
	}
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		raw := strings.TrimSpace(strings.ReplaceAll(field, " ", ""))
		if raw == "" {
			continue
		}
		runes := []rune(raw)
		for i, r := range runes {
			hit := r == 'x' || r == 'X'
			if keep(i, hit) {
				runes[i] = 'x'
			} else {
				runes[i] = '.'
			}
		}
		out = append(out, string(runes))
	}
	return strings.Join(out, " | ")
}

func compactPatternTokens(tokens []string) string {
	var out []string
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if token == "|" {
			if len(out) > 0 && out[len(out)-1] != "|" {
				out = append(out, "|")
			}
			continue
		}
		out = append(out, token)
	}
	return strings.TrimSpace(strings.Join(out, " "))
}

func sameRoleSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	counts := map[string]int{}
	for _, item := range a {
		counts[strings.ToLower(strings.TrimSpace(item))]++
	}
	for _, item := range b {
		key := strings.ToLower(strings.TrimSpace(item))
		counts[key]--
		if counts[key] < 0 {
			return false
		}
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}

func appendNames(groups ...[]string) []string {
	var out []string
	seen := map[string]bool{}
	for _, group := range groups {
		for _, name := range group {
			key := strings.ToLower(strings.TrimSpace(name))
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, name)
		}
	}
	return out
}

func hasText(text string, parts ...string) bool {
	for _, part := range parts {
		if part != "" && strings.Contains(text, strings.ToLower(strings.TrimSpace(part))) {
			return true
		}
	}
	return false
}

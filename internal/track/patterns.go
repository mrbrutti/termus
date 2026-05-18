package track

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	harmonyTokenRE  = regexp.MustCompile(`^(?:[A-G](?:#|b)?[A-Za-z0-9()+/#-]*|[ivIV]+[A-Za-z0-9()+/#-]*)$`)
	melodyTokenRE   = regexp.MustCompile(`^(?:[.\-|r]|[><^]?(?:#|b)?[0-9]+)$`)
	rhythmTokenRE   = regexp.MustCompile(`^(?:[a-zA-Z][a-zA-Z0-9_-]*:)?[x.\-]+$`)
	sceneTokenRE    = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	registerRE      = regexp.MustCompile(`^(?:sub|low|mid|mid-high|high|air)$`)
	// euclideanRE matches E(k,n) or E(k,n,rotate:r) optionally preceded by
	// a role prefix like "kick:".
	euclideanRE = regexp.MustCompile(
		`^([a-zA-Z][a-zA-Z0-9_-]*:)?E\(\s*(\d+)\s*,\s*(\d+)(?:\s*,\s*rotate\s*:\s*(-?\d+))?\s*\)$`,
	)
)

// expandRhythmToken expands a single rhythm token that may contain Euclidean
// syntax (E(k,n) or E(k,n,rotate:r)) into a literal x/. string.
// Returns the expanded token and nil on success, or "" and an error on
// parse failure. Non-Euclidean tokens are returned unchanged.
func expandRhythmToken(token string) (string, error) {
	m := euclideanRE.FindStringSubmatch(token)
	if m == nil {
		return token, nil
	}
	prefix := m[1] // e.g. "kick:" or ""
	k, err := strconv.Atoi(m[2])
	if err != nil {
		return "", fmt.Errorf("euclidean k: %w", err)
	}
	n, err := strconv.Atoi(m[3])
	if err != nil {
		return "", fmt.Errorf("euclidean n: %w", err)
	}
	rotate := 0
	if m[4] != "" {
		rotate, err = strconv.Atoi(m[4])
		if err != nil {
			return "", fmt.Errorf("euclidean rotate: %w", err)
		}
	}
	expanded, err := euclideanPatternString(k, n, rotate)
	if err != nil {
		return "", fmt.Errorf("E(%d,%d): %w", k, n, err)
	}
	return prefix + expanded, nil
}

// ExpandRhythmPattern expands any Euclidean E(...) tokens in a rhythm pattern
// string, returning the fully literal x/. pattern. Returns an error for any
// malformed E(...) token; non-Euclidean patterns pass through unchanged.
func ExpandRhythmPattern(pattern string) (string, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return "", nil
	}
	tokens := strings.Fields(strings.ReplaceAll(pattern, "|", " | "))
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token == "|" {
			out = append(out, token)
			continue
		}
		expanded, err := expandRhythmToken(token)
		if err != nil {
			return "", err
		}
		out = append(out, expanded)
	}
	return strings.Join(out, " "), nil
}

func validatePattern(pattern, kind string) error {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return nil
	}
	// For rhythm patterns, expand Euclidean syntax before validation.
	if kind == "rhythm" {
		expanded, err := ExpandRhythmPattern(pattern)
		if err != nil {
			return err
		}
		pattern = expanded
	}
	tokens := strings.Fields(strings.ReplaceAll(pattern, "|", " | "))
	var tokenRE *regexp.Regexp
	switch kind {
	case "harmony":
		tokenRE = harmonyTokenRE
	case "melody":
		tokenRE = melodyTokenRE
	case "rhythm":
		tokenRE = rhythmTokenRE
	case "scene":
		tokenRE = sceneTokenRE
	default:
		return fmt.Errorf("unknown pattern kind %q", kind)
	}
	for _, token := range tokens {
		if token == "|" {
			continue
		}
		if !tokenRE.MatchString(token) {
			return fmt.Errorf("invalid %s token %q", kind, token)
		}
	}
	return nil
}

func validateRole(name string, role Role) error {
	if err := validatePattern(role.Pattern, "rhythm"); err != nil {
		return fmt.Errorf("roles.%s.pattern: %w", name, err)
	}
	if err := validatePattern(role.Motif, "melody"); err != nil {
		return fmt.Errorf("roles.%s.motif: %w", name, err)
	}
	// Role.Harmony removed in SP8 (v1 dead field).
	if role.Register != "" && !registerRE.MatchString(role.Register) {
		return fmt.Errorf("roles.%s.register: invalid register %q", name, role.Register)
	}
	for phrase, block := range role.Phrases {
		if err := validatePattern(block.Pattern, "rhythm"); err != nil {
			return fmt.Errorf("roles.%s.phrases.%s.pattern: %w", name, phrase, err)
		}
		if err := validatePattern(block.Motif, "melody"); err != nil {
			return fmt.Errorf("roles.%s.phrases.%s.motif: %w", name, phrase, err)
		}
		// PhraseBlock.Harmony removed in SP8 (v1 dead field).
	}
	return nil
}

func validateOrchestrationRole(name string, role OrchestrationRole) error {
	if role.Register != "" && !registerRE.MatchString(role.Register) {
		return fmt.Errorf("orchestration.roles.%s.register: invalid register %q", name, role.Register)
	}
	return nil
}

func validateEvent(sectionIndex, eventIndex int, event Event) error {
	path := fmt.Sprintf("sections[%d].events[%d]", sectionIndex, eventIndex)
	kind := strings.ToLower(strings.TrimSpace(event.Kind))
	switch kind {
	case "drop", "stop", "fill", "pickup", "stab", "pedal", "swell", "double", "break", "tag", "ending",
		"crescendo", "decrescendo", "breath", "hold", "silence":
	default:
		return fmt.Errorf("%s.kind: unsupported event kind %q", path, event.Kind)
	}
	if event.Bar < 0 {
		return fmt.Errorf("%s.bar: must be >= 0", path)
	}
	if event.Bars < 0 {
		return fmt.Errorf("%s.bars: must be >= 0", path)
	}
	if event.Slot < 0 || event.Slot > authoredSlotsPerBar {
		return fmt.Errorf("%s.slot: must be within 0..%d", path, authoredSlotsPerBar)
	}
	if err := validatePattern(event.Pattern, "rhythm"); err != nil {
		return fmt.Errorf("%s.pattern: %w", path, err)
	}
	if err := validatePattern(event.Motif, "melody"); err != nil {
		return fmt.Errorf("%s.motif: %w", path, err)
	}
	for i, role := range event.Roles {
		role = strings.TrimSpace(role)
		if role == "" {
			return fmt.Errorf("%s.roles[%d]: role name cannot be empty", path, i)
		}
		if strings.Contains(role, " ") || strings.Contains(role, "\t") || strings.Contains(role, "|") {
			return fmt.Errorf("%s.roles[%d]: invalid role name %q", path, i, role)
		}
	}
	return nil
}

func baseRoleName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	idx := strings.LastIndex(name, "-")
	if idx <= 0 || idx == len(name)-1 {
		return name
	}
	if _, err := strconv.Atoi(name[idx+1:]); err == nil {
		return name[:idx]
	}
	return name
}

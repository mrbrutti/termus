package track

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	harmonyTokenRE = regexp.MustCompile(`^(?:[A-G](?:#|b)?[A-Za-z0-9()+/#-]*|[ivIV]+[A-Za-z0-9()+/#-]*)$`)
	melodyTokenRE  = regexp.MustCompile(`^(?:[.\-|r]|[><^]?(?:#|b)?[0-9]+)$`)
	rhythmTokenRE  = regexp.MustCompile(`^(?:[a-zA-Z][a-zA-Z0-9_-]*:)?[x.\-]+$`)
	sceneTokenRE   = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	registerRE     = regexp.MustCompile(`^(?:sub|low|mid|mid-high|high|air)$`)
)

func validatePattern(pattern, kind string) error {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return nil
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
	if err := validatePattern(role.Harmony, "harmony"); err != nil {
		return fmt.Errorf("roles.%s.harmony: %w", name, err)
	}
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
		if err := validatePattern(block.Harmony, "harmony"); err != nil {
			return fmt.Errorf("roles.%s.phrases.%s.harmony: %w", name, phrase, err)
		}
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

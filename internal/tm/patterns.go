package tm

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	harmonyTokenRE = regexp.MustCompile(`^(?:[A-Ga-g][#b]?[A-Za-z0-9()+/_-]*|[b#]?[ivIV]+[A-Za-z0-9()+/_-]*)$`)
	melodyTokenRE  = regexp.MustCompile(`^(?:[<>^vb#-]?\d+|[rR]|\.|-)$`)
	rhythmTokenRE  = regexp.MustCompile(`^(?:[xXoO~\.\-]+|[a-zA-Z0-9_]+:)$`)
	arrangeTokenRE = regexp.MustCompile(`^(?:[a-zA-Z][a-zA-Z0-9_-]*|[+][a-zA-Z][a-zA-Z0-9_-]*)$`)
)

func validatePattern(pattern string, kind string) error {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return nil
	}
	var tokenRE *regexp.Regexp
	switch kind {
	case "harmony":
		tokenRE = harmonyTokenRE
	case "melody":
		tokenRE = melodyTokenRE
	case "rhythm":
		tokenRE = rhythmTokenRE
	case "arrange":
		tokenRE = arrangeTokenRE
	default:
		return fmt.Errorf("unknown pattern kind %q", kind)
	}
	for _, tok := range strings.Fields(pattern) {
		if tok == "|" {
			continue
		}
		if !tokenRE.MatchString(tok) {
			return fmt.Errorf("invalid %s token %q", kind, tok)
		}
	}
	return nil
}

package track

import (
	"fmt"
	"strconv"
	"strings"
)

// ResolveMotifs takes the raw motif entries from a parsed file and returns a
// name→resolved-pattern map after applying transforms recursively.
//
// Resolution rules:
//   - If a motif has no based_on, its pattern is used as the starting point.
//   - If a motif has based_on, the resolved pattern of the parent is used.
//   - Transforms are applied in order: Retrograde → Transpose → Invert (stub).
//     Augment/Diminish are textual-duration multipliers (stub; TODO for future
//     work when a concrete duration notation is finalised).
//   - Cycles in based_on chains are detected and returned as an error.
func ResolveMotifs(entries []MotifEntry) (map[string]string, error) {
	// Build name lookup.
	byName := make(map[string]*MotifEntry, len(entries))
	for i := range entries {
		e := &entries[i]
		if e.Name == "" {
			return nil, fmt.Errorf("motif entry at index %d has no name", i)
		}
		if _, dup := byName[e.Name]; dup {
			return nil, fmt.Errorf("duplicate motif name %q", e.Name)
		}
		byName[e.Name] = e
	}

	resolved := make(map[string]string, len(entries))

	// DFS with cycle detection via a "visiting" set.
	visiting := make(map[string]bool)
	var resolve func(name string) (string, error)
	resolve = func(name string) (string, error) {
		if p, ok := resolved[name]; ok {
			return p, nil
		}
		e, ok := byName[name]
		if !ok {
			return "", fmt.Errorf("motif %q not found", name)
		}
		if visiting[name] {
			return "", fmt.Errorf("cycle detected in motif based_on chain involving %q", name)
		}
		visiting[name] = true

		pattern := e.Pattern
		if e.BasedOn != "" {
			parent, err := resolve(e.BasedOn)
			if err != nil {
				return "", err
			}
			pattern = parent
		}

		// Apply transforms.
		pattern = applyMotifTransforms(pattern, e)

		delete(visiting, name)
		resolved[name] = pattern
		return pattern, nil
	}

	for name := range byName {
		if _, err := resolve(name); err != nil {
			return nil, err
		}
	}
	return resolved, nil
}

// applyMotifTransforms applies the textual transforms from e to the given
// pattern string and returns the resulting pattern.
//
// Transform order: Retrograde → Transpose → Invert.
// Augment/Diminish are stubs (TODO).
func applyMotifTransforms(pattern string, e *MotifEntry) string {
	if e.Retrograde {
		pattern = retrogradePattern(pattern)
	}
	if e.Transpose != 0 {
		pattern = transposePattern(pattern, e.Transpose)
	}
	if e.Invert != 0 {
		// TODO: full inversion around a pivot scale degree is deferred until
		// the pattern notation is stable. Currently a no-op.
		_ = e.Invert
	}
	if e.Augment != 0 {
		// TODO: augment duration multiplier — deferred until duration notation
		// is finalised (the pattern string format does not yet encode absolute
		// durations in a way that makes multiplying them unambiguous).
		_ = e.Augment
	}
	if e.Diminish != 0 {
		// TODO: diminish duration multiplier — same deferral as Augment.
		_ = e.Diminish
	}
	return pattern
}

// retrogradePattern reverses the token order in a pattern string.
// Tokens are separated by whitespace; bar separators "|" are treated as tokens
// and reversed along with the rest.
func retrogradePattern(pattern string) string {
	tokens := strings.Fields(pattern)
	n := len(tokens)
	for i := 0; i < n/2; i++ {
		tokens[i], tokens[n-1-i] = tokens[n-1-i], tokens[i]
	}
	return strings.Join(tokens, " ")
}

// transposePattern shifts every numeric scale-degree token in the pattern by
// delta semitones. Non-numeric tokens (rests ".", bar markers "|", etc.) are
// left unchanged.
func transposePattern(pattern string, delta int) string {
	tokens := strings.Fields(pattern)
	for i, tok := range tokens {
		if n, err := strconv.Atoi(tok); err == nil {
			tokens[i] = strconv.Itoa(n + delta)
		}
	}
	return strings.Join(tokens, " ")
}

// ValidateNotePool checks that the weights in a NotePool sum to approximately
// 1.0 (within ±0.05). Returns a warning string if they do not, or "" if OK.
func ValidateNotePool(pool NotePool) string {
	if len(pool.Choices) == 0 {
		return ""
	}
	sum := 0.0
	for _, w := range pool.Choices {
		sum += w
	}
	if sum < 0.95 || sum > 1.05 {
		return fmt.Sprintf("note pool weights sum to %.3f, expected ~1.0", sum)
	}
	return ""
}

// ValidateChordMarkov checks that each row of a ChordMarkov table sums to
// approximately 1.0. Returns a slice of warning strings (empty = OK).
func ValidateChordMarkov(cm ChordMarkov) []string {
	var warns []string
	for state, row := range cm.Transitions {
		sum := 0.0
		for _, w := range row {
			sum += w
		}
		if sum < 0.95 || sum > 1.05 {
			warns = append(warns, fmt.Sprintf("chord_markov state %q weights sum to %.3f, expected ~1.0", state, sum))
		}
	}
	return warns
}

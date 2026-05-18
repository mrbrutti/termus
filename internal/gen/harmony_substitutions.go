package gen

// Package gen — harmonic substitution directives (SP7).
//
// ApplySubstitutions rewrites a chord progression according to a list of
// SubstitutionRules using a deterministic seed for probabilistic gates.
//
// Simplifications documented inline:
//   - Chord roots are parsed by a simple heuristic (up to two characters for
//     flats/sharps) rather than a full chord-symbol parser.
//   - Roman-numeral matching (ApplyTo / Of / Before) is compared
//     case-sensitively against the literal chord symbols in the progression.
//     Real diatonic analysis would require the key context, which is not
//     available in this compile-time rewriter.
//   - ii-V insertion uses hard-coded diatonic neighbours across all 12 keys
//     but always selects the tonic from a static lookup table.

import (
	"math/rand"
	"strings"
)

// SubstRule is the gen-package representation of a harmonic substitution
// directive (SP7). It mirrors track.SubstitutionRule but is defined here to
// avoid an import cycle (internal/gen must not import internal/track).
//
// Callers should convert track.SubstitutionRule → SubstRule at the
// compile/bridge layer.
type SubstRule struct {
	// Rule is one of: tritone_sub, ii_V_chain, secondary_dominant, deceptive.
	Rule string
	// ApplyTo constrains which chord role triggers the rule (e.g. "V", "I").
	ApplyTo string
	// Before is an optional anchor chord for ii_V_chain insertion.
	Before string
	// Of is the target chord for secondary_dominant (e.g. "ii").
	Of string
	// Probability is 0..1; when < 1 the rule is applied probabilistically.
	Probability float64
}

// ApplySubstitutions rewrites progression according to rules.
// seed is used to initialise the probabilistic gate so that the same
// input + seed always produces the same output.
func ApplySubstitutions(progression []string, rules []SubstRule, seed int64) []string {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec

	result := make([]string, len(progression))
	copy(result, progression)

	for _, rule := range rules {
		result = applyRule(result, rule, rng)
	}
	return result
}

func applyRule(prog []string, rule SubstRule, rng *rand.Rand) []string {
	switch rule.Rule {
	case "tritone_sub":
		return applyTritoneSubstitution(prog, rule, rng)
	case "ii_V_chain":
		return applyIIVChain(prog, rule, rng)
	case "secondary_dominant":
		return applySecondaryDominant(prog, rule, rng)
	case "deceptive":
		return applyDeceptiveCadence(prog, rule, rng)
	}
	return prog
}

// probabilityGate returns true if the rule should be applied for this
// occurrence. When Probability is 0 the rule is never applied; when 1 (or
// above) it is always applied.
//
// Note: a Probability of 0.0 (Go zero value) means "never apply" — authors
// must explicitly set a positive value. This matches the spec intent: every
// SubstRule should carry an explicit probability.
func probabilityGate(rule SubstRule, rng *rand.Rand) bool {
	if rule.Probability <= 0 {
		return false
	}
	if rule.Probability >= 1 {
		return true
	}
	return rng.Float64() < rule.Probability
}

// chromNoteNames maps pitch-class integers 0..11 to preferred flat labels.
var chromNoteNames = []string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}

// pitchClassOf returns the chromatic pitch class (0..11) for a note name
// string. Returns -1 if unrecognised.
func pitchClassOf(name string) int {
	name = strings.TrimSpace(name)
	switch strings.ToUpper(name) {
	case "C":
		return 0
	case "DB", "C#":
		return 1
	case "D":
		return 2
	case "EB", "D#":
		return 3
	case "E":
		return 4
	case "F":
		return 5
	case "GB", "F#":
		return 6
	case "G":
		return 7
	case "AB", "G#":
		return 8
	case "A":
		return 9
	case "BB", "A#":
		return 10
	case "B":
		return 11
	}
	return -1
}

// parseChordRoot splits a chord symbol into root (e.g. "G", "Db") and suffix
// (e.g. "7", "maj7", "m7").
//
// Simplification: handles up to two characters (letter + optional b/#).
func parseChordRoot(chord string) (root, suffix string) {
	if len(chord) == 0 {
		return "", ""
	}
	// Two-character root: letter + flat or sharp.
	if len(chord) >= 2 && (chord[1] == 'b' || chord[1] == '#') {
		return chord[:2], chord[2:]
	}
	return chord[:1], chord[1:]
}

// isDominant is a simple heuristic: chord ends with "7" or "9" (and is not
// major-seventh, i.e. not "maj7"). This covers dominant-seventh and ninth
// chords but not all cases. Documented simplification.
func isDominant(chord string) bool {
	_, suffix := parseChordRoot(chord)
	if suffix == "" {
		return false
	}
	suffix = strings.ToLower(suffix)
	// Reject major seventh.
	if strings.Contains(suffix, "maj") {
		return false
	}
	return strings.HasSuffix(suffix, "7") || strings.HasSuffix(suffix, "9")
}

// tritoneOf returns the tritone substitute of a dominant chord by adding 6
// semitones to the root. Retains the original suffix (e.g. G7 → Db7).
//
// Simplification: if the root is unrecognised, the original chord is returned
// unchanged.
func tritoneOf(chord string) string {
	root, suffix := parseChordRoot(chord)
	pc := pitchClassOf(root)
	if pc < 0 {
		return chord
	}
	newPC := (pc + 6) % 12
	return chromNoteNames[newPC] + suffix
}

// applyTritoneSubstitution replaces every dominant chord in the progression
// with its tritone substitute, subject to the probability gate.
//
// Simplification: all dominant-ish chords are candidates regardless of ApplyTo
// or the diatonic context. ApplyTo is compared against the chord symbol
// directly only as a filter when non-empty.
func applyTritoneSubstitution(prog []string, rule SubstRule, rng *rand.Rand) []string {
	out := make([]string, len(prog))
	copy(out, prog)
	for i, chord := range out {
		if !isDominant(chord) {
			continue
		}
		if rule.ApplyTo != "" && chord != rule.ApplyTo {
			continue
		}
		if !probabilityGate(rule, rng) {
			continue
		}
		out[i] = tritoneOf(chord)
	}
	return out
}

// diatonicSubstitutions maps a tonic chord to its ii and V.
//
// Simplification: mapping is static and covers common chord symbols. A full
// implementation would require the section key context.
var diatonicSubstitutions = map[string][2]string{
	// tonic → [ii, V]
	"Cmaj7":  {"Dm7", "G7"},
	"C":      {"Dm7", "G7"},
	"Fmaj7":  {"Gm7", "C7"},
	"F":      {"Gm7", "C7"},
	"Gmaj7":  {"Am7", "D7"},
	"G":      {"Am7", "D7"},
	"Amaj7":  {"Bm7", "E7"},
	"A":      {"Bm7", "E7"},
	"Dmaj7":  {"Em7", "A7"},
	"D":      {"Em7", "A7"},
	"Emaj7":  {"F#m7", "B7"},
	"E":      {"F#m7", "B7"},
	"Bbmaj7": {"Cm7", "F7"},
	"Bb":     {"Cm7", "F7"},
	"Ebmaj7": {"Fm7", "Bb7"},
	"Eb":     {"Fm7", "Bb7"},
	"Abmaj7": {"Bbm7", "Eb7"},
	"Ab":     {"Bbm7", "Eb7"},
	"Dbmaj7": {"Ebm7", "Ab7"},
	"Db":     {"Ebm7", "Ab7"},
	"Gbmaj7": {"Abm7", "Db7"},
	"Gb":     {"Abm7", "Db7"},
	"Bmaj7":  {"C#m7", "F#7"},
	"B":      {"C#m7", "F#7"},
}

// applyIIVChain inserts ii–V before the target chord (identified by the
// Before field). If a ii–V already precedes the target, insertion is skipped.
//
// Simplification: uses the static diatonicSubstitutions table. If the target
// chord is not in the table, no insertion occurs.
func applyIIVChain(prog []string, rule SubstRule, rng *rand.Rand) []string {
	target := rule.Before
	if target == "" {
		return prog
	}

	out := make([]string, 0, len(prog)+len(prog)/2)
	for i, chord := range prog {
		if chord == target {
			pair, known := diatonicSubstitutions[chord]
			if known && probabilityGate(rule, rng) {
				// Skip if ii–V already precedes.
				alreadyHas := false
				if i >= 2 && prog[i-2] == pair[0] && prog[i-1] == pair[1] {
					alreadyHas = true
				}
				if !alreadyHas {
					out = append(out, pair[0], pair[1])
				}
			}
		}
		out = append(out, chord)
	}
	return out
}

// secondaryDominantOf returns the dominant-seventh chord that resolves to the
// given target chord (its V7). We add 7 semitones to the target root (a fifth
// up = dominant).
//
// Simplification: always appends "7" as the suffix regardless of target type.
func secondaryDominantOf(target string) string {
	root, _ := parseChordRoot(target)
	pc := pitchClassOf(root)
	if pc < 0 {
		return ""
	}
	domPC := (pc + 7) % 12
	return chromNoteNames[domPC] + "7"
}

// applySecondaryDominant prepends the secondary dominant (V/Of) before every
// occurrence of the target chord (identified by the Of field).
//
// Simplification: ApplyTo is unused; Of identifies the target directly.
func applySecondaryDominant(prog []string, rule SubstRule, rng *rand.Rand) []string {
	target := rule.Of
	if target == "" {
		return prog
	}

	out := make([]string, 0, len(prog)+len(prog)/2)
	for _, chord := range prog {
		if chord == target {
			secDom := secondaryDominantOf(chord)
			if secDom != "" && probabilityGate(rule, rng) {
				out = append(out, secDom)
			}
		}
		out = append(out, chord)
	}
	return out
}

// deceptiveSubstitutes maps a tonic chord symbol to its relative-minor vi
// chord for deceptive cadences.
//
// Simplification: uses a static table rather than computing from key context.
var deceptiveSubstitutes = map[string]string{
	"Cmaj7":  "Am7",
	"C":      "Am7",
	"Fmaj7":  "Dm7",
	"F":      "Dm7",
	"Gmaj7":  "Em7",
	"G":      "Em7",
	"Amaj7":  "F#m7",
	"A":      "F#m7",
	"Dmaj7":  "Bm7",
	"D":      "Bm7",
	"Emaj7":  "C#m7",
	"E":      "C#m7",
	"Bbmaj7": "Gm7",
	"Bb":     "Gm7",
	"Ebmaj7": "Cm7",
	"Eb":     "Cm7",
	"Abmaj7": "Fm7",
	"Ab":     "Fm7",
	"Dbmaj7": "Bbm7",
	"Db":     "Bbm7",
	"Gbmaj7": "Ebm7",
	"Gb":     "Ebm7",
	"Bmaj7":  "G#m7",
	"B":      "G#m7",
}

// applyDeceptiveCadence scans for V–I pairs in the progression and replaces I
// with vi, subject to the probability gate.
//
// Simplification: V is identified as any isDominant chord immediately preceding
// a tonic chord in the deceptiveSubstitutes table. The rule's ApplyTo field
// is not checked (the dominant is structural here, not user-specified).
func applyDeceptiveCadence(prog []string, rule SubstRule, rng *rand.Rand) []string {
	out := make([]string, len(prog))
	copy(out, prog)

	for i := 1; i < len(out); i++ {
		prev := out[i-1]
		curr := out[i]
		vi, knownTonic := deceptiveSubstitutes[curr]
		if !knownTonic {
			continue
		}
		if !isDominant(prev) {
			continue
		}
		if !probabilityGate(rule, rng) {
			continue
		}
		out[i] = vi
	}
	return out
}

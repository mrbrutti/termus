package acestep

import (
	"fmt"
	"strings"
	"time"

	"github.com/mrbrutti/termus/internal/track"
)

// CompileV3 translates a parsed track.File with RenderEngine="acestep" into
// the RenderSpec ready for the Python service.
//
// The mapping is intentionally one-pass and total; nothing about the source
// File is mutated. Required fields produce a RenderSpec with sensible
// defaults; missing optional fields are left at the zero value.
//
// Behaviour summary:
//
//   - The natural-language style paragraph + tags are joined into Prompt.
//   - file.Key and file.Tempo populate Key and Tempo.
//   - Duration prefers file.TotalDuration (e.g. "3m"); otherwise sums each
//     AcestepSection's Bars × (4 beats × beat-duration-at-bpm).
//   - Scale: AcestepSpec.Scale if set, else inferred from file.Key
//     (presence of "min" → minor; otherwise major).
//   - TimeSignature: AcestepSpec.TimeSignature if set, else "4/4".
//   - Seed: AcestepSpec.SeedOverride wins; otherwise file.Seed.
//   - Sections: per-section Description (+ Dynamic) become
//     SectionDescriptions; Harmony strings concatenate into HarmonyChain.
//   - Each section description is also appended to the prompt so the model
//     sees per-section guidance in-line.
//   - Motif and InferenceSteps pass through unchanged.
//
// Returns an error if the file's RenderEngine is not "acestep" or if the
// Acestep block is missing.
func CompileV3(file *track.File) (RenderSpec, error) {
	if file == nil {
		return RenderSpec{}, fmt.Errorf("acestep: CompileV3 called with nil file")
	}
	if file.RenderEngine != track.RenderEngineACEStep {
		return RenderSpec{}, fmt.Errorf("acestep: CompileV3 requires render_engine=%q, got %q", track.RenderEngineACEStep, file.RenderEngine)
	}
	if file.Acestep == nil {
		return RenderSpec{}, fmt.Errorf("acestep: CompileV3 requires the 'acestep:' block to be set")
	}
	spec := file.Acestep

	// Combine style paragraph with tags, in order, on a single line. The
	// model's "caption" embedding tokenises both equally; rank-ordering the
	// tags after the prose keeps the genre signal near the start of the
	// joined string.
	prompt := strings.TrimSpace(spec.Style)
	if len(spec.Tags) > 0 {
		tagPart := strings.Join(spec.Tags, ", ")
		if prompt == "" {
			prompt = tagPart
		} else {
			prompt = prompt + ". Tags: " + tagPart
		}
	}

	// Per-section descriptions: collect for SectionDescriptions and inline
	// onto the end of the prompt so callers without rich per-section
	// support still see the per-section guidance.
	sectionDescs := make([]string, 0, len(spec.Sections))
	for _, sec := range spec.Sections {
		desc := strings.TrimSpace(sec.Description)
		if sec.Dynamic != "" {
			if desc == "" {
				desc = "(dynamic: " + sec.Dynamic + ")"
			} else {
				desc = desc + " (dynamic: " + sec.Dynamic + ")"
			}
		}
		if desc == "" {
			continue
		}
		sectionDescs = append(sectionDescs, desc)
	}
	if len(sectionDescs) > 0 {
		prompt = prompt + ". Sections: " + strings.Join(sectionDescs, "; ")
	}

	// Harmony chain: concatenate every section's Harmony left-to-right.
	// Empty Harmony fields are skipped.
	harmonyParts := make([]string, 0, len(spec.Sections))
	for _, sec := range spec.Sections {
		h := strings.TrimSpace(sec.Harmony)
		if h == "" {
			continue
		}
		harmonyParts = append(harmonyParts, h)
	}
	harmonyChain := strings.Join(harmonyParts, " ")

	// Scale: explicit override, else infer.
	scale := strings.TrimSpace(spec.Scale)
	if scale == "" {
		scale = inferScale(file.Key)
	}

	// TimeSignature default.
	timesig := strings.TrimSpace(spec.TimeSignature)
	if timesig == "" {
		timesig = "4/4"
	}

	// Seed: AcestepSpec.SeedOverride wins; otherwise file.Seed.
	seed := file.Seed
	if spec.SeedOverride != nil {
		seed = *spec.SeedOverride
	}

	// Tempo: parse the string ("86", "86 bpm") → int.
	tempoBPM := parseTempo(file.Tempo)

	// Duration:
	//   1. file.TotalDuration ("3m", "12m30s") via time.ParseDuration.
	//   2. Sum of section bars × bar duration at tempo.
	//   3. 0 (let the model choose).
	durationSeconds := 0.0
	if td := strings.TrimSpace(file.TotalDuration); td != "" {
		if d, err := time.ParseDuration(td); err == nil {
			durationSeconds = d.Seconds()
		}
	}
	if durationSeconds <= 0 {
		durationSeconds = sumSectionBarsSeconds(spec.Sections, tempoBPM, timesig)
	}

	out := RenderSpec{
		Prompt:              strings.TrimSpace(prompt),
		Tags:                append([]string(nil), spec.Tags...),
		Key:                 strings.TrimSpace(file.Key),
		Tempo:               tempoBPM,
		DurationSeconds:     durationSeconds,
		Scale:               scale,
		TimeSignature:       timesig,
		Seed:                seed,
		SectionDescriptions: sectionDescs,
		HarmonyChain:        harmonyChain,
		Motif:               strings.TrimSpace(spec.Motif),
		InferenceSteps:      spec.InferenceSteps,
	}
	// ReferenceAudio on the .tm side is a path; the wire side is a base64
	// blob. The CLI is responsible for the file → b64 read step; the
	// compiler leaves the field empty.
	return out, nil
}

// inferScale returns "minor" when the key string contains "min" or starts
// with a lowercase root with a trailing "m" (e.g. "Am", "Dm7"); otherwise
// "major". Empty input → empty string.
func inferScale(key string) string {
	k := strings.TrimSpace(key)
	if k == "" {
		return ""
	}
	low := strings.ToLower(k)
	if strings.Contains(low, "min") {
		return "minor"
	}
	if strings.Contains(low, "maj") {
		return "major"
	}
	// "Am", "Dm", "Bm7" → minor heuristic: tonic letter then "m" then
	// optionally a digit / suffix that does not begin with "aj".
	if len(k) >= 2 {
		// First char is the tonic letter; second char is "m" with no
		// "aj" following.
		if (k[1] == 'm' || k[1] == 'M') && (len(k) < 4 || !strings.HasPrefix(low[1:], "maj")) {
			if k[1] == 'm' {
				return "minor"
			}
		}
	}
	return "major"
}

// parseTempo reuses the same lenient parser shape as the rest of the
// codebase: "86", "86 bpm", "86.0" all → 86. Returns 0 on failure.
func parseTempo(raw string) int {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0
	}
	for _, p := range strings.Fields(s) {
		var n float64
		if _, err := fmt.Sscanf(p, "%f", &n); err == nil && n > 0 {
			return int(n + 0.5)
		}
	}
	return 0
}

// sumSectionBarsSeconds adds up each section's Bars × bar duration. Returns
// 0 when no section has Bars > 0 or when tempo is unknown.
func sumSectionBarsSeconds(sections []track.AcestepSection, bpm int, timesig string) float64 {
	if bpm <= 0 || len(sections) == 0 {
		return 0
	}
	beatsPerBar := beatsPerBarFromTimeSig(timesig)
	beatSeconds := 60.0 / float64(bpm)
	total := 0.0
	for _, s := range sections {
		if s.Bars > 0 {
			total += float64(s.Bars) * beatsPerBar * beatSeconds
		}
	}
	return total
}

// beatsPerBarFromTimeSig parses "N/D" → N. Returns 4 on unrecognised input.
func beatsPerBarFromTimeSig(ts string) float64 {
	parts := strings.Split(strings.TrimSpace(ts), "/")
	if len(parts) != 2 {
		return 4
	}
	var n int
	if _, err := fmt.Sscanf(parts[0], "%d", &n); err == nil && n > 0 {
		return float64(n)
	}
	return 4
}

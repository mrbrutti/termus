package track

import (
	"math"
	"sort"
	"strings"

	"github.com/mrbrutti/termus/internal/gen"
)

type ReviewMetrics struct {
	RepetitionFatigue      float64 `json:"repetition_fatigue"`
	SectionContrast        float64 `json:"section_contrast"`
	CadenceSpacing         float64 `json:"cadence_spacing"`
	LeadOccupancy          float64 `json:"lead_occupancy"`
	RegisterSpread         float64 `json:"register_spread"`
	HarmonicColorRetention float64 `json:"harmonic_color_retention"`
	EnsembleDiversity      float64 `json:"ensemble_diversity"`
}

type ReviewSection struct {
	ID        string        `json:"id,omitempty"`
	Title     string        `json:"title,omitempty"`
	Duration  string        `json:"duration"`
	Harmony   string        `json:"harmony,omitempty"`
	Scene     string        `json:"scene,omitempty"`
	Variation string        `json:"variation,omitempty"`
	RoleNames []string      `json:"roles,omitempty"`
	Events    []string      `json:"events,omitempty"`
	Metrics   ReviewMetrics `json:"metrics"`
}

type ReviewReport struct {
	Title      string          `json:"title"`
	Style      string          `json:"style"`
	Substyle   string          `json:"substyle,omitempty"`
	ListenMode string          `json:"listen_mode,omitempty"`
	Metrics    ReviewMetrics   `json:"metrics"`
	Sections   []ReviewSection `json:"sections"`
	Warnings   []Warning       `json:"warnings,omitempty"`
}

func Analyze(file *File, compiled *Compiled) ReviewReport {
	report := ReviewReport{}
	if file == nil {
		return report
	}
	report.Title = file.Title
	report.Style = file.Style
	pack := resolveStylePack(file.Style, file.Substyle, file.Title, file.Tags)
	report.Substyle = pack.Substyle
	report.ListenMode = file.ListenMode
	if compiled != nil {
		report.Warnings = append(report.Warnings, compiled.Warnings...)
	}

	sections, err := resolveSections(file)
	if err != nil {
		sections = append([]Section(nil), file.Sections...)
	}
	report.Sections = make([]ReviewSection, 0, len(sections))
	var planList []gen.AuthoredTrackPlan
	if compiled != nil {
		for _, track := range compiled.Playlist.Tracks {
			if plan, ok := compiled.Plans[playlistKey(track.Spec, track.Seed)]; ok {
				planList = append(planList, plan)
			}
		}
	}
	for idx, section := range sections {
		roles := resolvedSectionRoles(file, section)
		section, roles = applyStyleLibrary(pack, section, roles)
		sectionReport := ReviewSection{
			ID:        section.ID,
			Title:     firstNonBlank(section.Title, section.ID),
			Duration:  section.Duration,
			Harmony:   section.Harmony,
			Scene:     section.Scene,
			Variation: section.Variation,
			RoleNames: sortedActiveRoleNames(roles),
			Events:    reviewEventLabels(sectionEvents(section)),
		}
		if idx < len(planList) {
			sectionReport.Metrics = analyzePlan(section, planList[idx])
		}
		report.Sections = append(report.Sections, sectionReport)
	}
	report.Metrics = combineReviewMetrics(sections, planList, file)
	return report
}

func analyzePlan(section Section, plan gen.AuthoredTrackPlan) ReviewMetrics {
	return ReviewMetrics{
		RepetitionFatigue:      roundMetric(planRepetitionFatigue(plan)),
		LeadOccupancy:          roundMetric(planLeadOccupancy(plan)),
		RegisterSpread:         roundMetric(planRegisterSpread(plan)),
		HarmonicColorRetention: roundMetric(harmonicColorRetention(section.Harmony, plan.ChordSpans)),
		EnsembleDiversity:      roundMetric(planEnsembleDiversity(plan)),
	}
}

func combineReviewMetrics(sections []Section, plans []gen.AuthoredTrackPlan, file *File) ReviewMetrics {
	var out ReviewMetrics
	if len(plans) == 0 {
		return out
	}
	var (
		totalBars     int
		repSum        float64
		leadSum       float64
		spreadSum     float64
		harmonySum    float64
		diversitySum  float64
		contrastSum   float64
		contrastCount int
	)
	for idx, plan := range plans {
		weight := maxInt(1, plan.BarCount)
		totalBars += weight
		repSum += planRepetitionFatigue(plan) * float64(weight)
		leadSum += planLeadOccupancy(plan) * float64(weight)
		spreadSum += planRegisterSpread(plan) * float64(weight)
		diversitySum += planEnsembleDiversity(plan) * float64(weight)
		if idx < len(sections) {
			harmonySum += harmonicColorRetention(sections[idx].Harmony, plan.ChordSpans) * float64(weight)
		}
	}
	if totalBars > 0 {
		out.RepetitionFatigue = roundMetric(repSum / float64(totalBars))
		out.LeadOccupancy = roundMetric(leadSum / float64(totalBars))
		out.RegisterSpread = roundMetric(spreadSum / float64(totalBars))
		out.HarmonicColorRetention = roundMetric(harmonySum / float64(totalBars))
		out.EnsembleDiversity = roundMetric(diversitySum / float64(totalBars))
	}
	if len(sections) > 1 {
		for i := 1; i < len(sections); i++ {
			contrastSum += 1 - sectionSimilarity(file, sections[i-1], sections[i])
			contrastCount++
		}
	}
	if contrastCount > 0 {
		out.SectionContrast = roundMetric(contrastSum / float64(contrastCount))
	}
	out.CadenceSpacing = roundMetric(cadenceSectionPosition(sections))
	return out
}

func cadenceSectionPosition(sections []Section) float64 {
	last := -1
	for i, section := range sections {
		if sectionLooksCadential(section) {
			last = i
		}
	}
	if last < 0 || len(sections) == 0 {
		return 0
	}
	return clampMetric(float64(last+1) / float64(len(sections)))
}

func sectionLooksCadential(section Section) bool {
	text := strings.ToLower(strings.TrimSpace(strings.Join([]string{section.ID, section.Title, section.Scene, section.Variation}, " ")))
	if strings.Contains(text, "cadence") || strings.Contains(text, "outro") || strings.Contains(text, "release") {
		return true
	}
	for _, event := range sectionEvents(section) {
		switch strings.ToLower(strings.TrimSpace(event.Kind)) {
		case "ending", "tag", "hold", "silence":
			return true
		}
	}
	return false
}

func planRepetitionFatigue(plan gen.AuthoredTrackPlan) float64 {
	var total float64
	var count int
	for _, track := range plan.Tracks {
		if authoredRoleKind(track.Name, Role{Family: track.Family}) == "drum" {
			continue
		}
		nonRest := 0
		repeated := 0
		last := math.MinInt
		for _, note := range track.Notes {
			if note < 0 {
				continue
			}
			nonRest++
			if note == last {
				repeated++
			}
			last = note
		}
		if nonRest > 1 {
			total += float64(repeated) / float64(nonRest-1)
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return clampMetric(total / float64(count))
}

func planLeadOccupancy(plan gen.AuthoredTrackPlan) float64 {
	var melodySlots, activeSlots int
	for _, track := range plan.Tracks {
		if authoredRoleKind(track.Name, Role{Family: track.Family}) != "melody" {
			continue
		}
		for _, note := range track.Notes {
			activeSlots++
			if note >= 0 {
				melodySlots++
			}
		}
	}
	if activeSlots == 0 {
		return 0
	}
	return clampMetric(float64(melodySlots) / float64(activeSlots))
}

func planRegisterSpread(plan gen.AuthoredTrackPlan) float64 {
	minNote := math.MaxInt
	maxNote := math.MinInt
	for _, track := range plan.Tracks {
		if authoredRoleKind(track.Name, Role{Family: track.Family}) == "drum" {
			continue
		}
		for _, note := range track.Notes {
			if note < 0 {
				continue
			}
			if note < minNote {
				minNote = note
			}
			if note > maxNote {
				maxNote = note
			}
		}
	}
	if minNote == math.MaxInt || maxNote <= minNote {
		return 0
	}
	return clampMetric(float64(maxNote-minNote) / 36.0)
}

func harmonicColorRetention(src string, spans []gen.AuthoredChordSpan) float64 {
	sourceColor := colorfulChordCount(strings.Fields(strings.ReplaceAll(strings.TrimSpace(src), "|", " ")))
	if sourceColor == 0 {
		return 1
	}
	planLabels := make([]string, 0, len(spans))
	for _, span := range spans {
		planLabels = append(planLabels, span.Label)
	}
	planColor := colorfulChordCount(planLabels)
	return clampMetric(float64(minInt(sourceColor, planColor)) / float64(sourceColor))
}

func colorfulChordCount(labels []string) int {
	count := 0
	for _, label := range labels {
		lower := strings.ToLower(strings.TrimSpace(label))
		if lower == "" {
			continue
		}
		if strings.Contains(lower, "9") || strings.Contains(lower, "11") || strings.Contains(lower, "13") ||
			strings.Contains(lower, "sus") || strings.Contains(lower, "add") || strings.Contains(lower, "/") ||
			strings.Contains(lower, "b5") || strings.Contains(lower, "#11") || strings.Contains(lower, "maj7") {
			count++
		}
	}
	return count
}

func planEnsembleDiversity(plan gen.AuthoredTrackPlan) float64 {
	families := map[string]bool{}
	registers := map[string]bool{}
	for _, track := range plan.Tracks {
		if authoredRoleKind(track.Name, Role{Family: track.Family}) == "drum" {
			continue
		}
		if family := strings.ToLower(strings.TrimSpace(track.Family)); family != "" {
			families[family] = true
		}
		if register := strings.ToLower(strings.TrimSpace(track.Register)); register != "" {
			registers[register] = true
		}
	}
	if len(families) == 0 {
		return 0
	}
	score := (float64(len(families))/6.0 + float64(len(registers))/4.0) * 0.5
	return clampMetric(score)
}

func sortedActiveRoleNames(roles map[string]Role) []string {
	names := make([]string, 0, len(roles))
	for name, role := range roles {
		if role.Active != nil && !*role.Active {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func reviewEventLabels(events []Event) []string {
	out := make([]string, 0, len(events))
	for _, event := range events {
		label := strings.ToLower(strings.TrimSpace(event.Kind))
		if label == "" {
			continue
		}
		out = append(out, label)
	}
	sort.Strings(out)
	return out
}

func roundMetric(v float64) float64 {
	return math.Round(clampMetric(v)*1000) / 1000
}

func clampMetric(v float64) float64 {
	switch {
	case v < 0:
		return 0
	case v > 1:
		return 1
	default:
		return v
	}
}

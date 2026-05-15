package gen

import "math/rand"

type FormSectionKind string

const (
	FormIntro     FormSectionKind = "intro"
	FormA         FormSectionKind = "A"
	FormAprime    FormSectionKind = "A'"
	FormB         FormSectionKind = "B"
	FormBreakdown FormSectionKind = "breakdown"
	FormCadence   FormSectionKind = "cadence"
	FormOutro     FormSectionKind = "outro"
)

// FormSection describes one higher-level arrangement block above the per-bar
// note loops. It intentionally carries only coarse musical intent so
// algorithms can interpret it differently.
type FormSection struct {
	Kind            FormSectionKind
	Bars            int
	LeadLevel       int
	TextureLevel    int
	RhythmLevel     int
	CadenceStrength int
	RegisterLift    int
}

type FormPlan struct {
	barSamples int64
	sections   []FormSection
	totalBars  int
}

type ListeningMarker struct {
	Label  string `json:"label"`
	Sample int64  `json:"sample"`
}

type ListeningInspectable interface {
	ListeningMarkers() []ListeningMarker
}

func NewFormPlan(rng *rand.Rand, barSamples int64, profile string) FormPlan {
	sections := make([]FormSection, 0, 7)
	switch profile {
	case "jazz":
		sections = []FormSection{
			{Kind: FormIntro, Bars: 4, LeadLevel: 0, TextureLevel: 1, RhythmLevel: 1, CadenceStrength: 0},
			{Kind: FormA, Bars: 8, LeadLevel: 1, TextureLevel: 1, RhythmLevel: 1, CadenceStrength: 0},
			{Kind: FormAprime, Bars: 8, LeadLevel: 1, TextureLevel: 2, RhythmLevel: 1, CadenceStrength: 1, RegisterLift: 2},
			{Kind: FormB, Bars: 8, LeadLevel: 2, TextureLevel: 2, RhythmLevel: 2, CadenceStrength: 1, RegisterLift: 3},
			{Kind: FormBreakdown, Bars: 4, LeadLevel: 0, TextureLevel: 1, RhythmLevel: 0, CadenceStrength: 0},
			{Kind: FormCadence, Bars: 4, LeadLevel: 2, TextureLevel: 2, RhythmLevel: 2, CadenceStrength: 2, RegisterLift: 3},
			{Kind: FormOutro, Bars: 4, LeadLevel: 0, TextureLevel: 1, RhythmLevel: 0, CadenceStrength: 2},
		}
	case "classical":
		sections = []FormSection{
			{Kind: FormIntro, Bars: 4, LeadLevel: 1, TextureLevel: 1, RhythmLevel: 0, CadenceStrength: 0},
			{Kind: FormA, Bars: 8, LeadLevel: 1, TextureLevel: 1, RhythmLevel: 0, CadenceStrength: 0},
			{Kind: FormAprime, Bars: 8, LeadLevel: 2, TextureLevel: 2, RhythmLevel: 0, CadenceStrength: 1, RegisterLift: 2},
			{Kind: FormB, Bars: 8, LeadLevel: 2, TextureLevel: 2, RhythmLevel: 0, CadenceStrength: 1, RegisterLift: 3},
			{Kind: FormCadence, Bars: 4, LeadLevel: 2, TextureLevel: 2, RhythmLevel: 0, CadenceStrength: 2, RegisterLift: 2},
			{Kind: FormOutro, Bars: 4, LeadLevel: 0, TextureLevel: 1, RhythmLevel: 0, CadenceStrength: 2},
		}
	default: // lofi / chill
		sections = []FormSection{
			{Kind: FormIntro, Bars: 8, LeadLevel: 0, TextureLevel: 1, RhythmLevel: 1, CadenceStrength: 0},
			{Kind: FormA, Bars: 8, LeadLevel: 1, TextureLevel: 1, RhythmLevel: 1, CadenceStrength: 0},
			{Kind: FormAprime, Bars: 8, LeadLevel: 1, TextureLevel: 2, RhythmLevel: 1, CadenceStrength: 1, RegisterLift: 2},
			{Kind: FormB, Bars: 8, LeadLevel: 2, TextureLevel: 2, RhythmLevel: 2, CadenceStrength: 1, RegisterLift: 3},
			{Kind: FormBreakdown, Bars: 4, LeadLevel: 0, TextureLevel: 1, RhythmLevel: 0, CadenceStrength: 0},
			{Kind: FormCadence, Bars: 4, LeadLevel: 2, TextureLevel: 2, RhythmLevel: 2, CadenceStrength: 2, RegisterLift: 2},
			{Kind: FormOutro, Bars: 4, LeadLevel: 0, TextureLevel: 1, RhythmLevel: 0, CadenceStrength: 2},
		}
	}
	if rng != nil && len(sections) > 0 {
		// Small duration variation keeps the form from feeling like a rigid
		// trainer loop while preserving bar-aligned boundaries.
		for i := range sections {
			if sections[i].Kind == FormA || sections[i].Kind == FormAprime || sections[i].Kind == FormB {
				if rng.Float64() < 0.35 {
					sections[i].Bars += 4
				}
			}
		}
	}
	totalBars := 0
	for _, section := range sections {
		totalBars += section.Bars
	}
	return FormPlan{
		barSamples: barSamples,
		sections:   sections,
		totalBars:  totalBars,
	}
}

func (f FormPlan) SectionAt(samples int64) FormSection {
	if len(f.sections) == 0 || f.barSamples <= 0 || f.totalBars <= 0 {
		return FormSection{}
	}
	bar := sampleBarIndex(samples, f.barSamples) % f.totalBars
	acc := 0
	for _, section := range f.sections {
		acc += section.Bars
		if bar < acc {
			return section
		}
	}
	return f.sections[len(f.sections)-1]
}

func (f FormPlan) SectionBoundaryCrossed(prev, curr int64) bool {
	if len(f.sections) == 0 || f.barSamples <= 0 || f.totalBars <= 0 {
		return false
	}
	prevBar := sampleBarIndex(prev, f.barSamples) % f.totalBars
	currBar := sampleBarIndex(curr, f.barSamples) % f.totalBars
	if currBar == prevBar {
		return false
	}
	acc := 0
	for _, section := range f.sections {
		if acc == currBar {
			return true
		}
		acc += section.Bars
	}
	return currBar == 0
}

func (f FormPlan) TotalBars() int { return f.totalBars }

func (f FormPlan) BarAt(samples int64) int {
	if len(f.sections) == 0 || f.barSamples <= 0 || f.totalBars <= 0 {
		return 0
	}
	return sampleBarIndex(samples, f.barSamples)%f.totalBars + 1
}

func (f FormPlan) ListeningMarkers(cycles int) []ListeningMarker {
	if cycles < 1 {
		cycles = 1
	}
	if len(f.sections) == 0 || f.barSamples <= 0 {
		return nil
	}
	markers := []ListeningMarker{{Label: "bar:0", Sample: 0}}
	sample := int64(0)
	bar := 0
	for cycle := 0; cycle < cycles; cycle++ {
		for _, section := range f.sections {
			markers = append(markers, ListeningMarker{
				Label:  "section:" + string(section.Kind),
				Sample: sample,
			})
			if section.CadenceStrength > 0 {
				markers = append(markers, ListeningMarker{
					Label:  "cadence:" + string(section.Kind),
					Sample: sample,
				})
			}
			for i := 0; i < section.Bars; i++ {
				bar++
				sample += f.barSamples
				markers = append(markers, ListeningMarker{
					Label:  "bar",
					Sample: sample,
				})
			}
		}
	}
	return markers
}

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

type EpisodeMovement string

const (
	MovementEstablish EpisodeMovement = "establish"
	MovementDevelop   EpisodeMovement = "develop"
	MovementBreathe   EpisodeMovement = "breathe"
	MovementLift      EpisodeMovement = "lift"
	MovementReturn    EpisodeMovement = "return"
)

type FormEpisode struct {
	Movement  EpisodeMovement
	Sections  []FormSection
	StartBar  int
	TotalBars int
}

type EpisodePlan struct {
	barSamples int64
	profile    string
	rng        *rand.Rand
	episodes   []FormEpisode
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
	sections := planSections(rng, profile, MovementEstablish)
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

func NewEpisodePlan(rng *rand.Rand, barSamples int64, profile string) EpisodePlan {
	plan := EpisodePlan{
		barSamples: barSamples,
		profile:    profile,
		rng:        rng,
	}
	plan.ensureBars(1)
	return plan
}

func planSections(rng *rand.Rand, profile string, movement EpisodeMovement) []FormSection {
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
	default:
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
		for i := range sections {
			if sections[i].Kind == FormA || sections[i].Kind == FormAprime || sections[i].Kind == FormB {
				if rng.Float64() < 0.35 {
					sections[i].Bars += 4
				}
			}
		}
	}
	applyMovementContour(sections, movement)
	if movement != MovementEstablish && len(sections) > 0 && sections[0].Kind == FormIntro {
		sections = append([]FormSection(nil), sections[1:]...)
	}
	return sections
}

func applyMovementContour(sections []FormSection, movement EpisodeMovement) {
	switch movement {
	case MovementDevelop:
		for i := range sections {
			sections[i].TextureLevel++
			if sections[i].Kind == FormB || sections[i].Kind == FormCadence {
				sections[i].LeadLevel++
			}
		}
	case MovementBreathe:
		for i := range sections {
			if sections[i].Kind == FormA || sections[i].Kind == FormAprime {
				sections[i].RhythmLevel = maxInt(0, sections[i].RhythmLevel-1)
			}
			if sections[i].Kind == FormBreakdown || sections[i].Kind == FormOutro {
				sections[i].TextureLevel = maxInt(0, sections[i].TextureLevel-1)
			}
		}
	case MovementLift:
		for i := range sections {
			sections[i].RegisterLift += 1
			if sections[i].Kind != FormIntro && sections[i].Kind != FormBreakdown {
				sections[i].LeadLevel++
			}
		}
	case MovementReturn:
		for i := range sections {
			if sections[i].Kind == FormOutro {
				sections[i].TextureLevel++
			}
			if sections[i].Kind == FormCadence {
				sections[i].CadenceStrength++
			}
		}
	}
}

func (p *EpisodePlan) ensureBars(bar int) {
	if p == nil {
		return
	}
	for bar >= p.totalBars {
		movement := p.nextMovement(len(p.episodes))
		sections := planSections(p.rng, p.profile, movement)
		episode := FormEpisode{
			Movement: movement,
			Sections: sections,
			StartBar: p.totalBars,
		}
		for _, section := range sections {
			episode.TotalBars += section.Bars
		}
		p.totalBars += episode.TotalBars
		p.episodes = append(p.episodes, episode)
	}
}

func (p *EpisodePlan) nextMovement(idx int) EpisodeMovement {
	order := []EpisodeMovement{
		MovementEstablish,
		MovementDevelop,
		MovementBreathe,
		MovementLift,
		MovementReturn,
	}
	if p.rng != nil && idx > 0 && p.rng.Float64() < 0.25 {
		return order[p.rng.Intn(len(order))]
	}
	return order[idx%len(order)]
}

func (p *EpisodePlan) locateEpisode(bar int) (FormEpisode, int) {
	p.ensureBars(bar)
	for i := range p.episodes {
		ep := p.episodes[i]
		if bar >= ep.StartBar && bar < ep.StartBar+ep.TotalBars {
			return ep, i
		}
	}
	if len(p.episodes) == 0 {
		return FormEpisode{}, 0
	}
	return p.episodes[len(p.episodes)-1], len(p.episodes) - 1
}

func (p *EpisodePlan) SectionAt(samples int64) FormSection {
	if p == nil || p.barSamples <= 0 {
		return FormSection{}
	}
	bar := sampleBarIndex(samples, p.barSamples)
	ep, _ := p.locateEpisode(bar)
	relative := bar - ep.StartBar
	acc := 0
	for _, section := range ep.Sections {
		acc += section.Bars
		if relative < acc {
			return section
		}
	}
	if len(ep.Sections) == 0 {
		return FormSection{}
	}
	return ep.Sections[len(ep.Sections)-1]
}

func (p *EpisodePlan) SectionBoundaryCrossed(prev, curr int64) bool {
	if p == nil || p.barSamples <= 0 {
		return false
	}
	prevBar := sampleBarIndex(prev, p.barSamples)
	currBar := sampleBarIndex(curr, p.barSamples)
	if currBar == prevBar {
		return false
	}
	ep, _ := p.locateEpisode(currBar)
	relative := currBar - ep.StartBar
	if relative == 0 {
		return true
	}
	acc := 0
	for _, section := range ep.Sections {
		if acc == relative {
			return true
		}
		acc += section.Bars
	}
	return false
}

func (p *EpisodePlan) EpisodeBoundaryCrossed(prev, curr int64) bool {
	if p == nil || p.barSamples <= 0 {
		return false
	}
	prevBar := sampleBarIndex(prev, p.barSamples)
	currBar := sampleBarIndex(curr, p.barSamples)
	if currBar == prevBar {
		return false
	}
	prevEp, prevIdx := p.locateEpisode(prevBar)
	currEp, currIdx := p.locateEpisode(currBar)
	return currIdx != prevIdx || currEp.StartBar != prevEp.StartBar
}

func (p *EpisodePlan) BarAt(samples int64) int {
	if p == nil || p.barSamples <= 0 {
		return 0
	}
	bar := sampleBarIndex(samples, p.barSamples)
	ep, _ := p.locateEpisode(bar)
	return bar - ep.StartBar + 1
}

func (p *EpisodePlan) MovementAt(samples int64) EpisodeMovement {
	if p == nil || p.barSamples <= 0 {
		return MovementEstablish
	}
	bar := sampleBarIndex(samples, p.barSamples)
	ep, _ := p.locateEpisode(bar)
	return ep.Movement
}

func (p *EpisodePlan) ListeningMarkers(episodes int) []ListeningMarker {
	if p == nil || p.barSamples <= 0 {
		return nil
	}
	if episodes < 1 {
		episodes = 1
	}
	targetBars := 0
	for i := 0; i < episodes; i++ {
		p.ensureBars(targetBars)
		if i < len(p.episodes) {
			targetBars = p.episodes[i].StartBar + p.episodes[i].TotalBars
		}
	}
	markers := []ListeningMarker{{Label: "bar:0", Sample: 0}}
	for i := 0; i < episodes && i < len(p.episodes); i++ {
		ep := p.episodes[i]
		sample := int64(ep.StartBar) * p.barSamples
		markers = append(markers, ListeningMarker{
			Label:  "movement:" + string(ep.Movement),
			Sample: sample,
		})
		offset := 0
		for _, section := range ep.Sections {
			sectionSample := sample + int64(offset)*p.barSamples
			markers = append(markers, ListeningMarker{
				Label:  "section:" + string(section.Kind),
				Sample: sectionSample,
			})
			if section.CadenceStrength > 0 {
				markers = append(markers, ListeningMarker{
					Label:  "cadence:" + string(section.Kind),
					Sample: sectionSample,
				})
			}
			offset += section.Bars
		}
	}
	return markers
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

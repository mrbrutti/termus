package gen

func sectionTemplate(kind FormSectionKind) FormSection {
	switch kind {
	case FormIntro:
		return FormSection{Kind: kind, LeadLevel: 0, TextureLevel: 1, RhythmLevel: 0, CadenceStrength: 0}
	case FormA:
		return FormSection{Kind: kind, LeadLevel: 1, TextureLevel: 1, RhythmLevel: 0, CadenceStrength: 0}
	case FormAprime:
		return FormSection{Kind: kind, LeadLevel: 1, TextureLevel: 2, RhythmLevel: 0, CadenceStrength: 1, RegisterLift: 1}
	case FormB:
		return FormSection{Kind: kind, LeadLevel: 2, TextureLevel: 2, RhythmLevel: 0, CadenceStrength: 1, RegisterLift: 2}
	case FormBreakdown:
		return FormSection{Kind: kind, LeadLevel: 0, TextureLevel: 1, RhythmLevel: 0, CadenceStrength: 0}
	case FormCadence:
		return FormSection{Kind: kind, LeadLevel: 2, TextureLevel: 2, RhythmLevel: 0, CadenceStrength: 2, RegisterLift: 1}
	case FormOutro:
		return FormSection{Kind: kind, LeadLevel: 0, TextureLevel: 1, RhythmLevel: 0, CadenceStrength: 2}
	default:
		return FormSection{Kind: kind}
	}
}

func textureSectionForLayers(primaryOn, answerOn, cadence bool) FormSection {
	switch {
	case cadence && primaryOn && answerOn:
		return sectionTemplate(FormCadence)
	case primaryOn && answerOn:
		return sectionTemplate(FormAprime)
	case primaryOn:
		return sectionTemplate(FormA)
	case answerOn:
		return sectionTemplate(FormB)
	default:
		return sectionTemplate(FormBreakdown)
	}
}

func cycleTextureSection(idx, total int) FormSection {
	if total <= 1 {
		return sectionTemplate(FormA)
	}
	switch {
	case idx <= 0:
		return sectionTemplate(FormIntro)
	case idx >= total-1:
		return sectionTemplate(FormCadence)
	case idx == total-2:
		return sectionTemplate(FormB)
	case idx%2 == 0:
		return sectionTemplate(FormAprime)
	default:
		return sectionTemplate(FormA)
	}
}

func waltzTextureSection(bar, total int, ornament bool) FormSection {
	if total <= 0 {
		return sectionTemplate(FormA)
	}
	switch {
	case bar == 0:
		return sectionTemplate(FormIntro)
	case bar == total-1:
		return sectionTemplate(FormCadence)
	case bar >= total-2:
		return sectionTemplate(FormB)
	case ornament:
		return sectionTemplate(FormAprime)
	default:
		return sectionTemplate(FormA)
	}
}

package gen

import "testing"

func TestSectionSceneForDifferentiatesRoles(t *testing.T) {
	lead := SectionSceneFor(FormSection{Kind: FormCadence}, RoleLead)
	bass := SectionSceneFor(FormSection{Kind: FormCadence}, RoleBass)
	drums := SectionSceneFor(FormSection{Kind: FormBreakdown}, RoleDrums)
	texture := SectionSceneFor(FormSection{Kind: FormB}, RoleTexture)

	if lead.ExpressionDelta <= bass.ExpressionDelta {
		t.Fatalf("lead expression = %d, bass = %d; want lead > bass", lead.ExpressionDelta, bass.ExpressionDelta)
	}
	if bass.ReverbDelta >= lead.ReverbDelta {
		t.Fatalf("bass reverb = %d, lead = %d; want bass < lead", bass.ReverbDelta, lead.ReverbDelta)
	}
	if drums.BrightnessDelta < 0 {
		t.Fatalf("drum brightness = %d, want non-negative presence lift", drums.BrightnessDelta)
	}
	if texture.ReverbDelta <= 0 {
		t.Fatalf("texture reverb = %d, want positive halo", texture.ReverbDelta)
	}
}

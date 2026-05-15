package gen

type InstrumentRole string

const (
	RoleLead    InstrumentRole = "lead"
	RoleComp    InstrumentRole = "comp"
	RoleBass    InstrumentRole = "bass"
	RoleDrums   InstrumentRole = "drums"
	RoleTexture InstrumentRole = "texture"
)

type SectionScene struct {
	ExpressionDelta int32
	BrightnessDelta int32
	ReverbDelta     int32
}

// SectionSceneFor tailors a coarse form section into an instrument-specific
// scene. Lead and texture layers move more aggressively than bass and drums,
// so the arrangement can breathe without every channel changing in lockstep.
func SectionSceneFor(section FormSection, role InstrumentRole) SectionScene {
	base := SectionMixProfileFor(section)
	scene := SectionScene{
		ExpressionDelta: base.ExpressionDelta,
		BrightnessDelta: base.BrightnessDelta,
		ReverbDelta:     base.ReverbDelta,
	}
	switch role {
	case RoleLead:
		scene.ExpressionDelta += 6
		scene.BrightnessDelta += 4
		scene.ReverbDelta += 8
	case RoleComp:
		scene.ExpressionDelta /= 2
		scene.BrightnessDelta /= 2
		scene.ReverbDelta /= 3
	case RoleBass:
		scene.ExpressionDelta /= 3
		scene.BrightnessDelta = scene.BrightnessDelta/3 - 2
		scene.ReverbDelta = scene.ReverbDelta/4 - 4
	case RoleDrums:
		scene.ExpressionDelta = scene.ExpressionDelta / 4
		scene.BrightnessDelta = scene.BrightnessDelta/2 + 2
		if scene.BrightnessDelta < 0 {
			scene.BrightnessDelta = 0
		}
		scene.ReverbDelta = scene.ReverbDelta/3 - 2
	case RoleTexture:
		scene.ExpressionDelta = scene.ExpressionDelta/2 + 3
		scene.BrightnessDelta = scene.BrightnessDelta/2 + 2
		scene.ReverbDelta += 10
	}
	return scene
}

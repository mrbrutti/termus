package gen

// TimeLayerWindow describes how a render step crossed the higher-level musical
// layers above per-note scheduling. The note layer lives inside sf2_engine;
// this helper separates bar, section, and episode logic so long-form changes
// only happen on the right boundaries.
type TimeLayerWindow struct {
	PrevSamples    int64
	CurrSamples    int64
	PrevBar        int
	CurrBar        int
	BarChanged     bool
	SectionChanged bool
	EpisodeChanged bool
	Section        FormSection
	Movement       EpisodeMovement
	EpisodeIndex   int
}

func ComputeTimeLayerWindow(plan *EpisodePlan, prev, curr int64) TimeLayerWindow {
	window := TimeLayerWindow{
		PrevSamples: prev,
		CurrSamples: curr,
	}
	if plan == nil || plan.barSamples <= 0 {
		return window
	}
	window.PrevBar = sampleBarIndex(prev, plan.barSamples)
	window.CurrBar = sampleBarIndex(curr, plan.barSamples)
	window.BarChanged = window.CurrBar != window.PrevBar
	window.SectionChanged = plan.SectionBoundaryCrossed(prev, curr)
	window.EpisodeChanged = plan.EpisodeBoundaryCrossed(prev, curr)
	window.Section = plan.SectionAt(curr)
	window.Movement = plan.MovementAt(curr)
	window.EpisodeIndex = plan.EpisodeIndexAt(curr)
	return window
}

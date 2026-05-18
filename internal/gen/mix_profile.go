package gen

// SectionGainProvider lets an algorithm expose a dynamic section-level gain
// scalar for the current musical form block.
type SectionGainProvider interface {
	SectionGain() float64
}

// SectionMixProfile is a coarse, shared section-energy profile that
// algorithms can interpret with their own instrument-specific base values.
type SectionMixProfile struct {
	Gain            float64
	ExpressionDelta int32
	BrightnessDelta int32
	ReverbDelta     int32
}

// EffectiveOutputGain resolves the static per-algorithm trim and any dynamic
// section-level gain into one scalar used by both live playback and offline
// rendering.
func EffectiveOutputGain(algo Algorithm) float64 {
	if algo == nil {
		return 1.0
	}
	gain := defaultOutputTrim(algo.Name())
	if provider, ok := algo.(SectionGainProvider); ok {
		gain *= clampGain(provider.SectionGain())
	}
	return clampGain(gain)
}

// SectionMixProfileFor returns the shared musical contour for a form section.
// Intros and outros are intentionally softer and darker, while B and cadence
// sections get more projection and space.
func SectionMixProfileFor(section FormSection) SectionMixProfile {
	switch section.Kind {
	case FormIntro:
		return SectionMixProfile{Gain: 0.88, ExpressionDelta: -12, BrightnessDelta: -8, ReverbDelta: 6}
	case FormAprime:
		return SectionMixProfile{Gain: 1.02, ExpressionDelta: 4, BrightnessDelta: 2, ReverbDelta: 3}
	case FormB:
		return SectionMixProfile{Gain: 1.06, ExpressionDelta: 10, BrightnessDelta: 8, ReverbDelta: 10}
	case FormBreakdown:
		return SectionMixProfile{Gain: 0.82, ExpressionDelta: -18, BrightnessDelta: -12, ReverbDelta: -6}
	case FormCadence:
		return SectionMixProfile{Gain: 1.10, ExpressionDelta: 14, BrightnessDelta: 10, ReverbDelta: 14}
	case FormOutro:
		return SectionMixProfile{Gain: 0.78, ExpressionDelta: -22, BrightnessDelta: -12, ReverbDelta: 4}
	default:
		return SectionMixProfile{Gain: 1.0}
	}
}

// SectionCC applies a signed section delta to a MIDI CC-style 0..127 value.
func SectionCC(base, delta int32) int32 {
	value := base + delta
	switch {
	case value < 0:
		return 0
	case value > 127:
		return 127
	default:
		return value
	}
}

func clampGain(gain float64) float64 {
	switch {
	case gain < 0.25:
		return 0.25
	case gain > 2.0:
		return 2.0
	default:
		return gain
	}
}

func defaultOutputTrim(name string) float64 {
	switch name {
	case "ambient", "eno-drift":
		if name == "eno-drift" {
			return 0.75
		}
		return 1.00
	case "drone":
		return 1.15
	case "drone-bed":
		return 0.95
	case "bells":
		return 1.15
	case "glass-fm":
		return 0.85
	case "lullaby":
		return 1.00
	case "pentatonic-walk":
		return 0.80
	case "classical":
		return 0.95
	case "markov-melody":
		return 0.80
	case "phase":
		return 1.15
	case "chill":
		return 1.00
	case "jazz":
		return 1.20
	default:
		return 1.0
	}
}

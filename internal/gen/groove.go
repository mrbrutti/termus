package gen

// cyclicTimingOffset converts a millisecond timing pattern into a reusable
// per-slot offset callback. Negative values push ahead of the grid; positive
// values lay the slot behind it.
func cyclicTimingOffset(patternMS ...float64) func(int) float64 {
	if len(patternMS) == 0 {
		return nil
	}
	return func(slot int) float64 {
		idx := slot % len(patternMS)
		if idx < 0 {
			idx += len(patternMS)
		}
		return patternMS[idx] / 1000.0
	}
}

func jazzPlanTimingOffset(code int) float64 {
	switch code {
	case jazzPlanApproachAbove, jazzPlanApproachBelow, jazzPlanAnticipateNextRoot:
		return -0.014
	case jazzPlanResolveThird, jazzPlanRoot:
		return 0.008
	case jazzPlanSuspendFourth:
		return 0.004
	default:
		return 0.0
	}
}

func jazzSaxTiming(codeAt func(int) int) func(int) float64 {
	if codeAt == nil {
		return nil
	}
	return func(slot int) float64 {
		return jazzPlanTimingOffset(codeAt(slot))
	}
}

func chillPlanTimingOffset(code int) float64 {
	switch code {
	case chillPlanPickupAbove, chillPlanPickupBelow:
		return -0.016
	case chillPlanResolveThird, chillPlanRoot:
		return 0.010
	case chillPlanSuspendFourth:
		return 0.005
	default:
		return 0.0
	}
}

func chillLeadTiming(codeAt func(int) int) func(int) float64 {
	if codeAt == nil {
		return nil
	}
	return func(slot int) float64 {
		return chillPlanTimingOffset(codeAt(slot))
	}
}

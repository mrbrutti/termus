package track

// ValueAt returns the linearly interpolated value of the lane at the given
// progress fraction (0..1 maps to 0%..100% of the section duration).
//
// Behaviour:
//   - No breakpoints → returns 0.
//   - Single breakpoint → returns that breakpoint's value at all progress.
//   - progress < 0 → clamps to the first breakpoint's value.
//   - progress > 1 → clamps to the last breakpoint's value.
//   - Between two breakpoints → linear interpolation.
func (l AutomationLane) ValueAt(progress01 float64) float64 {
	bps := l.Breakpoints
	if len(bps) == 0 {
		return 0
	}
	if len(bps) == 1 {
		return bps[0].Value
	}

	// Convert progress01 (0..1) to percent (0..100) for comparison.
	pct := progress01 * 100.0

	// Clamp to ends.
	if pct <= bps[0].AtPercent {
		return bps[0].Value
	}
	if pct >= bps[len(bps)-1].AtPercent {
		return bps[len(bps)-1].Value
	}

	// Find the surrounding pair.
	for i := 1; i < len(bps); i++ {
		lo := bps[i-1]
		hi := bps[i]
		if pct >= lo.AtPercent && pct <= hi.AtPercent {
			span := hi.AtPercent - lo.AtPercent
			if span == 0 {
				return lo.Value
			}
			t := (pct - lo.AtPercent) / span
			return lo.Value + t*(hi.Value-lo.Value)
		}
	}

	// Unreachable if breakpoints are sorted, but be safe.
	return bps[len(bps)-1].Value
}

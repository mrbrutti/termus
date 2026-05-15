package gen

func swellCurve(start, peak, end int32, peakAt float64) SF2ExpressionCurve {
	return SF2ExpressionCurve{Start: start, Peak: peak, End: end, PeakAt01: peakAt}
}

func gentleVibratoCurve(start, peak, end int32) SF2ExpressionCurve {
	return swellCurve(start, peak, end, 0.45)
}

func brightnessBloomCurve(start, peak, end int32) SF2ExpressionCurve {
	return swellCurve(start, peak, end, 0.18)
}

func slotDetunePattern(pattern ...int32) func(int, int) int32 {
	if len(pattern) == 0 {
		pattern = []int32{0}
	}
	return func(slot int, _ int) int32 {
		idx := ((slot % len(pattern)) + len(pattern)) % len(pattern)
		return pattern[idx]
	}
}

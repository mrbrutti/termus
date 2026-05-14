package gen

import "math"

// pow2Impl is split into its own file purely so eno.go reads cleanly.
func pow2Impl(x float64) float64 { return math.Exp2(x) }

package gen

// GrooveTemplate describes a named per-16th-note timing and velocity recipe.
// TimingOffsetsSamples cycles per bar at 4/4 (index 0 = downbeat of bar).
// Positive values = laid back (later); negative = pushed (earlier).
// VelocityMultipliers scale each slot's velocity; 1.0 = unchanged.
type GrooveTemplate struct {
	Name string
	// Per-16th-note timing offsets in samples (positive = later). Length 16.
	TimingOffsetsSamples [16]int
	// Per-16th velocity multipliers. 1.0 = unchanged. Length 16.
	VelocityMultipliers [16]float64
}

// GrooveLibrary returns the 4 built-in named groove templates.
func GrooveLibrary() []GrooveTemplate {
	flat := func() [16]float64 {
		var v [16]float64
		for i := range v {
			v[i] = 1.0
		}
		return v
	}

	return []GrooveTemplate{
		{
			// straight: perfectly on-grid, no offsets, flat velocity.
			Name:                 "straight",
			TimingOffsetsSamples: [16]int{},
			VelocityMultipliers:  flat(),
		},
		{
			// swing_56: typical lofi MPC swing at ~54%.
			// Odd 16ths (index 1,3,5,7,9,11,13,15) are pushed ~6 samples late,
			// with a slight velocity accent on the downbeats.
			Name: "swing_56",
			TimingOffsetsSamples: [16]int{
				0, 6, 0, 6,
				0, 6, 0, 6,
				0, 6, 0, 6,
				0, 6, 0, 6,
			},
			VelocityMultipliers: [16]float64{
				1.05, 0.90, 1.00, 0.90,
				1.05, 0.90, 1.00, 0.90,
				1.05, 0.90, 1.00, 0.90,
				1.05, 0.90, 1.00, 0.90,
			},
		},
		{
			// dilla_late: J Dilla MPC feel — kick slightly early (index 0 = -3),
			// snare slightly late (index 4/12 = +8), hats stay straight.
			Name: "dilla_late",
			TimingOffsetsSamples: [16]int{
				-3, 0, 0, 0, // beat 1: kick early
				8, 0, 0, 0,  // beat 2: snare late
				-3, 0, 0, 0, // beat 3: kick early
				8, 0, 0, 0,  // beat 4: snare late
			},
			VelocityMultipliers: [16]float64{
				1.10, 0.85, 0.90, 0.80,
				1.05, 0.85, 0.90, 0.80,
				1.10, 0.85, 0.90, 0.80,
				1.05, 0.85, 0.90, 0.80,
			},
		},
		{
			// bossa_loose: bossa nova feel — alternate 16ths pushed 3-5 samples
			// ahead for an anticipating, forward-leaning character.
			Name: "bossa_loose",
			TimingOffsetsSamples: [16]int{
				0, -3, 0, -5,
				0, -3, 0, -4,
				0, -3, 0, -5,
				0, -4, 0, -3,
			},
			VelocityMultipliers: [16]float64{
				1.00, 0.95, 1.05, 0.90,
				1.00, 0.95, 1.05, 0.90,
				1.00, 0.95, 1.05, 0.90,
				1.00, 0.95, 1.05, 0.90,
			},
		},
	}
}

// GrooveByName resolves a groove template by name. Returns nil if not found.
func GrooveByName(name string) *GrooveTemplate {
	lib := GrooveLibrary()
	for i := range lib {
		if lib[i].Name == name {
			return &lib[i]
		}
	}
	return nil
}

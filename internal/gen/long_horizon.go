package gen

import "math/rand"

type LongHorizonState struct {
	Episode       int
	Movement      EpisodeMovement
	HarmonyFamily string
	MotifFamily   string
	TextureScene  string
	DensityBias   int
	RegisterBias  int
}

func NewLongHorizonState(rng *rand.Rand, profile string, movement EpisodeMovement) LongHorizonState {
	state := LongHorizonState{
		Episode:       0,
		Movement:      movement,
		HarmonyFamily: chooseProfileFamily(rng, profile, harmonyFamilies(profile)),
		MotifFamily:   chooseProfileFamily(rng, profile, motifFamilies(profile)),
		TextureScene:  chooseProfileFamily(rng, profile, textureScenes(profile)),
		DensityBias:   chooseBias(rng),
		RegisterBias:  chooseBias(rng),
	}
	applyMovementBias(&state, profile)
	return state
}

func AdvanceLongHorizonState(rng *rand.Rand, prev LongHorizonState, profile string, movement EpisodeMovement) LongHorizonState {
	next := prev
	next.Episode++
	next.Movement = movement
	if shouldShift(rng, 0.72) {
		next.HarmonyFamily = chooseDifferentFamily(rng, prev.HarmonyFamily, harmonyFamilies(profile))
	}
	if shouldShift(rng, 0.68) {
		next.MotifFamily = chooseDifferentFamily(rng, prev.MotifFamily, motifFamilies(profile))
	}
	if shouldShift(rng, 0.64) {
		next.TextureScene = chooseDifferentFamily(rng, prev.TextureScene, textureScenes(profile))
	}
	if shouldShift(rng, 0.58) {
		next.DensityBias = chooseBias(rng)
	}
	if shouldShift(rng, 0.52) {
		next.RegisterBias = chooseBias(rng)
	}
	applyMovementBias(&next, profile)
	return next
}

func applyMovementBias(state *LongHorizonState, profile string) {
	if state == nil {
		return
	}
	state.HarmonyFamily = movementHarmonyFamily(profile, state.Movement, state.HarmonyFamily)
	state.MotifFamily = movementMotifFamily(profile, state.Movement, state.MotifFamily)
	state.TextureScene = movementTextureScene(profile, state.Movement, state.TextureScene)
	switch state.Movement {
	case MovementDevelop:
		if state.DensityBias < 0 {
			state.DensityBias = 0
		}
	case MovementBreathe:
		state.DensityBias = -1
		if state.RegisterBias > 0 {
			state.RegisterBias = 0
		}
	case MovementLift:
		state.DensityBias = 1
		state.RegisterBias = 1
	case MovementReturn:
		state.DensityBias = 0
		state.RegisterBias = 0
	default:
		if state.DensityBias > 0 {
			state.DensityBias = 0
		}
		if state.RegisterBias > 0 {
			state.RegisterBias = 0
		}
	}
}

func chooseProfileFamily(rng *rand.Rand, profile string, values []string) string {
	if len(values) == 0 {
		return profile
	}
	if rng == nil {
		return values[0]
	}
	return values[rng.Intn(len(values))]
}

func chooseDifferentFamily(rng *rand.Rand, current string, values []string) string {
	if len(values) == 0 {
		return current
	}
	if len(values) == 1 {
		return values[0]
	}
	if rng == nil {
		for _, value := range values {
			if value != current {
				return value
			}
		}
		return values[0]
	}
	for tries := 0; tries < 6; tries++ {
		candidate := values[rng.Intn(len(values))]
		if candidate != current {
			return candidate
		}
	}
	return values[(indexOf(values, current)+1)%len(values)]
}

func shouldShift(rng *rand.Rand, p float64) bool {
	if rng == nil {
		return false
	}
	return rng.Float64() < p
}

func chooseBias(rng *rand.Rand) int {
	options := []int{-1, 0, 1}
	if rng == nil {
		return 0
	}
	return options[rng.Intn(len(options))]
}

func indexOf(values []string, current string) int {
	for i, value := range values {
		if value == current {
			return i
		}
	}
	return 0
}

func harmonyFamilies(profile string) []string {
	switch profile {
	case "jazz":
		return []string{"ii-v-cycle", "modal-minor", "dominant-chain", "turnaround"}
	case "classical":
		return []string{"period", "answer", "subdominant-arc", "cadential-return"}
	default:
		return []string{"warm-major", "minor-haze", "borrowed-loop", "modal-wander"}
	}
}

func motifFamilies(profile string) []string {
	switch profile {
	case "jazz":
		return []string{"guide-tone", "pickup-line", "arched-answer", "late-resolution"}
	case "classical":
		return []string{"triadic", "stepwise", "sentence", "answering"}
	default:
		return []string{"chime", "sigh", "answer", "hover"}
	}
}

func textureScenes(profile string) []string {
	switch profile {
	case "jazz":
		return []string{"combo", "horn-forward", "piano-open", "brushes-low"}
	case "classical":
		return []string{"chamber", "strings-open", "winds-answer", "cadence-bright"}
	default:
		return []string{"dusty", "wet-night", "narrow-room", "haze"}
	}
}

func movementHarmonyFamily(profile string, movement EpisodeMovement, fallback string) string {
	switch profile {
	case "jazz":
		switch movement {
		case MovementDevelop:
			return "dominant-chain"
		case MovementBreathe:
			return "modal-minor"
		case MovementLift:
			return "turnaround"
		case MovementReturn, MovementEstablish:
			return "ii-v-cycle"
		}
	case "classical":
		switch movement {
		case MovementDevelop:
			return "answer"
		case MovementBreathe:
			return "subdominant-arc"
		case MovementLift:
			return "cadential-return"
		case MovementReturn, MovementEstablish:
			return "period"
		}
	default:
		switch movement {
		case MovementDevelop:
			return "borrowed-loop"
		case MovementBreathe:
			return "minor-haze"
		case MovementLift:
			return "modal-wander"
		case MovementReturn, MovementEstablish:
			return "warm-major"
		}
	}
	return fallback
}

func movementMotifFamily(profile string, movement EpisodeMovement, fallback string) string {
	switch profile {
	case "jazz":
		switch movement {
		case MovementDevelop:
			return "arched-answer"
		case MovementBreathe:
			return "guide-tone"
		case MovementLift:
			return "pickup-line"
		case MovementReturn, MovementEstablish:
			return "late-resolution"
		}
	case "classical":
		switch movement {
		case MovementDevelop:
			return "sentence"
		case MovementBreathe:
			return "stepwise"
		case MovementLift:
			return "triadic"
		case MovementReturn, MovementEstablish:
			return "answering"
		}
	default:
		switch movement {
		case MovementDevelop:
			return "answer"
		case MovementBreathe:
			return "hover"
		case MovementLift:
			return "chime"
		case MovementReturn, MovementEstablish:
			return "sigh"
		}
	}
	return fallback
}

func movementTextureScene(profile string, movement EpisodeMovement, fallback string) string {
	switch profile {
	case "jazz":
		switch movement {
		case MovementDevelop:
			return "combo"
		case MovementBreathe:
			return "brushes-low"
		case MovementLift:
			return "horn-forward"
		case MovementReturn, MovementEstablish:
			return "piano-open"
		}
	case "classical":
		switch movement {
		case MovementDevelop:
			return "winds-answer"
		case MovementBreathe:
			return "chamber"
		case MovementLift:
			return "cadence-bright"
		case MovementReturn, MovementEstablish:
			return "strings-open"
		}
	default:
		switch movement {
		case MovementDevelop:
			return "wet-night"
		case MovementBreathe:
			return "narrow-room"
		case MovementLift:
			return "haze"
		case MovementReturn, MovementEstablish:
			return "dusty"
		}
	}
	return fallback
}

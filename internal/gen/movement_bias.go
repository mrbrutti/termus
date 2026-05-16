package gen

func movementArrangementBias(movement EpisodeMovement) (densityDelta int32, brightnessDelta int32) {
	switch movement {
	case MovementDevelop:
		return 4, 3
	case MovementBreathe:
		return -6, -5
	case MovementLift:
		return 8, 8
	case MovementReturn:
		return 0, -2
	default:
		return -2, -1
	}
}

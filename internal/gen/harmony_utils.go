package gen

func wrapPitchClass(v int) int {
	v %= 12
	if v < 0 {
		v += 12
	}
	return v
}

func pitchClassLabel(pc int) string {
	switch wrapPitchClass(pc) {
	case 0:
		return "C"
	case 1:
		return "Db"
	case 2:
		return "D"
	case 3:
		return "Eb"
	case 4:
		return "E"
	case 5:
		return "F"
	case 6:
		return "Gb"
	case 7:
		return "G"
	case 8:
		return "Ab"
	case 9:
		return "A"
	case 10:
		return "Bb"
	default:
		return "B"
	}
}

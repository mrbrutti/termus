package gen

type genreDensityPolicy struct {
	LeadFillCount         int
	AccentBonus           int
	SecondaryTextureFloor int
	TextureExpressionBias int32
}

func densityPolicyFor(name string, profile ControlProfile) genreDensityPolicy {
	profile = profileOrDefault(profile)
	centered := ProfileCentered(profile.Density)
	switch name {
	case "jazz":
		return genreDensityPolicy{
			LeadFillCount: 2 + maxInt(0, centered),
			AccentBonus:   1 + maxInt(0, centered),
		}
	case "lofi":
		return genreDensityPolicy{
			LeadFillCount: 1 + maxInt(0, centered),
		}
	case "ambient":
		return genreDensityPolicy{
			SecondaryTextureFloor: 3,
			TextureExpressionBias: -6,
		}
	case "bells":
		return genreDensityPolicy{
			SecondaryTextureFloor: 3,
			TextureExpressionBias: -8,
		}
	case "drone":
		return genreDensityPolicy{
			TextureExpressionBias: -6,
		}
	default:
		return genreDensityPolicy{}
	}
}

func densifyPhrase(src []int, rest int, replacements []int, fills int) []int {
	if len(src) == 0 || fills <= 0 || len(replacements) == 0 {
		return copyPhrase(src)
	}
	out := copyPhrase(src)
	rests := make([]int, 0, len(src))
	for i, v := range out {
		if v == rest {
			rests = append(rests, i)
		}
	}
	if len(rests) == 0 {
		return out
	}
	if fills > len(rests) {
		fills = len(rests)
	}
	for i := 0; i < fills; i++ {
		restIdx := rests[(i*len(rests))/fills]
		out[restIdx] = replacements[i%len(replacements)]
	}
	return out
}


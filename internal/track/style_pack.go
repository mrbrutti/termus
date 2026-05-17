package track

import (
	"sort"
	"strings"
)

type styleVariant struct {
	Name            string
	Keywords        []string
	DefaultBPM      float64
	DefaultRhythms  map[string]string
	DefaultMelody   map[string]string
	DrumLeadSurface string
	SoftLowEnd      *bool
	ExtendedComp    *bool
}

type stylePack struct {
	Name            string
	Substyle        string
	DefaultBPM      float64
	ShortPhraseBars int
	LongPhraseBars  int
	DefaultRhythms  map[string]string
	DefaultMelody   map[string]string
	DrumLeadSurface string
	SoftLowEnd      bool
	ExtendedComp    bool
	Substyles       map[string]styleVariant
}

var stylePacks = map[string]stylePack{
	"ambient": stylePackSpec(
		58, 2, 4, "none", true, false,
		map[string]string{"pad": "x... ....", "bass": "x... ....", "bells": "x... ...."},
		map[string]string{"lead": "5 . 3 . | 1 . . ."},
		map[string]styleVariant{
			"station-haze": {Name: "station-haze", Keywords: []string{"subway", "platform", "haze", "drift"}, DefaultMelody: map[string]string{"lead": "5 . 3 . | 1 . . ."}},
			"choir-fog":    {Name: "choir-fog", Keywords: []string{"choir", "fog", "mist"}, DefaultRhythms: map[string]string{"pad": "x... ....", "bells": ".... x..."}, DefaultMelody: map[string]string{"lead": "5 . . . | 3 . 1 ."}},
		},
	),
	"bells": stylePackSpec(
		54, 2, 2, "none", true, false,
		map[string]string{"bells": "x... ....", "celesta": "x... ....", "pad": "x... ...."},
		map[string]string{"bells": "5 . . 7 | 9 . 7 5", "lead": "5 . . 7 | 9 . 7 5"},
		map[string]styleVariant{
			"vespers-glass": {Name: "vespers-glass", Keywords: []string{"vespers", "glass", "chapel"}, DefaultMelody: map[string]string{"lead": "5 . . 7 | 9 . 7 5"}},
			"cloister-rain": {Name: "cloister-rain", Keywords: []string{"cloister", "rain", "chapel"}, DefaultRhythms: map[string]string{"celesta": ".... x..."}, DefaultMelody: map[string]string{"lead": "5 . 6 . | 7 . 5 ."}},
		},
	),
	"classical": stylePackSpec(
		92, 2, 4, "none", false, true,
		map[string]string{"piano": "x..x .x..", "strings": "x... ...."},
		map[string]string{"lead": "5 . 3 . | 1 . . ."},
		map[string]styleVariant{
			"nocturne-room":   {Name: "nocturne-room", Keywords: []string{"nocturne", "room", "lantern"}, DefaultRhythms: map[string]string{"piano": "x..x ...."}},
			"chamber-lantern": {Name: "chamber-lantern", Keywords: []string{"chamber", "loop", "lantern"}, DefaultRhythms: map[string]string{"strings": "x... x..."}, DefaultMelody: map[string]string{"lead": "5 . 4 . | 3 . 1 ."}},
		},
	),
	"drone": stylePackSpec(
		46, 2, 4, "none", true, false,
		map[string]string{"bed": "x... ....", "bass": "x... ...."},
		map[string]string{"lead": "5 . 3 . | 1 . . ."},
		map[string]styleVariant{
			"soft-static":   {Name: "soft-static", Keywords: []string{"static", "soft"}, DefaultRhythms: map[string]string{"bed": "x... ...."}},
			"cathedral-bed": {Name: "cathedral-bed", Keywords: []string{"cathedral", "hymn", "field"}, DefaultMelody: map[string]string{"lead": "5 . . . | 1 . . ."}},
		},
	),
	"jazz": stylePackSpec(
		126, 2, 4, "ride", false, true,
		map[string]string{"kick": "x... x...", "snare": ".... x...", "hat": "x.x.x.x.", "ride": "x.x. x.x.", "bass": "x... x...", "comp": "x..x .x..", "piano": "x..x .x.."},
		map[string]string{"lead": "5 . 6 7 | 9 . 7 3"},
		map[string]styleVariant{
			"trio-after-hours": {Name: "trio-after-hours", Keywords: []string{"after", "hours", "blue"}, DrumLeadSurface: "ride", DefaultBPM: 124},
			"organ-combo":      {Name: "organ-combo", Keywords: []string{"organ", "turnaround", "red-eye"}, DefaultBPM: 118, DrumLeadSurface: "hat", DefaultRhythms: map[string]string{"comp": "x... x...", "organ": "x... x..."}},
			"vibes-cellar":     {Name: "vibes-cellar", Keywords: []string{"vibes", "cellar", "basement"}, DefaultBPM: 132, DefaultMelody: map[string]string{"lead": "5 . 7 9 | 6 . 5 3"}},
		},
	),
	"lofi": stylePackSpec(
		78, 2, 4, "hat", false, true,
		map[string]string{"kick": "x... x...", "snare": ".... x...", "hat": "x.x.x.x.", "bass": "x... x...", "keys": "x..x .x..", "guitar": "x..x .x.."},
		map[string]string{"lead": "5 . . 7 | 9 . 7 5"},
		map[string]styleVariant{
			"dusty-rhodes":   {Name: "dusty-rhodes", Keywords: []string{"rain", "bus", "soft", "tape"}, DefaultBPM: 76, DrumLeadSurface: "hat"},
			"vibes-nocturne": {Name: "vibes-nocturne", Keywords: []string{"library", "vent", "vibes"}, DefaultBPM: 72, DefaultRhythms: map[string]string{"keys": "x... .x.."}, DefaultMelody: map[string]string{"lead": "5 . 6 . | 7 . 5 ."}},
			"guitar-neon":    {Name: "guitar-neon", Keywords: []string{"walkman", "streetlights", "neon", "window"}, DefaultBPM: 84, DefaultRhythms: map[string]string{"guitar": "x..x ..x."}, DefaultMelody: map[string]string{"lead": "5 . 7 . | 9 . 5 ."}},
		},
	),
	"lullaby": stylePackSpec(
		68, 2, 2, "none", true, false,
		map[string]string{"lead": "x... ....", "harp": "x..x ....", "pad": "x... ...."},
		map[string]string{"lead": "5 . 3 . | 1 . . ."},
		map[string]styleVariant{
			"paper-box":      {Name: "paper-box", Keywords: []string{"paper", "moon", "box"}, DefaultMelody: map[string]string{"lead": "5 . 3 . | 2 . 1 ."}},
			"staircase-song": {Name: "staircase-song", Keywords: []string{"staircase", "lullaby"}, DefaultRhythms: map[string]string{"harp": "x... x..."}, DefaultMelody: map[string]string{"lead": "5 . . 3 | 2 . 1 ."}},
		},
	),
	"phase": stylePackSpec(
		74, 2, 2, "none", true, false,
		map[string]string{"mallet-a": "x... x...", "mallet-b": ".... x...", "pad": "x... ...."},
		map[string]string{"lead": "5 . 3 . | 1 . . ."},
		map[string]styleVariant{
			"glass-steps":    {Name: "glass-steps", Keywords: []string{"glass", "steps", "mirror"}, DefaultRhythms: map[string]string{"mallet-b": "x... .x.."}},
			"warm-interlock": {Name: "warm-interlock", Keywords: []string{"warm", "interlock"}, DefaultRhythms: map[string]string{"pad": "x... x..."}, DefaultMelody: map[string]string{"lead": "5 . 4 . | 3 . 1 ."}},
		},
	),
}

func stylePackSpec(bpm float64, shortBars, longBars int, drumLeadSurface string, softLowEnd, extendedComp bool, rhythms, melody map[string]string, substyles map[string]styleVariant) stylePack {
	return stylePack{
		DefaultBPM:      bpm,
		ShortPhraseBars: shortBars,
		LongPhraseBars:  longBars,
		DefaultRhythms:  rhythms,
		DefaultMelody:   melody,
		DrumLeadSurface: drumLeadSurface,
		SoftLowEnd:      softLowEnd,
		ExtendedComp:    extendedComp,
		Substyles:       substyles,
	}
}

func stylePackFor(style string) stylePack {
	pack, ok := stylePacks[strings.ToLower(strings.TrimSpace(style))]
	if ok {
		pack.Name = strings.ToLower(strings.TrimSpace(style))
		return pack
	}
	return stylePack{
		Name:            style,
		DefaultBPM:      80,
		ShortPhraseBars: 2,
		LongPhraseBars:  4,
		DefaultRhythms:  map[string]string{},
		DefaultMelody:   map[string]string{},
		DrumLeadSurface: "hat",
		Substyles:       map[string]styleVariant{},
	}
}

func resolveStylePack(style, explicitSubstyle, title string, tags []string) stylePack {
	pack := stylePackFor(style)
	substyle := strings.ToLower(strings.TrimSpace(explicitSubstyle))
	if substyle == "" {
		substyle = inferSubstyle(pack, title, tags)
	}
	if variant, ok := pack.Substyles[substyle]; ok {
		pack.Substyle = variant.Name
		if variant.DefaultBPM > 0 {
			pack.DefaultBPM = variant.DefaultBPM
		}
		pack.DefaultRhythms = mergedStringMap(pack.DefaultRhythms, variant.DefaultRhythms)
		pack.DefaultMelody = mergedStringMap(pack.DefaultMelody, variant.DefaultMelody)
		if strings.TrimSpace(variant.DrumLeadSurface) != "" {
			pack.DrumLeadSurface = variant.DrumLeadSurface
		}
		if variant.SoftLowEnd != nil {
			pack.SoftLowEnd = *variant.SoftLowEnd
		}
		if variant.ExtendedComp != nil {
			pack.ExtendedComp = *variant.ExtendedComp
		}
	} else {
		pack.Substyle = substyle
	}
	return pack
}

func inferSubstyle(pack stylePack, title string, tags []string) string {
	text := strings.ToLower(strings.TrimSpace(title + " " + strings.Join(tags, " ")))
	bestKey := ""
	bestScore := 0
	keys := make([]string, 0, len(pack.Substyles))
	for key := range pack.Substyles {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		variant := pack.Substyles[key]
		score := 0
		for _, keyword := range variant.Keywords {
			if keyword != "" && strings.Contains(text, strings.ToLower(keyword)) {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			bestKey = key
		}
	}
	if bestScore > 0 {
		return bestKey
	}
	return ""
}

func mergedStringMap(base, override map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(override))
	for key, value := range base {
		out[key] = value
	}
	for key, value := range override {
		if strings.TrimSpace(value) != "" {
			out[key] = value
		}
	}
	return out
}

func (p stylePack) phraseBars(totalBars int) int {
	if totalBars <= 2 {
		return totalBars
	}
	if totalBars >= 8 && p.LongPhraseBars > 0 {
		return p.LongPhraseBars
	}
	if p.ShortPhraseBars > 0 {
		return p.ShortPhraseBars
	}
	return 2
}

func (p stylePack) defaultRhythm(role string) string {
	lower := strings.ToLower(strings.TrimSpace(role))
	if pattern, ok := p.DefaultRhythms[lower]; ok && strings.TrimSpace(pattern) != "" {
		return pattern
	}
	if lower == "keys" || lower == "piano" || lower == "guitar" || lower == "organ" {
		if pattern, ok := p.DefaultRhythms["comp"]; ok && strings.TrimSpace(pattern) != "" {
			return pattern
		}
	}
	return ""
}

func (p stylePack) defaultMelody(role string) string {
	lower := strings.ToLower(strings.TrimSpace(role))
	if pattern, ok := p.DefaultMelody[lower]; ok && strings.TrimSpace(pattern) != "" {
		return pattern
	}
	if pattern, ok := p.DefaultMelody["lead"]; ok && strings.TrimSpace(pattern) != "" {
		return pattern
	}
	return ""
}

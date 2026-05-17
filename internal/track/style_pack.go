package track

import "strings"

type stylePack struct {
	Name            string
	DefaultBPM      float64
	ShortPhraseBars int
	LongPhraseBars  int
	DefaultRhythms  map[string]string
	DefaultMelody   map[string]string
	DrumLeadSurface string
	SoftLowEnd      bool
	ExtendedComp    bool
}

var stylePacks = map[string]stylePack{
	"ambient": {
		Name:            "ambient",
		DefaultBPM:      58,
		ShortPhraseBars: 2,
		LongPhraseBars:  4,
		DefaultRhythms: map[string]string{
			"pad":   "x... ....",
			"bass":  "x... ....",
			"bells": "x... ....",
		},
		DefaultMelody: map[string]string{
			"lead": "5 . 3 . | 1 . . .",
		},
		DrumLeadSurface: "none",
		SoftLowEnd:      true,
	},
	"bells": {
		Name:            "bells",
		DefaultBPM:      54,
		ShortPhraseBars: 2,
		LongPhraseBars:  2,
		DefaultRhythms: map[string]string{
			"bells":   "x... ....",
			"celesta": "x... ....",
			"pad":     "x... ....",
		},
		DefaultMelody: map[string]string{
			"bells": "5 . . 7 | 9 . 7 5",
			"lead":  "5 . . 7 | 9 . 7 5",
		},
		DrumLeadSurface: "none",
		SoftLowEnd:      true,
	},
	"classical": {
		Name:            "classical",
		DefaultBPM:      92,
		ShortPhraseBars: 2,
		LongPhraseBars:  4,
		DefaultRhythms: map[string]string{
			"piano":   "x..x .x..",
			"strings": "x... ....",
		},
		DefaultMelody: map[string]string{
			"lead": "5 . 3 . | 1 . . .",
		},
		DrumLeadSurface: "none",
		ExtendedComp:    true,
	},
	"drone": {
		Name:            "drone",
		DefaultBPM:      46,
		ShortPhraseBars: 2,
		LongPhraseBars:  4,
		DefaultRhythms: map[string]string{
			"bed":  "x... ....",
			"bass": "x... ....",
		},
		DefaultMelody: map[string]string{
			"lead": "5 . 3 . | 1 . . .",
		},
		DrumLeadSurface: "none",
		SoftLowEnd:      true,
	},
	"jazz": {
		Name:            "jazz",
		DefaultBPM:      126,
		ShortPhraseBars: 2,
		LongPhraseBars:  4,
		DefaultRhythms: map[string]string{
			"kick":  "x... x...",
			"snare": ".... x...",
			"hat":   "x.x.x.x.",
			"ride":  "x.x. x.x.",
			"bass":  "x... x...",
			"comp":  "x..x .x..",
			"piano": "x..x .x..",
		},
		DefaultMelody: map[string]string{
			"lead": "5 . 6 7 | 9 . 7 3",
		},
		DrumLeadSurface: "ride",
		ExtendedComp:    true,
	},
	"lofi": {
		Name:            "lofi",
		DefaultBPM:      78,
		ShortPhraseBars: 2,
		LongPhraseBars:  4,
		DefaultRhythms: map[string]string{
			"kick":   "x... x...",
			"snare":  ".... x...",
			"hat":    "x.x.x.x.",
			"bass":   "x... x...",
			"keys":   "x..x .x..",
			"guitar": "x..x .x..",
		},
		DefaultMelody: map[string]string{
			"lead": "5 . . 7 | 9 . 7 5",
		},
		DrumLeadSurface: "hat",
		ExtendedComp:    true,
	},
	"lullaby": {
		Name:            "lullaby",
		DefaultBPM:      68,
		ShortPhraseBars: 2,
		LongPhraseBars:  2,
		DefaultRhythms: map[string]string{
			"lead": "x... ....",
			"harp": "x..x ....",
			"pad":  "x... ....",
		},
		DefaultMelody: map[string]string{
			"lead": "5 . 3 . | 1 . . .",
		},
		DrumLeadSurface: "none",
		SoftLowEnd:      true,
	},
	"phase": {
		Name:            "phase",
		DefaultBPM:      74,
		ShortPhraseBars: 2,
		LongPhraseBars:  2,
		DefaultRhythms: map[string]string{
			"mallet-a": "x... x...",
			"mallet-b": ".... x...",
			"pad":      "x... ....",
		},
		DefaultMelody: map[string]string{
			"lead": "5 . 3 . | 1 . . .",
		},
		DrumLeadSurface: "none",
		SoftLowEnd:      true,
	},
}

func stylePackFor(style string) stylePack {
	pack, ok := stylePacks[strings.ToLower(strings.TrimSpace(style))]
	if ok {
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
	}
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
	if lower == "keys" || lower == "piano" || lower == "guitar" {
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

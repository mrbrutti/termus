package gen

import (
	"math/rand"
	"sort"
)

var sf2MaxPalette = map[string]map[int32]string{
	"ambient": {
		0: "merlin-symphony",
		1: "arachno",
		2: "musescore-general",
		3: "fairy-tale",
		4: "fairy-tale",
		5: "fm-dx",
	},
	"drone": {
		0: "arachno",
		1: "merlin-symphony",
		2: "musescore-general",
		3: "fm-dx",
		4: "fatboy",
	},
	"bells": {
		0: "fairy-tale",
		1: "fairy-tale",
		2: "fairy-tale",
		3: "fairy-tale",
		4: "fairy-tale",
		5: "fairy-tale",
		6: "fairy-tale",
		7: "fairy-tale",
	},
	"lullaby": {
		0: "musescore-general",
		1: "fairy-tale",
		2: "fairy-tale",
		3: "fairy-tale",
		4: "timbres-of-heaven",
	},
	"classical": {
		0: "timbres-of-heaven",
		1: "merlin-symphony",
		2: "musescore-general",
		3: "dsound4",
		4: "timbres-of-heaven",
	},
	"phase": {
		0: "fm-dx",
		1: "musescore-general",
		2: "timbres-of-heaven",
		3: "fm-dx",
		4: "fatboy",
		5: "fairy-tale",
	},
	"lofi": {
		0: "fatboy",
		1: "sgm",
		2: "dsound4",
		3: "tyros4",
		4: "sgm",
		9: "fatboy",
	},
	"jazz": {
		0: "sgm",
		1: "sgm",
		2: "tyros4",
		9: "tyros4",
	},
}

var sf2MaxPaletteExtras = map[string][]string{
	"bells": {"arachno", "timbres-of-heaven"},
}

func applyMaxSF2Palette(core *sf2Core, algoName string) {
	if core == nil || !core.usingMaxPalette() {
		return
	}
	routes, ok := sf2MaxPalette[algoName]
	if !ok {
		return
	}
	for channel, preset := range routes {
		core.routeChannelPreset(channel, preset)
	}
}

func applyGlassMaxPalette(core *sf2Core, rng *rand.Rand) {
	if core == nil || !core.usingMaxPalette() {
		return
	}
	scene := glassMaxScenes[0]
	if rng != nil {
		scene = glassMaxScenes[rng.Intn(len(glassMaxScenes))]
	}
	for channel, preset := range scene {
		core.routeChannelPreset(channel, preset)
	}
}

var glassMaxScenes = []map[int32]string{
	{
		0: "fairy-tale",
		1: "fairy-tale",
		2: "fairy-tale",
		3: "fairy-tale",
		4: "fairy-tale",
		5: "fairy-tale",
		6: "fairy-tale",
		7: "fairy-tale",
	},
	{
		0: "fairy-tale",
		1: "fairy-tale",
		2: "fairy-tale",
		3: "fairy-tale",
		4: "arachno",
		5: "arachno",
		6: "fairy-tale",
		7: "fairy-tale",
	},
	{
		0: "fairy-tale",
		1: "fairy-tale",
		2: "fairy-tale",
		3: "fairy-tale",
		4: "fairy-tale",
		5: "fairy-tale",
		6: "timbres-of-heaven",
		7: "fairy-tale",
	},
}

func MaxSF2PresetsForSpec(spec AlgoSpec) []string {
	if !spec.RequiresSF2 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0)
	if spec.PreferredSF2 != "" {
		seen[spec.PreferredSF2] = true
		out = append(out, spec.PreferredSF2)
	}
	for _, preset := range sf2MaxPalette[spec.Name] {
		if preset == "" || seen[preset] {
			continue
		}
		seen[preset] = true
		out = append(out, preset)
	}
	for _, preset := range sf2MaxPaletteExtras[spec.Name] {
		if preset == "" || seen[preset] {
			continue
		}
		seen[preset] = true
		out = append(out, preset)
	}
	sort.Strings(out)
	return out
}

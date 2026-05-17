package gen

import (
	"math/rand"
	"sort"
)

func applyMaxSF2Palette(core *sf2Core, algoName string) {
	if core == nil || !core.usingMaxPalette() {
		return
	}
	runtime := currentSF2Runtime()
	routes, ok := runtime.routes[algoName]
	if !ok {
		return
	}
	for channel, preset := range routes {
		core.routeChannelPreset(channel, preset)
	}
}

func applyGlassMaxPalette(core *sf2Core, rng *rand.Rand) {
	// Glass no longer needs bespoke randomized scenes; it uses the same
	// inventory-backed role routing as every other style.
	applyMaxSF2Palette(core, "bells")
}

func MaxSF2PresetsForSpec(spec AlgoSpec) []string {
	if !spec.RequiresSF2 {
		return nil
	}
	selection := ResolveSF2Selection(spec, nil, "max", spec.PreferredSF2)
	return append([]string(nil), selection.Presets...)
}

func ProSF2PresetForSpec(spec AlgoSpec, fallback string) string {
	if !spec.RequiresSF2 {
		return ""
	}
	selection := ResolveSF2Selection(spec, nil, "pro", fallback)
	if selection.Primary != "" {
		return selection.Primary
	}
	return fallback
}

func MaxSF2RoutesForSpec(spec AlgoSpec, blueprint *TrackBlueprint, fallback string) map[int32]string {
	selection := ResolveSF2Selection(spec, blueprint, "max", fallback)
	out := make(map[int32]string, len(selection.Routes))
	for channel, preset := range selection.Routes {
		out[channel] = preset
	}
	return out
}

func SortedPresetNames(names []string) []string {
	out := append([]string(nil), names...)
	sort.Strings(out)
	return out
}

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/track"
)

func discoverTracks() ([]track.Entry, error) {
	return track.Discover()
}

func loadTrackSelection(entries []track.Entry, value string, defaultSeed int64, defaultListenMode gen.ListeningMode) (track.Entry, *track.Compiled, error) {
	if strings.TrimSpace(value) == "" {
		return track.Entry{}, nil, fmt.Errorf("empty track selection")
	}
	entry, ok := track.Resolve(entries, value)
	if !ok {
		return track.Entry{}, nil, fmt.Errorf("unknown track %q", value)
	}
	file, err := track.ParseFile(entry.Path)
	if err != nil {
		return track.Entry{}, nil, err
	}
	compiled, err := track.Compile(file, defaultSeed, defaultListenMode)
	if err != nil {
		return track.Entry{}, nil, err
	}
	return entry, compiled, nil
}

func mergeProfiles(base, overlay gen.ControlProfile) gen.ControlProfile {
	return gen.ControlProfile{
		Density:    clampProfile(base.Density + (overlay.Density - 2)),
		Brightness: clampProfile(base.Brightness + (overlay.Brightness - 2)),
		Motion:     clampProfile(base.Motion + (overlay.Motion - 2)),
		Reverb:     clampProfile(base.Reverb + (overlay.Reverb - 2)),
		Swing:      clampProfile(base.Swing + (overlay.Swing - 2)),
		DroneDepth: clampProfile(base.DroneDepth + (overlay.DroneDepth - 2)),
		Tempo:      clampProfile(base.Tempo + (overlay.Tempo - 2)),
		Phrase:     clampProfile(base.Phrase + (overlay.Phrase - 2)),
	}
}

func clampProfile(v int) int {
	switch {
	case v < 0:
		return 0
	case v > 4:
		return 4
	default:
		return v
	}
}

func logTrackWarnings(compiled *track.Compiled) {
	if compiled == nil {
		return
	}
	for _, warning := range compiled.Warnings {
		fmt.Fprintf(os.Stderr, "track lint %s: %s\n", warning.Path, warning.Message)
	}
}

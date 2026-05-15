package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/gen"
)

type playlistAlgoBuilder func(spec gen.AlgoSpec, seed int64) gen.Algorithm
type playlistRenderer func(path string, algo gen.Algorithm, seconds float64, volume int) (int, error)

type playlistManifest struct {
	Name           string                  `json:"name"`
	Mode           string                  `json:"mode"`
	TrackCount     int                     `json:"track_count"`
	TotalDurationS float64                 `json:"total_duration_s"`
	Tracks         []playlistManifestTrack `json:"tracks"`
}

type playlistManifestTrack struct {
	Index     int                   `json:"index"`
	Algo      string                `json:"algo"`
	Display   string                `json:"display"`
	Seed      int64                 `json:"seed"`
	Path      string                `json:"path"`
	MIDIPath  string                `json:"midi_path,omitempty"`
	StemDir   string                `json:"stem_dir,omitempty"`
	StemFiles []string              `json:"stem_files,omitempty"`
	Frames    int                   `json:"frames"`
	DurationS float64               `json:"duration_s"`
	Markers   []gen.ListeningMarker `json:"markers,omitempty"`
}

func renderPlaylistOut(outDir string, pl *gen.Playlist, volume int, build playlistAlgoBuilder, exportMIDI, exportStems bool) (*playlistManifest, error) {
	return renderPlaylistOutWith(outDir, pl, volume, build, audio.RenderToWAV, exportMIDI, exportStems)
}

func renderPlaylistOutWith(outDir string, pl *gen.Playlist, volume int, build playlistAlgoBuilder, render playlistRenderer, exportMIDI, exportStems bool) (*playlistManifest, error) {
	if pl == nil {
		return nil, fmt.Errorf("playlist is nil")
	}
	if len(pl.Tracks) == 0 {
		return nil, fmt.Errorf("playlist has no tracks")
	}
	if build == nil {
		return nil, fmt.Errorf("playlist builder is nil")
	}
	if render == nil {
		return nil, fmt.Errorf("playlist renderer is nil")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}

	digits := len(strconv.Itoa(len(pl.Tracks)))
	manifest := &playlistManifest{
		Name:       pl.Name,
		Mode:       playlistModeLabel(pl.Mode),
		TrackCount: len(pl.Tracks),
		Tracks:     make([]playlistManifestTrack, 0, len(pl.Tracks)),
	}

	for i, track := range pl.Tracks {
		algo := build(track.Spec, track.Seed)
		seconds := track.Duration.Seconds()
		base := fmt.Sprintf("%0*d-%s-%d.wav", digits, i+1, safeFileStem(track.Spec.Name), track.Seed)
		absPath := filepath.Join(outDir, base)
		frames, err := render(absPath, algo, seconds, volume)
		if err != nil {
			return nil, fmt.Errorf("render track %d (%s): %w", i+1, track.Spec.Name, err)
		}

		item := playlistManifestTrack{
			Index:     i + 1,
			Algo:      track.Spec.Name,
			Display:   track.Spec.Display,
			Seed:      track.Seed,
			Path:      base,
			Frames:    frames,
			DurationS: seconds,
		}
		if inspectable, ok := algo.(gen.ListeningInspectable); ok {
			item.Markers = trimMarkersToFrames(inspectable.ListeningMarkers(), frames)
		}
		if exportMIDI || exportStems {
			exportAlgo := build(track.Spec, track.Seed)
			if exporter, ok := exportAlgo.(gen.TuningExporter); ok {
				stemBase := strings.TrimSuffix(base, ".wav")
				if exportMIDI {
					item.MIDIPath = stemBase + ".mid"
					if err := exporter.ExportMIDI(filepath.Join(outDir, item.MIDIPath), seconds); err != nil {
						return nil, fmt.Errorf("export midi track %d (%s): %w", i+1, track.Spec.Name, err)
					}
				}
				if exportStems {
					item.StemDir = stemBase + "-stems"
					files, err := exporter.ExportStems(filepath.Join(outDir, item.StemDir), seconds, volume)
					if err != nil {
						return nil, fmt.Errorf("export stems track %d (%s): %w", i+1, track.Spec.Name, err)
					}
					item.StemFiles = make([]string, 0, len(files))
					for _, path := range files {
						rel, err := filepath.Rel(outDir, path)
						if err != nil {
							rel = path
						}
						item.StemFiles = append(item.StemFiles, rel)
					}
				}
			}
		}
		manifest.TotalDurationS += seconds
		manifest.Tracks = append(manifest.Tracks, item)
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(outDir, "manifest.json"), data, 0o644); err != nil {
		return nil, err
	}
	return manifest, nil
}

func playlistModeLabel(mode gen.PlaylistMode) string {
	switch mode {
	case gen.PlaylistSameGenre:
		return "same"
	case gen.PlaylistMixed:
		return "mixed"
	default:
		return "unknown"
	}
}

func safeFileStem(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastDash = false
		case r == '-' || r == '_' || r == ' ' || r == '/':
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "track"
	}
	return out
}

func trimMarkersToFrames(markers []gen.ListeningMarker, totalFrames int) []gen.ListeningMarker {
	out := make([]gen.ListeningMarker, 0, len(markers))
	for _, marker := range markers {
		if int(marker.Sample) <= totalFrames {
			out = append(out, marker)
		}
	}
	return out
}

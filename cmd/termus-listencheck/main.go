package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sinshu/go-meltysynth/meltysynth"

	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/sf2"
)

type corpusCase struct {
	Name    string
	Algo    string
	Seed    int64
	Seconds float64
}

type corpusResult struct {
	Name      string                `json:"name"`
	Algo      string                `json:"algo"`
	Seed      int64                 `json:"seed"`
	Path      string                `json:"path,omitempty"`
	MIDIPath  string                `json:"midi_path,omitempty"`
	StemDir   string                `json:"stem_dir,omitempty"`
	StemFiles []string              `json:"stem_files,omitempty"`
	Skipped   string                `json:"skipped,omitempty"`
	Markers   []gen.ListeningMarker `json:"markers,omitempty"`
	Frames    int                   `json:"frames"`
	DurationS float64               `json:"duration_s"`
}

func main() {
	outDir := flag.String("out", "listencheck", "output directory for WAVs and manifest")
	sf2Path := flag.String("sf2", "", "optional SoundFont path for SF2-backed corpus cases")
	sf2Preset := flag.String("sf2-preset", "general", "SoundFont preset to auto-fetch for SF2 corpus cases")
	includeSF2 := flag.Bool("include-sf2", false, "include SF2-backed lofi and jazz renders in the corpus")
	exportStems := flag.Bool("stems", false, "also export per-stem WAVs for SF2-backed corpus cases")
	exportMIDI := flag.Bool("midi", false, "also export captured MIDI files for SF2-backed corpus cases")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "mkdir:", err)
		os.Exit(1)
	}

	corpus := []corpusCase{
		{Name: "ambient-synth-42", Algo: "ambient-synth", Seed: 42, Seconds: 12},
		{Name: "classical-synth-99", Algo: "classical-synth", Seed: 99, Seconds: 14},
	}
	if *includeSF2 {
		corpus = append(corpus,
			corpusCase{Name: "lofi-42", Algo: "lofi", Seed: 42, Seconds: 16},
			corpusCase{Name: "jazz-77", Algo: "jazz", Seed: 77, Seconds: 16},
		)
	}

	var sharedSF *meltysynth.SoundFont
	var err error
	if *includeSF2 {
		sharedSF, err = loadSoundFont(*sf2Path, *sf2Preset)
		if err != nil {
			fmt.Fprintln(os.Stderr, "sf2:", err)
			os.Exit(1)
		}
	}

	results := make([]corpusResult, 0, len(corpus))
	for _, item := range corpus {
		spec, ok := gen.Resolve(item.Algo)
		if !ok {
			results = append(results, corpusResult{
				Name:    item.Name,
				Algo:    item.Algo,
				Seed:    item.Seed,
				Skipped: "unknown algorithm",
			})
			continue
		}
		if spec.RequiresSF2 && sharedSF == nil {
			results = append(results, corpusResult{
				Name:    item.Name,
				Algo:    item.Algo,
				Seed:    item.Seed,
				Skipped: "sf2 not configured",
			})
			continue
		}
		algo := spec.Build(sharedSF)
		algo.Seed(item.Seed)

		outPath := filepath.Join(*outDir, item.Name+".wav")
		frames, err := renderToWAV(outPath, algo, item.Seconds)
		if err != nil {
			fmt.Fprintf(os.Stderr, "render %s: %v\n", item.Name, err)
			os.Exit(1)
		}
		result := corpusResult{
			Name:      item.Name,
			Algo:      item.Algo,
			Seed:      item.Seed,
			Path:      outPath,
			Frames:    frames,
			DurationS: item.Seconds,
		}
		if inspectable, ok := algo.(gen.ListeningInspectable); ok {
			result.Markers = trimMarkers(inspectable.ListeningMarkers(), frames)
		}
		if *exportMIDI && spec.RequiresSF2 {
			midiAlgo := spec.Build(sharedSF)
			midiAlgo.Seed(item.Seed)
			if exporter, ok := midiAlgo.(gen.TuningExporter); ok {
				result.MIDIPath = filepath.Join(*outDir, item.Name+".mid")
				if err := exporter.ExportMIDI(result.MIDIPath, item.Seconds); err != nil {
					fmt.Fprintf(os.Stderr, "midi %s: %v\n", item.Name, err)
					os.Exit(1)
				}
			}
		}
		if *exportStems && spec.RequiresSF2 {
			stemAlgo := spec.Build(sharedSF)
			stemAlgo.Seed(item.Seed)
			if exporter, ok := stemAlgo.(gen.TuningExporter); ok {
				result.StemDir = filepath.Join(*outDir, item.Name+"-stems")
				files, err := exporter.ExportStems(result.StemDir, item.Seconds, 100)
				if err != nil {
					fmt.Fprintf(os.Stderr, "stems %s: %v\n", item.Name, err)
					os.Exit(1)
				}
				result.StemFiles = files
			}
		}
		results = append(results, result)
		fmt.Fprintf(os.Stderr, "rendered %s -> %s\n", item.Name, outPath)
	}

	manifestPath := filepath.Join(*outDir, "manifest.json")
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "manifest:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "manifest write:", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", manifestPath)
}

func loadSoundFont(path, preset string) (*meltysynth.SoundFont, error) {
	if strings.TrimSpace(path) == "" {
		resolved, err := sf2.EnsurePreset(preset, nil)
		if err != nil {
			return nil, err
		}
		path = resolved
	}
	return sf2.Open(path)
}

func renderToWAV(path string, algo gen.Algorithm, seconds float64) (int, error) {
	return audio.RenderToWAV(path, algo, seconds, 100)
}

func trimMarkers(markers []gen.ListeningMarker, totalFrames int) []gen.ListeningMarker {
	out := make([]gen.ListeningMarker, 0, len(markers))
	for _, marker := range markers {
		if int(marker.Sample) <= totalFrames {
			out = append(out, marker)
		}
	}
	return out
}

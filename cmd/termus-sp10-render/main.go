// cmd/termus-sp10-render renders all 12 SP10 authored tracks to WAV files
// under wavs/sp10/<genre>/<name>.wav and then runs DSP verification on the
// rendered files: RMS loudness, spectral centroid per genre, and pitch-wow
// check for lofi vs jazz tracks.
//
// Usage:
//
//	go run ./cmd/termus-sp10-render [--out wavs/sp10] [--seconds 45] [--sf2-preset general]
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/audiotest"
	"github.com/mrbrutti/termus/internal/gen"
	termsf2 "github.com/mrbrutti/termus/internal/sf2"
	"github.com/mrbrutti/termus/internal/track"
	"github.com/sinshu/go-meltysynth/meltysynth"
)

const sampleRate = 44100.0

type result struct {
	genre string
	name  string
	path  string
	err   error
	// DSP measurements
	rmsDB      float64
	peakDB     float64
	centroidHz float64
}

func main() {
	outRoot := flag.String("out", "wavs/sp10", "root output directory")
	seconds := flag.Float64("seconds", 45.0, "render duration in seconds")
	sf2Preset := flag.String("sf2-preset", termsf2.DefaultPreset, "SF2 preset name")
	sf2Strategy := flag.String("sf2-strategy", "pro", "single | pro | max")
	volume := flag.Int("volume", 78, "render volume 0..100")
	flag.Parse()

	// Discover all tracks from cwd/tracks/
	entries, err := track.Discover()
	if err != nil {
		fmt.Fprintf(os.Stderr, "discover: %v\n", err)
		os.Exit(1)
	}
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "no tracks found — run from repo root")
		os.Exit(1)
	}

	// Filter to our 24 SP19 target tracks across the 6 genres
	target := map[string]bool{
		"lofi/rainy-window-cafe":            true,
		"lofi/midnight-train-window":        true,
		"lofi/bookstore-quiet-aisle":        true,
		"lofi/autumn-walk-home":             true,
		"jazz/autumn-leaves-after-hours":    true,
		"jazz/blue-bossa-late-set":          true,
		"jazz/bourbon-street-blues":         true,
		"jazz/coltrane-modal-meditation":    true,
		"chill/coastal-cliff-morning":       true,
		"chill/mountain-fog-drift":          true,
		"chill/sunset-balcony-loop":         true,
		"chill/midnight-pool-blue":          true,
		"ambient/deep-sea-cathedral":        true,
		"ambient/forest-after-rain":         true,
		"ambient/glacial-slow-drift":        true,
		"ambient/stellar-aurora":            true,
		"blues/delta-crossroads":            true,
		"blues/twelve-bar-rain":             true,
		"blues/chicago-after-midnight":      true,
		"blues/mississippi-slow-drag":       true,
		"rock/garage-saturday-night":        true,
		"rock/highway-sunset-cruise":        true,
		"rock/basement-jam-session":         true,
		"rock/anthem-stadium-rise":          true,
	}

	var selected []track.Entry
	for _, e := range entries {
		if target[e.ID] {
			selected = append(selected, e)
		}
	}

	// Report any missing targets
	found := map[string]bool{}
	for _, e := range selected {
		found[e.ID] = true
	}
	for id := range target {
		if !found[id] {
			fmt.Fprintf(os.Stderr, "WARNING: target track %q not found in discovery\n", id)
		}
	}

	// Build SF2 cache
	sfCache := &sfontCache{fonts: map[string]*meltysynth.SoundFont{}}

	var results []result

	fmt.Fprintf(os.Stderr, "\n=== Rendering %d tracks (%.0fs each) ===\n\n", len(selected), *seconds)

	for _, entry := range selected {
		parts := strings.SplitN(entry.ID, "/", 2)
		genre := parts[0]
		name := parts[1]
		if len(parts) < 2 {
			genre = entry.Style
			name = entry.ID
		}

		outDir := filepath.Join(*outRoot, genre)
		wavPath := filepath.Join(outDir, name+".wav")

		if err := os.MkdirAll(outDir, 0o755); err != nil {
			results = append(results, result{genre: genre, name: name, path: wavPath, err: err})
			continue
		}

		res := renderTrack(entry, wavPath, *seconds, *sf2Preset, *sf2Strategy, *volume, sfCache)
		results = append(results, res)

		if res.err != nil {
			fmt.Fprintf(os.Stderr, "FAIL  %s/%s: %v\n", genre, name, res.err)
		} else {
			fmt.Fprintf(os.Stderr, "OK    %s/%s -> %s  (RMS %.1f dBFS  centroid %.0f Hz)\n",
				genre, name, wavPath, res.rmsDB, res.centroidHz)
		}
	}

	// Print verification report
	printReport(results)
}

func renderTrack(entry track.Entry, wavPath string, seconds float64, sf2Preset, sf2Strategy string, volume int, sfCache *sfontCache) result {
	genre := entry.Style
	parts := strings.SplitN(entry.ID, "/", 2)
	name := entry.ID
	if len(parts) == 2 {
		name = parts[1]
	}

	file, err := track.ParseFile(entry.Path)
	if err != nil {
		return result{genre: genre, name: name, path: wavPath, err: fmt.Errorf("parse: %w", err)}
	}

	compiled, err := track.Compile(file, 1, gen.ListeningModeEndless)
	if err != nil {
		return result{genre: genre, name: name, path: wavPath, err: fmt.Errorf("compile: %w", err)}
	}

	if len(compiled.Playlist.Tracks) == 0 {
		return result{genre: genre, name: name, path: wavPath, err: fmt.Errorf("playlist is empty")}
	}

	// SP17: render through the full section schedule of the single
	// compiled Track. We render up to `seconds` worth of audio,
	// allocating frames to each section in proportion to its authored
	// duration (clipping the last section if necessary). Algorithms swap
	// at sample-aligned boundaries with no crossfade.
	item := compiled.Playlist.Tracks[0]
	totalRequested := int(seconds * 44100.0)
	if totalRequested < 1 {
		totalRequested = int(item.Duration.Seconds() * 44100.0)
	}
	var stops []audio.SectionStop
	if len(item.Sections) == 0 {
		algo := sfCache.buildAlgo(compiled, item.Spec, item.Seed, sf2Strategy, sf2Preset)
		stops = []audio.SectionStop{{Algo: algo, Frames: totalRequested}}
	} else {
		// Compute frame allocations per section, clipping to the
		// requested total. Sections beyond the requested duration are
		// skipped.
		remaining := totalRequested
		stops = make([]audio.SectionStop, 0, len(item.Sections))
		for _, stop := range item.Sections {
			if remaining <= 0 {
				break
			}
			frames := int(stop.Duration.Seconds() * 44100.0)
			if frames > remaining {
				frames = remaining
			}
			seed := stop.Seed
			algo := sfCache.buildAlgo(compiled, item.Spec, seed, sf2Strategy, sf2Preset)
			stops = append(stops, audio.SectionStop{Algo: algo, Frames: frames})
			remaining -= frames
		}
		// If we have leftover requested time after walking all sections,
		// loop back through them so the full duration is rendered. Each
		// loop iteration offsets the algorithm seed to vary the rendered
		// events while still using the same authored plan.
		loopIter := 1
		for remaining > 0 && loopIter < 16 {
			for _, stop := range item.Sections {
				if remaining <= 0 {
					break
				}
				frames := int(stop.Duration.Seconds() * 44100.0)
				if frames > remaining {
					frames = remaining
				}
				algo := sfCache.buildAlgoWithBaseSeed(compiled, item.Spec, stop.Seed, stop.Seed+int64(loopIter)*1009, sf2Strategy, sf2Preset)
				stops = append(stops, audio.SectionStop{Algo: algo, Frames: frames})
				remaining -= frames
			}
			loopIter++
		}
	}

	frames, err := audio.RenderSectionsToWAV(wavPath, stops, volume)
	if err != nil {
		return result{genre: genre, name: name, path: wavPath, err: fmt.Errorf("render: %w", err)}
	}
	if frames == 0 {
		return result{genre: genre, name: name, path: wavPath, err: fmt.Errorf("zero frames rendered")}
	}

	// Load the rendered WAV back and compute DSP metrics
	mono, err := audio.ReadIR(wavPath)
	if err != nil {
		return result{genre: genre, name: name, path: wavPath, err: fmt.Errorf("read back WAV: %w", err)}
	}
	if len(mono) == 0 {
		return result{genre: genre, name: name, path: wavPath, err: fmt.Errorf("WAV is empty after render")}
	}

	m := audiotest.MeasureMono(mono, sampleRate)
	return result{
		genre:      genre,
		name:       name,
		path:       wavPath,
		rmsDB:      m.RMSDb,
		peakDB:     m.PeakDb,
		centroidHz: m.CentroidHz,
	}
}

func printReport(results []result) {
	fmt.Fprintln(os.Stderr, "\n========================================")
	fmt.Fprintln(os.Stderr, "SP10 RENDER + VERIFY REPORT")
	fmt.Fprintln(os.Stderr, "========================================")

	// Per-track table
	fmt.Fprintln(os.Stderr, "\nPer-Track Measurements:")
	fmt.Fprintf(os.Stderr, "%-8s  %-30s  %8s  %8s  %10s  %s\n",
		"Genre", "Track", "RMS dBFS", "Peak dBFS", "Centroid Hz", "Status")
	fmt.Fprintln(os.Stderr, strings.Repeat("-", 80))

	okCount := 0
	failCount := 0
	loudnessFails := 0

	genreRMS := map[string][]float64{}
	genreCentroid := map[string][]float64{}

	for _, r := range results {
		status := "OK"
		if r.err != nil {
			status = "FAIL: " + r.err.Error()
			failCount++
		} else {
			okCount++
			// Check loudness band [-30, -10] dBFS
			if r.rmsDB < -30 || r.rmsDB > -10 {
				status = fmt.Sprintf("LOUD_WARN(%.1f dBFS)", r.rmsDB)
				loudnessFails++
			}
			genreRMS[r.genre] = append(genreRMS[r.genre], r.rmsDB)
			genreCentroid[r.genre] = append(genreCentroid[r.genre], r.centroidHz)
		}
		fmt.Fprintf(os.Stderr, "%-8s  %-30s  %8.1f  %8.1f  %10.0f  %s\n",
			r.genre, r.name, r.rmsDB, r.peakDB, r.centroidHz, status)
	}

	// Per-genre means
	fmt.Fprintln(os.Stderr, "\nPer-Genre Means:")
	fmt.Fprintf(os.Stderr, "%-10s  %10s  %12s\n", "Genre", "Mean RMS", "Mean Centroid")
	fmt.Fprintln(os.Stderr, strings.Repeat("-", 36))
	genres := []string{"lofi", "jazz", "chill", "ambient"}
	for _, g := range genres {
		if len(genreCentroid[g]) == 0 {
			fmt.Fprintf(os.Stderr, "%-10s  %10s  %12s\n", g, "N/A", "N/A")
			continue
		}
		fmt.Fprintf(os.Stderr, "%-10s  %10.1f  %12.0f\n", g, mean(genreRMS[g]), mean(genreCentroid[g]))
	}

	// Spectral ordering check: lofi < jazz (lofi has LP at 7kHz, jazz is bright)
	lofiC := mean(genreCentroid["lofi"])
	jazzC := mean(genreCentroid["jazz"])
	ambC := mean(genreCentroid["ambient"])
	chillC := mean(genreCentroid["chill"])

	fmt.Fprintln(os.Stderr, "\nSpectral Ordering Check (lofi should be darkest):")
	if lofiC > 0 && jazzC > 0 {
		if lofiC < jazzC {
			fmt.Fprintf(os.Stderr, "  PASS: lofi (%.0f Hz) < jazz (%.0f Hz)\n", lofiC, jazzC)
		} else {
			fmt.Fprintf(os.Stderr, "  WARN: lofi (%.0f Hz) >= jazz (%.0f Hz) — expected lofi to be darker\n", lofiC, jazzC)
		}
	}
	if ambC > 0 {
		fmt.Fprintf(os.Stderr, "  ambient centroid: %.0f Hz  chill centroid: %.0f Hz\n", ambC, chillC)
	}

	// Pitch wow check: sample first lofi and first jazz WAV
	pitchWowCheck(results, "lofi", true)
	pitchWowCheck(results, "jazz", false)

	// Summary
	fmt.Fprintln(os.Stderr, "\n--- Summary ---")
	fmt.Fprintf(os.Stderr, "Rendered OK:         %d / %d\n", okCount, len(results))
	fmt.Fprintf(os.Stderr, "Render failures:     %d\n", failCount)
	fmt.Fprintf(os.Stderr, "Loudness warnings:   %d\n", loudnessFails)

	if failCount > 0 {
		fmt.Fprintln(os.Stderr, "\nStatus: DONE_WITH_CONCERNS (render failures)")
		os.Exit(1)
	} else {
		fmt.Fprintln(os.Stderr, "\nStatus: DONE")
	}
}

func pitchWowCheck(results []result, genre string, expectWow bool) {
	for _, r := range results {
		if r.genre != genre || r.err != nil {
			continue
		}
		mono, err := audio.ReadIR(r.path)
		if err != nil || len(mono) < int(sampleRate)*5 {
			fmt.Fprintf(os.Stderr, "\nPitch-wow check %s/%s: skipped (too short or unreadable)\n", genre, r.name)
			return
		}
		// Sample a 5-second window from 10s in (avoid intro transients)
		start := int(sampleRate * 10)
		end := start + int(sampleRate*5)
		if end > len(mono) {
			end = len(mono)
		}
		window := mono[start:end]

		// Use the strongest spectral bin as nominal pitch estimate
		centroid := audiotest.SpectralCentroidHz(window, sampleRate)
		// Use the centroid as a rough nominal; pitch tracker works on fundamental
		// For a broad mix we expect very short pitch-track sequences, so we just
		// report modulation depth and whether it is non-zero.
		nominal := centroid
		if nominal <= 0 {
			nominal = 220.0 // fallback A3
		}
		cents := audiotest.PitchTrack(window, sampleRate, nominal)
		if len(cents) < 4 {
			fmt.Fprintf(os.Stderr, "\nPitch-wow %s/%s: pitch-track returned %d entries (mix too polyphonic for zero-crossing tracker)\n",
				genre, r.name, len(cents))
			return
		}
		depth := audiotest.ModulationDepthCents(cents)
		label := "ABSENT"
		if depth > 2.0 {
			label = "PRESENT"
		}
		pass := (expectWow && depth > 2.0) || (!expectWow && depth <= 2.0)
		verdict := "PASS"
		if !pass {
			verdict = "WARN"
		}
		fmt.Fprintf(os.Stderr, "\nPitch-wow %s/%s: depth=%.2f cents [%s] -> %s\n",
			genre, r.name, depth, label, verdict)
		return // only check one track per genre
	}
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

// sfontCache mirrors the cache in termus-track-review to avoid loading the
// same SF2 file more than once.
type sfontCache struct {
	fonts map[string]*meltysynth.SoundFont
}

func (c *sfontCache) mustLoad(preset string) *meltysynth.SoundFont {
	if sf, ok := c.fonts[preset]; ok {
		return sf
	}
	fmt.Fprintf(os.Stderr, "  loading SF2 preset %q ...\n", preset)
	path, err := termsf2.EnsurePreset(preset, func(done, total int64) {
		if total > 0 {
			fmt.Fprintf(os.Stderr, "\r  downloading %s: %.0f%%", preset, float64(done)/float64(total)*100)
		}
	})
	if err != nil {
		panic(fmt.Sprintf("sf2 EnsurePreset(%q): %v", preset, err))
	}
	sf, err := termsf2.Open(path)
	if err != nil {
		panic(fmt.Sprintf("sf2 Open(%q): %v", path, err))
	}
	c.fonts[preset] = sf
	fmt.Fprintln(os.Stderr)
	return sf
}

func (c *sfontCache) buildAlgo(compiled *track.Compiled, spec gen.AlgoSpec, seed int64, strategy, fallbackPreset string) gen.Algorithm {
	return c.buildAlgoWithBaseSeed(compiled, spec, seed, seed, strategy, fallbackPreset)
}

// buildAlgoWithBaseSeed builds an authored algorithm where the plan/profile
// is looked up by planSeed but the runtime algorithm is reseeded with
// algoSeed. Used by SP17 long-form evolution so iterating loops can pick up
// the authored composition with varying RNG state.
func (c *sfontCache) buildAlgoWithBaseSeed(compiled *track.Compiled, spec gen.AlgoSpec, planSeed, algoSeed int64, strategy, fallbackPreset string) gen.Algorithm {
	key := fmt.Sprintf("%s:%d", spec.Name, planSeed)
	plan, ok := compiled.Plans[key]
	if !ok {
		algo := spec.Build(nil)
		algo.Seed(algoSeed)
		return algo
	}
	profile := gen.DefaultControlProfile()
	if got, ok := compiled.Profiles[key]; ok {
		profile = got
	}
	selection := gen.ResolveSF2SelectionForPlan(spec, &plan, strategy, fallbackPreset)
	if selection.Primary == "" {
		selection.Primary = fallbackPreset
	}
	primary := c.mustLoad(selection.Primary)
	runtimeFonts := map[string]*meltysynth.SoundFont{}
	for _, p := range selection.Presets {
		runtimeFonts[p] = c.mustLoad(p)
	}
	if len(runtimeFonts) == 0 {
		runtimeFonts[selection.Primary] = primary
	}
	gen.SetSF2RuntimeWithRoutes(strategy, runtimeFonts, map[string]map[int32]string{spec.Name: selection.Routes})
	resolvedPlan := plan
	if len(selection.Programs) > 0 {
		resolvedPlan.Tracks = append([]gen.AuthoredRenderTrack(nil), plan.Tracks...)
		for i := range resolvedPlan.Tracks {
			if program, ok := selection.Programs[resolvedPlan.Tracks[i].Channel]; ok {
				resolvedPlan.Tracks[i].Program = program
			}
		}
	}
	algo := gen.NewAuthoredTrack(spec, primary, resolvedPlan)
	algo = gen.ConfigureControlProfile(algo, profile)
	algo.Seed(algoSeed)
	return algo
}

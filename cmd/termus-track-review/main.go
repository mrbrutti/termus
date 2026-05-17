package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/gen"
	termsf2 "github.com/mrbrutti/termus/internal/sf2"
	"github.com/mrbrutti/termus/internal/track"
	"github.com/sinshu/go-meltysynth/meltysynth"
)

type reviewRunManifest struct {
	TrackCount int                `json:"track_count"`
	Tracks     []reviewTrackEntry `json:"tracks"`
}

type reviewTrackEntry struct {
	ID           string              `json:"id"`
	Style        string              `json:"style"`
	Substyle     string              `json:"substyle,omitempty"`
	Title        string              `json:"title"`
	Dir          string              `json:"dir"`
	ManifestPath string              `json:"manifest_path"`
	ReviewPath   string              `json:"review_path"`
	Metrics      track.ReviewMetrics `json:"metrics"`
}

type reviewPlaylistManifest struct {
	Name       string                `json:"name"`
	TrackCount int                   `json:"track_count"`
	Tracks     []reviewPlaylistTrack `json:"tracks"`
}

type reviewPlaylistTrack struct {
	Index     int                   `json:"index"`
	Algo      string                `json:"algo"`
	Title     string                `json:"title,omitempty"`
	Seed      int64                 `json:"seed"`
	Path      string                `json:"path"`
	Frames    int                   `json:"frames"`
	DurationS float64               `json:"duration_s"`
	Markers   []gen.ListeningMarker `json:"markers,omitempty"`
}

type soundFontCache struct {
	fonts map[string]*meltysynth.SoundFont
}

func main() {
	outDir := flag.String("out", "", "output directory for rendered review artifacts")
	styleFilter := flag.String("style", "", "limit review to one style")
	trackFilter := flag.String("track", "", "limit review to one track id/path")
	limit := flag.Int("limit", 0, "maximum number of tracks to render (0 = all)")
	sf2Strategy := flag.String("sf2-strategy", "pro", "single | pro | max")
	sf2Preset := flag.String("sf2-preset", termsf2.DefaultPreset, "fallback sf2 preset when a plan has no preferred bank")
	volume := flag.Int("volume", 78, "offline render volume 0..100")
	flag.Parse()

	if strings.TrimSpace(*outDir) == "" {
		fmt.Fprintln(os.Stderr, "--out is required")
		os.Exit(2)
	}

	entries, err := track.Discover()
	if err != nil {
		fmt.Fprintln(os.Stderr, "discover:", err)
		os.Exit(1)
	}
	selected, err := selectEntries(entries, *styleFilter, *trackFilter, *limit)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if len(selected) == 0 {
		fmt.Fprintln(os.Stderr, "no tracks selected")
		os.Exit(2)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "mkdir:", err)
		os.Exit(1)
	}

	cache := &soundFontCache{fonts: map[string]*meltysynth.SoundFont{}}
	run := reviewRunManifest{TrackCount: len(selected), Tracks: make([]reviewTrackEntry, 0, len(selected))}
	for _, entry := range selected {
		file, err := track.ParseFile(entry.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse %s: %v\n", entry.ID, err)
			os.Exit(1)
		}
		compiled, err := track.Compile(file, 1, gen.ListeningModeEndless)
		if err != nil {
			fmt.Fprintf(os.Stderr, "compile %s: %v\n", entry.ID, err)
			os.Exit(1)
		}
		report := track.Analyze(file, compiled)
		trackDir := filepath.Join(*outDir, safeDirName(entry.ID))
		if err := os.MkdirAll(trackDir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", trackDir, err)
			os.Exit(1)
		}
		builder := cache.builder(compiled, *sf2Strategy, *sf2Preset)
		playlistManifest, err := renderReviewPlaylist(trackDir, compiled.Playlist, *volume, builder)
		if err != nil {
			fmt.Fprintf(os.Stderr, "render %s: %v\n", entry.ID, err)
			os.Exit(1)
		}
		reviewPath := filepath.Join(trackDir, "review.json")
		if err := writeJSON(reviewPath, report); err != nil {
			fmt.Fprintf(os.Stderr, "write review %s: %v\n", entry.ID, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "reviewed %s -> %s\n", entry.ID, trackDir)
		run.Tracks = append(run.Tracks, reviewTrackEntry{
			ID:           entry.ID,
			Style:        entry.Style,
			Substyle:     entry.Substyle,
			Title:        entry.Title,
			Dir:          filepath.Base(trackDir),
			ManifestPath: filepath.Join(filepath.Base(trackDir), "manifest.json"),
			ReviewPath:   filepath.Join(filepath.Base(trackDir), "review.json"),
			Metrics:      report.Metrics,
		})
		_ = playlistManifest
	}
	if err := writeJSON(filepath.Join(*outDir, "index.json"), run); err != nil {
		fmt.Fprintln(os.Stderr, "write index:", err)
		os.Exit(1)
	}
}

func selectEntries(entries []track.Entry, styleFilter, trackFilter string, limit int) ([]track.Entry, error) {
	var selected []track.Entry
	if strings.TrimSpace(trackFilter) != "" {
		entry, ok := track.Resolve(entries, trackFilter)
		if !ok {
			return nil, fmt.Errorf("unknown track %q", trackFilter)
		}
		return []track.Entry{entry}, nil
	}
	for _, entry := range entries {
		if styleFilter != "" && !strings.EqualFold(entry.Style, styleFilter) {
			continue
		}
		selected = append(selected, entry)
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].ID < selected[j].ID })
	if limit > 0 && len(selected) > limit {
		selected = selected[:limit]
	}
	return selected, nil
}

func (c *soundFontCache) builder(compiled *track.Compiled, strategy, fallbackPreset string) func(gen.AlgoSpec, int64) gen.Algorithm {
	return func(spec gen.AlgoSpec, seed int64) gen.Algorithm {
		key := fmt.Sprintf("%s:%d", spec.Name, seed)
		plan, ok := compiled.Plans[key]
		if !ok {
			algo := spec.Build(nil)
			algo.Seed(seed)
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
		for _, preset := range selection.Presets {
			runtimeFonts[preset] = c.mustLoad(preset)
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
		algo.Seed(seed)
		return algo
	}
}

func (c *soundFontCache) mustLoad(preset string) *meltysynth.SoundFont {
	if sf, ok := c.fonts[preset]; ok {
		return sf
	}
	path, err := termsf2.EnsurePreset(preset, nil)
	if err != nil {
		panic(err)
	}
	sf, err := termsf2.Open(path)
	if err != nil {
		panic(err)
	}
	c.fonts[preset] = sf
	return sf
}

func renderReviewPlaylist(outDir string, pl gen.Playlist, volume int, build func(gen.AlgoSpec, int64) gen.Algorithm) (*reviewPlaylistManifest, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	manifest := &reviewPlaylistManifest{
		Name:       pl.Name,
		TrackCount: len(pl.Tracks),
		Tracks:     make([]reviewPlaylistTrack, 0, len(pl.Tracks)),
	}
	digits := len(fmt.Sprintf("%d", len(pl.Tracks)))
	for i, item := range pl.Tracks {
		algo := build(item.Spec, item.Seed)
		plan := audio.PlanRender(algo, item.Duration.Seconds())
		name := fmt.Sprintf("%0*d-%s-%d.wav", digits, i+1, safeFileStem(item.Spec.Name), item.Seed)
		absPath := filepath.Join(outDir, name)
		frames, err := audio.RenderToWAVWithPlan(absPath, algo, plan, volume)
		if err != nil {
			return nil, err
		}
		rendered := reviewPlaylistTrack{
			Index:     i + 1,
			Algo:      item.Spec.Name,
			Title:     item.Title,
			Seed:      item.Seed,
			Path:      name,
			Frames:    frames,
			DurationS: float64(frames) / 44100.0,
		}
		if inspectable, ok := algo.(gen.ListeningInspectable); ok {
			rendered.Markers = trimMarkersToFrames(inspectable.ListeningMarkers(), frames)
		}
		manifest.Tracks = append(manifest.Tracks, rendered)
	}
	if err := writeJSON(filepath.Join(outDir, "manifest.json"), manifest); err != nil {
		return nil, err
	}
	return manifest, nil
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

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
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

func safeDirName(id string) string {
	return strings.ReplaceAll(strings.Trim(id, "/"), "/", "__")
}

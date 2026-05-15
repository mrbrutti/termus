package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gopxl/beep/v2"
	"github.com/sinshu/go-meltysynth/meltysynth"

	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
	"github.com/mrbrutti/termus/internal/sf2"
	"github.com/mrbrutti/termus/internal/synth"
	"github.com/mrbrutti/termus/internal/tui"
)

func main() {
	seed := flag.Int64("seed", time.Now().UnixNano(), "RNG seed (default: time-based)")
	algoName := flag.String("algo", "ambient",
		"algorithm name. Genre names: ambient | drone | bells | lullaby | "+
			"classical | phase | lofi | jazz. Append -synth for the no-download "+
			"versions. Legacy names (eno, chill, glass, etc.) also accepted.")
	initialVol := flag.Int("volume", 70, "initial volume 0..100")
	sf2Path := flag.String("sf2", "", "path to SoundFont file (overrides --sf2-preset)")
	sf2Preset := flag.String("sf2-preset", "general",
		"SoundFont preset: 'general' (32 MB GeneralUser-GS, balanced) | 'sgm' (325 MB, much better piano/guitar/bass)")
	sf2Strategy := flag.String("sf2-strategy", "single",
		"how to pick SoundFonts: 'single' (use --sf2-preset for everything) | "+
			"'optimal' (download each algorithm's preferred preset; uses more disk)")
	irPath := flag.String("ir", "", "convolution IR: WAV file path, or preset name: room | hall | cathedral | plate")
	irWet := flag.Float64("ir-wet", 0.40, "convolution wet mix 0..1 when --ir is provided")
	playlistMode := flag.String("playlist", "",
		"playlist mode: 'same' (multiple seeds of --algo) | 'mixed' (random genres). "+
			"Default empty = single track.")
	playlistTracks := flag.Int("playlist-tracks", 6, "number of tracks in the playlist")
	playlistDur := flag.Duration("playlist-duration", 5*time.Minute,
		"how long each playlist track plays before the crossfade")
	outPath := flag.String("out", "", "render directly to a WAV file instead of launching live playback")
	playlistOut := flag.String("playlist-out", "", "render a playlist to a directory of WAVs plus manifest.json")
	exportStems := flag.Bool("stems", false, "with --out/--playlist-out, also export per-stem WAVs when supported")
	exportMIDI := flag.Bool("midi", false, "with --out/--playlist-out, also export captured MIDI files when supported")
	debugView := flag.Bool("debug", false, "show the musical debug inspector in the TUI")
	renderSeconds := flag.Float64("seconds", 180, "render duration in seconds when --out is provided")
	flag.Parse()

	if *initialVol < 0 || *initialVol > 100 {
		fmt.Fprintf(os.Stderr, "volume must be 0..100, got %d\n", *initialVol)
		os.Exit(2)
	}
	if *outPath != "" && *renderSeconds <= 0 {
		fmt.Fprintf(os.Stderr, "--seconds must be > 0 when --out is used, got %.3f\n", *renderSeconds)
		os.Exit(2)
	}
	if *outPath != "" && *playlistOut != "" {
		fmt.Fprintln(os.Stderr, "--out and --playlist-out are mutually exclusive")
		os.Exit(2)
	}
	if *outPath != "" && *playlistMode != "" {
		fmt.Fprintln(os.Stderr, "--out does not support --playlist; use --playlist-out for batch rendering")
		os.Exit(2)
	}
	if *playlistOut != "" && *playlistMode == "" {
		fmt.Fprintln(os.Stderr, "--playlist-out requires --playlist same|mixed")
		os.Exit(2)
	}
	if (*exportStems || *exportMIDI) && *outPath == "" && *playlistOut == "" {
		fmt.Fprintln(os.Stderr, "--stems/--midi require --out or --playlist-out")
		os.Exit(2)
	}
	if *playlistMode != "" && *playlistDur <= 0 {
		fmt.Fprintf(os.Stderr, "--playlist-duration must be > 0, got %s\n", *playlistDur)
		os.Exit(2)
	}

	spec, ok := gen.Resolve(*algoName)
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown algorithm %q (available: %v)\n",
			*algoName, gen.AllAlgoNames())
		os.Exit(2)
	}
	// sfByPreset is the lookup the build closure uses to pick the right
	// SoundFont per algorithm. With strategy=single it contains exactly one
	// entry, mapped under both the chosen preset name and a "" fallback for
	// algorithms whose PreferredSF2 is empty.
	sfByPreset := map[string]*meltysynth.SoundFont{}
	var sf *meltysynth.SoundFont // the initial-algo SF
	if spec.RequiresSF2 {
		if *sf2Path != "" {
			// User pinned a custom file — use it for everything.
			loaded, err := sf2.Open(*sf2Path)
			if err != nil {
				fmt.Fprintln(os.Stderr, "sf2 open failed:", err)
				os.Exit(1)
			}
			sfByPreset[""] = loaded
			sf = loaded
		} else {
			needed := neededPresets(*sf2Strategy, *sf2Preset, spec)
			paths, err := sf2.EnsureAll(os.Stderr, needed)
			if err != nil {
				fmt.Fprintln(os.Stderr, "sf2 setup failed:", err)
				os.Exit(1)
			}
			for name, path := range paths {
				loaded, err := sf2.Open(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "sf2 open %q failed: %v\n", path, err)
					os.Exit(1)
				}
				sfByPreset[name] = loaded
			}
			// Pick the SF for the initially-requested algo.
			sf = pickSF(sfByPreset, spec, *sf2Preset)
		}
	}
	algo := spec.Build(sf)
	algo.Seed(*seed)

	var ir []float64
	var irLabel string
	// Optional convolution IR.
	if *irPath != "" {
		rev, ok := algo.(gen.SF2Reverberator)
		if !ok {
			fmt.Fprintf(os.Stderr, "--ir requires an sf2-mode algorithm; ignoring\n")
		} else {
			loadedIR, label, err := loadIR(*irPath, *seed)
			if err != nil {
				fmt.Fprintln(os.Stderr, "ir load failed:", err)
				os.Exit(1)
			}
			ir = loadedIR
			irLabel = label
			rev.SetReverbIR(ir, *irWet)
			fmt.Fprintf(os.Stderr, "IR %s: %d samples (%.1f ms) at wet=%.2f\n",
				irLabel, len(ir), float64(len(ir))*1000.0/44100.0, *irWet)
		}
	}
	buildRenderedAlgo := func(s gen.AlgoSpec, algoSeed int64) gen.Algorithm {
		chosen := pickSF(sfByPreset, s, *sf2Preset)
		a := s.Build(chosen)
		a.Seed(algoSeed)
		if len(ir) > 0 {
			if rev, ok := a.(gen.SF2Reverberator); ok {
				rev.SetReverbIR(ir, *irWet)
			}
		}
		return a
	}
	if *outPath != "" {
		frames, err := audio.RenderToWAV(*outPath, algo, *renderSeconds, *initialVol)
		if err != nil {
			fmt.Fprintln(os.Stderr, "render failed:", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "rendered %s: %.1fs (%d frames) -> %s\n",
			algo.Name(), *renderSeconds, frames, *outPath)
		exportBase := strings.TrimSuffix(*outPath, filepath.Ext(*outPath))
		if *exportMIDI || *exportStems {
			exportAlgo := buildRenderedAlgo(spec, *seed)
			if exporter, ok := exportAlgo.(gen.TuningExporter); ok {
				if *exportMIDI {
					midiPath := exportBase + ".mid"
					if err := exporter.ExportMIDI(midiPath, *renderSeconds); err != nil {
						fmt.Fprintln(os.Stderr, "midi export failed:", err)
						os.Exit(1)
					}
					fmt.Fprintf(os.Stderr, "wrote MIDI -> %s\n", midiPath)
				}
				if *exportStems {
					stemDir := exportBase + "-stems"
					files, err := exporter.ExportStems(stemDir, *renderSeconds, *initialVol)
					if err != nil {
						fmt.Fprintln(os.Stderr, "stem export failed:", err)
						os.Exit(1)
					}
					fmt.Fprintf(os.Stderr, "wrote %d stems -> %s\n", len(files), stemDir)
				}
			} else {
				fmt.Fprintln(os.Stderr, "artifacts skipped: algorithm does not support MIDI/stem export")
			}
		}
		return
	}

	// Build the switchable algorithm list. If we have a SoundFont loaded
	// we expose all SF2-backed genres; otherwise we fall back to the
	// pure-synth variants so cycling never crashes on a nil SoundFont.
	genres, startIdx := buildGenreList(sf, spec.Name)

	// Closure used by the TUI to build a fresh algorithm on swap. We
	// re-seed with the original --seed so the same key stays deterministic
	// across switches.
	buildAlgo := buildRenderedAlgo
	presetLabel := func(s gen.AlgoSpec) string {
		if !s.RequiresSF2 {
			return "synth"
		}
		if *sf2Path != "" {
			return filepath.Base(*sf2Path)
		}
		return pickSFName(sfByPreset, s, *sf2Preset)
	}
	if *playlistOut != "" {
		pl, err := buildPlaylist(*playlistMode, spec, genres, *playlistTracks, *seed, *playlistDur)
		if err != nil {
			fmt.Fprintln(os.Stderr, "playlist:", err)
			os.Exit(2)
		}
		manifest, err := renderPlaylistOut(*playlistOut, pl, *initialVol, buildAlgo, *exportMIDI, *exportStems)
		if err != nil {
			fmt.Fprintln(os.Stderr, "playlist render failed:", err)
			os.Exit(1)
		}
		for _, track := range manifest.Tracks {
			fmt.Fprintf(os.Stderr, "rendered %02d/%02d %s -> %s\n",
				track.Index, manifest.TrackCount, track.Algo, track.Path)
			if track.MIDIPath != "" {
				fmt.Fprintf(os.Stderr, "  midi  -> %s\n", track.MIDIPath)
			}
			if track.StemDir != "" {
				fmt.Fprintf(os.Stderr, "  stems -> %s\n", track.StemDir)
			}
		}
		fmt.Fprintf(os.Stderr, "wrote %s (%d tracks, %.1fs total)\n",
			filepath.Join(*playlistOut, "manifest.json"), manifest.TrackCount, manifest.TotalDurationS)
		return
	}
	liveAlgo := gen.WrapDebugStatus(buildAlgo(spec, *seed), presetLabel(spec))
	ring := scope.NewRing(4096)
	root := audio.NewRoot(liveAlgo, ring)
	root.SetSeed(*seed)
	root.SetVolume(*initialVol)
	buildFn := func(s gen.AlgoSpec) gen.Algorithm {
		return gen.WrapDebugStatus(buildAlgo(s, *seed), presetLabel(s))
	}

	model := tui.New(ring, root, liveAlgo.Name(), "Cmin", *seed, *initialVol).
		WithDebug(*debugView).
		WithSwitcher(genres, startIdx, buildFn)

	// Optional playlist. When set, the TUI auto-advances tracks with a
	// crossfade. The first track is whichever genre is currently playing —
	// we leave the speaker alone and just arm the timer for the next swap.
	if *playlistMode != "" {
		pl, err := buildPlaylist(*playlistMode, spec, genres, *playlistTracks, *seed, *playlistDur)
		if err != nil {
			fmt.Fprintln(os.Stderr, "playlist:", err)
			os.Exit(2)
		}
		// 2s crossfade at 44.1 kHz between tracks.
		model = model.WithPlaylist(pl, 0, 88200)
		fmt.Fprintf(os.Stderr, "playlist %q · %d tracks · %s each\n",
			pl.Name, len(pl.Tracks), *playlistDur)
	}

	// Start the live audio backend asynchronously so a bad CoreAudio/default-
	// device state does not block the TUI from launching.
	sr := beep.SampleRate(44100)
	live := audio.StartLive(root, sr, sr.N(time.Second/20), 3*time.Second)
	defer live.Close()

	p := tea.NewProgram(model, tea.WithAltScreen())
	go func() {
		for state := range live.States() {
			p.Send(state)
		}
	}()
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tui error:", err)
		os.Exit(1)
	}
}

// neededPresets returns the deduped list of SoundFont preset names the app
// must have available given the strategy. In "single" mode, that's just the
// user's chosen preset. In "optimal", we collect every SF2-backed algorithm's
// PreferredSF2 so cycling and playlists can hot-swap to the right SF.
func neededPresets(strategy, fallbackPreset string, initial gen.AlgoSpec) []string {
	switch strategy {
	case "optimal":
		seen := map[string]bool{}
		out := []string{}
		// Ensure we have the initially-chosen preset even if no algo prefers it.
		if fallbackPreset != "" && !seen[fallbackPreset] {
			seen[fallbackPreset] = true
			out = append(out, fallbackPreset)
		}
		for _, name := range gen.AllAlgoNames() {
			s, _ := gen.Resolve(name)
			if !s.RequiresSF2 || s.PreferredSF2 == "" {
				continue
			}
			if !seen[s.PreferredSF2] {
				seen[s.PreferredSF2] = true
				out = append(out, s.PreferredSF2)
			}
		}
		return out
	default: // "single" or anything else
		return []string{fallbackPreset}
	}
}

// pickSF returns the SoundFont to use for `s` from the preloaded map. Falls
// back to the user's --sf2-preset when the algo's preferred SF isn't loaded,
// then to whatever's there. Empty map → nil (caller should have ensured the
// algo doesn't require an SF in that case).
func pickSF(by map[string]*meltysynth.SoundFont, s gen.AlgoSpec, fallback string) *meltysynth.SoundFont {
	if !s.RequiresSF2 {
		return nil
	}
	if sf, ok := by[s.PreferredSF2]; ok {
		return sf
	}
	if sf, ok := by[fallback]; ok {
		return sf
	}
	for _, sf := range by {
		return sf
	}
	return nil
}

func pickSFName(by map[string]*meltysynth.SoundFont, s gen.AlgoSpec, fallback string) string {
	if !s.RequiresSF2 {
		return "synth"
	}
	if _, ok := by[s.PreferredSF2]; ok && s.PreferredSF2 != "" {
		return s.PreferredSF2
	}
	if _, ok := by[fallback]; ok && fallback != "" {
		return fallback
	}
	for name := range by {
		if name != "" {
			return name
		}
	}
	if fallback != "" {
		return fallback
	}
	return "sf2"
}

// buildPlaylist constructs the requested playlist. mode is "same" or "mixed".
// In "same" mode, all tracks use the given spec with different seeds. In
// "mixed" mode, tracks are randomly drawn from the available genre list.
// The first track keeps the currently-playing seed/spec so playback doesn't
// jolt at startup.
func buildPlaylist(mode string, currentSpec gen.AlgoSpec, available []gen.AlgoSpec,
	count int, baseSeed int64, dur time.Duration) (*gen.Playlist, error) {
	if count < 1 {
		return nil, fmt.Errorf("--playlist-tracks must be >= 1, got %d", count)
	}
	var pl gen.Playlist
	switch mode {
	case "same":
		pl = gen.SameGenrePlaylist(currentSpec, count, baseSeed, dur)
	case "mixed":
		if len(available) == 0 {
			return nil, fmt.Errorf("no algorithms available for a mixed playlist")
		}
		pl = gen.MixedPlaylist(available, count, baseSeed, dur)
	default:
		return nil, fmt.Errorf("unknown --playlist mode %q (want 'same' or 'mixed')", mode)
	}
	// Pin track 0 to the currently-playing spec + seed so the initial swap
	// happens at the *first* duration boundary, not immediately at startup.
	pl.Tracks[0] = gen.Track{Spec: currentSpec, Seed: baseSeed, Duration: dur}
	return &pl, nil
}

// buildGenreList returns the ordered list of algorithms the TUI can cycle
// through, filtered by SoundFont availability, plus the index of the
// currently-playing algorithm. Falls back to the first entry if the current
// name isn't in the filtered list.
func buildGenreList(sf *meltysynth.SoundFont, currentName string) ([]gen.AlgoSpec, int) {
	var out []gen.AlgoSpec
	for _, name := range gen.AllAlgoNames() {
		spec, _ := gen.Resolve(name)
		if spec.RequiresSF2 && sf == nil {
			continue
		}
		// Hide the -synth siblings when a SoundFont is loaded — they
		// would just duplicate the genre list and confuse cycling.
		if sf != nil && !spec.RequiresSF2 {
			continue
		}
		out = append(out, spec)
	}
	idx := 0
	for i, s := range out {
		if s.Name == currentName {
			idx = i
			break
		}
	}
	return out, idx
}

// loadIR resolves an --ir argument to an actual impulse response. Accepts:
//   - "room"      — short synthetic room (~80 ms early reflections)
//   - "hall"      — synthetic ~1.5 s concert hall
//   - "cathedral" — synthetic ~3.5 s cathedral (longest tail)
//   - "plate"     — synthetic ~2 s plate-style smooth dense reverb
//   - "synthetic" — alias for "room" (legacy)
//   - any other string is treated as a path to a 16-bit PCM WAV file.
//
// Returns (IR samples, human-readable label, error).
func loadIR(arg string, seed int64) ([]float64, string, error) {
	switch arg {
	case "room", "synthetic":
		return synth.SyntheticRoomIR(0.08), "room", nil
	case "hall":
		return synth.SyntheticHallIR(seed), "hall", nil
	case "cathedral":
		return synth.SyntheticCathedralIR(seed), "cathedral", nil
	case "plate":
		return synth.SyntheticPlateIR(seed), "plate", nil
	default:
		ir, err := audio.ReadIR(arg)
		if err != nil {
			return nil, "", err
		}
		return ir, arg, nil
	}
}

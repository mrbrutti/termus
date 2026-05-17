package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gopxl/beep/v2"
	"github.com/sinshu/go-meltysynth/meltysynth"

	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
	"github.com/mrbrutti/termus/internal/sf2"
	"github.com/mrbrutti/termus/internal/synth"
	"github.com/mrbrutti/termus/internal/track"
	"github.com/mrbrutti/termus/internal/tui"
)

const (
	defaultRenderSeconds  = 180.0
	defaultPlaylistTracks = 6
)

type silentAlgorithm struct{}

func (silentAlgorithm) Name() string               { return "silent" }
func (silentAlgorithm) Seed(int64)                 {}
func (silentAlgorithm) Next(left, right []float64) {}

func startupLabel(spec gen.AlgoSpec) string {
	label := spec.Label()
	if spec.Name == "" || strings.EqualFold(label, spec.Name) {
		return label
	}
	if label == "" {
		return spec.Name
	}
	return fmt.Sprintf("%s · %s", label, spec.Name)
}

func mListeningModeLabel(mode gen.ListeningMode) string {
	switch mode {
	case gen.ListeningModeAlbumSide:
		return "album side"
	case gen.ListeningModeHourStream:
		return "hour stream"
	case gen.ListeningModeRadio:
		return "radio"
	default:
		return "endless"
	}
}

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
			"'pro' (load each algorithm's preferred preset) | "+
			"'max' (load the full curated catalog). 'optimal' is accepted as a legacy alias for 'pro'")
	irPath := flag.String("ir", "", "convolution IR: WAV file path, or preset name: room | hall | cathedral | plate")
	irWet := flag.Float64("ir-wet", 0.40, "convolution wet mix 0..1 when --ir is provided")
	playlistMode := flag.String("playlist", "",
		"playlist mode: 'same' (multiple seeds of --algo) | 'mixed' (random genres). "+
			"Default empty = single track.")
	playlistTracks := flag.Int("playlist-tracks", defaultPlaylistTracks, "number of tracks in the playlist")
	playlistDur := flag.Duration("playlist-duration", 5*time.Minute,
		"how long each playlist track plays before the crossfade")
	listenModeName := flag.String("listen-mode", string(gen.ListeningModeEndless),
		"listening mode: endless | album-side | hour-stream | radio")
	trackName := flag.String("track", "", "track id (e.g. lofi/soft-tape-rain-bus) or path to a .tm track file")
	completionShell := flag.String("completion", "", "print a shell completion script (zsh|bash)")
	completeTrackPrefix := flag.String("complete-track-prefix", "", "internal: print track ids matching a prefix")
	outPath := flag.String("out", "", "render directly to a WAV file instead of launching live playback")
	playlistOut := flag.String("playlist-out", "", "render a playlist to a directory of WAVs plus manifest.json")
	exportStems := flag.Bool("stems", false, "with --out/--playlist-out, also export per-stem WAVs when supported")
	exportMIDI := flag.Bool("midi", false, "with --out/--playlist-out, also export captured MIDI files when supported")
	debugView := flag.Bool("debug", false, "show the musical debug inspector in the TUI")
	renderSeconds := flag.Float64("seconds", defaultRenderSeconds, "render duration in seconds when --out is provided")
	flag.Parse()

	visited := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { visited[f.Name] = true })

	if visited["completion"] {
		if err := printCompletion(*completionShell); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		return
	}
	if visited["complete-track-prefix"] {
		entries, err := discoverTracks()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		for _, entry := range entries {
			if strings.HasPrefix(entry.ID, *completeTrackPrefix) {
				fmt.Println(entry.ID)
			}
		}
		return
	}

	listenMode, ok := gen.ResolveListeningMode(*listenModeName)
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown listen mode %q (available: endless, album-side, hour-stream, radio)\n", *listenModeName)
		os.Exit(2)
	}
	sfStrategy, ok := normalizeSF2Strategy(*sf2Strategy)
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown sf2 strategy %q (available: single, pro, max)\n", *sf2Strategy)
		os.Exit(2)
	}

	effectivePlaylistMode := *playlistMode
	effectivePlaylistTracks := *playlistTracks
	effectivePlaylistDuration := *playlistDur
	effectiveRenderSeconds := *renderSeconds
	if *outPath != "" && *renderSeconds == defaultRenderSeconds && listenMode.DefaultRenderSeconds > 0 {
		effectiveRenderSeconds = listenMode.DefaultRenderSeconds
	}
	if listenMode.AutoPlaylistMode != "" && effectivePlaylistMode == "" {
		effectivePlaylistMode = listenMode.AutoPlaylistMode
		if *playlistTracks == defaultPlaylistTracks && listenMode.DefaultPlaylistTracks > 0 {
			effectivePlaylistTracks = listenMode.DefaultPlaylistTracks
		}
		if *playlistDur == 5*time.Minute && listenMode.DefaultPlaylistDuration > 0 {
			effectivePlaylistDuration = listenMode.DefaultPlaylistDuration
		}
	}

	if *initialVol < 0 || *initialVol > 100 {
		fmt.Fprintf(os.Stderr, "volume must be 0..100, got %d\n", *initialVol)
		os.Exit(2)
	}
	if *outPath != "" && effectiveRenderSeconds <= 0 {
		fmt.Fprintf(os.Stderr, "--seconds must be > 0 when --out is used, got %.3f\n", effectiveRenderSeconds)
		os.Exit(2)
	}
	if *outPath != "" && *playlistOut != "" {
		fmt.Fprintln(os.Stderr, "--out and --playlist-out are mutually exclusive")
		os.Exit(2)
	}
	if *outPath != "" && effectivePlaylistMode != "" {
		fmt.Fprintln(os.Stderr, "--out does not support --playlist; use --playlist-out for batch rendering")
		os.Exit(2)
	}
	if *outPath != "" && listenMode.Name == gen.ListeningModeRadio {
		fmt.Fprintln(os.Stderr, "--listen-mode radio requires live playback or --playlist-out")
		os.Exit(2)
	}
	if *playlistOut != "" && effectivePlaylistMode == "" && *trackName == "" {
		fmt.Fprintln(os.Stderr, "--playlist-out requires --playlist same|mixed")
		os.Exit(2)
	}
	if (*exportStems || *exportMIDI) && *outPath == "" && *playlistOut == "" {
		fmt.Fprintln(os.Stderr, "--stems/--midi require --out or --playlist-out")
		os.Exit(2)
	}
	if effectivePlaylistMode != "" && effectivePlaylistDuration <= 0 {
		fmt.Fprintf(os.Stderr, "--playlist-duration must be > 0, got %s\n", effectivePlaylistDuration)
		os.Exit(2)
	}
	if *trackName != "" && effectivePlaylistMode != "" {
		fmt.Fprintln(os.Stderr, "-track already defines its own authored sections; do not combine it with --playlist")
		os.Exit(2)
	}
	if *trackName != "" && *outPath != "" {
		fmt.Fprintln(os.Stderr, "-track currently supports live playback or --playlist-out; use --playlist-out for batch rendering")
		os.Exit(2)
	}

	var (
		spec        gen.AlgoSpec
		activeTrack *track.Compiled
		err         error
	)
	trackEntries, _ := discoverTracks()
	defaultTrackBrowse := *trackName == "" && !visited["algo"] && effectivePlaylistMode == "" && *outPath == "" && *playlistOut == "" && len(trackEntries) > 0
	if *trackName != "" {
		_, activeTrack, err = loadTrackSelection(trackEntries, *trackName, *seed, listenMode.Name)
		if err != nil {
			fmt.Fprintln(os.Stderr, "track load failed:", err)
			os.Exit(2)
		}
		logTrackWarnings(activeTrack)
		if len(activeTrack.Playlist.Tracks) == 0 {
			fmt.Fprintln(os.Stderr, "track has no sections")
			os.Exit(2)
		}
		spec = activeTrack.Playlist.Tracks[0].Spec
		if resolved, ok := gen.ResolveListeningMode(string(activeTrack.Playlist.ListenMode)); ok {
			listenMode = resolved
		}
	} else {
		var ok bool
		spec, ok = gen.Resolve(*algoName)
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown algorithm %q (available: %v)\n",
				*algoName, gen.AllAlgoNames())
			os.Exit(2)
		}
	}
	loadSpec := spec
	if activeTrack != nil {
		for _, track := range activeTrack.Playlist.Tracks {
			if track.Spec.RequiresSF2 {
				loadSpec = track.Spec
				break
			}
		}
	}
	liveRun := *outPath == "" && *playlistOut == ""
	startupNeedsLibrary := defaultTrackBrowse && len(trackEntries) > 0
	startupNeedsSF2 := loadSpec.RequiresSF2 || startupNeedsLibrary
	liveStartupLoader := liveRun && startupNeedsSF2
	var (
		catalog *soundFontCatalog
		sf      *meltysynth.SoundFont
	)
	if liveRun {
		if startupNeedsSF2 && *sf2Path == "" {
			catalog = newSoundFontCatalog(sfStrategy, *sf2Preset)
			gen.SetSF2Runtime(sfStrategy, catalog.snapshot())
		}
	} else {
		catalog, sf, err = loadInitialSoundFontCatalog(loadSpec, sfStrategy, *sf2Preset, *sf2Path, os.Stderr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "sf2 setup failed:", err)
			os.Exit(1)
		}
	}
	var algo gen.Algorithm
	if liveStartupLoader {
		algo = silentAlgorithm{}
	} else {
		algo = spec.Build(sf)
		algo.Seed(*seed)
	}
	controlProfile := gen.DefaultControlProfile()
	profileFor := func(s gen.AlgoSpec, algoSeed int64) gen.ControlProfile {
		if activeTrack == nil {
			return controlProfile
		}
		if base, ok := activeTrack.Profiles[fmt.Sprintf("%s:%d", s.Name, algoSeed)]; ok {
			return mergeProfiles(base, controlProfile)
		}
		return controlProfile
	}
	planFor := func(s gen.AlgoSpec, algoSeed int64) *gen.AuthoredTrackPlan {
		if activeTrack == nil {
			return nil
		}
		if plan, ok := activeTrack.Plans[fmt.Sprintf("%s:%d", s.Name, algoSeed)]; ok {
			cloned := plan
			return &cloned
		}
		return nil
	}
	selectionFor := func(s gen.AlgoSpec, algoSeed int64) gen.SF2Selection {
		if plan := planFor(s, algoSeed); plan != nil {
			return gen.ResolveSF2SelectionForPlan(s, plan, sfStrategy, *sf2Preset)
		}
		return gen.ResolveSF2Selection(s, sfStrategy, *sf2Preset)
	}

	var ir []float64
	var irLabel string
	// Optional convolution IR.
	if *irPath != "" {
		if !spec.RequiresSF2 {
			fmt.Fprintf(os.Stderr, "--ir requires an sf2-mode algorithm; ignoring\n")
		} else {
			loadedIR, label, err := loadIR(*irPath, *seed)
			if err != nil {
				fmt.Fprintln(os.Stderr, "ir load failed:", err)
				os.Exit(1)
			}
			ir = loadedIR
			irLabel = label
			if !liveStartupLoader {
				if rev, ok := algo.(gen.SF2Reverberator); ok {
					rev.SetReverbIR(ir, *irWet)
				}
				fmt.Fprintf(os.Stderr, "IR %s: %d samples (%.1f ms) at wet=%.2f\n",
					irLabel, len(ir), float64(len(ir))*1000.0/44100.0, *irWet)
			}
		}
	}
	buildRenderedAlgo := func(s gen.AlgoSpec, algoSeed int64) gen.Algorithm {
		chosen := sf
		if catalog != nil && s.RequiresSF2 {
			selection := selectionFor(s, algoSeed)
			if err := catalog.ensurePresets(io.Discard, selection.Presets); err != nil {
				logCatalogEnsureError(s, err)
			}
			chosen = catalog.Lookup(selection.Primary)
			if chosen == nil {
				chosen = catalog.Pick(s)
			}
			gen.SetSF2RuntimeWithRoutes(sfStrategy, catalog.snapshot(), map[string]map[int32]string{s.Name: selection.Routes})
		}
		var a gen.Algorithm
		if plan := planFor(s, algoSeed); plan != nil {
			a = gen.NewAuthoredTrack(s, chosen, *plan)
		} else {
			a = s.Build(chosen)
		}
		a = gen.ConfigureControlProfile(a, profileFor(s, algoSeed))
		a.Seed(algoSeed)
		if len(ir) > 0 {
			if rev, ok := a.(gen.SF2Reverberator); ok {
				rev.SetReverbIR(ir, *irWet)
			}
		}
		return a
	}
	buildLiveAlgo := func(s gen.AlgoSpec, algoSeed int64) gen.Algorithm {
		return gen.ApplyControlProfile(buildRenderedAlgo(s, algoSeed), profileFor(s, algoSeed))
	}
	if *outPath != "" {
		renderAlgo := buildRenderedAlgo(spec, *seed)
		plan := audio.PlanRender(renderAlgo, effectiveRenderSeconds)
		frames, err := audio.RenderToWAVWithPlan(*outPath, renderAlgo, plan, *initialVol)
		if err != nil {
			fmt.Fprintln(os.Stderr, "render failed:", err)
			os.Exit(1)
		}
		actualSeconds := float64(frames) / 44100.0
		if plan.SnapLabel != "" {
			fmt.Fprintf(os.Stderr, "rendered %s: requested %.1fs, landed on %s, wrote %.1fs (%d frames) -> %s\n",
				renderAlgo.Name(), effectiveRenderSeconds, plan.SnapLabel, actualSeconds, frames, *outPath)
		} else {
			fmt.Fprintf(os.Stderr, "rendered %s: requested %.1fs, wrote %.1fs (%d frames) -> %s\n",
				renderAlgo.Name(), effectiveRenderSeconds, actualSeconds, frames, *outPath)
		}
		exportBase := strings.TrimSuffix(*outPath, filepath.Ext(*outPath))
		if *exportMIDI || *exportStems {
			exportAlgo := buildRenderedAlgo(spec, *seed)
			if exporter, ok := exportAlgo.(gen.TuningExporter); ok {
				if *exportMIDI {
					midiPath := exportBase + ".mid"
					if err := exporter.ExportMIDI(midiPath, plan.DurationSeconds()); err != nil {
						fmt.Fprintln(os.Stderr, "midi export failed:", err)
						os.Exit(1)
					}
					fmt.Fprintf(os.Stderr, "wrote MIDI -> %s\n", midiPath)
				}
				if *exportStems {
					stemDir := exportBase + "-stems"
					files, err := exporter.ExportStems(stemDir, plan.DurationSeconds(), *initialVol)
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
	genres, startIdx := buildGenreList(sf != nil || catalog != nil, spec.Name)

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
		selection := selectionFor(s, *seed)
		if sfStrategy == "max" {
			return "max"
		}
		if selection.Primary != "" {
			return selection.Primary
		}
		if catalog == nil {
			return "sf2"
		}
		return catalog.PickName(s)
	}
	if *playlistOut != "" {
		var pl *gen.Playlist
		if activeTrack != nil {
			pl = &activeTrack.Playlist
		} else {
			pl, err = buildPlaylist(effectivePlaylistMode, spec, genres, effectivePlaylistTracks, *seed, effectivePlaylistDuration, listenMode.Name)
			if err != nil {
				fmt.Fprintln(os.Stderr, "playlist:", err)
				os.Exit(2)
			}
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
	liveAlgo := gen.WrapDebugStatus(gen.ApplyControlProfile(algo, controlProfile), presetLabel(spec))
	ring := scope.NewRing(4096)
	root := audio.NewRoot(liveAlgo, ring)
	root.SetSeed(*seed)
	root.SetVolume(*initialVol)
	buildFn := func(s gen.AlgoSpec, algoSeed int64) gen.Algorithm {
		return gen.WrapDebugStatus(buildLiveAlgo(s, algoSeed), presetLabel(s))
	}
	trackLoader := func(id string) (*gen.Playlist, string, error) {
		entry, compiled, err := loadTrackSelection(trackEntries, id, *seed, listenMode.Name)
		if err != nil {
			return nil, "", err
		}
		_ = entry
		activeTrack = compiled
		logTrackWarnings(compiled)
		if len(compiled.Playlist.Tracks) == 0 {
			return nil, "", fmt.Errorf("track has no sections")
		}
		modeLabel := mListeningModeLabel(compiled.Playlist.ListenMode)
		return &compiled.Playlist, modeLabel, nil
	}
	trackNav := make([]tui.TrackNavEntry, 0, len(trackEntries))
	for _, entry := range trackEntries {
		trackNav = append(trackNav, tui.TrackNavEntry{
			ID:          entry.ID,
			Style:       entry.Style,
			Title:       entry.Title,
			Description: entry.Description,
			Tags:        append([]string(nil), entry.Tags...),
			Key:         entry.Key,
			Tempo:       entry.Tempo,
			ListenMode:  entry.ListenMode,
			Sections:    append([]string(nil), entry.Sections...),
		})
	}
	openTrackBrowser := defaultTrackBrowse

	model := tui.New(ring, root, spec.Label(), "Cmin", *seed, *initialVol).
		WithDebug(*debugView).
		WithListeningMode(listenMode.Label).
		WithControlProfile(&controlProfile).
		WithExportController(makeTUIExporter(buildAlgo, *initialVol)).
		WithSwitcher(genres, startIdx, buildFn).
		WithTrackBrowser(trackNav, trackLoader, openTrackBrowser)
	if liveStartupLoader {
		model = model.WithStartupLoading(startupLoadTitle(defaultTrackBrowse, sfStrategy, spec), "preparing soundfonts", 0)
	}

	// Optional playlist. When set, the TUI auto-advances tracks with a
	// crossfade. The first track is whichever genre is currently playing —
	// we leave the speaker alone and just arm the timer for the next swap.
	if activeTrack != nil {
		pl := &activeTrack.Playlist
		model = model.WithPlaylist(pl, 0, 88200)
		fmt.Fprintf(os.Stderr, "track %q · %d sections · mode %s\n",
			pl.Name, len(pl.Tracks), pl.ListenMode)
	} else if effectivePlaylistMode != "" {
		pl, err := buildPlaylist(effectivePlaylistMode, spec, genres, effectivePlaylistTracks, *seed, effectivePlaylistDuration, listenMode.Name)
		if err != nil {
			fmt.Fprintln(os.Stderr, "playlist:", err)
			os.Exit(2)
		}
		// 2s crossfade at 44.1 kHz between tracks.
		model = model.WithPlaylist(pl, 0, 88200)
		fmt.Fprintf(os.Stderr, "playlist %q · %d tracks · %s each · mode %s\n",
			pl.Name, len(pl.Tracks), effectivePlaylistDuration, listenMode.Label)
	}

	sr := beep.SampleRate(44100)
	var (
		liveMu sync.Mutex
		live   *audio.LiveBackend
	)
	model = model.WithAudioControl(&tui.AudioControl{
		Retry: func() {
			liveMu.Lock()
			current := live
			liveMu.Unlock()
			if current != nil {
				current.Retry()
			}
		},
		RenderOnly: func() {
			liveMu.Lock()
			current := live
			liveMu.Unlock()
			if current != nil {
				current.SetRenderOnly()
			}
		},
	})
	p := tea.NewProgram(model, tea.WithAltScreen())
	startLive := func() {
		next := audio.StartLive(root, sr, sr.N(time.Second/20), 3*time.Second)
		liveMu.Lock()
		live = next
		liveMu.Unlock()
		go func() {
			for state := range next.States() {
				p.Send(state)
			}
		}()
	}
	if liveStartupLoader {
		go func() {
			title := startupLoadTitle(defaultTrackBrowse, sfStrategy, spec)
			update := func(progress catalogLoadUpdate) {
				p.Send(tui.StartupLoadMsg{
					Title:   title,
					Detail:  progress.Detail,
					Percent: progress.Percent,
				})
			}
			if *sf2Path != "" {
				p.Send(tui.StartupLoadMsg{Title: title, Detail: "opening " + filepath.Base(*sf2Path), Percent: 0.15})
				loaded, err := sf2.Open(*sf2Path)
				if err != nil {
					p.Send(audio.BackendState{Kind: audio.BackendStateInitFailed, Detail: err.Error()})
					p.Send(tui.StartupLoadMsg{Title: title, Detail: "load failed: " + err.Error(), Percent: 0, Done: true})
					return
				}
				sf = loaded
				p.Send(tui.StartupLoadMsg{Title: title, Detail: "starting audio...", Percent: 1})
			} else {
				presets := startupPresetNames(sfStrategy, *sf2Preset)
				if startupNeedsLibrary && sfStrategy == "max" {
					presets = sf2.AllPresetNames()
				}
				if err := catalog.ensurePresetsParallel(presets, startupLoadConcurrency(sfStrategy), update); err != nil {
					p.Send(audio.BackendState{Kind: audio.BackendStateInitFailed, Detail: err.Error()})
					p.Send(tui.StartupLoadMsg{Title: title, Detail: "load failed: " + err.Error(), Percent: 0, Done: true})
					return
				}
				p.Send(tui.StartupLoadMsg{Title: title, Detail: "starting audio...", Percent: 1})
			}
			if !defaultTrackBrowse {
				readyAlgo := gen.WrapDebugStatus(buildLiveAlgo(spec, *seed), presetLabel(spec))
				root.SwapAlgorithmFade(readyAlgo, 0)
			}
			startLive()
		}()
	} else {
		startLive()
	}
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tui error:", err)
		os.Exit(1)
	}
	liveMu.Lock()
	if live != nil {
		live.Close()
	}
	liveMu.Unlock()
}

// neededPresets returns the deduped list of SoundFont preset names the app
// must have available given the strategy. In "single" mode, that's just the
// user's chosen preset. In "pro", we collect the inventory-resolved primary
// preset for every SF2-backed algorithm so cycling and playlists can hot-swap
// to the right bank. In "max", we load the full curated catalog.
func neededPresets(strategy, fallbackPreset string, initial gen.AlgoSpec) []string {
	switch strategy {
	case "max":
		return sf2.AllPresetNames()
	case "pro":
		seen := map[string]bool{}
		out := []string{}
		// Ensure we have the initially-chosen preset even if no algo prefers it.
		if fallbackPreset != "" && !seen[fallbackPreset] {
			seen[fallbackPreset] = true
			out = append(out, fallbackPreset)
		}
		for _, name := range gen.AllAlgoNames() {
			s, _ := gen.Resolve(name)
			if !s.RequiresSF2 {
				continue
			}
			preset := gen.ProSF2PresetForSpec(s, fallbackPreset)
			if preset == "" {
				continue
			}
			if !seen[preset] {
				seen[preset] = true
				out = append(out, preset)
			}
		}
		return out
	default: // "single"
		return []string{fallbackPreset}
	}
}

func normalizeSF2Strategy(name string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "single":
		return "single", true
	case "pro", "optimal":
		return "pro", true
	case "max":
		return "max", true
	default:
		return "", false
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
	if primary := gen.ProSF2PresetForSpec(s, fallback); primary != "" {
		if sf, ok := by[primary]; ok {
			return sf
		}
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
	if primary := gen.ProSF2PresetForSpec(s, fallback); primary != "" {
		if _, ok := by[primary]; ok {
			return primary
		}
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
	count int, baseSeed int64, dur time.Duration, listenMode gen.ListeningMode) (*gen.Playlist, error) {
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
	pl.ListenMode = listenMode
	return &pl, nil
}

// buildGenreList returns the ordered list of algorithms the TUI can cycle
// through, filtered by SoundFont availability, plus the index of the
// currently-playing algorithm. Falls back to the first entry if the current
// name isn't in the filtered list.
func buildGenreList(sfAvailable bool, currentName string) ([]gen.AlgoSpec, int) {
	var out []gen.AlgoSpec
	for _, name := range gen.AllAlgoNames() {
		spec, _ := gen.Resolve(name)
		if spec.RequiresSF2 && !sfAvailable {
			continue
		}
		// Hide the -synth siblings when a SoundFont is loaded — they
		// would just duplicate the genre list and confuse cycling.
		if sfAvailable && !spec.RequiresSF2 {
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

func startupLoadTitle(trackBrowse bool, strategy string, spec gen.AlgoSpec) string {
	if trackBrowse {
		switch strategy {
		case "max":
			return "Loading Track Library · MAX"
		case "pro":
			return "Loading Track Library · PRO"
		default:
			return "Loading Track Library"
		}
	}
	switch strategy {
	case "max":
		return fmt.Sprintf("Loading MAX palette · %s", startupLabel(spec))
	case "pro":
		return fmt.Sprintf("Loading PRO palette · %s", startupLabel(spec))
	default:
		return fmt.Sprintf("Loading SF2 · %s", startupLabel(spec))
	}
}

func startupPresetNames(strategy, fallback string) []string {
	switch strategy {
	case "max":
		return sf2.AllPresetNames()
	case "single":
		return []string{fallback}
	default:
		seen := map[string]bool{}
		out := make([]string, 0)
		for _, name := range gen.AllGenreNames() {
			spec, ok := gen.Resolve(name)
			if !ok || !spec.RequiresSF2 {
				continue
			}
			preset := gen.ProSF2PresetForSpec(spec, fallback)
			if preset == "" || seen[preset] {
				continue
			}
			seen[preset] = true
			out = append(out, preset)
		}
		if len(out) == 0 && fallback != "" {
			return []string{fallback}
		}
		return out
	}
}

func startupLoadConcurrency(strategy string) int {
	switch strategy {
	case "max":
		return 3
	case "pro":
		return 2
	default:
		return 1
	}
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

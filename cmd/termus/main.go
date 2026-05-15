package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
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
	irPath := flag.String("ir", "", "convolution IR: WAV file path, or preset name: room | hall | cathedral | plate")
	irWet := flag.Float64("ir-wet", 0.40, "convolution wet mix 0..1 when --ir is provided")
	flag.Parse()

	if *initialVol < 0 || *initialVol > 100 {
		fmt.Fprintf(os.Stderr, "volume must be 0..100, got %d\n", *initialVol)
		os.Exit(2)
	}

	spec, ok := gen.Resolve(*algoName)
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown algorithm %q (available: %v)\n",
			*algoName, gen.AllAlgoNames())
		os.Exit(2)
	}
	var sf *meltysynth.SoundFont
	if spec.RequiresSF2 {
		path := *sf2Path
		if path == "" {
			preset, ok := sf2.Presets[*sf2Preset]
			if !ok {
				fmt.Fprintf(os.Stderr, "unknown --sf2-preset %q\n", *sf2Preset)
				os.Exit(2)
			}
			fmt.Fprintf(os.Stderr, "preparing SoundFont preset %q (%s, ~%d MB on first run)...\n",
				preset.Name, preset.FileName, preset.SizeMB)
			p, err := sf2.EnsurePreset(*sf2Preset, func(done, total int64) {
				if total > 0 {
					fmt.Fprintf(os.Stderr, "\r  %d / %d bytes", done, total)
				}
			})
			fmt.Fprintln(os.Stderr)
			if err != nil {
				fmt.Fprintln(os.Stderr, "sf2 setup failed:", err)
				os.Exit(1)
			}
			path = p
		}
		var err error
		sf, err = sf2.Open(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "sf2 open failed:", err)
			os.Exit(1)
		}
	}
	algo := spec.Build(sf)
	algo.Seed(*seed)

	// Optional convolution IR.
	if *irPath != "" {
		rev, ok := algo.(gen.SF2Reverberator)
		if !ok {
			fmt.Fprintf(os.Stderr, "--ir requires an sf2-mode algorithm; ignoring\n")
		} else {
			ir, label, err := loadIR(*irPath, *seed)
			if err != nil {
				fmt.Fprintln(os.Stderr, "ir load failed:", err)
				os.Exit(1)
			}
			rev.SetReverbIR(ir, *irWet)
			fmt.Fprintf(os.Stderr, "IR %s: %d samples (%.1f ms) at wet=%.2f\n",
				label, len(ir), float64(len(ir))*1000.0/44100.0, *irWet)
		}
	}
	ring := scope.NewRing(4096)
	root := audio.NewRoot(algo, ring)
	root.SetSeed(*seed)
	root.SetVolume(*initialVol)

	// Initialize beep speaker. The buffer must be big enough that one Stream
	// call can be produced before the speaker drains its previous chunk.
	// time.Second/60 (≈17ms) was too tight for Eno's per-sample work on
	// real hardware — caused ~25% underrun → silent output. time.Second/20
	// (50ms) gives comfortable headroom. Latency is unnoticeable for ambient.
	sr := beep.SampleRate(44100)
	if err := speaker.Init(sr, sr.N(time.Second/20)); err != nil {
		fmt.Fprintln(os.Stderr, "audio init failed:", err)
		os.Exit(1)
	}
	defer speaker.Close()
	speaker.Play(root)

	// Build the switchable algorithm list. If we have a SoundFont loaded
	// we expose all SF2-backed genres; otherwise we fall back to the
	// pure-synth variants so cycling never crashes on a nil SoundFont.
	genres, startIdx := buildGenreList(sf, spec.Name)

	// Closure used by the TUI to build a fresh algorithm on swap. We
	// re-seed with the original --seed so the same key stays deterministic
	// across switches.
	buildFn := func(s gen.AlgoSpec) gen.Algorithm {
		a := s.Build(sf)
		a.Seed(*seed)
		// If we had an IR loaded, propagate it onto the new algorithm so
		// the room/hall/etc. carries across the switch.
		if *irPath != "" {
			if rev, ok := a.(gen.SF2Reverberator); ok {
				if ir, _, err := loadIR(*irPath, *seed); err == nil {
					rev.SetReverbIR(ir, *irWet)
				}
			}
		}
		return a
	}

	// Launch TUI.
	model := tui.New(ring, root, algo.Name(), "Cmin", *seed, *initialVol).
		WithSwitcher(genres, startIdx, buildFn)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tui error:", err)
		os.Exit(1)
	}
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

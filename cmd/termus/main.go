package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"

	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
	"github.com/mrbrutti/termus/internal/sf2"
	"github.com/mrbrutti/termus/internal/synth"
	"github.com/mrbrutti/termus/internal/tui"
)

func main() {
	seed := flag.Int64("seed", time.Now().UnixNano(), "RNG seed (default: time-based)")
	algoName := flag.String("algo", "eno",
		"algorithm: eno | drone | glass | pentatonic | markov | sf2 | "+
			"eno-sf2 | drone-sf2 | glass-sf2 | pentatonic-sf2 | markov-sf2 | phase | chill")
	initialVol := flag.Int("volume", 70, "initial volume 0..100")
	sf2Path := flag.String("sf2", "", "path to SoundFont file for the sf2 algorithm (default: auto-download GeneralUser-GS.sf2, ~32 MB)")
	irPath := flag.String("ir", "", "convolution IR: WAV file path, or preset name: room | hall | cathedral | plate")
	irWet := flag.Float64("ir-wet", 0.40, "convolution wet mix 0..1 when --ir is provided")
	flag.Parse()

	if *initialVol < 0 || *initialVol > 100 {
		fmt.Fprintf(os.Stderr, "volume must be 0..100, got %d\n", *initialVol)
		os.Exit(2)
	}

	var algo gen.Algorithm
	switch *algoName {
	case "eno":
		algo = gen.NewEno()
	case "drone":
		algo = gen.NewDrone()
	case "glass":
		algo = gen.NewGlass()
	case "pentatonic":
		algo = gen.NewPentatonic()
	case "markov":
		algo = gen.NewMarkov()
	case "sf2", "eno-sf2", "drone-sf2", "glass-sf2", "pentatonic-sf2", "markov-sf2", "phase", "chill":
		// Resolve the SoundFont: --sf2 overrides, otherwise auto-download.
		path := *sf2Path
		if path == "" {
			fmt.Fprintln(os.Stderr, "preparing SoundFont (GeneralUser-GS.sf2, ~32 MB on first run, cached thereafter)...")
			p, err := sf2.EnsureDefault(func(done, total int64) {
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
		sf, err := sf2.Open(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "sf2 open failed:", err)
			os.Exit(1)
		}
		switch *algoName {
		case "sf2":
			algo = gen.NewSF2(sf)
		case "eno-sf2":
			algo = gen.NewSF2Eno(sf)
		case "drone-sf2":
			algo = gen.NewSF2Drone(sf)
		case "glass-sf2":
			algo = gen.NewSF2Glass(sf)
		case "pentatonic-sf2":
			algo = gen.NewSF2Pentatonic(sf)
		case "markov-sf2":
			algo = gen.NewSF2Markov(sf)
		case "phase":
			algo = gen.NewPhase(sf)
		case "chill":
			algo = gen.NewChill(sf)
		}
	default:
		fmt.Fprintf(os.Stderr,
			"unknown algorithm %q (eno, drone, glass, pentatonic, markov, sf2, "+
				"eno-sf2, drone-sf2, glass-sf2, pentatonic-sf2, markov-sf2)\n", *algoName)
		os.Exit(2)
	}
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

	// Launch TUI.
	model := tui.New(ring, root, algo.Name(), "Cmin", *seed, *initialVol)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tui error:", err)
		os.Exit(1)
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

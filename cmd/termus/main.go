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
	"github.com/mrbrutti/termus/internal/tui"
)

func main() {
	seed := flag.Int64("seed", time.Now().UnixNano(), "RNG seed (default: time-based)")
	algoName := flag.String("algo", "eno", "algorithm name (v1: eno)")
	initialVol := flag.Int("volume", 70, "initial volume 0..100")
	flag.Parse()

	if *algoName != "eno" {
		fmt.Fprintf(os.Stderr, "unknown algorithm %q (v1 only supports 'eno')\n", *algoName)
		os.Exit(2)
	}
	if *initialVol < 0 || *initialVol > 100 {
		fmt.Fprintf(os.Stderr, "volume must be 0..100, got %d\n", *initialVol)
		os.Exit(2)
	}

	// Wire algorithm, scope, audio root.
	algo := gen.NewEno()
	algo.Seed(*seed)
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

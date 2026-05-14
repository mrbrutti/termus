// termus-debug: standalone diagnostic for the audio path.
// Renders 5 seconds of Eno-drift output to a WAV file with NO speaker, NO TUI.
// If the resulting WAV is audible, the synthesis path works.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/sf2"
	"github.com/mrbrutti/termus/internal/synth"
)

func main() {
	seed := flag.Int64("seed", 42, "seed")
	seconds := flag.Float64("seconds", 5.0, "duration to render")
	out := flag.String("out", "termus-debug.wav", "output WAV path")
	algoName := flag.String("algo", "eno",
		"algorithm: eno|drone|glass|pentatonic|markov|sf2|"+
			"eno-sf2|drone-sf2|glass-sf2|pentatonic-sf2|markov-sf2")
	sf2Path := flag.String("sf2", "", "SoundFont path for the sf2 algorithm (default: auto-download)")
	irPath := flag.String("ir", "", "convolution IR WAV path, or 'synthetic'")
	irWet := flag.Float64("ir-wet", 0.40, "convolution wet mix 0..1")
	flag.Parse()

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
	case "sf2", "eno-sf2", "drone-sf2", "glass-sf2", "pentatonic-sf2", "markov-sf2", "phase":
		path := *sf2Path
		if path == "" {
			p, err := sf2.EnsureDefault(nil)
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
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown algorithm %q\n", *algoName)
		os.Exit(2)
	}
	algo.Seed(*seed)

	if *irPath != "" {
		if rev, ok := algo.(gen.SF2Reverberator); ok {
			ir, label, err := loadIR(*irPath, *seed)
			if err != nil {
				fmt.Fprintln(os.Stderr, "ir load failed:", err)
				os.Exit(1)
			}
			rev.SetReverbIR(ir, *irWet)
			fmt.Fprintf(os.Stderr, "IR %s: %d samples (%.1f ms)\n",
				label, len(ir), float64(len(ir))*1000.0/44100.0)
		} else {
			fmt.Fprintf(os.Stderr, "warning: --ir requires an sf2-mode algorithm; ignoring\n")
		}
	}

	const sr = 44100
	totalFrames := int(*seconds * float64(sr))
	chunkFrames := 4410 // 100ms chunks

	w, err := audio.NewWAVWriter(*out, sr, 2)
	if err != nil {
		fmt.Fprintln(os.Stderr, "wav open:", err)
		os.Exit(1)
	}

	l := make([]float64, chunkFrames)
	r := make([]float64, chunkFrames)
	frames := make([][2]float64, chunkFrames)

	var maxL, maxR float64
	var sumSq float64
	written := 0
	for written < totalFrames {
		n := chunkFrames
		if totalFrames-written < n {
			n = totalFrames - written
		}
		algo.Next(l[:n], r[:n])
		for i := 0; i < n; i++ {
			frames[i][0] = l[i]
			frames[i][1] = r[i]
			if math.Abs(l[i]) > maxL {
				maxL = math.Abs(l[i])
			}
			if math.Abs(r[i]) > maxR {
				maxR = math.Abs(r[i])
			}
			sumSq += l[i]*l[i] + r[i]*r[i]
		}
		if err := w.Write(frames[:n]); err != nil {
			fmt.Fprintln(os.Stderr, "wav write:", err)
			os.Exit(1)
		}
		written += n
		fmt.Fprintf(os.Stderr, "%5.2fs  peak L=%.3f R=%.3f\n",
			float64(written)/float64(sr), maxL, maxR)
	}
	if err := w.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "wav close:", err)
		os.Exit(1)
	}

	rms := math.Sqrt(sumSq / float64(2*written))
	fmt.Fprintf(os.Stderr, "\nwrote %s — %d frames, peak L=%.3f R=%.3f RMS=%.4f\n",
		*out, written, maxL, maxR, rms)
	fmt.Fprintf(os.Stderr, "play with: afplay %s\n", *out)
}

// loadIR resolves an --ir argument; see cmd/termus for full docs.
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

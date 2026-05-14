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
)

func main() {
	seed := flag.Int64("seed", 42, "seed")
	seconds := flag.Float64("seconds", 5.0, "duration to render")
	out := flag.String("out", "termus-debug.wav", "output WAV path")
	algoName := flag.String("algo", "eno", "algorithm: eno | drone | glass")
	flag.Parse()

	var algo gen.Algorithm
	switch *algoName {
	case "eno":
		algo = gen.NewEno()
	case "drone":
		algo = gen.NewDrone()
	case "glass":
		algo = gen.NewGlass()
	default:
		fmt.Fprintf(os.Stderr, "unknown algorithm %q\n", *algoName)
		os.Exit(2)
	}
	algo.Seed(*seed)

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

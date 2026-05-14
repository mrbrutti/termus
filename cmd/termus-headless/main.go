// termus-headless: plays Eno-drift via the speaker for 10 seconds with NO TUI.
// If this plays sound but `./termus` does not, the TUI integration is the bug.
// If this is also silent, the audio buffer is too small for Eno on this machine.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"sync/atomic"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"

	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
	"github.com/mrbrutti/termus/internal/sf2"
	"github.com/mrbrutti/termus/internal/synth"
)

// debugStreamer wraps a beep.Streamer and records call statistics so we can
// verify whether the audio goroutine is actually pulling samples and what
// they look like when it does.
type debugStreamer struct {
	inner  beep.Streamer
	calls  int64
	frames int64
	peak   float64
	last   float64
	logged atomic.Bool
}

func (d *debugStreamer) Stream(samples [][2]float64) (int, bool) {
	n, ok := d.inner.Stream(samples)
	d.calls++
	d.frames += int64(n)
	for i := 0; i < n; i++ {
		a := math.Abs(samples[i][0])
		if a > d.peak {
			d.peak = a
		}
		d.last = samples[i][0]
	}
	// Log the first call so we can see what the first chunk looks like.
	if !d.logged.Load() {
		d.logged.Store(true)
		first := 0.0
		if n > 0 {
			first = samples[0][0]
		}
		fmt.Fprintf(os.Stderr, "first Stream() call: n=%d ok=%v first=%.6f last=%.6f\n",
			n, ok, first, samples[n-1][0])
	}
	return n, ok
}

func (d *debugStreamer) Err() error { return d.inner.Err() }

func main() {
	seed := flag.Int64("seed", 42, "seed")
	seconds := flag.Int("seconds", 10, "duration")
	bufDivisor := flag.Int("buf", 60, "buffer = SampleRate / this")
	algoName := flag.String("algo", "eno", "algorithm name")
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
	case "sf2", "eno-sf2", "drone-sf2", "glass-sf2", "pentatonic-sf2", "markov-sf2", "phase", "chill":
		p, err := sf2.EnsureDefault(nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, "sf2 setup:", err)
			os.Exit(1)
		}
		sf, err := sf2.Open(p)
		if err != nil {
			fmt.Fprintln(os.Stderr, "sf2 open:", err)
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
		fmt.Fprintf(os.Stderr, "unknown algo %q\n", *algoName)
		os.Exit(2)
	}
	algo.Seed(*seed)
	if *irPath != "" {
		if rev, ok := algo.(gen.SF2Reverberator); ok {
			ir, label, err := loadIR(*irPath, *seed)
			if err != nil {
				fmt.Fprintln(os.Stderr, "ir load:", err)
				os.Exit(1)
			}
			rev.SetReverbIR(ir, *irWet)
			fmt.Fprintf(os.Stderr, "IR %s: %d samples (%.1f ms)\n",
				label, len(ir), float64(len(ir))*1000.0/44100.0)
		}
	}
	ring := scope.NewRing(4096)
	root := audio.NewRoot(algo, ring)
	root.SetSeed(*seed)
	root.SetVolume(100)

	// Wrap root with an instrumented streamer that logs every call.
	dbg := &debugStreamer{inner: root}

	sr := beep.SampleRate(44100)
	bufSize := sr.N(time.Second / time.Duration(*bufDivisor))
	fmt.Fprintf(os.Stderr, "init speaker: sampleRate=%d bufferSize=%d samples (%.1f ms)\n",
		int(sr), bufSize, float64(bufSize)*1000.0/44100.0)
	if err := speaker.Init(sr, bufSize); err != nil {
		fmt.Fprintln(os.Stderr, "speaker init failed:", err)
		os.Exit(1)
	}
	defer speaker.Close()
	speaker.Play(dbg)
	fmt.Fprintf(os.Stderr, "playing %s for %d seconds at volume 100...\n", *algoName, *seconds)
	time.Sleep(time.Duration(*seconds) * time.Second)
	fmt.Fprintf(os.Stderr, "done. Stream called %d times, total %d frames, peak |sample|=%.3f, last sample=%.3f\n",
		dbg.calls, dbg.frames, dbg.peak, dbg.last)
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

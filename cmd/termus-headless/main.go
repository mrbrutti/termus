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
	flag.Parse()

	algo := gen.NewEno()
	algo.Seed(*seed)
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
	fmt.Fprintf(os.Stderr, "playing Eno for %d seconds at volume 100...\n", *seconds)
	time.Sleep(time.Duration(*seconds) * time.Second)
	fmt.Fprintf(os.Stderr, "done. Stream called %d times, total %d frames, peak |sample|=%.3f, last sample=%.3f\n",
		dbg.calls, dbg.frames, dbg.peak, dbg.last)
}

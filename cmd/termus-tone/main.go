// termus-tone: play a 440Hz sine for 3 seconds via beep+oto. No TUI.
// If you hear the tone, your audio path works.
package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

type tone struct {
	phase float64
}

func (t *tone) Stream(samples [][2]float64) (int, bool) {
	const f = 440.0
	const sr = 44100.0
	inc := f / sr
	for i := range samples {
		v := 0.4 * math.Sin(t.phase*2*math.Pi)
		samples[i][0] = v
		samples[i][1] = v
		t.phase += inc
		if t.phase >= 1 {
			t.phase -= 1
		}
	}
	return len(samples), true
}

func (t *tone) Err() error { return nil }

func main() {
	sr := beep.SampleRate(44100)
	if err := speaker.Init(sr, sr.N(time.Second/30)); err != nil {
		fmt.Fprintln(os.Stderr, "speaker init failed:", err)
		os.Exit(1)
	}
	defer speaker.Close()

	fmt.Println("playing 440Hz sine for 3 seconds...")
	speaker.Play(&tone{})
	time.Sleep(3 * time.Second)
	fmt.Println("done. did you hear the tone?")
}

// termus-sf2-smoke: verifies go-meltysynth can load an SF2, trigger notes,
// and produce non-zero audio. Writes a 5-second WAV to disk for sanity check.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/sinshu/go-meltysynth/meltysynth"

	"github.com/mrbrutti/termus/internal/audio"
)

func main() {
	sf2 := flag.String("sf2", "/tmp/TimGM6mb.sf2", "SoundFont file")
	out := flag.String("out", "termus-sf2-smoke.wav", "output WAV")
	program := flag.Int("program", 0, "GM program 0-127 (0 = Acoustic Grand Piano)")
	flag.Parse()

	f, err := os.Open(*sf2)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open sf2:", err)
		os.Exit(1)
	}
	defer f.Close()

	sf, err := meltysynth.NewSoundFont(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse sf2:", err)
		os.Exit(1)
	}

	settings := meltysynth.NewSynthesizerSettings(44100)
	syn, err := meltysynth.NewSynthesizer(sf, settings)
	if err != nil {
		fmt.Fprintln(os.Stderr, "make synth:", err)
		os.Exit(1)
	}

	// Program change on channel 0 to the requested GM program.
	// MIDI program change command = 0xC0 | channel.
	syn.ProcessMidiMessage(0, 0xC0, int32(*program), 0)

	// Trigger a C major chord on channel 0.
	syn.NoteOn(0, 60, 100) // C4
	syn.NoteOn(0, 64, 100) // E4
	syn.NoteOn(0, 67, 100) // G4

	w, err := audio.NewWAVWriter(*out, 44100, 2)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create wav:", err)
		os.Exit(1)
	}

	const blockSize = 512
	leftF := make([]float32, blockSize)
	rightF := make([]float32, blockSize)
	frames := make([][2]float64, blockSize)

	totalFrames := 5 * 44100
	var peakL, peakR float64
	var sumSq float64

	for written := 0; written < totalFrames; written += blockSize {
		// Trigger note-off at 2.5 seconds so we hear the release.
		if written == 2*44100+22050 {
			syn.NoteOff(0, 60)
			syn.NoteOff(0, 64)
			syn.NoteOff(0, 67)
		}
		syn.Render(leftF, rightF)
		for i := 0; i < blockSize; i++ {
			l := float64(leftF[i])
			r := float64(rightF[i])
			frames[i][0] = l
			frames[i][1] = r
			if math.Abs(l) > peakL {
				peakL = math.Abs(l)
			}
			if math.Abs(r) > peakR {
				peakR = math.Abs(r)
			}
			sumSq += l*l + r*r
		}
		if err := w.Write(frames); err != nil {
			fmt.Fprintln(os.Stderr, "write:", err)
			os.Exit(1)
		}
	}
	if err := w.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "close:", err)
		os.Exit(1)
	}

	rms := math.Sqrt(sumSq / float64(2*totalFrames))
	fmt.Fprintf(os.Stderr, "wrote %s — peakL=%.3f peakR=%.3f RMS=%.4f\n",
		*out, peakL, peakR, rms)
}

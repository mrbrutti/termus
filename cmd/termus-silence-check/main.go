// termus-silence-check reads a WAV file and prints per-second RMS windows.
// Used to diagnose silence gaps in rendered tracks.
//
// Usage:
//
//	termus-silence-check --in path/to.wav
//	termus-silence-check --dir /tmp/sp15-pre   # walks all WAVs recursively
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mrbrutti/termus/internal/audio"
)

const sampleRate = 44100

type gap struct {
	startSec float64
	endSec   float64
}

func main() {
	in := flag.String("in", "", "single WAV to analyse")
	dir := flag.String("dir", "", "directory tree to walk for WAVs")
	winSec := flag.Float64("win", 1.0, "RMS window in seconds")
	silenceDB := flag.Float64("silence-db", -40.0, "below this dB → window flagged as silence")
	minGapSec := flag.Float64("min-gap", 1.5, "minimum gap in seconds to report")
	flag.Parse()

	var files []string
	if strings.TrimSpace(*in) != "" {
		files = []string{*in}
	}
	if strings.TrimSpace(*dir) != "" {
		err := filepath.WalkDir(*dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if strings.EqualFold(filepath.Ext(path), ".wav") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "walk:", err)
			os.Exit(1)
		}
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "no WAVs to analyse (provide --in or --dir)")
		os.Exit(2)
	}
	sort.Strings(files)

	for _, f := range files {
		samples, err := audio.ReadIR(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", f, err)
			continue
		}
		fmt.Printf("=== %s (%d samples = %.2fs) ===\n", f, len(samples), float64(len(samples))/float64(sampleRate))
		win := int(*winSec * float64(sampleRate))
		if win <= 0 {
			win = sampleRate
		}
		var gaps []gap
		var inGap bool
		var gapStart float64
		// Slide non-overlapping windows for the printable timeline.
		for i := 0; i < len(samples); i += win {
			end := i + win
			if end > len(samples) {
				end = len(samples)
			}
			sumSq := 0.0
			n := end - i
			for j := i; j < end; j++ {
				sumSq += samples[j] * samples[j]
			}
			rms := math.Sqrt(sumSq / float64(n))
			db := -120.0
			if rms > 1e-12 {
				db = 20 * math.Log10(rms)
			}
			t := float64(i) / float64(sampleRate)
			fmt.Printf("  t=%5.1fs: %6.1f dB\n", t, db)
			if db < *silenceDB {
				if !inGap {
					inGap = true
					gapStart = t
				}
			} else if inGap {
				gaps = append(gaps, gap{gapStart, t})
				inGap = false
			}
		}
		if inGap {
			gaps = append(gaps, gap{gapStart, float64(len(samples)) / float64(sampleRate)})
		}
		// Report long gaps.
		var maxGap gap
		var maxDur float64
		for _, g := range gaps {
			dur := g.endSec - g.startSec
			if dur > maxDur {
				maxDur = dur
				maxGap = g
			}
		}
		fmt.Printf("  -- summary: %d silence windows (below %.0f dB), longest gap = %.2fs at t=%.2fs\n",
			len(gaps), *silenceDB, maxDur, maxGap.startSec)
		for _, g := range gaps {
			dur := g.endSec - g.startSec
			if dur >= *minGapSec {
				fmt.Printf("  !! silence gap: %.2fs starting at t=%.2fs\n", dur, g.startSec)
			}
		}
		fmt.Println()
	}
}

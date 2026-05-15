package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
	"github.com/mrbrutti/termus/internal/synth"
)

type listeningScore struct {
	Total             float64 `json:"total"`
	CadenceDensity    float64 `json:"cadence_density"`
	HarmonicMotion    float64 `json:"harmonic_motion"`
	SectionVariety    float64 `json:"section_variety"`
	Occupancy         float64 `json:"occupancy"`
	RepetitionPenalty float64 `json:"repetition_penalty"`
	CadenceCount      int     `json:"cadence_count"`
	ChordChanges      int     `json:"chord_changes"`
	UniqueSections    int     `json:"unique_sections"`
	LongestRepeatRun  int     `json:"longest_repeat_run"`
}

func renderToWAVWithSnapshots(path string, algo gen.Algorithm, seconds float64) (written int, snapshots []gen.DebugStatus, err error) {
	if seconds <= 0 {
		return 0, nil, fmt.Errorf("seconds must be > 0, got %.3f", seconds)
	}
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return 0, nil, err
		}
	}

	w, err := audio.NewWAVWriter(path, synth.SampleRate, 2)
	if err != nil {
		return 0, nil, err
	}
	defer func() {
		if closeErr := w.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	root := audio.NewRoot(algo, scope.NewRing(64))
	root.SetVolume(100)

	totalFrames := int(seconds * float64(synth.SampleRate))
	chunk := 4410
	frames := make([][2]float64, chunk)
	sampleEvery := synth.SampleRate / 2
	sinceSample := 0
	snapshots = append(snapshots, root.DebugStatus())

	for written < totalFrames {
		n := chunk
		if remain := totalFrames - written; remain < n {
			n = remain
		}
		if _, ok := root.Stream(frames[:n]); !ok {
			return written, snapshots, fmt.Errorf("audio stream ended after %d frames", written)
		}
		if err := w.Write(frames[:n]); err != nil {
			return written, snapshots, err
		}
		written += n
		sinceSample += n
		for sinceSample >= sampleEvery {
			snapshots = append(snapshots, root.DebugStatus())
			sinceSample -= sampleEvery
		}
	}
	return written, snapshots, nil
}

func scoreListeningResult(durationS float64, markers []gen.ListeningMarker, snapshots []gen.DebugStatus) listeningScore {
	if durationS <= 0 {
		durationS = 1
	}
	minutes := durationS / 60.0
	if minutes <= 0 {
		minutes = 1.0 / 60.0
	}

	cadenceCount := 0
	for _, marker := range markers {
		if strings.HasPrefix(marker.Label, "cadence:") {
			cadenceCount++
		}
	}
	cadenceDensity := clamp01(float64(cadenceCount) / (minutes * 2.0))

	uniqueSections := map[string]struct{}{}
	chordChanges := 0
	longestRun := 0
	currentRun := 0
	prevState := ""
	prevChord := ""
	occupancyTotal := 0.0
	validSamples := 0

	for _, snap := range snapshots {
		state := snap.Section + "|" + snap.Chord
		if state == prevState {
			currentRun++
		} else {
			currentRun = 1
			prevState = state
		}
		if currentRun > longestRun {
			longestRun = currentRun
		}
		if snap.Section != "" {
			uniqueSections[snap.Section] = struct{}{}
		}
		if snap.Chord != "" {
			if prevChord != "" && snap.Chord != prevChord {
				chordChanges++
			}
			prevChord = snap.Chord
		}
		if snap.Section != "" {
			occupancyTotal += sectionOccupancyWeight(snap.Section)
			validSamples++
		}
	}

	sectionVariety := 0.0
	if len(uniqueSections) > 0 {
		sectionVariety = clamp01(float64(len(uniqueSections)) / 5.0)
	}
	harmonicMotion := 0.0
	if len(snapshots) > 1 {
		harmonicMotion = clamp01(float64(chordChanges) / float64(len(snapshots)-1))
	}
	occupancy := 0.0
	if validSamples > 0 {
		occupancy = occupancyTotal / float64(validSamples)
	}
	repetitionPenalty := 0.0
	if len(snapshots) > 0 {
		repetitionPenalty = clamp01(float64(longestRun) / float64(len(snapshots)))
	}

	total := 0.26*cadenceDensity +
		0.24*harmonicMotion +
		0.18*sectionVariety +
		0.18*occupancy +
		0.14*(1.0-repetitionPenalty)

	return listeningScore{
		Total:             total,
		CadenceDensity:    cadenceDensity,
		HarmonicMotion:    harmonicMotion,
		SectionVariety:    sectionVariety,
		Occupancy:         occupancy,
		RepetitionPenalty: repetitionPenalty,
		CadenceCount:      cadenceCount,
		ChordChanges:      chordChanges,
		UniqueSections:    len(uniqueSections),
		LongestRepeatRun:  longestRun,
	}
}

func sectionOccupancyWeight(section string) float64 {
	switch strings.ToLower(strings.TrimSpace(section)) {
	case "", "pad", "drone", "intro":
		return 0.25
	case "breakdown":
		return 0.15
	case "outro":
		return 0.30
	case "a":
		return 0.65
	case "a'":
		return 0.72
	case "b":
		return 0.80
	case "cadence":
		return 0.90
	default:
		return 0.60
	}
}

func rankResults(results []corpusResult) {
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Skipped != "" || results[j].Skipped != "" {
			if results[i].Skipped == "" {
				return true
			}
			if results[j].Skipped == "" {
				return false
			}
			return results[i].Seed < results[j].Seed
		}
		if results[i].Score.Total == results[j].Score.Total {
			return results[i].Seed < results[j].Seed
		}
		return results[i].Score.Total > results[j].Score.Total
	})
	rank := 0
	for i := range results {
		if results[i].Skipped != "" {
			results[i].Rank = 0
			continue
		}
		rank++
		results[i].Rank = rank
	}
}

func clamp01(v float64) float64 {
	switch {
	case v < 0:
		return 0
	case v > 1:
		return 1
	default:
		return v
	}
}

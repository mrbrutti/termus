package audio

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
	"github.com/mrbrutti/termus/internal/synth"
)

const (
	renderOutroSeconds         = 2.75
	renderPreferredOutroWindow = 6.0
	renderCadenceSnapWindow    = 12.0
)

// RenderPlan describes how an offline export should land: the requested
// duration, any cadence/outro snap point, and the fade-out tail added after
// the musical boundary.
type RenderPlan struct {
	RequestedFrames int
	FadeStartFrame  int
	FadeFrames      int
	TotalFrames     int
	SnapLabel       string
}

func (p RenderPlan) DurationSeconds() float64 {
	if p.TotalFrames <= 0 {
		return 0
	}
	return float64(p.TotalFrames) / float64(synth.SampleRate)
}

// PlanRender extends abrupt exports with a short musical outro. When the
// algorithm exposes listening markers, the render snaps to the next nearby
// cadence or outro before the fade begins.
func PlanRender(algo gen.Algorithm, seconds float64) RenderPlan {
	requested := int(seconds * float64(synth.SampleRate))
	if requested < 1 {
		return RenderPlan{}
	}
	fadeFrames := int(renderOutroSeconds * float64(synth.SampleRate))
	fadeStart := requested
	snapLabel := ""
	if inspectable, ok := algo.(gen.ListeningInspectable); ok {
		markers := inspectable.ListeningMarkers()
		if snapped, label, ok := findMarkerSnap(markers, requested, renderPreferredOutroWindow,
			func(label string) bool {
				return label == "section:outro" || label == "cadence:outro"
			},
		); ok {
			fadeStart = snapped
			snapLabel = label
		} else if snapped, label, ok := findMarkerSnap(markers, requested, renderCadenceSnapWindow,
			func(label string) bool {
				return strings.HasPrefix(label, "cadence:") || label == "section:cadence"
			},
		); ok {
			fadeStart = snapped
			snapLabel = label
		}
	}
	if fadeStart < requested {
		fadeStart = requested
	}
	total := fadeStart + fadeFrames
	return RenderPlan{
		RequestedFrames: requested,
		FadeStartFrame:  fadeStart,
		FadeFrames:      fadeFrames,
		TotalFrames:     total,
		SnapLabel:       snapLabel,
	}
}

// RenderToWAV renders an algorithm offline to a WAV file without touching the
// live speaker backend. Volume uses the same 0..100 scaling as the TUI.
func RenderToWAV(path string, algo gen.Algorithm, seconds float64, volume int) (written int, err error) {
	return RenderToWAVWithPlan(path, algo, PlanRender(algo, seconds), volume)
}

// RenderToWAVWithPlan renders an algorithm offline using a caller-specified
// plan so sibling exports (WAV/MIDI/stems/manifest) can share the same outro.
func RenderToWAVWithPlan(path string, algo gen.Algorithm, plan RenderPlan, volume int) (written int, err error) {
	if plan.TotalFrames <= 0 {
		return 0, fmt.Errorf("render plan has no frames")
	}
	if plan.RequestedFrames <= 0 {
		return 0, fmt.Errorf("render plan must include requested frames")
	}
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return 0, err
		}
	}

	w, err := NewWAVWriter(path, synth.SampleRate, 2)
	if err != nil {
		return 0, err
	}
	defer func() {
		if closeErr := w.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	root := NewRoot(algo, scope.NewRing(64))
	root.SetVolume(volume)

	chunk := 4410
	frames := make([][2]float64, chunk)
	for written < plan.TotalFrames {
		n := chunk
		if remain := plan.TotalFrames - written; remain < n {
			n = remain
		}
		if _, ok := root.Stream(frames[:n]); !ok {
			return written, fmt.Errorf("audio stream ended after %d frames", written)
		}
		applyOutroFade(frames[:n], written, plan)
		if err := w.Write(frames[:n]); err != nil {
			return written, err
		}
		written += n
	}
	return written, nil
}

func findMarkerSnap(markers []gen.ListeningMarker, requested int, windowSeconds float64, match func(string) bool) (int, string, bool) {
	if len(markers) == 0 {
		return 0, "", false
	}
	maxFrame := requested + int(windowSeconds*float64(synth.SampleRate))
	for _, marker := range markers {
		frame := int(marker.Sample)
		if frame < requested || frame > maxFrame {
			continue
		}
		if match(marker.Label) {
			return frame, marker.Label, true
		}
	}
	return 0, "", false
}

func applyOutroFade(frames [][2]float64, frameOffset int, plan RenderPlan) {
	if plan.FadeFrames <= 0 || plan.TotalFrames <= plan.FadeStartFrame {
		return
	}
	denom := maxInt(1, plan.FadeFrames-1)
	for i := range frames {
		frame := frameOffset + i
		if frame < plan.FadeStartFrame {
			continue
		}
		progress := float64(frame-plan.FadeStartFrame) / float64(denom)
		if progress > 1 {
			progress = 1
		}
		gain := math.Cos(progress * math.Pi * 0.5)
		frames[i][0] *= gain
		frames[i][1] *= gain
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// SectionStop describes one entry in a multi-section offline render. Each
// stop owns its own pre-built algorithm and the number of frames it should
// render before the next stop's algorithm takes over (no crossfade is
// applied; the swap is sample-aligned and seamless).
type SectionStop struct {
	Algo   gen.Algorithm
	Frames int
}

// RenderSectionsToWAV renders a multi-section seamless composition to a WAV
// file. Stops are rendered back-to-back; at each boundary the active
// algorithm is replaced with the next stop's algorithm and rendering
// continues into the same WAV stream. After all stops, a short tail fade is
// applied.
func RenderSectionsToWAV(path string, stops []SectionStop, volume int) (written int, err error) {
	if len(stops) == 0 {
		return 0, fmt.Errorf("no section stops")
	}
	totalFrames := 0
	for _, s := range stops {
		if s.Algo == nil {
			return 0, fmt.Errorf("section stop has nil algo")
		}
		if s.Frames <= 0 {
			return 0, fmt.Errorf("section stop has non-positive frame count")
		}
		totalFrames += s.Frames
	}
	fadeFrames := int(renderOutroSeconds * float64(synth.SampleRate))
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return 0, err
		}
	}
	w, err := NewWAVWriter(path, synth.SampleRate, 2)
	if err != nil {
		return 0, err
	}
	defer func() {
		if closeErr := w.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()
	plan := RenderPlan{
		RequestedFrames: totalFrames,
		FadeStartFrame:  totalFrames,
		FadeFrames:      fadeFrames,
		TotalFrames:     totalFrames + fadeFrames,
	}
	chunk := 4410
	frames := make([][2]float64, chunk)
	for idx, stop := range stops {
		root := NewRoot(stop.Algo, scope.NewRing(64))
		root.SetVolume(volume)
		remaining := stop.Frames
		// On the last stop, render through to the end of the fade tail too
		// so the outro is included.
		if idx == len(stops)-1 {
			remaining += fadeFrames
		}
		for remaining > 0 {
			n := chunk
			if remaining < n {
				n = remaining
			}
			if _, ok := root.Stream(frames[:n]); !ok {
				return written, fmt.Errorf("audio stream ended after %d frames", written)
			}
			applyOutroFade(frames[:n], written, plan)
			if err := w.Write(frames[:n]); err != nil {
				return written, err
			}
			written += n
			remaining -= n
		}
	}
	return written, nil
}

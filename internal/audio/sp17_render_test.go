package audio

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// constLevelAlgo emits a constant DC level on both channels, identifying
// which "section" is active. Used to verify section-boundary alignment in
// RenderSectionsToWAV.
type constLevelAlgo struct {
	level float64
}

func (c *constLevelAlgo) Name() string { return "const" }
func (c *constLevelAlgo) Seed(int64)   {}
func (c *constLevelAlgo) Next(l, r []float64) {
	for i := range l {
		l[i] = c.level
		r[i] = c.level
	}
}

// TestRenderRespectsSectionBoundaries verifies that RenderSectionsToWAV
// swaps algorithms at the exact frame boundaries requested, with no audible
// silence (crossfade dip) between sections. Each section emits a distinct
// DC level; the test reads the resulting WAV back and checks the level
// distribution across the timeline.
func TestRenderRespectsSectionBoundaries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sections.wav")
	const framesPerSection = 4410 // 100 ms each
	stops := []SectionStop{
		{Algo: &constLevelAlgo{level: 0.2}, Frames: framesPerSection},
		{Algo: &constLevelAlgo{level: 0.4}, Frames: framesPerSection},
		{Algo: &constLevelAlgo{level: 0.6}, Frames: framesPerSection},
	}
	written, err := RenderSectionsToWAV(path, stops, 100)
	if err != nil {
		t.Fatalf("RenderSectionsToWAV: %v", err)
	}
	// 3 sections + outro fade tail (renderOutroSeconds * 44100).
	if written < 3*framesPerSection {
		t.Fatalf("frames written = %d, want >= %d", written, 3*framesPerSection)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// PCM payload starts at byte 44 in a standard 44-byte WAV header.
	pcm := data[44:]
	// 16-bit stereo, so 4 bytes per frame.
	if len(pcm) < 4*written {
		t.Fatalf("pcm short: %d bytes for %d frames", len(pcm), written)
	}

	// Sample the middle of each section (avoid any per-frame transient).
	// Pre-outro: section i = frames [i*framesPerSection, (i+1)*framesPerSection).
	levelAt := func(frame int) float64 {
		offset := 4 * frame
		left := int16(binary.LittleEndian.Uint16(pcm[offset : offset+2]))
		return float64(left) / 32767.0
	}

	type section struct {
		name     string
		frame    int
		expected float64
	}
	tests := []section{
		{"section 0 mid", framesPerSection / 2, 0.2},
		{"section 1 mid", framesPerSection + framesPerSection/2, 0.4},
		{"section 2 mid", 2*framesPerSection + framesPerSection/2, 0.6},
	}
	for _, tt := range tests {
		got := levelAt(tt.frame)
		// Allow generous tolerance: PCM quantization + WAV writer dither.
		if math.Abs(got-tt.expected) > 0.05 {
			t.Errorf("%s: level = %.3f, want ~%.3f", tt.name, got, tt.expected)
		}
	}

	// Verify there's no audible "gap" at the section boundary. The frame
	// just before and just after the boundary should both be within their
	// own section's level (no fade to silence).
	const tolerance = 0.02
	boundaryFrame := framesPerSection
	beforeBoundary := levelAt(boundaryFrame - 1)
	afterBoundary := levelAt(boundaryFrame)
	if math.Abs(beforeBoundary-0.2) > tolerance {
		t.Errorf("frame just before boundary 1: %.3f, want ~0.2 (no dip)", beforeBoundary)
	}
	if math.Abs(afterBoundary-0.4) > tolerance {
		t.Errorf("frame just after boundary 1: %.3f, want ~0.4 (immediate swap, no dip)", afterBoundary)
	}
}

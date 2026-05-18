package track

import (
	"testing"
	"time"

	"github.com/mrbrutti/termus/internal/gen"
)

const sp17FourSectionSrc = `
title: SP17 Four Section
style: lofi
listen_mode: hour-stream
seed: 7
roles:
  keys:
    family: electric_piano
    pattern: "x..x .x.."
  lead:
    family: reed_lead
    motif: "5 . 6 5 | 3 . 2 1"
sections:
  - id: intro
    title: opening
    duration: 12s
    harmony: "Dm9 G13 | Cmaj9 A7"
    scene: "intro sparse"
  - id: verse
    title: verse
    duration: 32s
    harmony: "Dm7 G7 | Cmaj7 A7"
    scene: "verse main"
  - id: bridge
    title: bridge
    duration: 24s
    harmony: "Fm7 Bb7 | Ebmaj7 C7"
    scene: "bridge lift"
  - id: outro
    title: closing
    duration: 20s
    harmony: "Dm9 G13 | Cmaj9"
    scene: "outro cadence"
`

func TestCompile_SingleTrackWithSections(t *testing.T) {
	file, err := Parse([]byte(sp17FourSectionSrc))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 1, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if got, want := len(compiled.Playlist.Tracks), 1; got != want {
		t.Fatalf("Tracks = %d, want %d (SP17: one seamless track)", got, want)
	}
	track := compiled.Playlist.Tracks[0]
	if got, want := len(track.Sections), 4; got != want {
		t.Fatalf("Sections = %d, want %d", got, want)
	}
	if got, want := track.Duration, 88*time.Second; got != want {
		t.Fatalf("Track.Duration = %v, want %v", got, want)
	}
	if !track.LoopForeverEvolving {
		t.Fatal("hour-stream listen mode should set LoopForeverEvolving=true")
	}
	if track.Title != "SP17 Four Section" {
		t.Fatalf("Track.Title = %q, want %q", track.Title, "SP17 Four Section")
	}
	if len(compiled.Plans) != 4 {
		t.Fatalf("Plans = %d, want 4", len(compiled.Plans))
	}
	// Every section's PlanKey must resolve into Plans.
	for i, stop := range track.Sections {
		if _, ok := compiled.Plans[stop.PlanKey]; !ok {
			t.Errorf("section %d (%s): PlanKey %q missing from Plans", i, stop.Title, stop.PlanKey)
		}
	}
}

func TestSectionStopSchedule(t *testing.T) {
	file, err := Parse([]byte(sp17FourSectionSrc))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 1, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	sections := compiled.Playlist.Tracks[0].Sections
	wantStarts := []time.Duration{0, 12 * time.Second, 44 * time.Second, 68 * time.Second}
	wantDurs := []time.Duration{12 * time.Second, 32 * time.Second, 24 * time.Second, 20 * time.Second}
	wantTitles := []string{"opening", "verse", "bridge", "closing"}
	if len(sections) != len(wantStarts) {
		t.Fatalf("section count = %d", len(sections))
	}
	for i, stop := range sections {
		if stop.StartTime != wantStarts[i] {
			t.Errorf("section[%d].StartTime = %v, want %v", i, stop.StartTime, wantStarts[i])
		}
		if stop.Duration != wantDurs[i] {
			t.Errorf("section[%d].Duration = %v, want %v", i, stop.Duration, wantDurs[i])
		}
		if stop.Title != wantTitles[i] {
			t.Errorf("section[%d].Title = %q, want %q", i, stop.Title, wantTitles[i])
		}
	}
}

func TestCompile_SingleSectionStillWorks(t *testing.T) {
	const src = `
title: Solo
style: ambient
roles:
  pad:
    family: pad
sections:
  - id: only
    title: the only one
    duration: 20s
    harmony: "Cmaj7"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 1, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if got, want := len(compiled.Playlist.Tracks), 1; got != want {
		t.Fatalf("Tracks = %d, want %d", got, want)
	}
	track := compiled.Playlist.Tracks[0]
	if got, want := len(track.Sections), 1; got != want {
		t.Fatalf("Sections = %d, want %d", got, want)
	}
	if track.LoopForeverEvolving {
		t.Fatal("endless listen mode should NOT set LoopForeverEvolving")
	}
}

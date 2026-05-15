package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mrbrutti/termus/internal/gen"
)

type inspectableStubAlgo struct {
	name    string
	markers []gen.ListeningMarker
}

func (a *inspectableStubAlgo) Name() string { return a.name }
func (a *inspectableStubAlgo) Seed(int64)   {}
func (a *inspectableStubAlgo) Next(l, r []float64) {
	for i := range l {
		l[i], r[i] = 0.1, 0.1
	}
}
func (a *inspectableStubAlgo) ListeningMarkers() []gen.ListeningMarker {
	return append([]gen.ListeningMarker(nil), a.markers...)
}
func (a *inspectableStubAlgo) ExportMIDI(path string, seconds float64) error {
	return os.WriteFile(path, []byte("midi:"+a.name), 0o644)
}
func (a *inspectableStubAlgo) ExportStems(dir string, seconds float64, volume int) ([]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "stem.wav")
	if err := os.WriteFile(path, []byte("stem:"+a.name), 0o644); err != nil {
		return nil, err
	}
	return []string{path}, nil
}

func TestRenderPlaylistOutWithWritesManifest(t *testing.T) {
	dir := t.TempDir()
	pl := &gen.Playlist{
		Name: "Night Drive",
		Mode: gen.PlaylistMixed,
		Tracks: []gen.Track{
			{
				Spec:     gen.AlgoSpec{Name: "ambient-synth", Display: "Ambient"},
				Seed:     7,
				Duration: 1500 * time.Millisecond,
			},
			{
				Spec:     gen.AlgoSpec{Name: "bells/synth", Display: "Bells"},
				Seed:     8,
				Duration: 2 * time.Second,
			},
		},
	}

	build := func(spec gen.AlgoSpec, seed int64) gen.Algorithm {
		return &inspectableStubAlgo{
			name: spec.Name,
			markers: []gen.ListeningMarker{
				{Label: "keep", Sample: 100},
				{Label: "drop", Sample: 999999},
			},
		}
	}
	render := func(path string, algo gen.Algorithm, seconds float64, volume int) (int, error) {
		if volume != 70 {
			t.Fatalf("volume = %d, want 70", volume)
		}
		if err := os.WriteFile(path, []byte(algo.Name()), 0o644); err != nil {
			return 0, err
		}
		return int(seconds * 44100), nil
	}

	manifest, err := renderPlaylistOutWith(dir, pl, 70, build, render, true, true)
	if err != nil {
		t.Fatalf("renderPlaylistOutWith: %v", err)
	}
	if manifest.Mode != "mixed" || manifest.TrackCount != 2 {
		t.Fatalf("manifest summary = %+v", manifest)
	}
	if manifest.Tracks[0].Path != "1-ambient-synth-7.wav" {
		t.Fatalf("track 0 path = %q", manifest.Tracks[0].Path)
	}
	if manifest.Tracks[1].Path != "2-bells-synth-8.wav" {
		t.Fatalf("track 1 path = %q", manifest.Tracks[1].Path)
	}
	if manifest.Tracks[0].MIDIPath != "1-ambient-synth-7.mid" {
		t.Fatalf("track 0 midi path = %q", manifest.Tracks[0].MIDIPath)
	}
	if manifest.Tracks[0].StemDir != "1-ambient-synth-7-stems" {
		t.Fatalf("track 0 stem dir = %q", manifest.Tracks[0].StemDir)
	}
	if len(manifest.Tracks[0].StemFiles) != 1 {
		t.Fatalf("track 0 stem files = %+v", manifest.Tracks[0].StemFiles)
	}
	if len(manifest.Tracks[0].Markers) != 1 || manifest.Tracks[0].Markers[0].Label != "keep" {
		t.Fatalf("trimmed markers = %+v", manifest.Tracks[0].Markers)
	}

	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatalf("manifest.json: %v", err)
	}
	var decoded playlistManifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("manifest decode: %v", err)
	}
	if decoded.Name != pl.Name || len(decoded.Tracks) != 2 {
		t.Fatalf("decoded manifest = %+v", decoded)
	}
	if _, err := os.Stat(filepath.Join(dir, decoded.Tracks[0].Path)); err != nil {
		t.Fatalf("track file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, decoded.Tracks[0].MIDIPath)); err != nil {
		t.Fatalf("midi file missing: %v", err)
	}
}

func TestSafeFileStem(t *testing.T) {
	if got := safeFileStem(" Bells / Synth "); got != "bells-synth" {
		t.Fatalf("safeFileStem = %q", got)
	}
}

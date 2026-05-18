package acestep

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mrbrutti/termus/internal/track"
)

// helper to build a minimal valid acestep File for testing.
func newACEFile() *track.File {
	return &track.File{
		Title:        "test",
		Key:          "Dmin",
		Tempo:        "86",
		Seed:         42,
		RenderEngine: track.RenderEngineACEStep,
		Acestep: &track.AcestepSpec{
			Style: "warm lo-fi rhodes in a quiet bookstore",
			Tags:  []string{"lofi", "rhodes", "rainy"},
		},
	}
}

func TestCompileV3_BasicFields(t *testing.T) {
	f := newACEFile()
	f.TotalDuration = "3m"

	spec, err := CompileV3(f)
	if err != nil {
		t.Fatalf("CompileV3: unexpected error: %v", err)
	}
	if spec.Key != "Dmin" {
		t.Errorf("Key = %q, want %q", spec.Key, "Dmin")
	}
	if spec.Tempo != 86 {
		t.Errorf("Tempo = %d, want 86", spec.Tempo)
	}
	if spec.DurationSeconds != 180 {
		t.Errorf("DurationSeconds = %v, want 180", spec.DurationSeconds)
	}
	if spec.Seed != 42 {
		t.Errorf("Seed = %d, want 42", spec.Seed)
	}
	if spec.TimeSignature != "4/4" {
		t.Errorf("TimeSignature = %q, want default 4/4", spec.TimeSignature)
	}
}

func TestCompileV3_StyleAndTags_Joined(t *testing.T) {
	f := newACEFile()
	spec, err := CompileV3(f)
	if err != nil {
		t.Fatalf("CompileV3: %v", err)
	}
	// caption must contain the prose and every tag.
	if !strings.Contains(spec.Prompt, "warm lo-fi rhodes") {
		t.Errorf("Prompt missing style prose: %q", spec.Prompt)
	}
	for _, tag := range []string{"lofi", "rhodes", "rainy"} {
		if !strings.Contains(spec.Prompt, tag) {
			t.Errorf("Prompt missing tag %q: %q", tag, spec.Prompt)
		}
	}
	// Tags slice round-trip on the wire as well.
	if len(spec.Tags) != 3 {
		t.Errorf("Tags len = %d, want 3 (%v)", len(spec.Tags), spec.Tags)
	}
}

func TestCompileV3_RejectsNonAcestepEngine(t *testing.T) {
	f := newACEFile()
	f.RenderEngine = track.RenderEngineSF2
	_, err := CompileV3(f)
	if err == nil {
		t.Fatalf("expected error when render_engine=sf2, got nil")
	}
	if !strings.Contains(err.Error(), "render_engine") {
		t.Errorf("error should mention render_engine, got: %v", err)
	}
}

func TestCompileV3_RejectsMissingAcestepBlock(t *testing.T) {
	f := newACEFile()
	f.Acestep = nil
	_, err := CompileV3(f)
	if err == nil {
		t.Fatalf("expected error when acestep block is nil, got nil")
	}
}

func TestCompileV3_RejectsNilFile(t *testing.T) {
	if _, err := CompileV3(nil); err == nil {
		t.Fatalf("expected error on nil file")
	}
}

func TestCompileV3_InfersScaleFromKey(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"Dmin", "minor"},
		{"Cmin", "minor"},
		{"Cmaj", "major"},
		{"Am", "minor"},
		{"Bm7", "minor"},
		{"C", "major"},
		{"F#maj7", "major"},
	}
	for _, tc := range tests {
		f := newACEFile()
		f.Key = tc.key
		// no explicit scale on the AcestepSpec
		spec, err := CompileV3(f)
		if err != nil {
			t.Fatalf("CompileV3 key=%q: %v", tc.key, err)
		}
		if spec.Scale != tc.expected {
			t.Errorf("Key=%q: Scale = %q, want %q", tc.key, spec.Scale, tc.expected)
		}
	}
}

func TestCompileV3_ExplicitScaleWins(t *testing.T) {
	f := newACEFile()
	f.Key = "Dmin" // would normally infer "minor"
	f.Acestep.Scale = "dorian"
	spec, err := CompileV3(f)
	if err != nil {
		t.Fatalf("CompileV3: %v", err)
	}
	if spec.Scale != "dorian" {
		t.Errorf("Scale = %q, want %q (explicit override)", spec.Scale, "dorian")
	}
}

func TestCompileV3_HarmonyConcatenation(t *testing.T) {
	f := newACEFile()
	f.Acestep.Sections = []track.AcestepSection{
		{ID: "intro", Bars: 8, Description: "soft intro", Harmony: "Dm7 Am7"},
		{ID: "head", Bars: 16, Description: "main theme", Harmony: "Bbmaj7 Gm7"},
		{ID: "outro", Bars: 8, Description: "fadeout", Harmony: "Dm7"},
	}
	spec, err := CompileV3(f)
	if err != nil {
		t.Fatalf("CompileV3: %v", err)
	}
	want := "Dm7 Am7 Bbmaj7 Gm7 Dm7"
	if spec.HarmonyChain != want {
		t.Errorf("HarmonyChain = %q, want %q", spec.HarmonyChain, want)
	}
	// sections appear, in order, in the prompt.
	if !strings.Contains(spec.Prompt, "soft intro") || !strings.Contains(spec.Prompt, "main theme") {
		t.Errorf("Prompt missing section descriptions: %q", spec.Prompt)
	}
	if len(spec.SectionDescriptions) != 3 {
		t.Errorf("SectionDescriptions len = %d, want 3", len(spec.SectionDescriptions))
	}
}

func TestCompileV3_SeedOverride(t *testing.T) {
	f := newACEFile()
	f.Seed = 100
	v := int64(999)
	f.Acestep.SeedOverride = &v
	spec, err := CompileV3(f)
	if err != nil {
		t.Fatalf("CompileV3: %v", err)
	}
	if spec.Seed != 999 {
		t.Errorf("Seed = %d, want 999 (acestep override)", spec.Seed)
	}
}

func TestCompileV3_DurationFromSectionBars(t *testing.T) {
	// 4 bars at 120 BPM, 4 beats per bar = 4 * 4 * (60/120) = 8 seconds.
	f := newACEFile()
	f.Tempo = "120"
	f.TotalDuration = "" // force fallback
	f.Acestep.Sections = []track.AcestepSection{
		{ID: "intro", Bars: 2, Description: "a"},
		{ID: "outro", Bars: 2, Description: "b"},
	}
	spec, err := CompileV3(f)
	if err != nil {
		t.Fatalf("CompileV3: %v", err)
	}
	if spec.DurationSeconds != 8 {
		t.Errorf("DurationSeconds = %v, want 8", spec.DurationSeconds)
	}
}

func TestCompileV3_TimeSignatureOverride(t *testing.T) {
	f := newACEFile()
	f.Acestep.TimeSignature = "3/4"
	spec, err := CompileV3(f)
	if err != nil {
		t.Fatalf("CompileV3: %v", err)
	}
	if spec.TimeSignature != "3/4" {
		t.Errorf("TimeSignature = %q, want 3/4", spec.TimeSignature)
	}
}

func TestCompileV3_MotifAndInferenceSteps(t *testing.T) {
	f := newACEFile()
	f.Acestep.Motif = "stepwise minor descent"
	f.Acestep.InferenceSteps = 12
	spec, err := CompileV3(f)
	if err != nil {
		t.Fatalf("CompileV3: %v", err)
	}
	if spec.Motif != "stepwise minor descent" {
		t.Errorf("Motif = %q", spec.Motif)
	}
	if spec.InferenceSteps != 12 {
		t.Errorf("InferenceSteps = %d, want 12", spec.InferenceSteps)
	}
}

// TestCompileV3_ParsesReferenceTrack exercises the v3 fixture authored under
// tracks/lofi/bookstore-rainy-night-v3.tm. The fixture is also the canonical
// example for end-users; if it stops parsing the docs are stale.
func TestCompileV3_ParsesReferenceTrack(t *testing.T) {
	// The test runs from internal/acestep; the fixture is in tracks/.
	// We walk up to the repo root by looking for go.mod.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	root := wd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		root = filepath.Dir(root)
	}
	path := filepath.Join(root, "tracks/lofi/bookstore-rainy-night-v3.tm")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("reference track not present at %s: %v", path, err)
	}
	f, err := track.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile(%q): %v", path, err)
	}
	if f.RenderEngine != track.RenderEngineACEStep {
		t.Fatalf("reference track render_engine = %q, want acestep", f.RenderEngine)
	}
	spec, err := CompileV3(f)
	if err != nil {
		t.Fatalf("CompileV3: %v", err)
	}
	if spec.Prompt == "" {
		t.Errorf("compiled Prompt is empty")
	}
	if spec.Key == "" {
		t.Errorf("compiled Key is empty")
	}
	if spec.DurationSeconds <= 0 {
		t.Errorf("compiled DurationSeconds = %v, want > 0", spec.DurationSeconds)
	}
}

// Package acestep contains the Go client for the local ACE-Step inference
// service (services/acestep/server.py) and the v3 prompt-compiler that turns
// a parsed track.File into a wire-shaped RenderSpec.
//
// The Python side is a separate process. termus-stream and any future
// consumer talk to it over plain HTTP+JSON.
//
// UNTESTED: the Python service is mocked out in the Go tests
// (client_test.go uses httptest). Real ACE-Step inference has not been
// exercised in this PR.
package acestep

// RenderSpec is the wire shape sent to POST /render on the Python service.
// JSON field names are snake_case and match the Pydantic RenderRequest in
// services/acestep/server.py exactly. Adding or renaming a field requires a
// matching change on both sides.
type RenderSpec struct {
	// Prompt is the natural-language style description. Becomes the main
	// "caption" passed to the ACE-Step model. < 512 chars after composition.
	Prompt string `json:"prompt"`

	// Tags are rank-ordered descriptors. ACE-Step's documentation says the
	// first tag should be the genre.
	Tags []string `json:"tags"`

	// Key is the musical key in human notation, e.g. "Cmin", "C major", "Am".
	Key string `json:"key"`

	// Tempo is BPM. 0 lets the model choose.
	Tempo int `json:"tempo"`

	// DurationSeconds is the target length. 0 lets the model choose.
	DurationSeconds float64 `json:"duration_seconds"`

	// Scale: "minor", "major", "dorian", ...
	Scale string `json:"scale"`

	// TimeSignature: "4/4", "3/4", "6/8".
	TimeSignature string `json:"time_signature"`

	// Seed: reproducibility seed. -1 means random.
	Seed int64 `json:"seed"`

	// ReferenceAudioB64 is optional base64-encoded reference audio for style
	// transfer / cover tasks. Empty = none.
	ReferenceAudioB64 string `json:"reference_audio_b64"`

	// SectionDescriptions is an optional per-section natural-language list
	// that gets folded into the caption.
	SectionDescriptions []string `json:"section_descriptions,omitempty"`

	// HarmonyChain is the concatenated chord progression across all sections,
	// e.g. "Am7 Fmaj7 Dm7 G7sus".
	HarmonyChain string `json:"harmony_chain,omitempty"`

	// Motif is a natural-language description of the motif.
	Motif string `json:"motif,omitempty"`

	// InferenceSteps is the diffusion step count. 8 is the turbo default.
	InferenceSteps int `json:"inference_steps,omitempty"`
}

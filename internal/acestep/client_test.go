package acestep

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// minimalWAV returns a 12-byte-header valid RIFF/WAVE blob padded out with
// silence. Suitable for the httptest mock; not a real audio file.
func minimalWAV(t *testing.T) []byte {
	t.Helper()
	// 44-byte standard PCM header + 0 data bytes is enough to pass the
	// client's RIFF magic check and a wav reader's "I can parse this" check.
	hdr := make([]byte, 0, 44)
	hdr = append(hdr, []byte("RIFF")...)
	hdr = binary.LittleEndian.AppendUint32(hdr, 36) // file size - 8
	hdr = append(hdr, []byte("WAVE")...)
	hdr = append(hdr, []byte("fmt ")...)
	hdr = binary.LittleEndian.AppendUint32(hdr, 16) // fmt chunk size
	hdr = binary.LittleEndian.AppendUint16(hdr, 1)  // PCM
	hdr = binary.LittleEndian.AppendUint16(hdr, 1)  // mono
	hdr = binary.LittleEndian.AppendUint32(hdr, 48000)
	hdr = binary.LittleEndian.AppendUint32(hdr, 96000)
	hdr = binary.LittleEndian.AppendUint16(hdr, 2)
	hdr = binary.LittleEndian.AppendUint16(hdr, 16)
	hdr = append(hdr, []byte("data")...)
	hdr = binary.LittleEndian.AppendUint32(hdr, 0)
	return hdr
}

func TestClient_Health(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("unexpected path %q", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HealthResponse{
			Loaded:          true,
			Backend:         "mlx",
			ModelName:       "acestep-v15-turbo",
			LMModelName:     "acestep-5Hz-lm-1.7B",
			MockMode:        false,
			LoadTimeSeconds: 27.5,
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, time.Second)
	h, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	if !h.Loaded {
		t.Errorf("Loaded = false; want true")
	}
	if h.Backend != "mlx" {
		t.Errorf("Backend = %q; want mlx", h.Backend)
	}
	if h.ModelName != "acestep-v15-turbo" {
		t.Errorf("ModelName = %q", h.ModelName)
	}
}

func TestClient_HealthBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, time.Second)
	_, err := c.Health(context.Background())
	if err == nil {
		t.Fatalf("expected error on 500, got nil")
	}
	if !strings.Contains(err.Error(), "status=500") {
		t.Errorf("error missing status code: %v", err)
	}
}

func TestClient_Render(t *testing.T) {
	wantWAV := minimalWAV(t)
	var gotSpec RenderSpec

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s; want POST", r.Method)
		}
		if r.URL.Path != "/render" {
			t.Errorf("path = %s; want /render", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q; want application/json", ct)
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &gotSpec); err != nil {
			t.Fatalf("unmarshal request: %v", err)
		}
		w.Header().Set("Content-Type", "audio/wav")
		w.Header().Set("X-Generation-Time-Seconds", "0.5")
		_, _ = w.Write(wantWAV)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, 2*time.Second)
	spec := RenderSpec{
		Prompt:          "warm lofi rhodes",
		Tags:            []string{"lofi", "rhodes", "downtempo"},
		Key:             "Cmin",
		Tempo:           78,
		DurationSeconds: 90,
		Scale:           "minor",
		TimeSignature:   "4/4",
		Seed:            71003,
		HarmonyChain:    "Am7 Fmaj7 Dm7 G7sus",
		Motif:           "stepwise minor 5 7 5 3",
		InferenceSteps:  8,
	}
	wav, err := c.Render(context.Background(), spec)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(wav) != len(wantWAV) {
		t.Errorf("body length = %d; want %d", len(wav), len(wantWAV))
	}
	if string(wav[0:4]) != "RIFF" || string(wav[8:12]) != "WAVE" {
		t.Errorf("response is not a valid WAV header")
	}

	// Verify the wire payload made the trip intact.
	if gotSpec.Prompt != spec.Prompt {
		t.Errorf("Prompt round-trip: got %q want %q", gotSpec.Prompt, spec.Prompt)
	}
	if len(gotSpec.Tags) != len(spec.Tags) {
		t.Errorf("Tags round-trip: got %v want %v", gotSpec.Tags, spec.Tags)
	}
	if gotSpec.Tempo != spec.Tempo {
		t.Errorf("Tempo round-trip: got %d want %d", gotSpec.Tempo, spec.Tempo)
	}
	if gotSpec.HarmonyChain != spec.HarmonyChain {
		t.Errorf("HarmonyChain round-trip: got %q want %q", gotSpec.HarmonyChain, spec.HarmonyChain)
	}
	if gotSpec.InferenceSteps != spec.InferenceSteps {
		t.Errorf("InferenceSteps round-trip: got %d want %d", gotSpec.InferenceSteps, spec.InferenceSteps)
	}
}

func TestClient_RenderRejectsNonWAVBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/wav")
		_, _ = w.Write([]byte("not actually a wav"))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, time.Second)
	_, err := c.Render(context.Background(), RenderSpec{Prompt: "x"})
	if err == nil {
		t.Fatalf("expected error on non-WAV body, got nil")
	}
	if !strings.Contains(err.Error(), "not WAV") {
		t.Errorf("error should mention not WAV: %v", err)
	}
}

func TestClient_RenderServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"detail":"model not loaded"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, time.Second)
	_, err := c.Render(context.Background(), RenderSpec{Prompt: "x"})
	if err == nil {
		t.Fatalf("expected error on 503, got nil")
	}
	if !strings.Contains(err.Error(), "status=503") {
		t.Errorf("missing status code in error: %v", err)
	}
}

func TestClient_RenderContextCancel(t *testing.T) {
	// done is closed by the test once the client call has returned. The
	// handler waits on done OR the request context, whichever fires first.
	// Without the test-owned channel, srv.Close() can block forever on
	// macOS when the client-side disconnect doesn't propagate promptly to
	// the server-side r.Context().
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-done:
		}
	}))
	defer srv.Close()
	defer close(done)
	// Aggressively drop any in-flight connections so Close() doesn't hang
	// even if the handler is still running.
	defer srv.CloseClientConnections()

	c := NewClient(srv.URL, 5*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	_, err := c.Render(ctx, RenderSpec{Prompt: "x"})
	if err == nil {
		t.Fatalf("expected cancellation error")
	}
}

func TestNewClient_TrimsTrailingSlash(t *testing.T) {
	c := NewClient("http://example.com:7790/", 0)
	if c.baseURL != "http://example.com:7790" {
		t.Errorf("baseURL = %q; want http://example.com:7790", c.baseURL)
	}
}

func TestNewClient_DefaultTimeout(t *testing.T) {
	c := NewClient("http://example.com", 0)
	if c.http.Timeout <= 0 {
		t.Errorf("expected positive default timeout, got %v", c.http.Timeout)
	}
}

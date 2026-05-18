package acestep

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is the HTTP client for the local ACE-Step daemon. It is safe for
// concurrent use; the streaming engine calls Render from a background
// generator goroutine while the main goroutine is still playing earlier
// buffers.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient builds a client pointing at the Python service.
//
// baseURL is typically "http://localhost:7790" (the service default).
//
// timeout applies to each request. The user should size it generously -
// generation on the 2B turbo model takes ~10s per track on M-series Macs
// once the model is warm, plus ~30s for the initial /health-driven warmup.
// A timeout of 5 minutes is reasonable.
func NewClient(baseURL string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

// HealthResponse mirrors the JSON shape of GET /health.
type HealthResponse struct {
	Loaded            bool    `json:"loaded"`
	Backend           string  `json:"backend"`
	ModelName         string  `json:"model_name"`
	LMModelName       string  `json:"lm_model_name"`
	MockMode          bool    `json:"mock_mode"`
	Error             string  `json:"error,omitempty"`
	LoadTimeSeconds   float64 `json:"load_time_seconds"`
}

// Health calls GET /health. Returns the full response and a non-nil error
// only on transport/parse failure - a service that responds with loaded=false
// is reported as a non-error result with Loaded=false.
func (c *Client) Health(ctx context.Context) (HealthResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return HealthResponse{}, fmt.Errorf("acestep: build health request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return HealthResponse{}, fmt.Errorf("acestep: GET %s/health: %w", c.baseURL, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return HealthResponse{}, fmt.Errorf("acestep: read health body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return HealthResponse{}, fmt.Errorf("acestep: GET /health: status=%d body=%s", resp.StatusCode, truncate(string(body), 200))
	}
	var h HealthResponse
	if err := json.Unmarshal(body, &h); err != nil {
		return HealthResponse{}, fmt.Errorf("acestep: parse health JSON: %w (body=%s)", err, truncate(string(body), 200))
	}
	return h, nil
}

// Render posts the spec and returns the WAV bytes. Returns an error if the
// service is not loaded, the request fails, or the response body is not WAV.
func (c *Client) Render(ctx context.Context, spec RenderSpec) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("acestep: nil client")
	}
	buf, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("acestep: marshal spec: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/render", bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("acestep: build render request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/wav")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("acestep: POST %s/render: %w", c.baseURL, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("acestep: read render body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		// Try to extract a {"detail": "..."} JSON error payload.
		return nil, fmt.Errorf("acestep: POST /render status=%d body=%s", resp.StatusCode, truncate(string(body), 400))
	}
	// Sanity check: the body should start with a RIFF/WAVE header.
	if len(body) < 12 || string(body[0:4]) != "RIFF" || string(body[8:12]) != "WAVE" {
		return nil, fmt.Errorf("acestep: response is not WAV (got %d bytes, first 12=%q)", len(body), truncate(string(body), 12))
	}
	return body, nil
}

// truncate returns s shortened to at most n bytes, with an ellipsis when truncated.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

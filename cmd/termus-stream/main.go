// termus-stream is the streaming-playback CLI for the SP21 ACE-Step engine.
//
// It loads a v3 .tm file, verifies render_engine=acestep, builds an
// internal/acestep.Client pointed at the local Python service, and feeds
// generated WAV bytes into an internal/audio.Streamer that maintains a
// look-ahead queue plus equal-power crossfades.
//
// UNTESTED end-to-end: the Python ACE-Step service requires a 5-10 GB model
// download via services/acestep/install.sh. Without that, the CLI's
// health-check fails fast and reports the cause. The Go side is exercised
// by unit tests with mocks.
//
// Usage:
//
//	termus-stream <track.tm>
//	  --service-url http://localhost:7790
//	  --crossfade 3
//	  --max-tracks N
//	  --queue-depth 2
//	  --output-dir <dir>
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/mrbrutti/termus/internal/acestep"
	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/track"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "termus-stream:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		serviceURL = flag.String("service-url", "http://localhost:7790", "ACE-Step service URL")
		crossfade  = flag.Float64("crossfade", 3.0, "crossfade duration in seconds")
		maxTracks  = flag.Int("max-tracks", 0, "stop after N tracks (0 = infinite)")
		queueDepth = flag.Int("queue-depth", 2, "look-ahead queue depth")
		outputDir  = flag.String("output-dir", "", "optional: also save each generated WAV here")
		timeout    = flag.Duration("render-timeout", 5*time.Minute, "per-render HTTP timeout")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: termus-stream [flags] <track.tm>\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		return fmt.Errorf("missing <track.tm>")
	}
	trackPath := flag.Arg(0)

	file, err := track.ParseFile(trackPath)
	if err != nil {
		return fmt.Errorf("parse %s: %w", trackPath, err)
	}
	if file.RenderEngine != track.RenderEngineACEStep {
		return fmt.Errorf("track %s has render_engine=%q; termus-stream only supports %q. "+
			"Use the standard termus binary for SF2-engine tracks.",
			trackPath, file.RenderEngine, track.RenderEngineACEStep)
	}
	if file.Acestep == nil {
		return fmt.Errorf("track %s sets render_engine=acestep but the 'acestep:' block is missing", trackPath)
	}

	if *outputDir != "" {
		if err := os.MkdirAll(*outputDir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", *outputDir, err)
		}
	}

	client := acestep.NewClient(*serviceURL, *timeout)

	// Health-check before going further; otherwise the first /render call
	// produces a confusing connection error.
	healthCtx, healthCancel := context.WithTimeout(context.Background(), 10*time.Second)
	health, err := client.Health(healthCtx)
	healthCancel()
	if err != nil {
		return fmt.Errorf("ACE-Step service health check failed at %s: %w. "+
			"Did you start services/acestep/server.py? See services/acestep/README.md.",
			*serviceURL, err)
	}
	if !health.Loaded {
		return fmt.Errorf("ACE-Step service at %s is reachable but not loaded "+
			"(backend=%s, mock=%v, error=%q). Wait for warmup or check service logs.",
			*serviceURL, health.Backend, health.MockMode, health.Error)
	}
	fmt.Fprintf(os.Stderr, "termus-stream: connected to ACE-Step service at %s "+
		"(backend=%s, model=%s, mock=%v, load_time=%.1fs)\n",
		*serviceURL, health.Backend, health.ModelName, health.MockMode, health.LoadTimeSeconds)

	// Producer: each call increments the seed offset so successive
	// renders are variations rather than bit-identical clones.
	baseSpec, err := acestep.CompileV3(file)
	if err != nil {
		return fmt.Errorf("compile v3: %w", err)
	}
	prod := &renderingProducer{
		client:    client,
		baseSpec:  baseSpec,
		outputDir: *outputDir,
		trackName: trackNameFromPath(trackPath),
	}

	// Build the streamer. Pass a real speaker sink (the default).
	s := audio.NewStreamer(audio.StreamerConfig{
		Producer:     prod,
		QueueDepth:   *queueDepth,
		CrossfadeSec: *crossfade,
		MaxTracks:    *maxTracks,
		Logger:       os.Stderr,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// SIGINT → graceful stop.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		fmt.Fprintf(os.Stderr, "termus-stream: received %v, stopping...\n", sig)
		cancel()
	}()

	if err := s.Start(ctx); err != nil {
		return fmt.Errorf("start streamer: %w", err)
	}

	// Heartbeat: print streamer status every 5s while running.
	statusDone := make(chan struct{})
	go statusHeartbeat(ctx, s, statusDone)

	<-ctx.Done()
	fmt.Fprintln(os.Stderr, "termus-stream: shutting down...")
	s.Stop()
	<-statusDone

	if last := s.Status().LastError; last != nil {
		return fmt.Errorf("streamer ended with error: %w", last)
	}
	fmt.Fprintln(os.Stderr, "termus-stream: done.")
	return nil
}

// renderingProducer turns each Produce(seq) call into an HTTP /render call
// against the ACE-Step service. Seed is offset by seq so successive tracks
// are variations.
type renderingProducer struct {
	client    *acestep.Client
	baseSpec  acestep.RenderSpec
	outputDir string
	trackName string
	count     atomic.Int64
}

func (p *renderingProducer) Produce(ctx context.Context, seq int) ([]byte, error) {
	spec := p.baseSpec
	// Seed offset per track so the queue contents differ from each other.
	// We avoid clobbering an explicit seed of -1 (which means "random
	// each time" on the service side).
	if spec.Seed >= 0 {
		spec.Seed += int64(seq)
	}
	t0 := time.Now()
	data, err := p.client.Render(ctx, spec)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "termus-stream: rendered seq=%d (%d bytes, %.1fs)\n",
		seq, len(data), time.Since(t0).Seconds())

	if p.outputDir != "" {
		path := filepath.Join(p.outputDir, fmt.Sprintf("%s-%03d.wav", p.trackName, seq))
		if err := os.WriteFile(path, data, 0o644); err != nil {
			// Saving is a side-effect; don't kill the stream if disk
			// is full.
			fmt.Fprintf(os.Stderr, "termus-stream: warning: write %s: %v\n", path, err)
		}
	}
	p.count.Add(1)
	return data, nil
}

func statusHeartbeat(ctx context.Context, s *audio.Streamer, done chan<- struct{}) {
	defer close(done)
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			st := s.Status()
			fmt.Fprintf(os.Stderr, "termus-stream: status: playing=%v seq=%d queue=%d\n",
				st.Playing, st.CurrentSeq, st.QueueDepth)
		}
	}
}

// trackNameFromPath returns the filename without extension for use in
// output-dir filenames.
func trackNameFromPath(p string) string {
	base := filepath.Base(p)
	ext := filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return base
}

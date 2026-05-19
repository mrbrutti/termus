package acestep

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// ProgressFunc receives per-event progress updates parsed from the daemon's
// stderr stream. percent is in [0, 1]; detail is a short human-readable label.
// May be called from any goroutine. Implementations must be cheap or buffer.
type ProgressFunc func(percent float64, detail string)

// renderProgressLine matches the structured progress lines server.py emits:
//
//	RENDER_PROGRESS: 0.512 Preparing inputs...
//
// The format is intentionally minimal so the parser stays simple. If the
// server-side format changes, update both ends together.
var renderProgressLine = regexp.MustCompile(`^RENDER_PROGRESS:\s+([0-9]*\.?[0-9]+)\s*(.*)$`)

// progressTee returns an io.Writer that:
//
//  1. passes every byte through to dst (for transparent debug logging), and
//  2. parses any complete line for RENDER_PROGRESS markers, calling fn with
//     the parsed (percent, detail) tuple.
//
// The returned writer is line-buffered: incomplete trailing bytes are
// buffered until a newline arrives. Goroutine-safe per writer instance,
// but not safe for concurrent writers (matches io.Writer's documented
// contract).
func progressTee(dst io.Writer, fn ProgressFunc) io.Writer {
	if fn == nil {
		// No callback wired — short-circuit to a plain passthrough.
		if dst == nil {
			return io.Discard
		}
		return dst
	}
	pr, pw := io.Pipe()
	if dst == nil {
		dst = io.Discard
	}
	go func() {
		defer pr.Close()
		scanner := bufio.NewScanner(pr)
		// Allow long progress lines (some include long descriptions).
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			if m := renderProgressLine.FindStringSubmatch(line); m != nil {
				if v, err := strconv.ParseFloat(m[1], 64); err == nil {
					fn(clamp01(v), strings.TrimSpace(m[2]))
				}
			}
		}
	}()
	return &teeWriter{dst: dst, parse: pw}
}

// teeWriter writes to both dst and parse. Errors on parse (pipe closed) are
// ignored so a slow parser never blocks the producer.
type teeWriter struct {
	dst   io.Writer
	parse io.Writer
}

func (t *teeWriter) Write(p []byte) (int, error) {
	n, err := t.dst.Write(p)
	if n < len(p) {
		// Best-effort: still forward what dst accepted to the parser.
		_, _ = t.parse.Write(p[:n])
		return n, err
	}
	_, _ = t.parse.Write(p)
	return n, err
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

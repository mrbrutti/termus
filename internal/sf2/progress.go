package sf2

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// ProgressBar is a tiny, dependency-free CLI progress bar. It writes
// carriage-return updates to w (typically stderr) at most once per ~50 ms
// so terminals don't redraw every chunk.
type ProgressBar struct {
	w           io.Writer
	label       string
	width       int
	lastWrite   time.Time
	finished    bool
	total       int64
	lastPercent int
}

// NewProgressBar returns a bar that draws into w with `label` on the left and
// `width` of bar characters. Use Update() during a download and Finish() to
// emit the trailing newline.
func NewProgressBar(w io.Writer, label string, width int) *ProgressBar {
	if width < 8 {
		width = 8
	}
	return &ProgressBar{w: w, label: label, width: width, lastPercent: -1}
}

// Update reports byte progress. total == -1 means "unknown size" — the bar
// will display a spinning marker instead of a percentage.
func (p *ProgressBar) Update(done, total int64) {
	if p.finished {
		return
	}
	p.total = total
	now := time.Now()
	if !p.lastWrite.IsZero() && now.Sub(p.lastWrite) < 50*time.Millisecond {
		return
	}
	p.lastWrite = now
	p.render(done, total)
}

func (p *ProgressBar) render(done, total int64) {
	var bar string
	var trail string
	if total > 0 {
		pct := int(done * 100 / total)
		if pct < 0 {
			pct = 0
		}
		if pct > 100 {
			pct = 100
		}
		p.lastPercent = pct
		filled := pct * p.width / 100
		bar = strings.Repeat("█", filled) + strings.Repeat("░", p.width-filled)
		trail = fmt.Sprintf(" %3d%% %s / %s",
			pct, fmtBytes(done), fmtBytes(total))
	} else {
		// Unknown total: spinning marker.
		const ticks = "|/-\\"
		i := int(time.Now().UnixMilli()/100) % len(ticks)
		bar = strings.Repeat("░", p.width)
		trail = fmt.Sprintf(" %c %s", ticks[i], fmtBytes(done))
	}
	fmt.Fprintf(p.w, "\r%s [%s]%s ", p.label, bar, trail)
}

// Finish forces a final render at 100% and emits a newline so the next bar
// starts on a fresh line.
func (p *ProgressBar) Finish() {
	if p.finished {
		return
	}
	p.finished = true
	if p.total > 0 {
		p.render(p.total, p.total)
	}
	fmt.Fprintln(p.w)
}

func fmtBytes(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

// EnsureAll downloads and verifies each of the requested presets, returning a
// map of presetName → on-disk path. Already-cached presets resolve instantly
// (no bar is drawn). Each fresh download gets its own progress bar on `w`
// (typically os.Stderr).
func EnsureAll(w io.Writer, presets []string) (map[string]string, error) {
	out := make(map[string]string, len(presets))
	seen := make(map[string]bool, len(presets))
	for _, name := range presets {
		if seen[name] {
			continue
		}
		seen[name] = true
		p, ok := Presets[name]
		if !ok {
			return nil, fmt.Errorf("unknown sf2 preset %q", name)
		}
		// Lazy bar: only create + draw if a download actually starts.
		var bar *ProgressBar
		path, err := EnsurePreset(name, func(done, total int64) {
			if bar == nil {
				bar = NewProgressBar(w,
					fmt.Sprintf("%s (~%d MB)", p.FileName, p.SizeMB), 24)
			}
			bar.Update(done, total)
		})
		if bar != nil {
			bar.Finish()
		}
		if err != nil {
			return nil, err
		}
		out[name] = path
	}
	return out, nil
}

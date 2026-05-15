package sf2

import (
	"bytes"
	"strings"
	"testing"
)

func TestProgressBarRendersPercent(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(&buf, "test.sf2", 20)
	// Force-render even though we just created the bar.
	bar.lastWrite = bar.lastWrite.Add(-1) // make idle interval pass
	bar.Update(50, 100)
	out := buf.String()
	if !strings.Contains(out, "50%") {
		t.Errorf("missing percent in %q", out)
	}
	if !strings.Contains(out, "test.sf2") {
		t.Errorf("missing label in %q", out)
	}
	bar.Finish()
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Errorf("Finish should emit a trailing newline; got %q", buf.String())
	}
}

func TestFmtBytes(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{2048, "2.0 KB"},
		{2 << 20, "2.0 MB"},
		{int64(2) << 30, "2.0 GB"},
	}
	for _, c := range cases {
		if got := fmtBytes(c.in); got != c.want {
			t.Errorf("fmtBytes(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

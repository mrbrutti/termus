package audio

import "testing"

func TestSF2AlgoForACEStepPath(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"tracks/ambient/glacial-slow-drift.tm", "ambient"},
		{"tracks/blues/crossroads-prayer.tm", "jazz"},
		{"tracks/chill/sunset-balcony-loop.tm", "lofi"},
		{"tracks/jazz/smoke-and-mirrors.tm", "jazz"},
		{"tracks/lofi/late-night-letter.tm", "lofi"},
		{"tracks/rock/highway-static.tm", "lofi"},
		{"/abs/path/to/tracks/Jazz/Upper-Case.tm", "jazz"}, // case-insensitive
		{"tracks/unknown/foo.tm", ""},
		{"", ""},
	}
	for _, c := range cases {
		got := SF2AlgoForACEStepPath(c.path)
		if got != c.want {
			t.Errorf("SF2AlgoForACEStepPath(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

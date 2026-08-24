package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrbrutti/termus/internal/gen"
)

func splashTestModel() Model {
	ambient, _ := gen.Resolve("ambient")
	drone, _ := gen.Resolve("drone")
	m := Model{width: 118, height: 32, algo: ambient.Label(), cmd: &tuiCommanderStub{}}
	m.genres = []gen.AlgoSpec{ambient, drone}
	m.genreIdx = 0
	m.buildFn = func(spec gen.AlgoSpec, seed int64) gen.Algorithm { return &tuiAlgoStub{name: spec.Name} }
	m.splashVisible = true
	m.splashUntil = time.Now().Add(5 * time.Second)
	return m
}

func TestSplashScreenShowsWordmarkAndDial(t *testing.T) {
	m := splashTestModel()
	out := splashScreen(m, 118, 32, DefaultTheme(), time.Now())
	if !strings.Contains(out, "█") {
		t.Fatalf("splash missing block wordmark: %q", out)
	}
	if !strings.Contains(out, "a terminal music instrument") {
		t.Fatalf("splash missing tagline: %q", out)
	}
	if !strings.Contains(out, "NIGHT DRIFT") {
		t.Fatalf("splash missing selected station: %q", out)
	}
	if !strings.Contains(out, "deep field") {
		t.Fatalf("splash missing other stations: %q", out)
	}
	if !strings.Contains(out, "[←→] choose a station · [enter] begin · [t] authored tracks") {
		t.Fatalf("splash missing footer: %q", out)
	}
}

func TestSplashArrowKeysSwitchStationWithoutDismissing(t *testing.T) {
	m := splashTestModel()
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	updated := next.(Model)
	if !updated.splashVisible {
		t.Fatal("right arrow should not dismiss the splash")
	}
	if updated.genreIdx != 1 {
		t.Fatalf("genreIdx = %d, want 1 after right arrow", updated.genreIdx)
	}
	next2, _ := updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if next2.(Model).splashVisible {
		t.Fatal("enter should dismiss the splash")
	}
}

// The dial re-arms the auto-dismiss timer so browsing stations does not race
// the 5s countdown.
func TestSplashArrowKeysRearmAutoDismiss(t *testing.T) {
	m := splashTestModel()
	m.splashUntil = time.Now().Add(200 * time.Millisecond)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	updated := next.(Model)
	if !updated.splashUntil.After(time.Now().Add(3 * time.Second)) {
		t.Fatalf("left arrow should re-arm the splash timer, got %v", updated.splashUntil)
	}
	if updated.genreIdx != 1 {
		t.Fatalf("genreIdx = %d, want 1 after left arrow (wrap)", updated.genreIdx)
	}
}

func TestStartupLoadingShowsProgressOnSplash(t *testing.T) {
	m := splashTestModel()
	m.startupLoading = true
	m.startupTitle = "Loading termus"
	m.startupDetail = "loading SoundFont preset"
	m.startupPercent = 0.42
	out := splashScreen(m, 118, 32, DefaultTheme(), time.Now())
	if !strings.Contains(out, "42%") || !strings.Contains(out, "loading SoundFont preset") {
		t.Fatalf("loading splash missing progress: %q", out)
	}
}

// Keys stay swallowed while the startup loader is up: the dial must not move
// the station out from under a load that is already in flight.
func TestStartupLoadingSwallowsDialKeys(t *testing.T) {
	m := splashTestModel()
	m.startupLoading = true
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	got := next.(Model)
	if got.genreIdx != 0 {
		t.Fatalf("genreIdx = %d, want 0 — arrows are swallowed while loading", got.genreIdx)
	}
	if !got.splashVisible || !got.startupLoading {
		t.Fatal("startup loading should keep the splash up until loading completes")
	}
}

// TestSplashScreenNeverExceedsSize sweeps every runnable width at four
// heights, loading and idle, with and without stations. The splash is a
// full-screen centered stack, so at h=10 the full stack (wordmark, tagline,
// blanks, dial, braille bar, progress, footer) does not fit — it has to
// degrade rather than push rows off screen or wrap a row past the width.
func TestSplashScreenNeverExceedsSize(t *testing.T) {
	theme := DefaultTheme()
	now := time.Unix(0, 0)
	for _, h := range []int{10, 14, 18, 32} {
		for _, loading := range []bool{false, true} {
			for _, withGenres := range []bool{false, true} {
				for w := 40; w <= 200; w++ {
					m := splashTestModel()
					if !withGenres {
						m.genres = nil
					}
					if loading {
						m.startupLoading = true
						m.startupTitle = "Loading MAX palette · Dusty Swing · jazz"
						m.startupDetail = "ready 1/2 · last sgm, tyros4, timbres-of-heaven"
						m.startupPercent = 0.5
						m.aceContextTitle = "A Rather Long Authored Track Title For Width Testing"
						m.aceContextGenre = "lofi hip hop"
						m.aceContextStyle = "warm rhodes, dusty tape saturation, brushed kit, upright bass, slow swing"
						m.aceContextTags = []string{"warm", "dusty", "slow", "rhodes", "tape", "brushes", "night"}
					}
					out := splashScreen(m, w, h, theme, now)
					lines := strings.Split(out, "\n")
					if len(lines) > h {
						t.Fatalf("splash is %d rows tall at w=%d h=%d loading=%v genres=%v, want <= %d\n%s",
							len(lines), w, h, loading, withGenres, h, out)
					}
					for i, line := range lines {
						if got := lipgloss.Width(line); got > w {
							t.Fatalf("splash line %d overflows at w=%d h=%d loading=%v genres=%v: rendered %d cols\n%q",
								i, w, h, loading, withGenres, got, line)
						}
					}
				}
			}
		}
	}
}

// The wordmark and the footer are the last survivors of the degradation
// ladder: at the app's minimum size the identity and the way out still show.
func TestSplashKeepsWordmarkAndFooterAtMinimumSize(t *testing.T) {
	m := splashTestModel()
	out := splashScreen(m, 40, 10, DefaultTheme(), time.Unix(0, 0))
	if !strings.Contains(out, "█") {
		t.Fatalf("wordmark should survive at 40x10:\n%s", out)
	}
	if !strings.Contains(out, "[enter] begin") {
		t.Fatalf("footer should survive at 40x10:\n%s", out)
	}
	if !strings.Contains(out, "NIGHT DRIFT") {
		t.Fatalf("current station should survive at 40x10:\n%s", out)
	}
}

// While loading, the braille bar and the percent survive the squeeze even at
// the minimum height — that pair is the whole point of the loading screen.
func TestSplashKeepsLoadingProgressAtMinimumSize(t *testing.T) {
	m := splashTestModel()
	m.startupLoading = true
	m.startupTitle = "Loading termus"
	m.startupDetail = "loading SoundFont preset"
	m.startupPercent = 0.5
	out := splashScreen(m, 40, 10, DefaultTheme(), time.Unix(0, 0))
	if !strings.Contains(out, "50%") {
		t.Fatalf("percent should survive at 40x10:\n%s", out)
	}
	if !strings.ContainsAny(out, "⠄⡀⠤⠶⠒⠂⠦") {
		t.Fatalf("braille bar should survive at 40x10:\n%s", out)
	}
}

package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// AdaptiveUI captures the terminal capabilities that shape the startup UI.
// Themes are filtered/reordered so light terminals and low-color terminals
// don't start with unreadable palettes or controls that can't do anything.
type AdaptiveUI struct {
	ColorProfile    termenv.Profile
	DarkBackground  bool
	Themes          []ColorTheme
	DefaultThemeIdx int
}

// DetectAdaptiveUI inspects the current terminal and derives a sensible
// palette/controls profile for the TUI.
func DetectAdaptiveUI() AdaptiveUI {
	dark := lipgloss.HasDarkBackground()
	if noColor := os.Getenv("NO_COLOR"); noColor != "" {
		return detectAdaptiveUIWith(termenv.Ascii, dark)
	}
	return detectAdaptiveUIWith(termenv.EnvColorProfile(), dark)
}

func detectAdaptiveUIWith(profile termenv.Profile, dark bool) AdaptiveUI {
	switch profile {
	case termenv.Ascii:
		return AdaptiveUI{
			ColorProfile:    profile,
			DarkBackground:  dark,
			Themes:          []ColorTheme{themeNamed("mono")},
			DefaultThemeIdx: 0,
		}
	case termenv.ANSI:
		if dark {
			return AdaptiveUI{
				ColorProfile:    profile,
				DarkBackground:  true,
				Themes:          []ColorTheme{themeNamed("matrix"), themeNamed("amber"), themeNamed("mono")},
				DefaultThemeIdx: 0,
			}
		}
		return AdaptiveUI{
			ColorProfile:    profile,
			DarkBackground:  false,
			Themes:          []ColorTheme{themeNamed("amber"), themeNamed("mono")},
			DefaultThemeIdx: 0,
		}
	case termenv.ANSI256:
		if dark {
			return AdaptiveUI{
				ColorProfile:    profile,
				DarkBackground:  true,
				Themes:          []ColorTheme{themeNamed("indigo"), themeNamed("matrix"), themeNamed("amber"), themeNamed("mono")},
				DefaultThemeIdx: 0,
			}
		}
		return AdaptiveUI{
			ColorProfile:    profile,
			DarkBackground:  false,
			Themes:          []ColorTheme{themeNamed("amber"), themeNamed("indigo"), themeNamed("mono")},
			DefaultThemeIdx: 0,
		}
	default:
		if dark {
			return AdaptiveUI{
				ColorProfile:    profile,
				DarkBackground:  true,
				Themes:          append([]ColorTheme(nil), Themes...),
				DefaultThemeIdx: 0,
			}
		}
		return AdaptiveUI{
			ColorProfile:   profile,
			DarkBackground: false,
			Themes: []ColorTheme{
				themeNamed("amber"),
				themeNamed("indigo"),
				themeNamed("rainbow"),
				themeNamed("mono"),
			},
			DefaultThemeIdx: 0,
		}
	}
}

func themeNamed(name string) ColorTheme {
	for _, theme := range Themes {
		if theme.Name == name {
			return theme
		}
	}
	return DefaultTheme()
}

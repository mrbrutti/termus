package gen

import (
	"sort"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

// AlgoBuilder constructs an Algorithm given an optional SoundFont. sf is
// nil if RequiresSF2 is false.
type AlgoBuilder func(*meltysynth.SoundFont) Algorithm

// AlgoSpec describes one algorithm by its public name, requirements, and
// constructor.
type AlgoSpec struct {
	Name        string      // canonical CLI name (genre-style, e.g. "ambient")
	Display     string      // human label for UI (e.g. "Ambient")
	Tagline     string      // one-line description for help/playlists
	RequiresSF2 bool
	// PreferredSF2 names the SoundFont preset that sounds best for this
	// algorithm (e.g. "sgm" for piano-heavy genres). Empty string means
	// "no strong preference" — use whatever the user/default picks.
	PreferredSF2 string
	Build        AlgoBuilder
}

// Registry maps genre-style names to algorithm specs. The keys are the
// preferred public names; the aliases map below provides backwards-compat
// for the older internal names.
var registry = map[string]AlgoSpec{
	"ambient": {
		Name: "ambient", Display: "Ambient", RequiresSF2: true, PreferredSF2: "general",
		Tagline: "Music for Airports — pad-bell on incommensurate loops, sampled",
		Build:   func(sf *meltysynth.SoundFont) Algorithm { return NewSF2Eno(sf) },
	},
	"ambient-synth": {
		Name: "ambient-synth", Display: "Ambient (pure synth)", RequiresSF2: false,
		Tagline: "Same as ambient but no SoundFont download — synthesized voices",
		Build:   func(_ *meltysynth.SoundFont) Algorithm { return NewEno() },
	},
	"drone": {
		Name: "drone", Display: "Drone", RequiresSF2: true, PreferredSF2: "general",
		Tagline: "Stars of the Lid — held strings + flute shimmer over deep bed",
		Build:   func(sf *meltysynth.SoundFont) Algorithm { return NewSF2Drone(sf) },
	},
	"drone-synth": {
		Name: "drone-synth", Display: "Drone (pure synth)", RequiresSF2: false,
		Tagline: "Drone without SoundFont — synthesized sustained voices",
		Build:   func(_ *meltysynth.SoundFont) Algorithm { return NewDrone() },
	},
	"bells": {
		Name: "bells", Display: "Bells", RequiresSF2: true, PreferredSF2: "general",
		Tagline: "Tubular bells + crystal pad — bright, late-night focus",
		Build:   func(sf *meltysynth.SoundFont) Algorithm { return NewSF2Glass(sf) },
	},
	"bells-synth": {
		Name: "bells-synth", Display: "Bells (FM synth)", RequiresSF2: false,
		Tagline: "FM-synthesized bell tones — Aphex Twin SAW II",
		Build:   func(_ *meltysynth.SoundFont) Algorithm { return NewGlass() },
	},
	"lullaby": {
		Name: "lullaby", Display: "Lullaby", RequiresSF2: true, PreferredSF2: "sgm",
		Tagline: "Pentatonic random walk — piano, harp, kalimba — never clashes",
		Build:   func(sf *meltysynth.SoundFont) Algorithm { return NewSF2Pentatonic(sf) },
	},
	"lullaby-synth": {
		Name: "lullaby-synth", Display: "Lullaby (pure synth)", RequiresSF2: false,
		Tagline: "Pentatonic walk over synthesized pad-bell voices",
		Build:   func(_ *meltysynth.SoundFont) Algorithm { return NewPentatonic() },
	},
	"classical": {
		Name: "classical", Display: "Classical", RequiresSF2: true, PreferredSF2: "sgm",
		Tagline: "Markov melody on piano + strings + clarinet — feels composed",
		Build:   func(sf *meltysynth.SoundFont) Algorithm { return NewSF2Markov(sf) },
	},
	"classical-synth": {
		Name: "classical-synth", Display: "Classical (pure synth)", RequiresSF2: false,
		Tagline: "Markov melody on synthesized pad-bell voices",
		Build:   func(_ *meltysynth.SoundFont) Algorithm { return NewMarkov() },
	},
	"phase": {
		Name: "phase", Display: "Phase", RequiresSF2: true, PreferredSF2: "general",
		Tagline: "Reich-style — two vibraphones drift in tempo, ever-changing pattern",
		Build:   func(sf *meltysynth.SoundFont) Algorithm { return NewPhase(sf) },
	},
	"lofi": {
		Name: "lofi", Display: "Lo-fi", RequiresSF2: true, PreferredSF2: "sgm",
		Tagline: "Hip-hop drums + Rhodes EP + walking bass + sax + nylon guitar",
		Build:   func(sf *meltysynth.SoundFont) Algorithm { return NewChill(sf) },
	},
	"jazz": {
		Name: "jazz", Display: "Jazz", RequiresSF2: true, PreferredSF2: "sgm",
		Tagline: "Piano + strings + warm pad + bass + flute melody, A-B chord form",
		Build:   func(sf *meltysynth.SoundFont) Algorithm { return NewSF2(sf) },
	},
}

// aliases maps legacy algorithm names to their new genre names. The CLI
// keeps accepting old names but resolves them to the new specs.
var aliases = map[string]string{
	"eno":            "ambient-synth",
	"eno-sf2":        "ambient",
	"drone-sf2":      "drone",
	"glass":          "bells-synth",
	"glass-sf2":      "bells",
	"pentatonic":     "lullaby-synth",
	"pentatonic-sf2": "lullaby",
	"markov":         "classical-synth",
	"markov-sf2":     "classical",
	"chill":          "lofi",
	"sf2":            "jazz",
}

// Resolve returns the AlgoSpec for a name (genre name or legacy alias).
func Resolve(name string) (AlgoSpec, bool) {
	if alias, ok := aliases[name]; ok {
		name = alias
	}
	spec, ok := registry[name]
	return spec, ok
}

// AllAlgoNames returns the genre names sorted alphabetically. Used for
// help text and UI listings.
func AllAlgoNames() []string {
	names := make([]string, 0, len(registry))
	for k := range registry {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// AllGenreNames returns just the SF2-backed genre names — the "main"
// listing without the -synth fallbacks or the legacy aliases. These are
// the names a user normally interacts with.
func AllGenreNames() []string {
	out := make([]string, 0)
	for k, spec := range registry {
		if spec.RequiresSF2 && !endsWith(k, "-synth") {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

package main

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/sinshu/go-meltysynth/meltysynth"

	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/sf2"
)

type soundFontCatalog struct {
	strategy string
	fallback string

	mu     sync.RWMutex
	loadMu sync.Mutex
	byName map[string]*meltysynth.SoundFont

	warmOnce sync.Once
}

func newSoundFontCatalog(strategy, fallback string) *soundFontCatalog {
	return &soundFontCatalog{
		strategy: strategy,
		fallback: fallback,
		byName:   make(map[string]*meltysynth.SoundFont),
	}
}

func newPinnedSoundFontCatalog(path string) (*soundFontCatalog, *meltysynth.SoundFont, error) {
	loaded, err := sf2.Open(path)
	if err != nil {
		return nil, nil, err
	}
	c := newSoundFontCatalog("single", "")
	c.byName[""] = loaded
	gen.SetSF2Runtime("single", c.snapshot())
	return c, loaded, nil
}

func (c *soundFontCatalog) EnsureForSpec(spec gen.AlgoSpec, progress io.Writer) (*meltysynth.SoundFont, error) {
	if !spec.RequiresSF2 {
		return nil, nil
	}
	var presets []string
	switch c.strategy {
	case "max":
		presets = gen.MaxSF2PresetsForSpec(spec)
	case "pro":
		presets = neededPresets("pro", c.fallback, spec)
	default:
		presets = []string{c.fallback}
	}
	if err := c.ensurePresets(progress, presets); err != nil {
		return nil, err
	}
	return c.pick(spec), nil
}

func (c *soundFontCatalog) WarmMaxAsync() {
	if c == nil || c.strategy != "max" {
		return
	}
	c.warmOnce.Do(func() {
		go func() {
			_ = c.ensurePresets(io.Discard, sf2.AllPresetNames())
		}()
	})
}

func (c *soundFontCatalog) WarmCurrentSpecMaxAsync(spec gen.AlgoSpec, onReady func()) {
	if c == nil || c.strategy != "max" || !spec.RequiresSF2 {
		return
	}
	go func() {
		if err := c.ensurePresets(io.Discard, gen.MaxSF2PresetsForSpec(spec)); err != nil {
			logCatalogEnsureError(spec, err)
			return
		}
		if onReady != nil {
			onReady()
		}
	}()
}

func (c *soundFontCatalog) Pick(spec gen.AlgoSpec) *meltysynth.SoundFont {
	if c == nil {
		return nil
	}
	return c.pick(spec)
}

func (c *soundFontCatalog) PickName(spec gen.AlgoSpec) string {
	if c == nil || !spec.RequiresSF2 {
		return "synth"
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if _, ok := c.byName[spec.PreferredSF2]; ok && spec.PreferredSF2 != "" {
		return spec.PreferredSF2
	}
	if _, ok := c.byName[c.fallback]; ok && c.fallback != "" {
		return c.fallback
	}
	for name := range c.byName {
		if name != "" {
			return name
		}
	}
	if c.fallback != "" {
		return c.fallback
	}
	return "sf2"
}

func (c *soundFontCatalog) snapshot() map[string]*meltysynth.SoundFont {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cloned := make(map[string]*meltysynth.SoundFont, len(c.byName))
	for name, loaded := range c.byName {
		cloned[name] = loaded
	}
	return cloned
}

func (c *soundFontCatalog) pick(spec gen.AlgoSpec) *meltysynth.SoundFont {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return pickSF(c.byName, spec, c.fallback)
}

func (c *soundFontCatalog) ensurePresets(progress io.Writer, presets []string) error {
	missing := c.missingPresets(presets)
	if len(missing) == 0 {
		return nil
	}
	c.loadMu.Lock()
	defer c.loadMu.Unlock()

	missing = c.missingPresets(presets)
	if len(missing) == 0 {
		return nil
	}
	if progress == nil {
		progress = io.Discard
	}
	paths, err := sf2.EnsureAll(progress, missing)
	if err != nil {
		return err
	}
	loaded := make(map[string]*meltysynth.SoundFont, len(paths))
	for name, path := range paths {
		sf, err := sf2.Open(path)
		if err != nil {
			return err
		}
		loaded[name] = sf
	}
	c.mu.Lock()
	for name, sf := range loaded {
		c.byName[name] = sf
	}
	c.mu.Unlock()
	gen.SetSF2Runtime(c.strategy, c.snapshot())
	return nil
}

func (c *soundFontCatalog) missingPresets(presets []string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	seen := make(map[string]bool, len(presets))
	out := make([]string, 0, len(presets))
	for _, name := range presets {
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		if _, ok := c.byName[name]; ok {
			continue
		}
		out = append(out, name)
	}
	return out
}

func loadInitialSoundFontCatalog(spec gen.AlgoSpec, strategy, fallbackPreset, customPath string, progress io.Writer) (*soundFontCatalog, *meltysynth.SoundFont, error) {
	if !spec.RequiresSF2 {
		return nil, nil, nil
	}
	if customPath != "" {
		return newPinnedSoundFontCatalog(customPath)
	}
	catalog := newSoundFontCatalog(strategy, fallbackPreset)
	loaded, err := catalog.ensureForStartup(spec, progress)
	if err != nil {
		return nil, nil, err
	}
	return catalog, loaded, nil
}

func (c *soundFontCatalog) ensureForStartup(spec gen.AlgoSpec, progress io.Writer) (*meltysynth.SoundFont, error) {
	if !spec.RequiresSF2 {
		return nil, nil
	}
	var presets []string
	switch c.strategy {
	case "max":
		presets = startupPresetsForMax(spec, c.fallback)
	default:
		presets = neededPresets(c.strategy, c.fallback, spec)
	}
	if err := c.ensurePresets(progress, presets); err != nil {
		return nil, err
	}
	if c.strategy == "max" {
		if c.fallback != "" {
			c.mu.RLock()
			fallback := c.byName[c.fallback]
			c.mu.RUnlock()
			if fallback != nil {
				return fallback, nil
			}
		}
	}
	return c.pick(spec), nil
}

func startupPresetsForMax(spec gen.AlgoSpec, fallback string) []string {
	if fallback != "" {
		return []string{fallback}
	}
	if spec.PreferredSF2 != "" {
		return []string{spec.PreferredSF2}
	}
	return gen.MaxSF2PresetsForSpec(spec)
}

func logCatalogEnsureError(spec gen.AlgoSpec, err error) {
	if err == nil {
		return
	}
	_, _ = io.WriteString(os.Stderr, "sf2 warm for "+spec.Name+" failed: "+err.Error()+"\n")
}

func fastForwardAlgorithm(algo gen.Algorithm, elapsed time.Duration) {
	if algo == nil || elapsed <= 0 {
		return
	}
	const sampleRate = 44100
	const block = 2048
	const maxCatchup = 8 * time.Second
	if elapsed > maxCatchup {
		elapsed = maxCatchup
	}
	frames := int(float64(sampleRate) * elapsed.Seconds())
	left := make([]float64, block)
	right := make([]float64, block)
	for frames > 0 {
		n := block
		if frames < n {
			n = frames
		}
		algo.Next(left[:n], right[:n])
		frames -= n
	}
}

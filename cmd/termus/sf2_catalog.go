package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"

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

type catalogLoadUpdate struct {
	Ready   int
	Total   int
	Percent float64
	Detail  string
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
			_ = c.ensurePresetsParallel(sf2.AllPresetNames(), 2, nil)
		}()
	})
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

func (c *soundFontCatalog) ensurePresetsParallel(presets []string, concurrency int, update func(catalogLoadUpdate)) error {
	missing := c.missingPresets(presets)
	if len(missing) == 0 {
		if update != nil {
			update(catalogLoadUpdate{Ready: 0, Total: 0, Percent: 1.0, Detail: "ready"})
		}
		return nil
	}
	c.loadMu.Lock()
	defer c.loadMu.Unlock()

	missing = c.missingPresets(presets)
	if len(missing) == 0 {
		if update != nil {
			update(catalogLoadUpdate{Ready: 0, Total: 0, Percent: 1.0, Detail: "ready"})
		}
		return nil
	}
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > len(missing) {
		concurrency = len(missing)
	}
	type result struct {
		name string
		sf   *meltysynth.SoundFont
		err  error
	}
	totalWeight := 0
	for _, name := range missing {
		totalWeight += presetLoadWeight(name)
	}
	sort.Strings(missing)
	if update != nil {
		update(catalogLoadUpdate{
			Ready:   0,
			Total:   len(missing),
			Percent: 0,
			Detail:  "loading " + strings.Join(missing, ", "),
		})
	}
	jobs := make(chan string, len(missing))
	results := make(chan result, len(missing))
	for i := 0; i < concurrency; i++ {
		go func() {
			for name := range jobs {
				path, err := sf2.EnsurePreset(name, nil)
				if err != nil {
					results <- result{name: name, err: err}
					continue
				}
				loaded, err := sf2.Open(path)
				results <- result{name: name, sf: loaded, err: err}
			}
		}()
	}
	for _, name := range missing {
		jobs <- name
	}
	close(jobs)

	ready := 0
	doneWeight := 0
	for range missing {
		res := <-results
		if res.err != nil {
			return res.err
		}
		c.mu.Lock()
		c.byName[res.name] = res.sf
		c.mu.Unlock()
		gen.SetSF2Runtime(c.strategy, c.snapshot())
		ready++
		doneWeight += presetLoadWeight(res.name)
		if update != nil {
			update(catalogLoadUpdate{
				Ready:   ready,
				Total:   len(missing),
				Percent: float64(doneWeight) / float64(totalWeight),
				Detail:  fmt.Sprintf("ready %d/%d · last %s", ready, len(missing), res.name),
			})
		}
	}
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
	loaded, err := catalog.EnsureForSpec(spec, progress)
	if err != nil {
		return nil, nil, err
	}
	return catalog, loaded, nil
}

func logCatalogEnsureError(spec gen.AlgoSpec, err error) {
	if err == nil {
		return
	}
	_, _ = io.WriteString(os.Stderr, "sf2 warm for "+spec.Name+" failed: "+err.Error()+"\n")
}

func presetLoadWeight(name string) int {
	if preset, ok := sf2.Presets[name]; ok && preset.SizeMB > 0 {
		return preset.SizeMB
	}
	return 1
}

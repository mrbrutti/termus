package track

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func DefaultRoots() []string {
	var roots []string
	if cwd, err := os.Getwd(); err == nil {
		roots = append(roots, filepath.Join(cwd, "tracks"))
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		roots = append(roots, filepath.Join(dir, "tracks"))
		roots = append(roots, filepath.Join(filepath.Dir(dir), "tracks"))
	}
	return dedupePaths(roots)
}

func Discover(roots ...string) ([]Entry, error) {
	if len(roots) == 0 {
		roots = DefaultRoots()
	}
	seen := map[string]bool{}
	var entries []Entry
	for _, root := range roots {
		info, err := os.Stat(root)
		if err != nil || !info.IsDir() {
			continue
		}
		err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".tm" {
				return nil
			}
			entry, err := loadEntry(root, path)
			if err != nil {
				return err
			}
			if seen[entry.ID] {
				return nil
			}
			seen[entry.ID] = true
			entries = append(entries, entry)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	return entries, nil
}

func Resolve(entries []Entry, input string) (Entry, bool) {
	input = strings.TrimSpace(input)
	for _, entry := range entries {
		if entry.ID == input || entry.Path == input {
			return entry, true
		}
	}
	if input != "" && filepath.Ext(input) == ".tm" {
		if info, err := os.Stat(input); err == nil && !info.IsDir() {
			file, err := ParseFile(input)
			if err != nil {
				return Entry{}, false
			}
			id := strings.TrimSuffix(filepath.Base(input), ".tm")
			style := file.Style
			if style == "" {
				style = filepath.Base(filepath.Dir(input))
			}
			pack := resolveStylePack(style, file.Substyle, file.Title, file.Tags)
			sections, structure, ensemble, eventCount, complexity := buildEntrySummary(file, pack)
			return Entry{
				ID:           id,
				Path:         input,
				Style:        style,
				Substyle:     pack.Substyle,
				Title:        file.Title,
				Description:  file.Description,
				Tags:         append([]string(nil), file.Tags...),
				Key:          file.Key,
				Tempo:        file.Tempo,
				ListenMode:   file.ListenMode,
				SectionCount: len(sections),
				Sections:     sections,
				Ensemble:     ensemble,
				EventCount:   eventCount,
				Complexity:   complexity,
				Structure:    structure,
			}, true
		}
	}
	if filepath.Ext(input) == ".tm" {
		for _, entry := range entries {
			if filepath.Base(entry.Path) == filepath.Base(input) {
				return entry, true
			}
		}
	}
	return Entry{}, false
}

func loadEntry(root, path string) (Entry, error) {
	file, err := ParseFile(path)
	if err != nil {
		return Entry{}, fmt.Errorf("parse track %s: %w", path, err)
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return Entry{}, err
	}
	rel = strings.TrimSuffix(filepath.ToSlash(rel), ".tm")
	style := file.Style
	if style == "" {
		style = filepath.Dir(rel)
	}
	pack := resolveStylePack(style, file.Substyle, file.Title, file.Tags)
	sections, structure, ensemble, eventCount, complexity := buildEntrySummary(file, pack)
	return Entry{
		ID:           rel,
		Path:         path,
		Style:        style,
		Substyle:     pack.Substyle,
		Title:        file.Title,
		Description:  file.Description,
		Tags:         append([]string(nil), file.Tags...),
		Key:          file.Key,
		Tempo:        file.Tempo,
		ListenMode:   file.ListenMode,
		SectionCount: len(sections),
		Sections:     sections,
		Ensemble:     ensemble,
		EventCount:   eventCount,
		Complexity:   complexity,
		Structure:    structure,
	}, nil
}

func buildEntrySummary(file *File, pack stylePack) ([]string, []EntrySection, []string, int, string) {
	// SP19: if the file is form-only (no explicit sections) expand the form
	// template here so the Discover summary captures structure metadata.
	// Compile() does the same expansion later — this duplication keeps the
	// summary in sync.
	if len(file.Sections) == 0 && strings.TrimSpace(file.Form) != "" {
		if template, ok := ResolveForm(file.Form); ok {
			bpm := resolveBPMHint(file.Tempo, template.DefaultBPM)
			file.Sections = expandFormTemplate(template, bpm)
		}
	}
	sections, err := resolveSections(file)
	if err != nil {
		sections = append([]Section(nil), file.Sections...)
	}
	titles := make([]string, 0, len(sections))
	structure := make([]EntrySection, 0, len(sections))
	ensembleSeen := map[string]bool{}
	ensemble := make([]string, 0, 8)
	totalEvents := 0
	for _, section := range sections {
		roles := resolvedSectionRoles(file, section)
		section, roles = applyStyleLibrary(pack, section, roles)
		label := firstNonBlank(section.Title, section.ID)
		if strings.TrimSpace(label) != "" {
			titles = append(titles, label)
		}
		roleNames := entryRoleLabels(roles)
		for _, roleName := range roleNames {
			if ensembleSeen[roleName] {
				continue
			}
			ensembleSeen[roleName] = true
			ensemble = append(ensemble, roleName)
		}
		events := reviewEventLabels(sectionEvents(section))
		totalEvents += len(events)
		structure = append(structure, EntrySection{
			ID:        section.ID,
			Label:     label,
			Harmony:   section.Harmony,
			RoleNames: roleNames,
			Events:    append([]string(nil), events...),
		})
	}
	return titles, structure, ensemble, totalEvents, entryComplexity(len(structure), totalEvents, len(ensemble))
}

func sectionTitles(sections []Section) []string {
	out := make([]string, 0, len(sections))
	for _, section := range sections {
		label := strings.TrimSpace(section.Title)
		if label == "" {
			label = strings.TrimSpace(section.ID)
		}
		if label == "" {
			continue
		}
		out = append(out, label)
	}
	return out
}

func entryRoleLabels(roles map[string]Role) []string {
	names := sortedActiveRoleNames(roles)
	seen := map[string]bool{}
	out := make([]string, 0, len(names))
	for _, name := range names {
		label := compactRoleLabel(name, roles[name])
		if label == "" || seen[label] {
			continue
		}
		seen[label] = true
		out = append(out, label)
	}
	return out
}

func compactRoleLabel(name string, role Role) string {
	family := strings.ToLower(strings.TrimSpace(role.Family))
	switch {
	case family == "electric_piano":
		return "ep"
	case family == "acoustic_piano" || strings.Contains(family, "piano"):
		return "pno"
	case family == "guitar" || strings.Contains(family, "guitar"):
		return "gtr"
	case family == "organ":
		return "org"
	case family == "bass" || family == "synth_bass" || strings.Contains(family, "bass"):
		return "bass"
	case family == "drums":
		return "drums"
	case family == "choir":
		return "choir"
	case family == "strings" || family == "string_ensemble":
		return "strings"
	case family == "pad" || strings.Contains(family, "pad"):
		return "pad"
	case family == "bells" || family == "glock" || family == "celesta":
		return "bells"
	case family == "harp":
		return "harp"
	case family == "mallet" || family == "vibraphone":
		return "mallet"
	case family == "flute" || family == "clarinet" || family == "reed_lead" || family == "sax":
		return "reed"
	case family == "trumpet" || family == "brass":
		return "brass"
	}
	switch authoredRoleKind(name, role) {
	case "drum":
		return "drums"
	case "bass":
		return "bass"
	case "pad":
		return "pad"
	case "melody":
		return "lead"
	default:
		if family != "" {
			return family
		}
		return strings.ToLower(strings.TrimSpace(name))
	}
}

func entryComplexity(sectionCount, eventCount, ensembleCount int) string {
	score := sectionCount*2 + eventCount + ensembleCount
	switch {
	case score >= 18:
		return "through"
	case score >= 11:
		return "arranged"
	default:
		return "lean"
	}
}

func dedupePaths(paths []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		out = append(out, path)
	}
	return out
}

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
			return Entry{
				ID:          id,
				Path:        input,
				Style:       style,
				Title:       file.Title,
				Description: file.Description,
				Tags:        append([]string(nil), file.Tags...),
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
	return Entry{
		ID:          rel,
		Path:        path,
		Style:       style,
		Title:       file.Title,
		Description: file.Description,
		Tags:        append([]string(nil), file.Tags...),
		Key:         file.Key,
		Tempo:       file.Tempo,
		ListenMode:  file.ListenMode,
		Sections:    sectionTitles(file.Sections),
	}, nil
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

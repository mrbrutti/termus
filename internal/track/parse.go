package track

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// arrangementYAML is the shape used to decode `arrangement:` on a Section.
// The field accepts two forms:
//   - events-list (legacy):     arrangement: { events: [...] }
//   - role-schedule map (SP18): arrangement: { rhodes: { enter_bar: 1 }, ... }
// We use this private struct to peek and then route accordingly.
type arrangementYAML struct {
	// Events is the legacy SP1-SP16 events list.
	Events []Event `yaml:"events"`
}

// UnmarshalYAML on Section accepts the dual-shape arrangement field and any
// other SP18 fields. We use a private alias struct + explicit decode of the
// arrangement node to keep both code paths working.
func (s *Section) UnmarshalYAML(node *yaml.Node) error {
	type sectionAlias Section
	var alias sectionAlias
	if err := node.Decode(&alias); err != nil {
		return err
	}
	*s = Section(alias)
	// Walk the node's mapping to find the "arrangement" key, then decode it
	// once in each shape to pick the right one.
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		k := node.Content[i]
		v := node.Content[i+1]
		if k.Value != "arrangement" {
			continue
		}
		if v.Kind != yaml.MappingNode {
			continue
		}
		// Detect by scanning the inner keys: if any inner value is a mapping
		// (e.g. `rhodes: { enter_bar: 1 }`) and not "events", treat as SP18
		// role-schedule. Otherwise, if there is an "events" key, leave the
		// legacy parse alone.
		eventsKeyPresent := false
		sp18Like := false
		for j := 0; j+1 < len(v.Content); j += 2 {
			ik := v.Content[j]
			iv := v.Content[j+1]
			if ik.Value == "events" {
				eventsKeyPresent = true
				continue
			}
			if iv.Kind == yaml.MappingNode {
				sp18Like = true
			}
		}
		if eventsKeyPresent && !sp18Like {
			// Legacy: already decoded into s.Arrangement.Events via alias.
			break
		}
		// SP18 role-schedule form (possibly mixed with events).
		schedules := map[string]RoleSchedule{}
		// Decode role-by-role to avoid clobbering Events parsing.
		for j := 0; j+1 < len(v.Content); j += 2 {
			ik := v.Content[j]
			iv := v.Content[j+1]
			if ik.Value == "events" {
				continue
			}
			if iv.Kind != yaml.MappingNode {
				continue
			}
			var rs RoleSchedule
			if err := iv.Decode(&rs); err != nil {
				return fmt.Errorf("section arrangement role %q: %w", ik.Value, err)
			}
			schedules[ik.Value] = rs
		}
		if len(schedules) > 0 {
			s.Arrangement18 = schedules
		}
		break
	}
	return nil
}

// UnmarshalYAML allows ChordSpec to be authored as either a plain string
// (symbol only) or as a full mapping with optional voice-leading directives.
//
//	# string form:
//	harmony_chords: [Cmaj9, Am7]
//
//	# map form:
//	harmony_chords:
//	  - {chord: Cmaj9, voicing: drop2, top: "9"}
//	  - {chord: Am7, smooth: true}
func (c *ChordSpec) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		c.Symbol = strings.TrimSpace(node.Value)
		return nil
	case yaml.MappingNode:
		// Use a local alias to avoid recursion.
		type chordSpecAlias ChordSpec
		var alias chordSpecAlias
		if err := node.Decode(&alias); err != nil {
			return err
		}
		*c = ChordSpec(alias)
		return nil
	default:
		return fmt.Errorf("chord spec must be a string or mapping")
	}
}

func ParseFile(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

func Parse(data []byte) (*File, error) {
	var file File
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

func (m *MacroValue) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case 0:
		return nil
	case yaml.ScalarNode:
		m.set = true
		m.raw = strings.TrimSpace(node.Value)
		return nil
	default:
		return fmt.Errorf("macro value must be scalar")
	}
}

func (m MacroValue) Resolve() (int, bool, error) {
	if !m.set || m.raw == "" {
		return 0, false, nil
	}
	if n, err := strconv.Atoi(m.raw); err == nil {
		if n < 0 || n > 4 {
			return 0, false, fmt.Errorf("macro %q out of range 0..4", m.raw)
		}
		return n, true, nil
	}
	v, ok := macroAliases[strings.ToLower(m.raw)]
	if !ok {
		return 0, false, fmt.Errorf("unknown macro value %q", m.raw)
	}
	return v, true, nil
}

var macroAliases = map[string]int{
	"off": 0, "low": 0, "dry": 0, "dark": 0, "still": 0, "straight": 0, "short": 0,
	"sparse": 1, "warm": 1, "light": 1, "room": 1, "slow": 1, "gentle": 1, "soft": 1,
	"steady": 2, "balanced": 2, "natural": 2, "medium": 2, "center": 2, "normal": 2,
	"moving": 3, "bright": 3, "busy": 3, "halo": 3, "groove": 3, "long": 3,
	"full": 4, "glass": 4, "wash": 4, "restless": 4, "cathedral": 4, "heavy": 4,
}

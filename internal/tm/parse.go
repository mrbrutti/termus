package tm

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseFile(path string) (*Score, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

func Parse(data []byte) (*Score, error) {
	var score Score
	if err := yaml.Unmarshal(data, &score); err != nil {
		return nil, err
	}
	return &score, nil
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
	"sparse": 1, "warm": 1, "light": 1, "room": 1, "slow": 1, "gentle": 1,
	"steady": 2, "balanced": 2, "natural": 2, "medium": 2, "center": 2, "normal": 2,
	"moving": 3, "bright": 3, "busy": 3, "halo": 3, "groove": 3, "long": 3,
	"full": 4, "glass": 4, "wash": 4, "restless": 4, "cathedral": 4, "heavy": 4,
}

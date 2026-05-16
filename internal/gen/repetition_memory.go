package gen

import (
	"strconv"
	"strings"
)

type recentPatternMemory struct {
	limit   int
	entries []string
}

func newRecentPatternMemory(limit int) recentPatternMemory {
	if limit < 1 {
		limit = 1
	}
	return recentPatternMemory{limit: limit}
}

func (m *recentPatternMemory) remember(signature string) {
	if m == nil || signature == "" {
		return
	}
	m.entries = append(m.entries, signature)
	if len(m.entries) > m.limit {
		m.entries = append([]string(nil), m.entries[len(m.entries)-m.limit:]...)
	}
}

func (m *recentPatternMemory) penalty(signature string) int {
	if m == nil || signature == "" || len(m.entries) == 0 {
		return 0
	}
	score := 0
	for i, entry := range m.entries {
		if entry != signature {
			continue
		}
		recency := len(m.entries) - i
		score += 1 + recency
	}
	return score
}

func phraseSignature(phrase []int) string {
	if len(phrase) == 0 {
		return ""
	}
	var b strings.Builder
	for i, v := range phrase {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(v))
	}
	return b.String()
}

func boolSignature(bits []bool) string {
	if len(bits) == 0 {
		return ""
	}
	var b strings.Builder
	for _, bit := range bits {
		if bit {
			b.WriteByte('1')
		} else {
			b.WriteByte('0')
		}
	}
	return b.String()
}

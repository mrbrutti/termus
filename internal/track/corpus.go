package track

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed listening_corpus.yaml
var bundledListeningCorpus []byte

type ListeningCorpus struct {
	Genres map[string]GenreCorpus `yaml:"genres"`
}

type GenreCorpus struct {
	Canonical string       `yaml:"canonical"`
	Corpus    []string     `yaml:"corpus"`
	AB        []CorpusPair `yaml:"ab"`
}

type CorpusPair struct {
	Label   string  `yaml:"label"`
	A       string  `yaml:"a"`
	B       string  `yaml:"b"`
	Seconds float64 `yaml:"seconds,omitempty"`
}

func LoadBundledCorpus() (*ListeningCorpus, error) {
	var corpus ListeningCorpus
	if err := yaml.Unmarshal(bundledListeningCorpus, &corpus); err != nil {
		return nil, err
	}
	if corpus.Genres == nil {
		corpus.Genres = map[string]GenreCorpus{}
	}
	return &corpus, nil
}

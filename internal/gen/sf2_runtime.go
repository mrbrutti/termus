package gen

import (
	"sync"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

type sf2RuntimeState struct {
	strategy string
	fonts    map[string]*meltysynth.SoundFont
}

var sf2Runtime struct {
	mu    sync.RWMutex
	state sf2RuntimeState
}

func SetSF2Runtime(strategy string, fonts map[string]*meltysynth.SoundFont) {
	sf2Runtime.mu.Lock()
	defer sf2Runtime.mu.Unlock()
	cloned := make(map[string]*meltysynth.SoundFont, len(fonts))
	for name, sf := range fonts {
		cloned[name] = sf
	}
	sf2Runtime.state = sf2RuntimeState{
		strategy: strategy,
		fonts:    cloned,
	}
}

func currentSF2Runtime() sf2RuntimeState {
	sf2Runtime.mu.RLock()
	defer sf2Runtime.mu.RUnlock()
	cloned := make(map[string]*meltysynth.SoundFont, len(sf2Runtime.state.fonts))
	for name, sf := range sf2Runtime.state.fonts {
		cloned[name] = sf
	}
	return sf2RuntimeState{
		strategy: sf2Runtime.state.strategy,
		fonts:    cloned,
	}
}

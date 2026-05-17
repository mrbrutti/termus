package gen

import (
	"sync"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

type sf2RuntimeState struct {
	strategy string
	fonts    map[string]*meltysynth.SoundFont
	routes   map[string]map[int32]string
}

var sf2Runtime struct {
	mu    sync.RWMutex
	state sf2RuntimeState
}

func SetSF2Runtime(strategy string, fonts map[string]*meltysynth.SoundFont) {
	SetSF2RuntimeWithRoutes(strategy, fonts, nil)
}

func SetSF2RuntimeWithRoutes(strategy string, fonts map[string]*meltysynth.SoundFont, routes map[string]map[int32]string) {
	sf2Runtime.mu.Lock()
	defer sf2Runtime.mu.Unlock()
	cloned := make(map[string]*meltysynth.SoundFont, len(fonts))
	for name, sf := range fonts {
		cloned[name] = sf
	}
	clonedRoutes := make(map[string]map[int32]string, len(routes))
	for algo, route := range routes {
		inner := make(map[int32]string, len(route))
		for channel, preset := range route {
			inner[channel] = preset
		}
		clonedRoutes[algo] = inner
	}
	sf2Runtime.state = sf2RuntimeState{
		strategy: strategy,
		fonts:    cloned,
		routes:   clonedRoutes,
	}
}

func currentSF2Runtime() sf2RuntimeState {
	sf2Runtime.mu.RLock()
	defer sf2Runtime.mu.RUnlock()
	cloned := make(map[string]*meltysynth.SoundFont, len(sf2Runtime.state.fonts))
	for name, sf := range sf2Runtime.state.fonts {
		cloned[name] = sf
	}
	clonedRoutes := make(map[string]map[int32]string, len(sf2Runtime.state.routes))
	for algo, route := range sf2Runtime.state.routes {
		inner := make(map[int32]string, len(route))
		for channel, preset := range route {
			inner[channel] = preset
		}
		clonedRoutes[algo] = inner
	}
	return sf2RuntimeState{
		strategy: sf2Runtime.state.strategy,
		fonts:    cloned,
		routes:   clonedRoutes,
	}
}

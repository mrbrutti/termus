package gen

import "time"

type ListeningMode string

const (
	ListeningModeEndless    ListeningMode = "endless"
	ListeningModeAlbumSide  ListeningMode = "album-side"
	ListeningModeHourStream ListeningMode = "hour-stream"
	ListeningModeRadio      ListeningMode = "radio"
)

type ListeningModeSpec struct {
	Name                    ListeningMode
	Label                   string
	DefaultRenderSeconds    float64
	AutoPlaylistMode        string
	DefaultPlaylistTracks   int
	DefaultPlaylistDuration time.Duration
}

var listeningModes = map[ListeningMode]ListeningModeSpec{
	ListeningModeEndless: {
		Name:                 ListeningModeEndless,
		Label:                "endless",
		DefaultRenderSeconds: 180,
	},
	ListeningModeAlbumSide: {
		Name:                 ListeningModeAlbumSide,
		Label:                "album side",
		DefaultRenderSeconds: 24 * 60,
	},
	ListeningModeHourStream: {
		Name:                 ListeningModeHourStream,
		Label:                "hour stream",
		DefaultRenderSeconds: 60 * 60,
	},
	ListeningModeRadio: {
		Name:                    ListeningModeRadio,
		Label:                   "radio",
		DefaultRenderSeconds:    0,
		AutoPlaylistMode:        "mixed",
		DefaultPlaylistTracks:   8,
		DefaultPlaylistDuration: 7*time.Minute + 30*time.Second,
	},
}

func ResolveListeningMode(name string) (ListeningModeSpec, bool) {
	spec, ok := listeningModes[ListeningMode(name)]
	return spec, ok
}

func DefaultListeningMode() ListeningModeSpec {
	return listeningModes[ListeningModeEndless]
}

package audio

import "strings"

// SF2AlgoForACEStepPath maps an ACE-Step .tm file path to the SF2 algorithm
// name that best matches the track's genre, for the SF2->ACE-Step pre-roll
// bridge. SF2 keeps playing while the (slow) diffusion render is in flight
// so the user hears music in <1s of picking a track instead of staring at
// a silent loader for ~30-45s.
//
// Tracks live under tracks/<genre>/*.tm; the genre directory drives the
// mapping. Returns "" when no good match exists — caller should skip the
// bridge and just tear SF2 down immediately.
//
// Mapping:
//
//	ambient/* -> ambient (Night Drift)
//	blues/*   -> jazz    (existing registry alias)
//	chill/*   -> lofi    (closest tonal match)
//	jazz/*    -> jazz
//	lofi/*    -> lofi
//	rock/*    -> lofi    (existing registry alias)
func SF2AlgoForACEStepPath(path string) string {
	p := strings.ToLower(path)
	switch {
	case strings.Contains(p, "/ambient/"):
		return "ambient"
	case strings.Contains(p, "/blues/"):
		return "jazz"
	case strings.Contains(p, "/chill/"):
		return "lofi"
	case strings.Contains(p, "/jazz/"):
		return "jazz"
	case strings.Contains(p, "/lofi/"):
		return "lofi"
	case strings.Contains(p, "/rock/"):
		return "lofi"
	}
	return ""
}

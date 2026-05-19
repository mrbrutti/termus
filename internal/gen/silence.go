package gen

// SilenceAlgo is a no-op Algorithm that emits stereo silence. Used by the
// SF2->ACE-Step pre-roll bridge: after ACE-Step's first track is ready,
// Playback hot-swaps the running SF2 Root's algorithm to SilenceAlgo via
// SwapAlgorithmFade. The Root's existing fade-out machinery handles the
// taper; once the fade completes the LiveBackend is closed.
type SilenceAlgo struct{}

// NewSilence returns a fresh SilenceAlgo. (It's stateless, but Algorithm
// constructors throughout termus follow the New* pattern.)
func NewSilence() Algorithm { return &SilenceAlgo{} }

func (SilenceAlgo) Name() string { return "silence" }

func (SilenceAlgo) Seed(int64) {}

func (SilenceAlgo) Next(left, right []float64) {
	for i := range left {
		left[i] = 0
		right[i] = 0
	}
}

package synth

// ModSource is a runtime modulation source (LFO, envelope, MIDI CC follower,
// etc.). Any value implementing this interface can drive a destination through
// a ModMatrix.
type ModSource interface {
	Value() float64
}

// ModDest is a callable that consumes a modulation value (usually scaling or
// offsetting a parameter). Multiple sources may route to the same logical
// destination by sharing state inside their closures.
type ModDest func(value float64)

// ModRoute connects one source to one destination with a fixed scaling amount.
// The source's Value() is multiplied by Amount before being handed to Dest.
type ModRoute struct {
	Source ModSource
	Dest   ModDest
	Amount float64
}

// ModMatrix holds a slice of routes and updates all destinations on each
// Tick(). Destinations are called in insertion order; when multiple sources
// route to the same parameter, callers should accumulate inside the ModDest
// closure.
type ModMatrix struct {
	routes []ModRoute
}

// NewModMatrix returns an empty ModMatrix ready for use.
func NewModMatrix() *ModMatrix {
	return &ModMatrix{}
}

// AddRoute appends a route to the matrix.
func (m *ModMatrix) AddRoute(r ModRoute) {
	m.routes = append(m.routes, r)
}

// Tick evaluates every route: calls route.Dest(route.Source.Value() * route.Amount).
func (m *ModMatrix) Tick() {
	for _, r := range m.routes {
		r.Dest(r.Source.Value() * r.Amount)
	}
}

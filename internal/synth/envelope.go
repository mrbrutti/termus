package synth

type envStage int

const (
	stageIdle envStage = iota
	stageAttack
	stageDecay
	stageSustain
	stageRelease
)

// Envelope is a classic ADSR envelope generator.
type Envelope struct {
	a, d, s, r float64 // seconds, seconds, level, seconds
	value      float64
	stage      envStage
}

func NewEnvelope(attack, decay, sustain, release float64) *Envelope {
	return &Envelope{a: attack, d: decay, s: sustain, r: release}
}

// Gate(true) starts/restarts the attack stage; Gate(false) enters release.
func (e *Envelope) Gate(on bool) {
	if on {
		e.stage = stageAttack
	} else {
		e.stage = stageRelease
	}
}

// Tick advances by one sample and returns the current envelope value.
func (e *Envelope) Tick() float64 {
	const dt = 1.0 / float64(SampleRate)
	switch e.stage {
	case stageAttack:
		if e.a <= 0 {
			e.value = 1
		} else {
			e.value += dt / e.a
		}
		if e.value >= 1 {
			e.value = 1
			e.stage = stageDecay
		}
	case stageDecay:
		if e.d <= 0 {
			e.value = e.s
		} else {
			e.value -= dt * (1 - e.s) / e.d
		}
		if e.value <= e.s {
			e.value = e.s
			e.stage = stageSustain
		}
	case stageSustain:
		e.value = e.s
	case stageRelease:
		if e.r <= 0 {
			e.value = 0
		} else {
			e.value -= dt * e.s / e.r
		}
		if e.value <= 0 {
			e.value = 0
			e.stage = stageIdle
		}
	}
	return e.value
}

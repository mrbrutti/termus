title: Dusty Swing / After Hours
description: Small-group late set with a quiet count-in, brighter middle, and a soft-closing coda.
style: jazz
listen_mode: album-side
seed: 7319
tags: [jazz, swing, small-group, late-set]
key: Cmaj
tempo: 128
globals:
  density: steady
  brightness: balanced
  swing: groove
  phrase: long
roles:
  keys:
    family: acoustic_piano
    tone: [clear, present]
    articulation: comp
    register: mid
    prominence: support
    pattern: "x..x .x.. | .x.. x..."
  bass:
    family: bass
    tone: [woody, round]
    articulation: walk
    register: low
    prominence: anchor
    pattern: "x.x.x.x. x.x.x.x."
  lead:
    family: reed_lead
    tone: [present, live]
    articulation: lyrical
    register: mid-high
    prominence: lead
    motif: "9 . 7 5 | 3 . 2 1"
  drums:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: anchor
    pattern: "x.x. x.x. | .x.. .x.."
sections:
  - id: intro
    title: count-in
    duration: 90s
    harmony: "Dm7 G7 | Cmaj7 A7"
    scene: "intro small-room"
    variation: "establish"
    profile:
      density: light
      brightness: warm
      swing: groove
    roles:
      lead:
        active: false
  - id: head
    title: house theme
    duration: 180s
    harmony: "Dm7 G7 | Cmaj7 A7 | Dm7 G7 | Cmaj7 Cmaj7"
    scene: "head full-combo"
    variation: "head-statement"
    roles:
      lead:
        active: true
        motif: "5 . 6 7 | 9 . 7 3 | 5 . 6 5 | 3 . 2 1"
      keys:
        pattern: "x..x .x.. | x... .x.x"
  - id: bridge
    title: back booth solo
    duration: 210s
    harmony: "Dm7 Db7 | Cmaj7 A7 | Fmaj7 E7 | Dm7 G7"
    scene: "bridge brighter pull"
    variation: "turnaround-lift"
    profile:
      density: busy
      brightness: bright
      swing: heavy
    roles:
      lead:
        motif: "9 . b9 7 | 5 . 3 1 | 9 . 7 5 | 3 . 2 1"
      drums:
        pattern: "x.x. x.x. | .x.. .x.. | x.x. xx.."
  - id: outro
    title: last call
    duration: 120s
    harmony: "Dm7 G7 | Cmaj7 A7 | Dm7 G7 | Cmaj7 Cmaj7"
    scene: "outro soft-landing"
    variation: "cadence"
    profile:
      density: light
      brightness: warm
      swing: groove
    roles:
      lead:
        motif: "3 . 2 1 | 1 . . ."
      drums:
        pattern: "x.x. x.x. | .x.. .x.."


title: Dusty Swing / Red Eye Turnaround
description: Leaner head, sharper bridge reharm, and a quieter runway close.
style: jazz
listen_mode: album-side
seed: 18421
tags: [jazz, turnaround, midnight, quartet]
key: Fmaj
tempo: 124
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
    register: high
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
    title: terminal lights
    duration: 75s
    harmony: "Gm7 C7 | Fmaj7 D7"
    scene: "intro restrained"
    variation: "establish"
    profile:
      density: light
      brightness: warm
    roles:
      lead:
        active: false
  - id: head
    title: aisle melody
    duration: 150s
    harmony: "Gm7 C7 | Fmaj7 D7 | Gm7 C7 | Fmaj7 Fmaj7"
    scene: "head statement"
    variation: "head-statement"
    roles:
      lead:
        active: true
        motif: "5 . 6 7 | 9 . 7 3 | 5 . 6 5 | 3 . 2 1"
  - id: bridge
    title: gate change
    duration: 195s
    harmony: "Gm7 Gb7 | Fmaj7 D7 | Bbmaj7 A7 | Gm7 C7"
    scene: "bridge sharper-turn"
    variation: "reharm"
    profile:
      density: busy
      brightness: bright
      swing: heavy
    roles:
      lead:
        motif: "9 . b9 7 | 5 . 3 1 | 11 . 9 7 | 5 . 2 1"
      keys:
        pattern: "x... .x.x | x..x .x.."
  - id: outro
    title: wheels down
    duration: 105s
    harmony: "Gm7 C7 | Fmaj7 D7 | Gm7 C7 | Fmaj7 Fmaj7"
    scene: "outro quiet-runway"
    variation: "cadence"
    roles:
      lead:
        motif: "3 . 2 1 | 1 . . ."


title: Dusty Swing / Basement Blue Hour
description: Vibes-and-tenor basement set with a wider middle chorus and a slow-closing downstairs fade.
style: jazz
listen_mode: album-side
seed: 31777
tags: [jazz, vibes, tenor, basement, nocturne]
key: Bbmaj
tempo: 118
globals:
  density: steady
  brightness: warm
  swing: groove
  phrase: long
roles:
  piano:
    family: acoustic_piano
    tone: [clear, present]
    articulation: comp
    register: mid
    prominence: support
    pattern: ".x...... | ..x...x."
  bass:
    family: bass
    tone: [woody, round]
    articulation: walk
    register: low
    prominence: anchor
    pattern: "x...x.x. | x...x..x"
  ride:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: support
    pattern: "x..x.x.. | x..x.xx."
  kick:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: anchor
    pattern: "x....... | ....x..."
  snare:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: anchor
    pattern: "........ | ..x.x..."
  rim:
    family: drums
    tone: [live, dry]
    articulation: swing
    prominence: support
    pattern: "...x.... | ....x..."
  tenor:
    family: reed_lead
    tone: [present, round]
    articulation: lyrical
    register: mid-high
    prominence: lead
    motif: "5 . 6 7 9 . 7 5 | 3 . 2 . 1 . . ."
  vibes:
    family: mallet
    tone: [clear, soft]
    articulation: comp
    register: mid-high
    prominence: support
    pattern: ".x..x... | ..x.x..."
sections:
  - id: intro
    title: blue stairwell
    duration: 14s
    harmony: "Cm7 F7 | Bbmaj7 G7 | Cm7 F7 | Bbmaj7 Bbmaj7"
    scene: "intro downstairs"
    variation: "establish"
    roles:
      tenor:
        active: true
        motif: "5 . . . 7 . . 5 | . . 3 . 2 . 1 ."
      piano:
        active: false
      rim:
        pattern: "........ | ....x..."
  - id: head
    title: first set
    duration: 50s
    harmony: "Cm7 F7 | Bbmaj7 G7 | Cm7 F7 | Bbmaj7 Bbmaj7 | Dm7 G7 | Cm7 F7 | Dm7 G7 | Cm7 F7"
    scene: "head relaxed"
    variation: "statement"
    roles:
      tenor:
        active: true
        motif: "9 . 7 5 6 . 5 3 | 5 . 2 . 1 . . ."
      piano:
        active: false
      vibes:
        pattern: "x..x.x.. | .x..x..x"
  - id: chorus
    title: room opens
    duration: 55s
    harmony: "Dm7 G7 | Cm7 F7 | Bbmaj7 G7 | Cm7 F7 | Ebmaj7 D7 | Cm7 F7 | Dm7 G7 | Cm7 F7"
    scene: "chorus wider"
    variation: "lift"
    profile:
      density: busy
      brightness: balanced
      swing: heavy
    roles:
      tenor:
        motif: "11 . 9 7 5 . 3 1 | 9 . 7 5 6 . 5 3"
      vibes:
        pattern: "x.x.x..x | .x..x.x."
      piano:
        active: true
        pattern: ".x..x... | ..x.x..."
      snare:
        pattern: "....x... | ..x.xx.."
  - id: outro
    title: bar stools empty
    duration: 35s
    harmony: "Cm7 F7 | Bbmaj7 G7 | Cm7 F7 | Bbmaj7 Bbmaj7"
    scene: "outro downstairs"
    variation: "cadence"
    roles:
      tenor:
        motif: "3 . 2 . 1 . . . | . . . . . . . ."

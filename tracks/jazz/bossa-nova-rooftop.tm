title: Bossa Nova Rooftop
description: Bossa-tinted jazz trio with loose groove, secondary dominants, and guitar comp.
style: jazz
mix_bus: jazz
listen_mode: album-side
seed: 44190
tags: [jazz, bossa, guitar, rooftop, trio]
key: Gmaj
tempo: 104
globals: {density: steady, brightness: balanced, motion: gentle, phrase: long}
roles:
  guitar:
    family: guitar
    tone: [warm, soft]
    articulation: comp
    register: mid
    prominence: support
    pattern: ".x..x... | ..x.x..."
  bass:
    family: bass
    tone: [woody, round]
    articulation: walk
    register: low
    prominence: anchor
    pattern: "x...x... | x...x..."
  kick:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: anchor
    pattern: "x....... | x......."
  snare:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: support
    pattern: "....x... | ....x..."
  hat:
    family: drums
    tone: [live, dry]
    articulation: swing
    prominence: support
    pattern: "x.x.x.x. | x.x.x.x."
  flute:
    family: flute
    tone: [airy, soft]
    register: high
    prominence: lead
    motif: "5 . 7 9 | 3 . 2 1"
sections:
  - id: intro
    title: warm-up breeze
    duration: 12s
    harmony: "Gmaj7 Em7"
    scene: "intro hush"
    variation: "establish"
    groove: bossa_loose
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.4}
          - {at: 100, value: 0.65}
  - id: verse
    title: open terrace
    duration: 40s
    harmony: "Gmaj7 Em7 | Am7 D7"
    scene: "head glide"
    variation: "statement"
    groove: bossa_loose
    substitutions:
      - {rule: secondary_dominant, of: ii, probability: 0.7}
  - id: bridge
    title: moon-rise shift
    duration: 22s
    harmony: "Cmaj7 B7 | Em7 A7 | Am7 D7 | Gmaj7 E7"
    scene: "bridge lift"
    variation: "open-register"
    groove: bossa_loose
    substitutions:
      - {rule: secondary_dominant, of: ii, probability: 1.0}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.65}
          - {at: 60, value: 0.9}
          - {at: 100, value: 0.7}
  - id: outro
    title: last balcony
    duration: 18s
    harmony: "Am7 D7 | Gmaj7 Gmaj7"
    scene: "outro cadence"
    variation: "cadence"
    groove: bossa_loose
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.7}
          - {at: 100, value: 0.3}

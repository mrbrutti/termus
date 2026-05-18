title: Dusty Swing / After Hours
description: Bop AABA head with walking bass and tenor lead, ii-V chains on every cadence.
style: jazz
mix_bus: jazz
listen_mode: album-side
seed: 31440
tags: [jazz, swing, bop, tenor, trio]
key: Bbmaj
tempo: 122
globals: {density: steady, brightness: warm, swing: groove, phrase: long}
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
  tenor:
    family: reed_lead
    tone: [present, round]
    articulation: lyrical
    register: mid-high
    prominence: lead
    motif: "5 . 6 7 9 . 7 5 | 3 . 2 . 1 . . ."
sections:
  - id: head-a1
    title: first chorus
    duration: 18s
    harmony: "Cm7 F7 | Bbmaj7 G7 | Cm7 F7 | Bbmaj7 Bbmaj7"
    scene: "head relaxed"
    variation: "statement"
    groove: swing_56
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 0.8}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.6}
          - {at: 100, value: 0.8}
  - id: head-a2
    title: second a
    duration: 18s
    harmony: "Cm7 F7 | Bbmaj7 G7 | Dm7 G7 | Cm7 F7"
    scene: "head glide"
    variation: "sequence-up"
    groove: swing_56
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 1.0}
  - id: head-b
    title: bridge climb
    duration: 16s
    harmony: "Ebmaj7 D7 | Dm7 G7 | Cm7 F7 | Bbmaj7 G7"
    scene: "bridge lift"
    variation: "open-register"
    groove: swing_56
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 0.5}
  - id: outro
    title: bar stools empty
    duration: 16s
    harmony: "Cm7 F7 | Bbmaj7 G7 | Cm7 F7 | Bbmaj7 Bbmaj7"
    scene: "outro cadence"
    variation: "cadence"
    groove: swing_56
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.8}
          - {at: 100, value: 0.3}

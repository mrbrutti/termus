title: Dusty Swing / After Hours
description: Bop AABA head with walking bass every beat, spang-a-lang ride, jazz comping, ghost kicks, and tenor lead.
style: jazz
mix_bus: jazz
listen_mode: album-side
seed: 31440
tags: [jazz, swing, bop, tenor, trio, walking, ride]
key: Bbmaj
tempo: 138
globals: {density: full, brightness: bright, motion: restless, phrase: long}
roles:
  piano:
    family: acoustic_piano
    tone: [clear, present]
    articulation: comp
    register: mid
    prominence: support
  bass:
    family: bass
    tone: [woody, round]
    articulation: walk
    register: low
    prominence: anchor
  ride:
    family: drums
    tone: [live, bright]
    articulation: swing
    prominence: support
  kick:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: anchor
  snare:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: support
  tenor:
    family: reed_lead
    tone: [present, round]
    articulation: lyrical
    register: mid-high
    prominence: lead
    motif: "5 . 6 7 9 . 7 5 | 3 . 2 . 1 . . ."
  vibes:
    family: mallet
    tone: [glass, warm]
    register: mid-high
    prominence: air
sections:
  - id: head-a1
    title: first chorus
    duration: 16s
    harmony: "Cm7 F7 | Bbmaj7 G7 | Cm7 F7 | Bbmaj7 Bbmaj7"
    scene: "head relaxed"
    variation: "statement"
    groove: swing_56
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 0.9}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.6}
          - {at: 100, value: 0.85}
  - id: head-a2
    title: second a
    duration: 16s
    harmony: "Cm7 F7 | Bbmaj7 G7 | Dm7 G7 | Cm7 F7"
    scene: "head glide"
    variation: "sequence-up"
    groove: swing_56
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 1.0}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.8}
          - {at: 100, value: 0.9}
  - id: head-b
    title: bridge climb
    duration: 16s
    harmony: "Ebmaj7 D7 | Dm7 G7 | Cm7 F7 | Bbmaj7 G7"
    scene: "bridge lift"
    variation: "open-register"
    groove: swing_56
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 0.7}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.85}
          - {at: 60, value: 1.0}
          - {at: 100, value: 0.9}
  - id: head-a3
    title: final a
    duration: 16s
    harmony: "Cm7 F7 | Bbmaj7 G7 | Cm7 F7 | Bbmaj7 Bbmaj7"
    scene: "head glide"
    variation: "statement"
    groove: swing_56
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 1.0}
  - id: outro
    title: bar stools empty
    duration: 16s
    harmony: "Cm7 F7 | Bbmaj7 G7 | Cm7 F7 | Bbmaj7 Bbmaj7"
    scene: "outro cadence"
    variation: "cadence"
    groove: swing_56
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 0.7}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.85}
          - {at: 100, value: 0.3}

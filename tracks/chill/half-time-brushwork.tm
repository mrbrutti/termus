title: Half-Time Brushwork
description: Half-time brushed chill with dense electric piano comping, syncopated bass, pad sustain, guitar texture, and sax lead.
style: chill
mix_bus: chill
listen_mode: hour-stream
seed: 27845
tags: [chill, halftime, brushes, swing, moody, pad, sax]
key: Emin
tempo: 75
globals: {density: full, brightness: warm, motion: moving, reverb: halo}
roles:
  keys:
    family: electric_piano
    tone: [warm, dusty]
    register: mid
    prominence: support
  pad:
    family: pad
    tone: [soft, wide]
    register: mid
    prominence: air
  bass:
    family: bass
    tone: [round, woody]
    register: low
    prominence: anchor
  kick:
    family: drums
    tone: [soft, dusty]
    prominence: anchor
  snare:
    family: drums
    tone: [soft]
    prominence: support
  hat:
    family: drums
    tone: [dry, tight]
    prominence: support
  ride:
    family: drums
    tone: [live, soft]
    prominence: support
  guitar:
    family: guitar
    tone: [warm, soft]
    register: mid
    prominence: support
  lead:
    family: reed_lead
    tone: [breathy, intimate]
    register: mid-high
    prominence: lead
    motif: "5 . 7 . | 9 . 7 5"
sections:
  - id: intro
    title: brushed arrival
    duration: 14s
    harmony: "Em9 Cmaj9 | Am7 D7"
    scene: "intro hush"
    variation: "establish"
    groove: swing_56
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.35}
          - {at: 100, value: 0.65}
  - id: verse
    title: slow-motion glide
    duration: 32s
    harmony: "Em9 Cmaj9 | Am7 Dsus4"
    scene: "head glide"
    variation: "statement"
    groove: swing_56
    substitutions:
      - {rule: secondary_dominant, of: ii, probability: 0.7}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.65}
          - {at: 50, value: 0.7}
          - {at: 100, value: 0.65}
  - id: bridge
    title: middle distance
    duration: 22s
    harmony: "Fmaj9 E7 | Am9 Dsus4"
    scene: "bridge tilt"
    variation: "open-register"
    groove: swing_56
    substitutions:
      - {rule: secondary_dominant, of: ii, probability: 1.0}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.7}
          - {at: 60, value: 0.9}
          - {at: 100, value: 0.8}
  - id: verse2
    title: amber return
    duration: 24s
    harmony: "Em9 Cmaj9 | Am7 D7"
    scene: "head glide"
    variation: "sequence-up"
    groove: swing_56
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 0.8}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.8}
          - {at: 100, value: 0.75}
  - id: outro
    title: last light fades
    duration: 18s
    harmony: "Em9 Cmaj9 | Em6"
    scene: "outro hush"
    variation: "cadence"
    groove: swing_56
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 1.0}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.75}
          - {at: 100, value: 0.2}

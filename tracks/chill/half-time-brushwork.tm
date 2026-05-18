title: Half-Time Brushwork
description: Half-time brushed groove with swing pocket and slow pan drift.
style: chill
mix_bus: chill
listen_mode: album-side
seed: 27845
tags: [chill, halftime, brushes, swing, moody]
key: Emin
tempo: 72
globals: {density: steady, brightness: warm, motion: gentle, reverb: room}
roles:
  keys:
    family: electric_piano
    tone: [warm, dusty]
    register: mid
    prominence: support
    pattern: "x..x .... x..x ...."
  bass:
    family: bass
    tone: [round, woody]
    register: low
    prominence: anchor
    pattern: "x... .... x..x ...."
  kick:
    family: drums
    tone: [soft, dusty]
    prominence: anchor
    pattern: "x....... | ....x..."
  snare:
    family: drums
    tone: [soft]
    prominence: support
    pattern: "........ | ....x..."
  hat:
    family: drums
    tone: [dry]
    prominence: support
    pattern: "x.x.x.x. | x.x.x.x."
  guitar:
    family: guitar
    tone: [warm, soft]
    register: mid
    prominence: air
    pattern: "..x. .... ..x. ...."
  lead:
    family: reed_lead
    tone: [breathy]
    register: mid-high
    prominence: lead
    motif: "5 . 7 . | 9 . 7 5"
sections:
  - id: intro
    title: brushed arrival
    duration: 16s
    harmony: "Em9 Cmaj9"
    scene: "intro hush"
    variation: "establish"
    groove: swing_56
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.35}
          - {at: 100, value: 0.6}
  - id: verse
    title: slow-motion glide
    duration: 38s
    harmony: "Em9 Cmaj9 | Am7 Dsus4"
    scene: "head glide"
    variation: "statement"
    groove: swing_56
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.6}
          - {at: 50, value: 0.55}
          - {at: 100, value: 0.55}
  - id: bridge
    title: middle distance
    duration: 24s
    harmony: "Fmaj9 E7 | Am9 Dsus4"
    scene: "bridge tilt"
    variation: "open-register"
    groove: swing_56
    substitutions:
      - {rule: secondary_dominant, of: ii, probability: 0.7}
  - id: outro
    title: last light fades
    duration: 18s
    harmony: "Em9 Cmaj9 | Em6"
    scene: "outro hush"
    variation: "cadence"
    groove: swing_56
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.55}
          - {at: 100, value: 0.2}

title: Sunday Afternoon Drive
description: 4/4 ambient-pop with pad chords and slow cutoff opening on a straight groove.
style: chill
mix_bus: chill
listen_mode: album-side
seed: 19334
tags: [chill, pop, pad, afternoon, drive]
key: Amin
tempo: 88
globals: {density: steady, brightness: balanced, motion: gentle, reverb: room}
roles:
  keys:
    family: electric_piano
    tone: [warm, soft]
    register: mid
    prominence: support
    pattern: "x..x ..x. x..x ...."
  pad:
    family: pad
    tone: [soft, wide]
    register: mid
    prominence: air
    pattern: "x....... | ........"
  bass:
    family: synth_bass
    tone: [round]
    register: low
    prominence: anchor
    pattern: "x... x... x... x..."
  kick:
    family: drums
    tone: [soft]
    prominence: anchor
    pattern: "x... .... x... ...."
  snare:
    family: drums
    tone: [soft]
    prominence: support
    pattern: ".... x... .... x..."
  hat:
    family: drums
    tone: [dry]
    prominence: support
    pattern: "x.x. x.x. x.x. x.x."
  lead:
    family: guitar
    tone: [warm, soft]
    register: mid-high
    prominence: lead
    motif: "5 . 7 9 | 3 . 2 1"
sections:
  - id: intro
    title: open road
    duration: 14s
    harmony: "Am9 Fmaj9"
    scene: "intro hush"
    variation: "establish"
    groove: straight
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.2}
          - {at: 100, value: 0.55}
  - id: verse
    title: window down
    duration: 40s
    harmony: "Am9 Fmaj9 | Cmaj9 Gsus4"
    scene: "head glide"
    variation: "statement"
    groove: straight
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.55}
          - {at: 50, value: 0.75}
          - {at: 100, value: 0.65}
  - id: bridge
    title: golden-hour stretch
    duration: 22s
    harmony: "Fmaj9 Em7 | Am9 Gsus4"
    scene: "bridge lift"
    variation: "open-register"
    groove: straight
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.65}
          - {at: 100, value: 0.85}
  - id: outro
    title: home at dusk
    duration: 16s
    harmony: "Am9 Fmaj9 | Am6"
    scene: "outro hush"
    variation: "cadence"
    groove: straight
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.7}
          - {at: 100, value: 0.3}

title: Sunday Afternoon Drive
description: Half-time chill with sustained pad, syncopated bass, 16th-note hats, Rhodes chops, guitar lead, and choir halo.
style: chill
mix_bus: chill
listen_mode: album-side
seed: 19334
tags: [chill, pop, pad, afternoon, drive, halftime, choir]
key: Amin
tempo: 88
globals: {density: full, brightness: balanced, motion: moving, reverb: halo}
roles:
  keys:
    family: electric_piano
    tone: [warm, soft]
    register: mid
    prominence: support
    pattern: "x.x..x.. x.x..x.."
  pad:
    family: pad
    tone: [soft, wide]
    register: mid
    prominence: air
    pattern: "x..............."
  choir:
    family: choir
    tone: [soft, airy]
    register: high
    prominence: air
    pattern: "x..............."
  bass:
    family: synth_bass
    tone: [round, soft]
    register: low
    prominence: anchor
    pattern: "x..x x..x x.x. x..."
  kick:
    family: drums
    tone: [soft, deep]
    prominence: anchor
    pattern: "x.......x......."
  snare:
    family: drums
    tone: [soft]
    prominence: support
    pattern: "....x.......x..."
  hat:
    family: drums
    tone: [dry, tight]
    prominence: support
    pattern: "x.x.x.x.x.x.x.x."
  ride:
    family: drums
    tone: [live, soft]
    prominence: support
    pattern: "....x.......x..."
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
    harmony: "Am9 Fmaj9 | Cmaj9 G7"
    scene: "intro hush"
    variation: "establish"
    groove: straight
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.2}
          - {at: 100, value: 0.6}
  - id: verse
    title: window down
    duration: 34s
    harmony: "Am9 Fmaj9 | Cmaj9 Gsus4"
    scene: "head glide"
    variation: "statement"
    groove: straight
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 0.7}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.6}
          - {at: 50, value: 0.78}
          - {at: 100, value: 0.68}
  - id: bridge
    title: golden-hour stretch
    duration: 24s
    harmony: "Fmaj9 Em7 | Am9 Gsus4"
    scene: "bridge lift"
    variation: "open-register"
    groove: straight
    substitutions:
      - {rule: secondary_dominant, of: ii, probability: 0.8}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.7}
          - {at: 60, value: 0.9}
          - {at: 100, value: 0.8}
  - id: verse2
    title: highway merge
    duration: 26s
    harmony: "Am9 Fmaj9 | Cmaj9 G7"
    scene: "head glide"
    variation: "sequence-up"
    groove: straight
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 0.9}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.8}
          - {at: 100, value: 0.75}
  - id: outro
    title: home at dusk
    duration: 16s
    harmony: "Am9 Fmaj9 | Am6"
    scene: "outro hush"
    variation: "cadence"
    groove: straight
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 1.0}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.75}
          - {at: 100, value: 0.25}

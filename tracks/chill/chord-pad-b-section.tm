title: Chord Pad B Section
description: Ambient-pop chord pad cycle with ii-V chain inserts in B sections.
style: chill
mix_bus: chill
listen_mode: album-side
seed: 33710
tags: [chill, pad, chords, ambient-pop]
key: Abmaj
tempo: 80
globals: {density: steady, brightness: balanced, motion: gentle, reverb: warm}
roles:
  pad:
    family: pad
    tone: [soft, wide]
    register: mid
    prominence: support
    pattern: "x....... | ........"
  keys:
    family: electric_piano
    tone: [warm]
    register: mid
    prominence: air
    pattern: "..x. .... ..x. ...."
  bass:
    family: synth_bass
    tone: [round, soft]
    register: low
    prominence: anchor
    pattern: "x... .... x... ...."
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
    pattern: "x.x. .... x.x. ...."
  lead:
    family: reed_lead
    tone: [airy, soft]
    register: mid-high
    prominence: lead
    motif: "5 . 7 9 | 3 . 2 1"
sections:
  - id: intro
    title: morning pad
    duration: 14s
    harmony: "Abmaj9 Fm9"
    scene: "intro hush"
    variation: "establish"
    groove: straight
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.25}
          - {at: 100, value: 0.5}
  - id: verse-a
    title: floating chord
    duration: 36s
    harmony: "Abmaj9 Fm9 | Bbm7 Eb7"
    scene: "head glide"
    variation: "statement"
    groove: straight
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.5}
          - {at: 100, value: 0.65}
  - id: verse-b
    title: ii-V sequence
    duration: 26s
    harmony: "Bbm7 Eb7 | Abmaj9 Fm9"
    scene: "answer lift"
    variation: "sequence-up"
    groove: straight
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 1.0}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.65}
          - {at: 100, value: 0.8}
  - id: outro
    title: cloud return
    duration: 16s
    harmony: "Abmaj9 Fm9 | Abmaj7"
    scene: "outro hush"
    variation: "cadence"
    groove: straight
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.7}
          - {at: 100, value: 0.3}

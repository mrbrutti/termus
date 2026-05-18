title: Chord Pad B Section
description: Ambient-pop chord pad cycle with layered choir, syncopated bass, hat groove, vibes sparkle, and a reed lead.
style: chill
mix_bus: chill
listen_mode: hour-stream
seed: 33710
tags: [chill, pad, chords, ambient-pop, choir, vibes]
key: Abmaj
tempo: 96
globals: {density: full, brightness: balanced, motion: moving, reverb: halo}
roles:
  pad:
    family: pad
    tone: [soft, wide]
    register: mid
    prominence: air
  choir:
    family: choir
    tone: [soft, airy]
    register: high
    prominence: air
  keys:
    family: electric_piano
    tone: [warm, soft]
    register: mid
    prominence: support
  vibes:
    family: mallet
    tone: [glass, warm]
    register: mid-high
    prominence: support
  bass:
    family: synth_bass
    tone: [round, soft]
    register: low
    prominence: anchor
  kick:
    family: drums
    tone: [soft, deep]
    prominence: anchor
  snare:
    family: drums
    tone: [soft]
    prominence: support
  hat:
    family: drums
    tone: [dry, tight]
    prominence: support
  lead:
    family: reed_lead
    tone: [airy, soft]
    register: mid-high
    prominence: lead
    motif: "5 . 7 9 | 3 . 2 1"
sections:
  - id: intro
    title: morning pad
    duration: 12s
    harmony: "Abmaj9 Fm9 | Bbm7 Eb7"
    scene: "intro hush"
    variation: "establish"
    groove: straight
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.25}
          - {at: 100, value: 0.55}
  - id: verse-a
    title: floating chord
    duration: 30s
    harmony: "Abmaj9 Fm9 | Bbm7 Eb7"
    scene: "head glide"
    variation: "statement"
    groove: straight
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 0.8}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.55}
          - {at: 100, value: 0.7}
  - id: verse-b
    title: ii-V sequence
    duration: 24s
    harmony: "Bbm7 Eb7 | Abmaj9 Fm9"
    scene: "answer lift"
    variation: "sequence-up"
    groove: straight
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 1.0}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.7}
          - {at: 60, value: 0.9}
          - {at: 100, value: 0.8}
  - id: bridge
    title: upper voice shift
    duration: 24s
    harmony: "Dbmaj9 C7 | Fm9 Bbm7"
    scene: "bridge tilt"
    variation: "open-register"
    groove: straight
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 0.8}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.8}
          - {at: 100, value: 0.85}
  - id: outro
    title: cloud return
    duration: 18s
    harmony: "Abmaj9 Fm9 | Abmaj7"
    scene: "outro hush"
    variation: "cadence"
    groove: straight
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 1.0}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.8}
          - {at: 100, value: 0.25}

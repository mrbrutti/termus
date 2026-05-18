title: Bell-Struck Skylight
description: Layered bell motifs over dual sustained pads, strings halo, and slow choir swell — no drums.
style: ambient
mix_bus: ambient
listen_mode: album-side
seed: 24567
tags: [ambient, bells, sparse, high, crystalline, pad, strings]
key: Cmaj
tempo: 48
globals: {density: busy, brightness: balanced, motion: slow, reverb: cathedral}
roles:
  pad:
    family: pad
    tone: [soft, wide, dreamy]
    register: mid
    prominence: anchor
    pattern: "x..............."
  strings:
    family: strings
    tone: [soft, warm]
    register: mid-high
    prominence: support
    pattern: "x..............."
  choir:
    family: choir
    tone: [soft, airy]
    register: high
    prominence: air
    pattern: "x..............."
  texture:
    family: bells
    tone: [glass, sparkle]
    register: high
    prominence: air
    pattern: "..x.....x..x...."
  bells:
    family: bells
    tone: [glass, soft, luminous]
    register: high
    prominence: lead
    personality: bell_struck
    room: cathedral_large
    reverb_send_db: -8
    motif: "5 . . . 9 . . . | 3 . . . . . . ."
sections:
  - id: intro
    title: first toll
    duration: 20s
    harmony: "Cmaj9 Am9"
    scene: "intro still"
    variation: "establish"
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.3}
          - {at: 100, value: 0.6}
  - id: middle
    title: bell cascade
    duration: 32s
    harmony: "Cmaj9 Am9 | Fmaj9 Gsus4"
    scene: "head drift"
    variation: "statement"
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.6}
          - {at: 50, value: 0.85}
          - {at: 100, value: 0.7}
  - id: sparse
    title: echoes only
    duration: 26s
    harmony: "Am9 Fmaj9 | Cmaj9 Cmaj9"
    scene: "bridge still"
    variation: "open-register"
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 0.5}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.65}
          - {at: 100, value: 0.88}
  - id: swell
    title: choir rising
    duration: 22s
    harmony: "Fmaj9 Cmaj9 | Am9 Gsus4"
    scene: "bridge lift"
    variation: "sequence-up"
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.7}
          - {at: 60, value: 1.0}
          - {at: 100, value: 0.8}
  - id: fade
    title: last ring
    duration: 18s
    harmony: "Cmaj9 Cmaj9"
    scene: "outro still"
    variation: "cadence"
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.7}
          - {at: 100, value: 0.1}

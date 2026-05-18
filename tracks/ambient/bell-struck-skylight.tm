title: Bell-Struck Skylight
description: Sparse bell-like motifs with high crystalline attacks over a sustained pad.
style: ambient
mix_bus: ambient
listen_mode: album-side
seed: 24567
tags: [ambient, bells, sparse, high, crystalline]
key: Cmaj
tempo: 48
globals: {density: sparse, brightness: balanced, motion: still, reverb: cathedral}
roles:
  bells:
    family: bells
    tone: [glass, soft]
    register: high
    prominence: lead
    personality: bell_struck
    room: cathedral_large
    reverb_send_db: -8
    motif: "5 . . . 9 . . . | 3 . . . . . . ."
  pad:
    family: pad
    tone: [soft, wide]
    register: mid
    prominence: anchor
    pattern: "x....... | ........"
  harp:
    family: harp
    tone: [soft, warm]
    register: mid-high
    prominence: support
    pattern: "..x. .... ..x. ...."
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
          - {at: 100, value: 0.55}
  - id: middle
    title: bell cascade
    duration: 36s
    harmony: "Cmaj9 Am9 | Fmaj9 Gsus4"
    scene: "head drift"
    variation: "statement"
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.55}
          - {at: 50, value: 0.8}
          - {at: 100, value: 0.65}
  - id: sparse
    title: echoes only
    duration: 24s
    harmony: "Am9 Fmaj9 | Cmaj9 Cmaj9"
    scene: "bridge still"
    variation: "open-register"
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 0.5}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.65}
          - {at: 100, value: 0.85}
  - id: fade
    title: last ring
    duration: 18s
    harmony: "Cmaj9 Cmaj9"
    scene: "outro still"
    variation: "cadence"
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.65}
          - {at: 100, value: 0.1}

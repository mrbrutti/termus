title: Deep Pad Breath
description: Deep pad wash with sustained strings, choir halo, bell texture, and a slow brass lead — no drums.
style: ambient
mix_bus: ambient
listen_mode: album-side
seed: 38901
tags: [ambient, pad, brass, breath, deep, strings, choir]
key: Dmin
tempo: 60
globals: {density: busy, brightness: warm, motion: slow, reverb: cathedral}
roles:
  pad:
    family: pad
    tone: [soft, wide, deep]
    register: low
    prominence: anchor
  strings:
    family: strings
    tone: [soft, warm]
    register: mid
    prominence: support
  choir:
    family: choir
    tone: [soft, airy]
    register: high
    prominence: air
  texture:
    family: bells
    tone: [glass, sparkle, soft]
    register: high
    prominence: air
  lead:
    family: brass
    tone: [soft, airy]
    register: mid-high
    prominence: lead
    personality: brass_breath
    room: cathedral_large
    reverb_send_db: -6
    motif: "5 . . . . . 7 . | 9 . . . . . . ."
sections:
  - id: breath-in
    title: first breath
    duration: 20s
    harmony: "Dm9 Dm9"
    scene: "intro still"
    variation: "establish"
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.15}
          - {at: 100, value: 0.55}
  - id: swell
    title: ascending tide
    duration: 30s
    harmony: "Dm9 Bbmaj9 | Fmaj9 C6"
    scene: "head drift"
    variation: "statement"
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 0.5}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.55}
          - {at: 60, value: 0.92}
          - {at: 100, value: 0.78}
  - id: hold
    title: sustained presence
    duration: 28s
    harmony: "Gm9 Fmaj9 | Dm9 Bbmaj9"
    scene: "bridge still"
    variation: "open-register"
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.55}
          - {at: 100, value: 0.78}
  - id: peak
    title: full resonance
    duration: 22s
    harmony: "Bbmaj9 Am7 | Gm9 Fmaj9"
    scene: "bridge lift"
    variation: "sequence-up"
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.8}
          - {at: 60, value: 1.0}
          - {at: 100, value: 0.88}
  - id: release
    title: breath out
    duration: 20s
    harmony: "Dm9 Dm9"
    scene: "outro still"
    variation: "cadence"
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.78}
          - {at: 100, value: 0.1}

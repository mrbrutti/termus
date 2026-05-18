title: Deep Pad Breath
description: Deep pad wash with breath-noise brass lead and long slow swells.
style: ambient
mix_bus: ambient
listen_mode: album-side
seed: 38901
tags: [ambient, pad, brass, breath, deep]
key: Dmin
tempo: 44
globals: {density: sparse, brightness: warm, motion: still, reverb: cathedral}
roles:
  pad:
    family: pad
    tone: [soft, wide, deep]
    register: low
    prominence: anchor
    pattern: "x....... | ........"
  lead:
    family: brass
    tone: [soft, airy]
    register: mid-high
    prominence: lead
    personality: brass_breath
    room: cathedral_large
    reverb_send_db: -6
    motif: "5 . . . . . 7 . | 9 . . . . . . ."
  strings:
    family: strings
    tone: [soft, warm]
    register: mid
    prominence: support
    pattern: "x....... | ........"
  choir:
    family: choir
    tone: [soft, airy]
    register: high
    prominence: air
    pattern: "x....... | ........"
sections:
  - id: breath-in
    title: first breath
    duration: 22s
    harmony: "Dm9 Dm9"
    scene: "intro still"
    variation: "establish"
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.15}
          - {at: 100, value: 0.5}
  - id: swell
    title: ascending tide
    duration: 36s
    harmony: "Dm9 Bbmaj9 | Fmaj9 C6"
    scene: "head drift"
    variation: "statement"
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.5}
          - {at: 60, value: 0.9}
          - {at: 100, value: 0.75}
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 0.5}
  - id: hold
    title: sustained presence
    duration: 28s
    harmony: "Gm9 Fmaj9 | Dm9 Bbmaj9"
    scene: "bridge still"
    variation: "open-register"
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.5}
          - {at: 100, value: 0.75}
  - id: release
    title: breath out
    duration: 20s
    harmony: "Dm9 Dm9"
    scene: "outro still"
    variation: "cadence"
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.75}
          - {at: 100, value: 0.1}

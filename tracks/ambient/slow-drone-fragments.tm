title: Slow Drone Fragments
description: Slow modal drone with motif fragments drifting over a deep pad bed. No rhythm section.
style: ambient
mix_bus: ambient
listen_mode: album-side
seed: 11023
tags: [ambient, drone, modal, slow, pad]
key: Amin
tempo: 52
globals: {density: sparse, brightness: warm, motion: still, reverb: halo}
roles:
  pad:
    family: pad
    tone: [soft, wide]
    register: mid
    prominence: anchor
    pattern: "x....... | ........"
  strings:
    family: strings
    tone: [soft, warm]
    register: mid-high
    prominence: support
    pattern: "x....... | ........"
  lead:
    family: flute
    tone: [airy, soft]
    register: high
    prominence: lead
    motif: "5 . . . 7 . . . | 9 . . . . . . ."
sections:
  - id: open
    title: still water
    duration: 24s
    harmony: "Am9 Am9"
    scene: "intro still"
    variation: "establish"
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.15}
          - {at: 100, value: 0.4}
  - id: drift
    title: motif rising
    duration: 36s
    harmony: "Am9 Fmaj9 | Am9 G6"
    scene: "head drift"
    variation: "statement"
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.4}
          - {at: 50, value: 0.6}
          - {at: 100, value: 0.5}
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 0.6}
  - id: deep
    title: lower tide
    duration: 30s
    harmony: "Am9 Fmaj9 | Em9 Am9"
    scene: "bridge still"
    variation: "open-register"
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.6}
          - {at: 100, value: 0.8}
  - id: close
    title: fog return
    duration: 20s
    harmony: "Am9 Am9"
    scene: "outro still"
    variation: "cadence"
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.5}
          - {at: 100, value: 0.1}

title: Slow Drone Fragments
description: Layered drone with low pad, mid strings, high choir, and flute motif fragments drifting over sustained harmony.
style: ambient
mix_bus: ambient
listen_mode: album-side
seed: 11023
tags: [ambient, drone, modal, slow, pad, strings, choir]
key: Amin
tempo: 52
globals: {density: busy, brightness: warm, motion: slow, reverb: cathedral}
roles:
  pad:
    family: pad
    tone: [soft, wide, deep]
    register: low
    prominence: anchor
    pattern: "x..............."
  strings:
    family: strings
    tone: [soft, warm]
    register: mid
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
  lead:
    family: woodwind
    tone: [airy, soft]
    register: high
    prominence: lead
    motif: "5 . . . 7 . . . | 9 . . . . . . ."
sections:
  - id: open
    title: still water
    duration: 22s
    harmony: "Am9 Am9"
    scene: "intro still"
    variation: "establish"
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.15}
          - {at: 100, value: 0.45}
  - id: drift
    title: motif rising
    duration: 32s
    harmony: "Am9 Fmaj9 | Am9 G6"
    scene: "head drift"
    variation: "statement"
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 0.6}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.45}
          - {at: 50, value: 0.65}
          - {at: 100, value: 0.55}
  - id: deep
    title: lower tide
    duration: 28s
    harmony: "Am9 Fmaj9 | Em9 Am9"
    scene: "bridge still"
    variation: "open-register"
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.65}
          - {at: 60, value: 0.85}
          - {at: 100, value: 0.8}
  - id: swell
    title: choir breath
    duration: 22s
    harmony: "Fmaj9 Am9 | Em9 Am9"
    scene: "bridge lift"
    variation: "sequence-up"
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.8}
          - {at: 60, value: 1.0}
          - {at: 100, value: 0.85}
  - id: close
    title: fog return
    duration: 20s
    harmony: "Am9 Am9"
    scene: "outro still"
    variation: "cadence"
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.55}
          - {at: 100, value: 0.1}

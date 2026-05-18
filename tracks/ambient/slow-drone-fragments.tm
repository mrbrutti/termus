title: Slow Drone Fragments
description: A-minor ambient — intent-driven layered drones, bell arpeggios, no drums. SP16 reference.
style: ambient
mix_bus: ambient
listen_mode: album-side
seed: 11023
tags: [ambient, drone, modal, sp16]
key: Amin
tempo: 60
globals: {density: sparse, brightness: warm, motion: slow, reverb: cathedral}

# 4-bar drone progression. Pad layers and bell motif are all intent-driven —
# the engine generates sustained voicings and bell arpeggios from each chord.

roles:
  drone_pad:
    family: pad
    voice: ambient_pad_dark
    auto_voice: pad_crossfade
    register: low
    prominence: anchor
    humanize: {timing_ms: 2, velocity: 4}
    chain: {reverb_send: 0.65, compress: glue, tape_drive_db: 0.4}

  drone_choir:
    family: choir
    voice: ambient_drone_choir
    auto_voice: pad_sustain
    register: mid
    prominence: support
    humanize: {timing_ms: 2, velocity: 4}
    chain: {reverb_send: 0.70, compress: glue}

  strings:
    family: strings
    voice: ambient_strings_soft
    auto_voice: pad_sustain
    register: mid
    prominence: air
    humanize: {timing_ms: 3, velocity: 5}
    chain: {reverb_send: 0.55, compress: glue}

  # Sparse bell motif — explicit events only on the "drift" section middle.
  bell_motif:
    family: bells
    voice: bell_struck_bright
    register: high
    prominence: air
    humanize: {timing_ms: 8, velocity: 10, accent: phrase_arc}
    chain: {reverb_send: 0.60, compress: "off"}
    loop_bars: 4
    events:
      - {beat: 2.0,  pitch: A5, dur: 3.0, vel: 50}
      - {beat: 8.0,  pitch: E5, dur: 3.0, vel: 46}
      - {beat: 13.0, pitch: G5, dur: 3.0, vel: 44}

sections:
  - id: emerge
    title: still water
    duration: 24s
    harmony: "Am9 Am11 | Em9 Am9"
    scene: "intro emerge"
    variation: "establish"
    intensity: 0.35
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.15}
          - {at: 100, value: 0.45}

  - id: drift
    title: motif rising
    duration: 32s
    harmony: "Am9 Fmaj7 | Cmaj7 Em9 | Am9 Fmaj7 | Em9 Am9"
    scene: "head drift"
    variation: "statement"
    intensity: 0.65
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 0.6}

  - id: recede
    title: fog return
    duration: 20s
    harmony: "Am9 Am9 | Em9 Am9"
    scene: "outro still"
    variation: "cadence"
    intensity: 0.3
    fill_at_end: true

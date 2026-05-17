title: Deep Field / Relay Signal Canon
description: Slower relay-tone drone with delayed lead entries, brighter shimmer at the midpoint, and a cold-air close.
style: drone
listen_mode: hour-stream
seed: 12884
tags: [drone, relay, signal, cold, slow]
key: Aminor
tempo: 42
globals:
  density: light
  brightness: balanced
  motion: still
  reverb: halo
  phrase: long
roles:
  bed:
    family: pad
    tone: [wide, soft]
    articulation: sustain
    register: mid
    prominence: support
    pattern: "x....... | ....x..."
  strings:
    family: strings
    tone: [soft, floating]
    articulation: sustain
    register: high
    prominence: support
    pattern: "....x... | x......."
  choir:
    family: choir
    tone: [airy, soft]
    articulation: sustain
    register: high
    prominence: air
    pattern: "x....... | ....x..."
  shimmer:
    family: lead
    tone: [icy, shimmer]
    articulation: bloom
    register: air
    prominence: air
    pattern: "........ | ....x..."
  bass:
    family: synth_bass
    tone: [warm, deep]
    articulation: sustain
    register: sub
    prominence: anchor
    pattern: "x....... | x......."
  lead:
    family: woodwind
    tone: [soft, breathy]
    articulation: breath
    register: mid-high
    prominence: lead
    motif: "5 . . 6 . . 7 9 | 7 . . 5 . 3 2 1"
sections:
  - id: establish
    title: relay hum
    duration: 75s
    harmony: "Am11 Fmaj9 | Cmaj9 G6 | Am11 Em7 | Fmaj9 G6"
    scene: "entry dim"
    variation: "settle"
    roles:
      lead:
        active: false
  - id: approach
    title: frost receiver
    duration: 80s
    harmony: "Am11 G6 | Fmaj9 Cmaj9 | Am11 Em7 | Fmaj9 G6"
    scene: "approach widen"
    variation: "foreground-shift"
    roles:
      lead:
        active: true
        motif: "9 . . 7 . . 5 3 | 5 . . 6 . . 7 5"
  - id: midpoint
    title: bright interference
    duration: 85s
    harmony: "Fmaj9 Cmaj9 | G6 Em7 | Am11 Fmaj9 | G6 G6"
    scene: "midpoint shimmer"
    variation: "lift"
    profile:
      density: medium
      brightness: bright
    roles:
      shimmer:
        pattern: "x....... | ....x..."
      choir:
        active: true
  - id: close
    title: cold-air fade
    duration: 70s
    harmony: "Am11 G6 | Fmaj9 Cmaj9 | Am11 Em7 | Fmaj9 G6"
    scene: "outro cold"
    variation: "cadence"
    roles:
      lead:
        motif: "3 . . 2 . . 1 . | . . . . . . . ."

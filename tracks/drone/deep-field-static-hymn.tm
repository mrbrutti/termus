title: Deep Field / Static Hymn
description: Low-frequency field hymn with choral openings, a stripped shadow center, and a wider return.
style: drone
listen_mode: hour-stream
seed: 64003
tags: [drone, field, hymn, low, longform]
key: Dminor
tempo: 44
globals:
  density: light
  brightness: warm
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
    pattern: "x....... | ....x..."
  choir:
    family: choir
    tone: [airy, soft]
    articulation: sustain
    register: high
    prominence: air
    pattern: "....x... | x......."
  shimmer:
    family: lead
    tone: [icy, shimmer]
    articulation: bloom
    register: air
    prominence: air
    pattern: "........ | ....x..."
  bass:
    family: synth_bass
    tone: [warm, dark]
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
    motif: "5 . . 7 . . 9 7 | 5 . . 3 . 2 1 ."
sections:
  - id: antenna
    title: antenna glow
    duration: 80s
    harmony: "Dm11 Bbmaj9 | Fmaj9 Cmaj9 | Dm11 Gm9 | Cmaj9 Cmaj9"
    scene: "establish wide"
    variation: "settle"
    roles:
      lead:
        active: false
      shimmer:
        active: false
  - id: cloud
    title: cloud cover
    duration: 85s
    harmony: "Dm11 Cmaj9 | Bbmaj9 Fmaj9 | Dm11 Gm9 | Cmaj9 Cmaj9"
    scene: "shadow darker"
    variation: "thin"
    profile:
      density: sparse
      brightness: warm
      motion: still
    roles:
      choir:
        active: false
      shimmer:
        active: false
      lead:
        motif: "5 . . . 3 . . . | 1 . . . . . . ."
  - id: horizon
    title: tower horizon
    duration: 90s
    harmony: "Dm11 Bbmaj9 | Fmaj9 Cmaj9 | Dm11 Gm9 | Cmaj9 Cmaj9"
    scene: "return wider"
    variation: "lift"
    profile:
      density: medium
      brightness: balanced
      motion: gentle
    roles:
      choir:
        active: true
      shimmer:
        active: true
        pattern: "x....... | ....x..."
      lead:
        active: true
        motif: "9 . . 7 . . 5 3 | 5 . . 6 . . 7 5"
  - id: fade
    title: long receiver
    duration: 75s
    harmony: "Dm11 Cmaj9 | Bbmaj9 Fmaj9 | Dm11 Gm9 | Cmaj9 Cmaj9"
    scene: "outro dissolve"
    variation: "cadence"
    roles:
      lead:
        motif: "3 . . 2 . . 1 . | . . . . . . . ."

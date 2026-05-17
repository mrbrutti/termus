title: Slow Signal / Harbor Phase Studies
description: Slightly warmer phase study with lower mallet cells, choir curtains, and a final lifted shimmer line.
style: phase
listen_mode: album-side
seed: 28763
tags: [phase, harbor, studies, mallet, signal]
key: Dminor
tempo: 72
globals:
  density: steady
  brightness: warm
  motion: moving
  reverb: halo
  phrase: long
roles:
  mallet-a:
    family: mallet
    tone: [glass, metallic]
    articulation: pulse
    register: mid
    prominence: lead
    motif: "5 . 3 . 5 . 7 5 | 3 . 2 . 1 . 2 3"
  mallet-b:
    family: mallet
    tone: [glass, metallic]
    articulation: pulse
    register: mid-high
    prominence: lead
    motif: ". 5 . 7 . 9 . 7 | . 3 . 5 . 2 . 1"
  pad:
    family: pad
    tone: [soft, wide]
    articulation: sustain
    register: mid
    prominence: support
    pattern: "x....... | ....x..."
  bass:
    family: synth_bass
    tone: [warm, direct]
    articulation: sustain
    register: low
    prominence: anchor
    pattern: "x...x... | x......."
  shimmer:
    family: bells
    tone: [sparkle, delicate]
    articulation: bloom
    register: air
    prominence: air
    pattern: "........ | ....x..."
  choir:
    family: choir
    tone: [airy, soft]
    articulation: sustain
    register: high
    prominence: air
    pattern: "....x... | x......."
sections:
  - id: measure-one
    title: harbor count
    duration: 40s
    harmony: "Dm9 Bbmaj9 | Fmaj9 Cmaj9"
    scene: "entry measured"
    variation: "establish"
    roles:
      choir:
        active: false
  - id: weave
    title: reflected buoys
    duration: 55s
    harmony: "Dm9 Cmaj9 | Bbmaj9 Fmaj9 | Dm9 Gm9 | Cmaj9 Cmaj9"
    scene: "weave active"
    variation: "interlock"
    roles:
      mallet-a:
        motif: "9 . 7 . 5 . 6 5 | 3 . 2 . 1 . 2 3"
      mallet-b:
        motif: ". 9 . 7 . 5 . 6 | . 3 . 2 . 1 . 2"
    events:
      - kind: pickup
        bar: 8
        roles: [mallet-a]
        motif: "3 5 6 9"
      - kind: stop
        bar: 10
        bars: 1
        roles: [mallet-a, bass]
  - id: curtain
    title: choir curtain
    duration: 45s
    harmony: "Bbmaj9 Fmaj9 | Dm9 Dm9 | Gm9 Cmaj9 | Cmaj9 Cmaj9"
    scene: "middle suspended"
    variation: "thin"
    profile:
      density: light
      brightness: warm
      motion: still
    roles:
      choir:
        active: true
      mallet-b:
        motif: ". 5 . . . 3 . . | . 1 . . . . . ."
      shimmer:
        active: false
    events:
      - kind: drop
        bar: 6
        bars: 1
        roles: [mallet-a]
  - id: lift
    title: signal rise
    duration: 45s
    harmony: "Dm9 Bbmaj9 | Fmaj9 Cmaj9 | Dm9 Gm9 | Cmaj9 Cmaj9"
    scene: "return brighter"
    variation: "lift"
    profile:
      density: busy
      brightness: bright
      motion: moving
    roles:
      shimmer:
        active: true
        pattern: "x....... | ....x..."
      mallet-a:
        motif: "11 . 9 . 7 . 9 7 | 5 . 3 . 2 . 1 ."
    events:
      - kind: pickup
        bar: 9
        roles: [mallet-b]
        motif: "5 6 7 9"
  - id: fade
    title: harbor empty
    duration: 35s
    harmony: "Dm9 Bbmaj9 | Fmaj9 Cmaj9"
    scene: "outro resolve"
    variation: "cadence"
    roles:
      mallet-a:
        motif: "3 . 2 . 1 . . . | . . . . . . . ."
      mallet-b:
        active: false

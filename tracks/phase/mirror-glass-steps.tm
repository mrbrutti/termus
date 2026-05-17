title: Slow Signal / Mirror Glass Steps
description: Interlocking mallet figure with a low anchor, choral swells, and a brighter mirrored return.
style: phase
listen_mode: album-side
seed: 22831
tags: [phase, glass, mirror, pulse, signal]
key: Aminor
tempo: 76
globals:
  density: steady
  brightness: balanced
  motion: moving
  reverb: halo
  phrase: long
roles:
  mallet-a:
    family: mallet
    tone: [glass, metallic]
    articulation: pulse
    register: mid-high
    prominence: lead
    motif: "5 . 7 . 9 . 7 5 | 3 . 5 . 2 . 1 ."
  mallet-b:
    family: mallet
    tone: [glass, metallic]
    articulation: pulse
    register: high
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
    pattern: "....x... | ..x....."
  choir:
    family: choir
    tone: [airy, soft]
    articulation: sustain
    register: high
    prominence: air
    pattern: "........ | x......."
sections:
  - id: intro
    title: mirrored count
    duration: 40s
    harmony: "Am9 G6 | Fmaj9 Em7"
    scene: "intro pulse"
    variation: "establish"
    roles:
      choir:
        active: false
      shimmer:
        pattern: "........ | ....x..."
  - id: weave
    title: stairwell braid
    duration: 55s
    harmony: "Am9 G6 | Fmaj9 Em7 | Dm9 Cmaj9 | Fmaj9 Em7"
    scene: "weave active"
    variation: "interlock"
    roles:
      mallet-a:
        motif: "9 . 7 . 5 . 6 5 | 3 . 2 . 1 . 2 3"
      mallet-b:
        motif: ". 9 . 7 . 5 . 6 | . 3 . 2 . 1 . 2"
  - id: shadow
    title: lower glass
    duration: 45s
    harmony: "Fmaj9 Em7 | Am9 Am9 | Dm9 Cmaj9 | Fmaj9 Em7"
    scene: "shadow thin"
    variation: "subtract"
    profile:
      density: light
      brightness: warm
      motion: still
    roles:
      choir:
        active: false
      shimmer:
        active: false
      mallet-b:
        motif: ". 5 . . . 3 . . | . 1 . . . . . ."
  - id: return
    title: bright landing
    duration: 45s
    harmony: "Am9 G6 | Fmaj9 Em7 | Dm9 Cmaj9 | Fmaj9 Em7"
    scene: "return lift"
    variation: "higher-register"
    profile:
      density: busy
      brightness: bright
      motion: moving
    roles:
      shimmer:
        pattern: "x...x... | ..x...x."
      choir:
        active: true
  - id: outro
    title: empty steps
    duration: 35s
    harmony: "Am9 G6 | Fmaj9 Em7"
    scene: "outro resolve"
    variation: "cadence"
    roles:
      mallet-a:
        motif: "3 . 2 . 1 . . . | . . . . . . . ."
      mallet-b:
        active: false

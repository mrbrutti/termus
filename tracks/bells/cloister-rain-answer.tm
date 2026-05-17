title: Glass Chapel / Cloister Rain Answer
description: Lower bell study with music-box replies, wider strings in the center, and a rain-lit coda.
style: bells
listen_mode: album-side
seed: 11908
tags: [bells, cloister, rain, celesta, sacred]
key: Dminor
tempo: 54
globals:
  density: sparse
  brightness: warm
  motion: gentle
  reverb: halo
  phrase: long
roles:
  bells:
    family: bells
    tone: [glass, soft]
    articulation: bloom
    register: mid-high
    prominence: lead
    motif: "5 . . 6 . . 7 5 | 3 . . 2 . . 1 ."
  celesta:
    family: mallet
    tone: [delicate, warm]
    articulation: echo
    register: high
    prominence: air
    pattern: "..x..... | ....x..."
  box:
    family: music_box
    tone: [delicate, sparkle]
    articulation: echo
    register: air
    prominence: air
    pattern: "....x... | ..x....."
  pad:
    family: pad
    tone: [soft, wide]
    articulation: sustain
    register: mid
    prominence: support
    pattern: "x....... | ....x..."
  choir:
    family: choir
    tone: [airy, soft]
    articulation: sustain
    register: high
    prominence: support
    pattern: "....x... | x......."
  strings:
    family: strings
    tone: [soft, floating]
    articulation: sustain
    register: mid-high
    prominence: support
    pattern: "x....... | ....x..."
  bass:
    family: bass
    tone: [soft, round]
    articulation: sustain
    register: low
    prominence: anchor
    pattern: "x....... | x......."
sections:
  - id: porch
    title: cloister drip
    duration: 40s
    harmony: "Dm7 Bbmaj7 | Fmaj7 Cmaj7 | Dm7 Gm7 | A7 A7"
    scene: "entry rain"
    variation: "establish"
    roles:
      choir:
        active: false
      strings:
        active: false
  - id: walkway
    title: stone walkway
    duration: 50s
    harmony: "Dm7 Bbmaj7 | Fmaj7 Cmaj7 | Gm7 Dm7 | A7 Dm7"
    scene: "walkway answer"
    variation: "statement"
    roles:
      bells:
        motif: "9 . . 7 . . 5 3 | 5 . . 6 . . 7 5"
      box:
        pattern: "x....... | ....x..."
    events:
      - kind: pickup
        bar: 8
        roles: [bells]
        motif: "5 6 7 9"
      - kind: stop
        bar: 10
        bars: 1
        roles: [bells, pad, bass]
  - id: gallery
    title: gallery blue
    duration: 45s
    harmony: "Gm7 Dm7 | Bbmaj7 Cmaj7 | Fmaj7 Cmaj7 | A7 A7"
    scene: "middle widen"
    variation: "lift"
    profile:
      density: light
      brightness: balanced
      motion: moving
    roles:
      strings:
        active: true
      choir:
        active: true
      bells:
        motif: "11 . . 9 . . 7 5 | 9 . . 7 . . 5 3"
    events:
      - kind: stab
        bar: 8
        roles: [pad]
        pattern: "x... ...."
      - kind: drop
        bar: 10
        bars: 1
        roles: [box]
  - id: fade
    title: rain-lit coda
    duration: 40s
    harmony: "Dm7 Bbmaj7 | Fmaj7 Cmaj7 | Dm7 Gm7 | A7 Dm7"
    scene: "outro warm"
    variation: "cadence"
    roles:
      bells:
        motif: "3 . . 2 . . 1 . | . . . . . . . ."

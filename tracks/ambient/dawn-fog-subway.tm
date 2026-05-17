title: Night Drift / Dawn Fog Subway
description: Cooler ambient drift with string swells, a reed-like center line, and a slowly waking closing lift.
style: ambient
listen_mode: hour-stream
seed: 11842
tags: [ambient, dawn, fog, subway, drift]
key: Bminor
tempo: 60
globals:
  density: light
  brightness: balanced
  motion: gentle
  reverb: halo
  phrase: long
roles:
  pad:
    family: pad
    tone: [dreamy, wide]
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
  texture:
    family: bells
    tone: [sparkle, delicate]
    articulation: shimmer
    register: air
    prominence: air
    pattern: "..x..... | ....x..."
  flute:
    family: woodwind
    tone: [soft, breathy]
    articulation: breath
    register: high
    prominence: lead
    motif: "5 . . 6 . . 7 9 | 7 . . 5 . 3 2 1"
  bass:
    family: synth_bass
    tone: [warm, direct]
    articulation: sustain
    register: low
    prominence: anchor
    pattern: "x....... | x......."
sections:
  - id: first-light
    title: service tunnel
    duration: 65s
    harmony: "Bm11 Gmaj9 | Dmaj9 Asus4 | Bm11 F#m7 | Gmaj9 Asus4"
    scene: "entry cool"
    variation: "establish"
    roles:
      flute:
        active: false
  - id: platform
    title: first commuters
    duration: 80s
    harmony: "Bm11 Amaj9 | Gmaj9 Dmaj9 | Bm11 F#m7 | Gmaj9 Asus4"
    scene: "platform broaden"
    variation: "foreground-shift"
    roles:
      flute:
        active: true
        motif: "9 . . 7 . . 5 3 | 5 . . 6 . . 7 5"
  - id: underpass
    title: tunnel bloom
    duration: 75s
    harmony: "Gmaj9 Dmaj9 | Amaj9 F#m7 | Bm11 Gmaj9 | Asus4 Asus4"
    scene: "middle suspended"
    variation: "thin"
    profile:
      density: sparse
      brightness: warm
      motion: still
    roles:
      choir:
        active: false
      texture:
        pattern: "........ | ....x..."
      flute:
        motif: "5 . . . 3 . . . | 1 . . . . . . ."
  - id: waking
    title: station waking
    duration: 75s
    harmony: "Bm11 Gmaj9 | Dmaj9 Asus4 | Bm11 F#m7 | Gmaj9 Asus4"
    scene: "return brighter"
    variation: "lift"
    profile:
      density: medium
      brightness: bright
      motion: moving
    roles:
      texture:
        pattern: "x....... | ....x..."
      choir:
        active: true
      flute:
        motif: "11 . . 9 . . 7 5 | 9 . . 7 . . 5 3"
  - id: fade
    title: stairwell light
    duration: 60s
    harmony: "Bm11 Amaj9 | Gmaj9 Dmaj9 | Bm11 Gmaj9 | Asus4 Asus4"
    scene: "outro settle"
    variation: "cadence"
    roles:
      flute:
        motif: "3 . . 2 . . 1 . | . . . . . . . ."

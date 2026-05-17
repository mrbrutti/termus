title: Night Drift / Platform Weather
description: Slow station piece with low bass fog, a darker center passage, and a bright final platform return.
style: ambient
listen_mode: hour-stream
seed: 99317
tags: [ambient, station, weather, slow, endless]
key: Eminor
tempo: 56
globals:
  density: light
  brightness: warm
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
    pattern: "....x... | x......."
  choir:
    family: choir
    tone: [airy, soft]
    articulation: sustain
    register: high
    prominence: air
    pattern: "x....... | ....x..."
  texture:
    family: bells
    tone: [glass, sparkle]
    articulation: shimmer
    register: air
    prominence: air
    pattern: "....x... | ..x....."
  lead:
    family: woodwind
    tone: [soft, breathy]
    articulation: breath
    register: mid-high
    prominence: lead
    motif: "5 . . 7 . . 9 7 | 5 . . 3 . 2 1 ."
  bass:
    family: synth_bass
    tone: [warm, wide]
    articulation: sustain
    register: low
    prominence: anchor
    pattern: "x....... | x......."
  shimmer:
    family: lead
    tone: [soft, icy]
    articulation: bloom
    register: air
    prominence: air
    pattern: "........ | ....x..."
sections:
  - id: entry
    title: empty platform
    duration: 70s
    harmony: "Em11 Cmaj9 | G6 Dsus4 | Em11 Bm7 | Cmaj9 Dsus4"
    scene: "entry fog"
    variation: "establish"
    roles:
      lead:
        active: false
      shimmer:
        active: false
  - id: movement-a
    title: trackside weather
    duration: 85s
    harmony: "Em11 Dmaj9 | Cmaj9 G6 | Em11 Bm7 | Cmaj9 Dsus4"
    scene: "movement drift"
    variation: "foreground-shift"
    roles:
      lead:
        active: true
        motif: "9 . . 7 . . 5 3 | 5 . . 6 . . 7 5"
      texture:
        pattern: "..x..... | ....x..."
    events:
      - kind: pickup
        bar: 8
        roles: [lead]
        motif: "5 6 7 9"
      - kind: stop
        bar: 12
        bars: 1
        roles: [pad, bass, lead]
  - id: shadow
    title: tannoy distance
    duration: 75s
    harmony: "Cmaj9 G6 | Dsus4 Bm7 | Em11 Cmaj9 | Dsus4 Dsus4"
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
    events:
      - kind: drop
        bar: 6
        bars: 2
        roles: [strings, bass]
  - id: return
    title: last train glow
    duration: 80s
    harmony: "Em11 Cmaj9 | G6 Dsus4 | Em11 Bm7 | Cmaj9 Dsus4"
    scene: "return lift"
    variation: "higher-register"
    profile:
      density: medium
      brightness: balanced
      motion: moving
    roles:
      lead:
        motif: "11 . . 9 . . 7 5 | 9 . . 7 . . 5 3"
      shimmer:
        active: true
        pattern: "x....... | ....x..."
    events:
      - kind: pickup
        bar: 10
        roles: [lead]
        motif: "7 9 11 9"
      - kind: stab
        bar: 14
        roles: [pad]
        pattern: "x... ...."
  - id: exit
    title: tunnel air
    duration: 65s
    harmony: "Em11 Dmaj9 | Cmaj9 G6 | Em11 Cmaj9 | Dsus4 Dsus4"
    scene: "exit settle"
    variation: "cadence"
    roles:
      lead:
        motif: "3 . . 2 . . 1 . | . . . . . . . ."
      shimmer:
        pattern: "........ | ....x..."

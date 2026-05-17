title: Night Drift / Platform Weather
description: Slow platform ambience with a darker center movement and a higher final return.
style: ambient
listen_mode: hour-stream
seed: 99317
tags: [ambient, station, slow, endless]
key: Eminor
tempo: 58
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
    pattern: "x... .... | x... ...."
  choir:
    family: choir
    tone: [airy, soft]
    articulation: sustain
    register: high
    prominence: air
    pattern: ".... x... | .... x..."
  texture:
    family: bells
    tone: [glass, sparkle]
    articulation: shimmer
    register: air
    prominence: air
    pattern: "x... .... | ..x. ...."
  lead:
    family: woodwind
    tone: [soft]
    articulation: breath
    register: mid-high
    prominence: lead
    motif: "5 . . 7 | 9 . 7 5"
  bass:
    family: synth_bass
    tone: [warm]
    articulation: sustain
    register: low
    prominence: anchor
    pattern: "x... .... | x... ...."
sections:
  - id: intro
    title: empty platform
    duration: 180s
    harmony: "Em11 Cmaj9 | Gmaj9 Dsus4"
    scene: "intro fog"
    variation: "establish"
    roles:
      lead:
        active: false
  - id: middle
    title: weather report
    duration: 240s
    harmony: "Em11 Dmaj9 | Cmaj9 G6"
    scene: "middle darker"
    variation: "shadow"
    profile:
      density: sparse
      brightness: warm
      motion: still
    roles:
      texture:
        pattern: ".... x... | .... ..x."
      choir:
        active: false
  - id: return
    title: last train glow
    duration: 210s
    harmony: "Em11 Cmaj9 | Gmaj9 Dsus4"
    scene: "return lift"
    variation: "higher-register"
    roles:
      lead:
        active: true
        motif: "9 . 7 5 | 11 . 9 7"
      choir:
        active: true


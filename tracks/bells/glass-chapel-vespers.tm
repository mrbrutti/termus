title: Glass Chapel / Vespers
description: Sparse bell-lit movements with a dim middle aisle and a luminous closing cadence.
style: bells
listen_mode: album-side
seed: 8801
tags: [bells, glass, sacred, sparse]
key: Amin
tempo: 52
globals:
  density: sparse
  brightness: bright
  motion: gentle
  reverb: halo
  phrase: long
roles:
  bells:
    family: bells
    tone: [glass, luminous]
    articulation: bloom
    register: high
    prominence: lead
    motif: "5 . . 7 | 9 . 7 5"
  celesta:
    family: mallet
    tone: [sparkle, delicate]
    articulation: echo
    register: air
    prominence: air
    pattern: "x... .... | ..x. ...."
  pad:
    family: pad
    tone: [soft, wide]
    articulation: sustain
    register: mid
    prominence: support
    pattern: "x... .... | x... ...."
  choir:
    family: choir
    tone: [airy]
    articulation: sustain
    register: mid-high
    prominence: support
    pattern: "x... .... | .... x..."
  bass:
    family: bass
    tone: [soft]
    articulation: sustain
    register: low
    prominence: anchor
    pattern: "x... .... | x... ...."
sections:
  - id: intro
    title: first light
    duration: 150s
    harmony: "Am7 Gmaj7 | Dm7 Am7"
    scene: "intro chapel hush"
    variation: "establish"
    roles:
      bells:
        motif: "5 . . 7 | 9 . 7 5"
      choir:
        active: false
  - id: nave
    title: nave echo
    duration: 180s
    harmony: "Am7 Gmaj7 | Dm7 Am7"
    scene: "nave wider"
    variation: "answer"
    profile:
      density: light
      brightness: balanced
    roles:
      bells:
        motif: "9 . 7 5 | 3 . 2 1"
      choir:
        active: true
  - id: aisle
    title: stone halo
    duration: 120s
    harmony: "Dm7 Am7 | Gmaj7 Am7"
    scene: "aisle dim"
    variation: "subtract"
    profile:
      density: sparse
      brightness: warm
      motion: still
    roles:
      bells:
        motif: "5 . . . | 3 . . ."
      celesta:
        active: false
      choir:
        active: false
  - id: outro
    title: vesper close
    duration: 150s
    harmony: "Am7 Gmaj7 | Dm7 Am7"
    scene: "outro luminous"
    variation: "cadence"
    roles:
      bells:
        motif: "3 . 2 1 | 1 . . ."
      choir:
        active: true


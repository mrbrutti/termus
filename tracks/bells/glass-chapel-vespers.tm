title: Glass Chapel / Vespers
description: Bell-led chapel suite with a dim aisle center, choir returns, and a luminous final cadence.
style: bells
listen_mode: album-side
seed: 8801
tags: [bells, chapel, sacred, glass, luminous]
key: Amin
tempo: 50
globals:
  density: sparse
  brightness: balanced
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
    motif: "5 . . 7 . . 9 7 | 5 . . 3 . 2 1 ."
  celesta:
    family: mallet
    tone: [sparkle, delicate]
    articulation: echo
    register: high
    prominence: air
    pattern: "..x..... | ....x..."
  glock:
    family: bells
    tone: [glass, sparkle]
    articulation: echo
    register: air
    prominence: air
    pattern: "....x... | ..x....."
  box:
    family: music_box
    tone: [delicate, soft]
    articulation: echo
    register: air
    prominence: air
    pattern: "........ | ....x..."
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
    pattern: "x....... | ....x..."
  strings:
    family: strings
    tone: [soft, floating]
    articulation: sustain
    register: mid-high
    prominence: support
    pattern: "....x... | x......."
  bass:
    family: bass
    tone: [soft, round]
    articulation: sustain
    register: low
    prominence: anchor
    pattern: "x....... | x......."
sections:
  - id: first-light
    title: first light
    duration: 45s
    harmony: "Am7 Gmaj7 | Dm7 Am7 | Fmaj7 Em7 | Dm7 E7"
    scene: "intro chapel"
    variation: "establish"
    roles:
      choir:
        active: false
      strings:
        active: false
  - id: nave
    title: nave echo
    duration: 55s
    harmony: "Am7 Gmaj7 | Dm7 Am7 | Cmaj7 Gmaj7 | Dm7 E7"
    scene: "nave widen"
    variation: "answer"
    roles:
      bells:
        motif: "9 . . 7 . . 5 3 | 5 . . 6 . . 7 5"
      choir:
        active: true
  - id: aisle
    title: stone aisle
    duration: 45s
    harmony: "Dm7 Am7 | Gmaj7 Am7 | Fmaj7 Em7 | Dm7 E7"
    scene: "aisle dim"
    variation: "subtract"
    profile:
      density: sparse
      brightness: warm
      motion: still
    roles:
      bells:
        motif: "5 . . . 3 . . . | 1 . . . . . . ."
      glock:
        active: false
      choir:
        active: false
  - id: apse
    title: colored glass
    duration: 50s
    harmony: "Fmaj7 Em7 | Am7 Gmaj7 | Dm7 Am7 | Dm7 E7"
    scene: "apse brighter"
    variation: "lift"
    profile:
      density: light
      brightness: bright
      motion: moving
    roles:
      strings:
        active: true
      glock:
        active: true
        pattern: "x....... | ....x..."
      bells:
        motif: "11 . . 9 . . 7 5 | 9 . . 7 . . 5 3"
  - id: close
    title: vesper close
    duration: 45s
    harmony: "Am7 Gmaj7 | Dm7 Am7 | Fmaj7 Em7 | Dm7 E7"
    scene: "outro luminous"
    variation: "cadence"
    roles:
      bells:
        motif: "3 . . 2 . . 1 . | . . . . . . . ."

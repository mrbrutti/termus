title: Sleep Walk / Staircase Lullaby
description: Celesta-led cradle tune with harp replies, a suspended middle room, and a soft room-light close.
style: lullaby
listen_mode: album-side
seed: 44701
tags: [lullaby, celesta, harp, soft, night]
key: Gmajor
tempo: 68
globals:
  density: light
  brightness: warm
  motion: gentle
  reverb: halo
  phrase: long
roles:
  lead:
    family: mallet
    tone: [delicate, warm]
    articulation: lyrical
    register: high
    prominence: lead
    motif: "5 . 6 . 7 . 6 5 | 3 . 2 . 1 . . ."
  harp:
    family: strings
    tone: [soft, airy]
    articulation: answer
    register: mid-high
    prominence: support
    pattern: ".x..x... | ..x...x."
  choir:
    family: choir
    tone: [airy, soft]
    articulation: sustain
    register: high
    prominence: support
    pattern: "x....... | ....x..."
  box:
    family: music_box
    tone: [delicate, sparkle]
    articulation: echo
    register: high
    prominence: air
    pattern: "....x... | ..x....."
  pad:
    family: pad
    tone: [soft, wide]
    articulation: sustain
    register: mid
    prominence: support
    pattern: "x....... | ....x..."
sections:
  - id: intro
    title: upstairs hush
    duration: 35s
    harmony: "Gmaj7 D/F# | Em9 Cmaj9 | Gmaj7 D/F# | Cmaj9 D7"
    scene: "intro hush"
    variation: "establish"
    roles:
      choir:
        active: false
      pad:
        active: false
  - id: verse-a
    title: blanket fold
    duration: 50s
    harmony: "Gmaj7 D/F# | Em9 Cmaj9 | Gmaj7 Bm7 | Cmaj9 D7"
    scene: "theme cradle"
    variation: "statement"
    roles:
      lead:
        motif: "9 . 7 . 5 . 6 5 | 3 . 2 . 1 . . ."
  - id: middle
    title: hallway moon
    duration: 45s
    harmony: "Em9 Cmaj9 | Gmaj7 Gmaj7 | Am9 D7 | Gmaj7 D7"
    scene: "middle suspended"
    variation: "subtract"
    profile:
      density: sparse
      brightness: warm
      motion: still
    roles:
      lead:
        motif: "5 . . . 3 . . . | 1 . . . . . . ."
      box:
        pattern: "x....... | ....x..."
      choir:
        active: false
  - id: return
    title: room-light answer
    duration: 45s
    harmony: "Gmaj7 D/F# | Em9 Cmaj9 | Gmaj7 Bm7 | Cmaj9 D7"
    scene: "return fuller"
    variation: "lift"
    roles:
      choir:
        active: true
      pad:
        active: true
      lead:
        motif: "11 . 9 . 7 . 6 5 | 3 . 2 . 1 . . ."
  - id: outro
    title: stairlight dim
    duration: 35s
    harmony: "Gmaj7 D/F# | Em9 Cmaj9 | Gmaj7 D/F# | Cmaj9 D7"
    scene: "outro close"
    variation: "cadence"
    roles:
      lead:
        motif: "3 . 2 . 1 . . . | . . . . . . . ."

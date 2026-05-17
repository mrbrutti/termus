title: Sleep Walk / Paper Moon Box
description: Music-box lullaby with suspended choir pads, longer harp answers, and a tender moonlit close.
style: lullaby
listen_mode: album-side
seed: 22119
tags: [lullaby, moon, music-box, soft, cradle]
key: Emajor
tempo: 66
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
    motif: "5 . 6 . 7 . 9 7 | 5 . 3 . 2 . 1 ."
  harp:
    family: strings
    tone: [airy, soft]
    articulation: answer
    register: high
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
    register: air
    prominence: air
    pattern: "..x..... | ....x..."
  pad:
    family: pad
    tone: [soft, wide]
    articulation: sustain
    register: mid
    prominence: support
    pattern: "x....... | ....x..."
sections:
  - id: opening
    title: paper moon
    duration: 35s
    harmony: "Emaj7 B/D# | C#m9 Amaj9 | Emaj7 B/D# | Amaj9 B7"
    scene: "intro hush"
    variation: "establish"
    roles:
      choir:
        active: false
      pad:
        active: false
  - id: cradle
    title: cradle line
    duration: 50s
    harmony: "Emaj7 B/D# | C#m9 Amaj9 | Emaj7 G#m7 | Amaj9 B7"
    scene: "theme cradle"
    variation: "statement"
    roles:
      lead:
        motif: "9 . 7 . 5 . 6 5 | 3 . 2 . 1 . . ."
  - id: middle
    title: window curtain
    duration: 45s
    harmony: "C#m9 Amaj9 | Emaj7 Emaj7 | F#m9 B7 | Emaj7 B7"
    scene: "middle suspended"
    variation: "thin"
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
    title: moonlit answer
    duration: 45s
    harmony: "Emaj7 B/D# | C#m9 Amaj9 | Emaj7 G#m7 | Amaj9 B7"
    scene: "return fuller"
    variation: "lift"
    roles:
      choir:
        active: true
      pad:
        active: true
      lead:
        motif: "11 . 9 . 7 . 6 5 | 3 . 2 . 1 . . ."
  - id: close
    title: moon box close
    duration: 35s
    harmony: "Emaj7 B/D# | C#m9 Amaj9 | Emaj7 B/D# | Amaj9 B7"
    scene: "outro settle"
    variation: "cadence"
    roles:
      lead:
        motif: "3 . 2 . 1 . . . | . . . . . . . ."

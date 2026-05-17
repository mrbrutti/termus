title: Soft Tape / Window Seat Letters
description: Mid-tempo notebook beat with guitar replies, a denser back half, and a soft handwritten outro.
style: lofi
listen_mode: album-side
seed: 51203
tags: [lofi, letters, train, study, warm]
key: Cmin
tempo: 76
globals:
  density: steady
  brightness: balanced
  motion: gentle
  reverb: room
  phrase: long
roles:
  keys:
    family: electric_piano
    tone: [warm, dusty, soft]
    articulation: stab
    register: mid
    prominence: support
    pattern: "x...x..x | x.x...x."
  bass:
    family: bass
    tone: [round, woody]
    articulation: legato
    register: low
    prominence: anchor
    pattern: "x...x... | x..xx..."
  kick:
    family: drums
    tone: [dusty, soft]
    articulation: pocket
    prominence: anchor
    pattern: "x...x... | x...x..x"
  snare:
    family: drums
    tone: [soft, dusty]
    articulation: pocket
    prominence: anchor
    pattern: "....x... | ....x..."
  hat:
    family: drums
    tone: [dry, light]
    articulation: pocket
    prominence: support
    pattern: "x.x.x.x. | x.x.xx.."
  guitar:
    family: guitar
    tone: [soft, warm]
    articulation: answer
    register: mid
    prominence: support
    pattern: ".x....x. | ..x.x..."
  vibes:
    family: mallet
    tone: [soft, glass]
    articulation: halo
    register: high
    prominence: air
    pattern: "....x... | ..x....."
  lead:
    family: reed_lead
    tone: [breathy, intimate]
    articulation: lyrical
    register: mid-high
    prominence: lead
    motif: "5 . . 6 7 . 9 7 | 5 . 3 . 2 . 1 ."
sections:
  - id: intro
    title: folded paper
    duration: 40s
    harmony: "Cm9 F13 | Bbmaj9 G7 | Cm9 Abmaj9 | F13 G7"
    scene: "intro hush"
    variation: "establish"
    profile:
      density: sparse
      brightness: warm
      motion: still
    roles:
      lead:
        active: false
      guitar:
        active: false
      hat:
        pattern: "x...x... | x...x..."
  - id: verse-a
    title: margin note
    duration: 55s
    harmony: "Cm9 F13 | Bbmaj9 G7 | Ebmaj9 D7 | Gm9 C7"
    scene: "head note"
    variation: "introduce-hook"
    roles:
      lead:
        motif: "9 . 7 . 5 . 6 5 | 3 . 2 . 1 . . ."
  - id: verse-b
    title: underlined answer
    duration: 55s
    harmony: "Gm9 C13 | Fmaj9 D7 | Cm9 F13 | Bbmaj9 G7"
    scene: "reply denser"
    variation: "answer-lift"
    profile:
      density: busy
      motion: moving
    roles:
      lead:
        motif: "5 . 6 . 7 . 9 b9 | 7 . 5 . 3 . 2 1"
      guitar:
        active: true
        pattern: ".x..x..x | ..x..x.."
      kick:
        pattern: "x..xx..x | x...x..x"
  - id: bridge
    title: radiator paragraph
    duration: 50s
    harmony: "Abmaj9 G7 | Cm9 Cm9 | Ebmaj9 D7 | Gm9 C7"
    scene: "bridge intimate"
    variation: "subtract"
    profile:
      density: light
      brightness: warm
      motion: still
      reverb: halo
    roles:
      lead:
        motif: "11 . 9 . 7 . . . | 5 . 3 . 1 . . ."
      guitar:
        active: false
      vibes:
        pattern: "x....... | ....x..."
      snare:
        pattern: "........ | ....x..."
  - id: outro
    title: last paragraph
    duration: 45s
    harmony: "Cm9 F13 | Bbmaj9 G7 | Cm9 Abmaj9 | F13 G7"
    scene: "outro fade"
    variation: "cadence"
    roles:
      lead:
        motif: "3 . 2 . 1 . . . | . . . . . . . ."
      guitar:
        active: false

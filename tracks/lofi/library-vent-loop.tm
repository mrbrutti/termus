title: Soft Tape / Library Vent Loop
description: Study-floor loop with brushed energy, sparse top lines, and a muted fluorescent middle.
style: lofi
listen_mode: album-side
seed: 70301
tags: [lofi, library, study, fluorescent, loop]
key: Amin
tempo: 72
globals:
  density: steady
  brightness: warm
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
    pattern: "x...x... | x..x.x.."
  bass:
    family: bass
    tone: [round, woody]
    articulation: legato
    register: low
    prominence: anchor
    pattern: "x...x... | x...x..x"
  kick:
    family: drums
    tone: [dusty, soft]
    articulation: pocket
    prominence: anchor
    pattern: "x...x... | x...x..."
  snare:
    family: drums
    tone: [soft, dry]
    articulation: pocket
    prominence: anchor
    pattern: "....x... | ....x..."
  hat:
    family: drums
    tone: [dry, light]
    articulation: pocket
    prominence: support
    pattern: "x...x.x. | x...x.x."
  vibes:
    family: mallet
    tone: [soft, glass]
    articulation: halo
    register: high
    prominence: air
    pattern: "....x... | ..x....."
  guitar:
    family: guitar
    tone: [soft, warm]
    articulation: answer
    register: mid
    prominence: support
    pattern: ".x....x. | ....x..x"
  lead:
    family: reed_lead
    tone: [breathy, intimate]
    articulation: lyrical
    register: mid-high
    prominence: lead
    motif: "5 . . 7 . . 9 7 | 5 . . 3 . 2 1 ."
sections:
  - id: intro
    title: desk lamp
    duration: 35s
    harmony: "Am9 D13 | Gmaj9 E7 | Am9 Fmaj9 | D13 E7"
    scene: "intro study"
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
  - id: verse-a
    title: highlighted page
    duration: 55s
    harmony: "Am9 D13 | Gmaj9 E7 | Cmaj9 B7 | Em9 A7"
    scene: "head repeat"
    variation: "introduce-hook"
    roles:
      lead:
        motif: "9 . 7 . 5 . 6 5 | 3 . 2 . 1 . . ."
      hat:
        pattern: "x.x.x.x. | x...x.x."
  - id: middle
    title: fluorescent blur
    duration: 50s
    harmony: "Fmaj9 E7 | Am9 Am9 | Cmaj9 B7 | Em9 A7"
    scene: "middle thin"
    variation: "subtract"
    profile:
      density: light
      brightness: warm
      motion: still
      reverb: halo
    roles:
      lead:
        motif: "5 . . . 3 . . . | 1 . . . . . . ."
      guitar:
        active: false
      hat:
        pattern: "x...x... | ....x..."
      vibes:
        pattern: "x....... | ....x..."
  - id: verse-b
    title: margin return
    duration: 55s
    harmony: "Em9 A13 | Dmaj9 B7 | Am9 D13 | Gmaj9 E7"
    scene: "reply denser"
    variation: "answer-lift"
    profile:
      density: busy
      brightness: balanced
      motion: moving
    roles:
      lead:
        motif: "5 . 6 7 9 . 7 5 | 3 . 2 . 1 . 7 5"
      guitar:
        active: true
        pattern: ".x..x..x | ..x...x."
      kick:
        pattern: "x..xx..x | x...x..x"
  - id: outro
    title: air vent fade
    duration: 40s
    harmony: "Am9 D13 | Gmaj9 E7 | Am9 Fmaj9 | D13 E7"
    scene: "outro close"
    variation: "cadence"
    roles:
      lead:
        motif: "3 . 2 . 1 . . . | . . . . . . . ."

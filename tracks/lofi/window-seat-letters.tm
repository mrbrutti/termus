title: Window Seat Letters
description: Notebook lofi with upright piano comp, clarinet lead fragments, and a suspended radiator middle.
style: lofi
listen_mode: album-side
seed: 51203
tags: [lofi, letters, train, piano, clarinet]
key: Cmin
tempo: 74
globals:
  density: steady
  brightness: warm
  motion: gentle
  reverb: room
  phrase: long
roles:
  piano:
    family: acoustic_piano
    tone: [warm, soft]
    articulation: stab
    register: mid
    prominence: support
    pattern: "x...x... | .x..x..x"
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
    pattern: "x.x...x. | x...x.x."
  clarinet:
    family: woodwind
    tone: [soft, breathy]
    articulation: lyrical
    register: mid-high
    prominence: lead
    motif: "5 . . 6 . . 7 9 | 7 . . 5 . 3 2 1"
  guitar:
    family: guitar
    tone: [soft, warm]
    articulation: answer
    register: mid
    prominence: support
    pattern: "..x...x. | .x....x."
  pad:
    family: pad
    tone: [soft, wide]
    articulation: sustain
    register: mid
    prominence: air
    pattern: "x....... | ....x..."
sections:
  - id: intro
    title: folded paper
    duration: 14s
    harmony: "Cm9 F13 | Bbmaj9 G7 | Cm9 Abmaj9 | F13 G7"
    scene: "intro hush"
    variation: "establish"
    profile:
      density: sparse
      brightness: warm
      motion: still
    roles:
      clarinet:
        active: true
        motif: "5 . . . 3 . . . | 2 . . . 1 . . ."
      guitar:
        active: false
      pad:
        active: false
      piano:
        pattern: "x....... | .x..x..."
  - id: verse-a
    title: margin note
    duration: 55s
    harmony: "Cm9 F13 | Bbmaj9 G7 | Ebmaj9 D7 | Gm9 C7"
    scene: "head note"
    variation: "introduce-hook"
    roles:
      clarinet:
        active: true
        motif: "9 . . 7 . . 5 3 | 5 . . 6 . . 7 5"
    events:
      - kind: pickup
        bar: 8
        roles: [clarinet]
        motif: "3 5 6 9"
      - kind: fill
        bar: 8
        roles: [hat, snare]
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
      clarinet:
        motif: "11 . . 9 . . 7 5 | 7 . . 5 . 3 2 1"
      guitar:
        active: true
        pattern: ".x..x..x | ..x...x."
      kick:
        pattern: "x..xx..x | x...x..x"
    events:
      - kind: stab
        bar: 9
        roles: [piano]
        pattern: "x... ...."
      - kind: fill
        bar: 12
        roles: [hat, snare, kick]
  - id: bridge
    title: radiator glow
    duration: 50s
    harmony: "Abmaj9 G7 | Cm9 Cm9 | Ebmaj9 D7 | Gm9 C7"
    scene: "bridge suspended"
    variation: "subtract"
    profile:
      density: light
      brightness: warm
      motion: still
      reverb: halo
    roles:
      clarinet:
        motif: "5 . . . 3 . . . | 1 . . . . . . ."
      guitar:
        active: false
      hat:
        pattern: "x...x... | ....x..."
      pad:
        active: true
    events:
      - kind: drop
        bar: 4
        bars: 1
        roles: [bass]
  - id: outro
    title: last paragraph
    duration: 40s
    harmony: "Cm9 F13 | Bbmaj9 G7 | Cm9 Abmaj9 | F13 G7"
    scene: "outro fade"
    variation: "cadence"
    roles:
      clarinet:
        motif: "3 . . 2 . . 1 . | . . . . . . . ."

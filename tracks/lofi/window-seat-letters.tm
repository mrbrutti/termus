title: Soft Tape / Window Seat Letters
description: Brighter chord turns, patient head melody, and a brushed late-section lift.
style: lofi
listen_mode: album-side
seed: 51203
tags: [lofi, mellow, letters, window, study]
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
    pattern: "x..x .x.. | x... .x.x"
  bass:
    family: bass
    tone: [round, woody]
    articulation: legato
    register: low
    prominence: anchor
    pattern: "x... x... | x..x x..."
  lead:
    family: reed_lead
    tone: [breathy, intimate]
    articulation: lyrical
    register: mid-high
    prominence: lead
    motif: "5 . . 7 | 9 . 7 5 | 3 . 2 1 | . . . ."
  texture:
    family: mallet
    tone: [glass, soft]
    articulation: halo
    register: air
    prominence: air
    pattern: "x... .... | .x.. ...."
  guitar:
    family: guitar
    tone: [soft, warm]
    articulation: answer
    register: mid
    prominence: support
    pattern: ".x.. ...x | ..x. .x.."
  drums:
    family: drums
    tone: [dusty, tight]
    articulation: pocket
    prominence: anchor
    pattern: "x... x... | ..x. ..x."
sections:
  - id: intro
    title: folded paper
    duration: 75s
    harmony: "Cm9 F13 Bbmaj9 G7 | Cm9 Abmaj9 F13 G7"
    scene: "intro sparse hush"
    variation: "settle"
    profile:
      density: sparse
      brightness: warm
      motion: still
    roles:
      lead:
        active: false
      guitar:
        active: false
  - id: a
    title: window seat
    duration: 135s
    harmony: "Cm9 F13 Bbmaj9 G7 | Ebmaj9 D7 Gm9 C7"
    scene: "head letterwriting sway"
    variation: "introduce-hook"
    roles:
      lead:
        active: true
        motif: "9 . b9 7 | 5 . 6 5 | 3 . 2 1 | . 9 7 5"
      drums:
        pattern: "x... x..x | ..x. ..x. | x.x.x.x."
  - id: a-prime
    title: margin reply
    duration: 120s
    harmony: "Gm9 C13 Fmaj9 D7 | Cm9 F13 Bbmaj9 G7"
    scene: "lift denser pocket"
    variation: "answer-lift"
    profile:
      density: busy
      motion: moving
    roles:
      lead:
        motif: "5 . 6 7 | 9 . b9 7 | 5 - 3 1 | . 9 7 3"
      guitar:
        active: true
      drums:
        pattern: "x..x x..x | ..x. ..x. | x.x.x.x."
  - id: breakdown
    title: radiator glow
    duration: 105s
    harmony: "Abmaj9 G7 Cm9 Cm9 | Ebmaj9 D7 Gm9 C7"
    scene: "breakdown thin suspended"
    variation: "subtract"
    profile:
      density: light
      brightness: warm
      motion: still
      reverb: halo
    roles:
      lead:
        motif: "11 . 9 7 | 5 . . 3 | 1 . . . | . . . ."
      guitar:
        active: false
      keys:
        pattern: "x... .... | ..x. ...."
  - id: outro
    title: last paragraph
    duration: 90s
    harmony: "Cm9 F13 Bbmaj9 G7 | Cm9 Abmaj9 F13 G7"
    scene: "outro home fade"
    variation: "cadence"
    roles:
      lead:
        motif: "3 . 2 1 | . 9 7 5 | 3 . . 1 | . . . ."
      guitar:
        active: false


title: Soft Tape / Library Vent Loop
description: Study-floor repetition with richer turnarounds, a muted midsection, and a warmer closing cadence.
style: lofi
listen_mode: album-side
seed: 70301
tags: [lofi, library, study, repeat, nocturne]
key: Amin
tempo: 74
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
    pattern: "x..x .x.. | x... .x.."
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
    motif: "5 . . 7 | 9 . 7 5 | 3 . . 1 | . . . ."
  texture:
    family: mallet
    tone: [glass, soft]
    articulation: halo
    register: air
    prominence: air
    pattern: "x... .... | ..x. ...."
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
    title: desk lamp
    duration: 75s
    harmony: "Am9 D13 Gmaj9 E7 | Am9 Fmaj9 D13 E7"
    scene: "intro sparse study"
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
    title: highlighted page
    duration: 120s
    harmony: "Am9 D13 Gmaj9 E7 | Cmaj9 B7 Em9 A7"
    scene: "head warm loop"
    variation: "introduce-hook"
    roles:
      lead:
        active: true
        motif: "9 . b9 7 | 5 . 6 5 | 3 . 2 1 | . 9 7 5"
      drums:
        pattern: "x... x..x | ..x. ..x. | x.x.x.x."
  - id: a-prime
    title: air-duct answer
    duration: 135s
    harmony: "Em9 A13 Dmaj9 B7 | Am9 D13 Gmaj9 E7"
    scene: "lift denser reply"
    variation: "sequence-up"
    profile:
      density: busy
      brightness: balanced
      motion: moving
    roles:
      lead:
        motif: "5 . 6 7 | 9 . b9 7 | 5 - 3 1 | . 9 7 3"
      guitar:
        active: true
      drums:
        pattern: "x..x x..x | ..x. ..x. | x.x.x.x."
  - id: breakdown
    title: stack shadows
    duration: 90s
    harmony: "Fmaj9 E7 Am9 Am9 | Cmaj9 B7 Em9 A7"
    scene: "breakdown thin muted"
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
    title: last underline
    duration: 90s
    harmony: "Am9 D13 Gmaj9 E7 | Am9 Fmaj9 D13 E7"
    scene: "outro home hush"
    variation: "cadence"
    roles:
      lead:
        motif: "3 . 2 1 | . 9 7 5 | 3 . . 1 | . . . ."


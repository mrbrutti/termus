title: Library Vent Loop
description: Vibraphone-front lofi with lighter drums, sparse electric bass, and a glassy midsection.
style: lofi
listen_mode: album-side
seed: 70301
tags: [lofi, library, vibraphone, study, loop]
key: Amin
tempo: 70
globals:
  density: steady
  brightness: balanced
  motion: gentle
  reverb: room
  phrase: long
roles:
  guitar:
    family: guitar
    tone: [warm, muted]
    articulation: answer
    register: mid
    prominence: support
    pattern: "x... .... | ..x...x."
  bass:
    family: synth_bass
    tone: [direct, warm]
    articulation: legato
    register: low
    prominence: anchor
    pattern: "x....... | x...x..."
  kick:
    family: drums
    tone: [dusty, soft]
    articulation: pocket
    prominence: anchor
    pattern: "x....... | x...x..."
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
    pattern: "x...x.x. | x...x..."
  vibes:
    family: mallet
    tone: [soft, glass]
    articulation: lyrical
    register: high
    prominence: lead
    motif: "5 . 7 . 9 . 7 5 | 3 . 2 . 1 . . ."
  texture:
    family: pad
    tone: [soft, wide]
    articulation: sustain
    register: air
    prominence: air
    pattern: "........ | ....x..."
sections:
  - id: intro
    title: desk lamp
    duration: 14s
    harmony: "Am9 D13 | Gmaj9 E7 | Am9 Fmaj9 | D13 E7"
    scene: "intro study"
    variation: "establish"
    profile:
      density: sparse
      brightness: warm
      motion: still
    roles:
      vibes:
        active: true
        motif: "5 . . 7 . . 5 . | 3 . . 2 . . 1 ."
      guitar:
        pattern: "x....... | ....x..."
  - id: verse-a
    title: highlighted page
    duration: 55s
    harmony: "Am9 D13 | Gmaj9 E7 | Cmaj9 B7 | Em9 A7"
    scene: "head repeat"
    variation: "introduce-hook"
    roles:
      vibes:
        active: true
        motif: "9 . 7 . 5 . 6 5 | 3 . 2 . 1 . . ."
      hat:
        pattern: "x.x...x. | x...x.x."
    events:
      - kind: pickup
        bar: 8
        roles: [vibes]
        motif: "5 6 7 9"
      - kind: fill
        bar: 8
        roles: [hat, snare]
  - id: middle
    title: fluorescent blur
    duration: 45s
    harmony: "Fmaj9 E7 | Am9 Am9 | Cmaj9 B7 | Em9 A7"
    scene: "middle glassy"
    variation: "thin"
    profile:
      density: light
      brightness: balanced
      motion: still
      reverb: halo
    roles:
      vibes:
        motif: "5 . . . 3 . . . | 1 . . . . . . ."
      texture:
        active: true
        pattern: "x....... | ....x..."
      hat:
        pattern: "x...x... | ....x..."
    events:
      - kind: drop
        bar: 5
        bars: 1
        roles: [bass, kick]
  - id: verse-b
    title: page turn
    duration: 55s
    harmony: "Em9 A13 | Dmaj9 B7 | Am9 D13 | Gmaj9 E7"
    scene: "reply brighter"
    variation: "answer-lift"
    profile:
      density: busy
      brightness: bright
      motion: moving
    roles:
      vibes:
        motif: "11 . 9 . 7 . 9 7 | 5 . 3 . 2 . 1 ."
      kick:
        pattern: "x...x..x | x...x..."
      hat:
        pattern: "x.x.x.x. | x.xxx.x."
    events:
      - kind: stab
        bar: 9
        roles: [guitar]
        pattern: "x... ...."
      - kind: fill
        bar: 12
        roles: [hat, snare, kick]
  - id: outro
    title: air vent fade
    duration: 40s
    harmony: "Am9 D13 | Gmaj9 E7 | Am9 Fmaj9 | D13 E7"
    scene: "outro close"
    variation: "cadence"
    roles:
      vibes:
        motif: "3 . . 2 . . 1 . | . . . . . . . ."

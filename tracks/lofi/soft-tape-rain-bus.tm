title: Soft Tape / Rain Bus
description: Late-night bus ride with patient chord motion, a brighter middle, and a quiet platform return.
style: lofi
listen_mode: album-side
seed: 42017
tags: [lofi, warm, late-night, dusty, ride]
key: Dmin
tempo: 78
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
    pattern: "x..x .x.. | x.x. .x.."
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
    motif: "5 . . 9 | b9 7 . 5 | 3 . 2 1 | . . . ."
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
    register: low
    prominence: anchor
    pattern: "x... x..x | ..x. ..x."
sections:
  - id: intro
    title: curbside intro
    duration: 75s
    harmony: "Dm9 G13 Cmaj9 A7 | Dm9 Bbmaj9 G13 A7"
    scene: "intro sparse nocturne"
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
    title: aisle sway
    duration: 135s
    harmony: "Dm9 G13 Cmaj9 A7 | Bbmaj9 A7 Dm9 G13"
    scene: "head warm moving"
    variation: "introduce-hook"
    profile:
      density: steady
      motion: gentle
    roles:
      lead:
        active: true
        motif: "9 . b9 7 | 5 . 6 5 | 3 . 2 1 | . 9 7 5"
      drums:
        pattern: "x... x..x | ..x. ..x. | x.x. x.x."
  - id: a-prime
    title: tunnel answer
    duration: 120s
    harmony: "Fm9 Bb13 Ebmaj9 C7 | Dm9 G13 Cmaj9 A7"
    scene: "lift denser pulse"
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
        pattern: ".x.. .x.x | ..x. .x.."
      drums:
        pattern: "x..x x..x | ..x. ..x. | x.x.x.x."
  - id: bridge
    title: orange transfer
    duration: 120s
    harmony: "Bbmaj9 A7 Dm9 G13 | Fmaj9 Em7b5 A7 Dm9"
    scene: "bridge brighter push"
    variation: "open-register"
    profile:
      density: busy
      brightness: balanced
      motion: moving
    roles:
      lead:
        motif: "11 . 9 7 | b9 7 5 3 | 5 . 6 7 | 9 . 7 3"
      keys:
        pattern: "x.x. .x.x | x..x .x.."
      drums:
        pattern: "x..x x..x | ..x. ..x. | x.x.x.x."
  - id: outro
    title: platform return
    duration: 90s
    harmony: "Dm9 G13 Cmaj9 A7 | Dm9 Bbmaj9 G13 A7"
    scene: "outro thin home"
    variation: "cadence"
    profile:
      density: light
      brightness: warm
      motion: gentle
    roles:
      lead:
        motif: "3 . 2 1 | . 9 7 5 | 3 . . 1 | . . . ."
      guitar:
        active: false


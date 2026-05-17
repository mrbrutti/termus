title: Soft Tape / Corner Store Neon
description: Slightly brighter chords, steadier hook returns, and a late bridge that opens the pocket.
style: lofi
listen_mode: album-side
seed: 88119
tags: [lofi, neon, storefront, mellow, loop]
key: Fmin
tempo: 80
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
    pattern: "x... x..x | x... x..."
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
    pattern: "x... x..x | ..x. ..x."
sections:
  - id: intro
    title: sign hum
    duration: 60s
    harmony: "Fm9 Bb13 Ebmaj9 C7 | Fm9 Dbmaj9 Bb13 C7"
    scene: "intro sparse glow"
    variation: "settle"
    profile:
      density: light
      brightness: warm
      motion: still
    roles:
      lead:
        active: false
      guitar:
        active: false
  - id: a
    title: aisle reflection
    duration: 120s
    harmony: "Fm9 Bb13 Ebmaj9 C7 | Abmaj9 G7 Cm9 F7"
    scene: "head warm loop"
    variation: "introduce-hook"
    roles:
      lead:
        active: true
        motif: "9 . b9 7 | 5 . 6 5 | 3 . 2 1 | . 9 7 5"
  - id: a-prime
    title: freezer answer
    duration: 120s
    harmony: "Cm9 F13 Bbmaj9 G7 | Fm9 Bb13 Ebmaj9 C7"
    scene: "lift reply"
    variation: "sequence-up"
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
  - id: bridge
    title: parking lot
    duration: 105s
    harmony: "Dbmaj9 C7 Fm9 Fm9 | Abmaj9 G7 Cm9 F7"
    scene: "bridge open-air"
    variation: "open-register"
    roles:
      lead:
        motif: "11 . 9 7 | 5 . . 3 | 1 . . . | . 9 7 5"
      keys:
        pattern: "x... .x.. | x..x .x.."
  - id: outro
    title: shutters down
    duration: 75s
    harmony: "Fm9 Bb13 Ebmaj9 C7 | Fm9 Dbmaj9 Bb13 C7"
    scene: "outro fade home"
    variation: "cadence"
    roles:
      lead:
        motif: "3 . 2 1 | . 9 7 5 | 3 . . 1 | . . . ."
      guitar:
        active: false


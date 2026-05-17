title: Soft Tape / Corner Store Neon
description: Brighter storefront groove with tighter drums, flute-like hooks, and a lifted chorus section.
style: lofi
listen_mode: album-side
seed: 88119
tags: [lofi, neon, storefront, brighter, late]
key: Fmin
tempo: 82
globals:
  density: steady
  brightness: balanced
  motion: moving
  reverb: room
  phrase: long
roles:
  keys:
    family: electric_piano
    tone: [warm, soft]
    articulation: stab
    register: mid
    prominence: support
    pattern: "x.x...x. | x..x.x.."
  bass:
    family: bass
    tone: [round, woody]
    articulation: legato
    register: low
    prominence: anchor
    pattern: "x...x..x | x.x...x."
  kick:
    family: drums
    tone: [tight, dusty]
    articulation: pocket
    prominence: anchor
    pattern: "x..xx..x | x...x..x"
  snare:
    family: drums
    tone: [tight, soft]
    articulation: pocket
    prominence: anchor
    pattern: "....x... | ....x..."
  hat:
    family: drums
    tone: [dry, direct]
    articulation: pocket
    prominence: support
    pattern: "x.xxx.x. | x.x.xx.x"
  vibes:
    family: mallet
    tone: [glass, delicate]
    articulation: halo
    register: high
    prominence: air
    pattern: "..x..x.. | ....x..."
  lead:
    family: reed_lead
    tone: [present, intimate]
    articulation: lyrical
    register: high
    prominence: lead
    motif: "5 . 7 . 9 . 11 9 | 7 . 5 . 3 . 2 1"
  pad:
    family: pad
    tone: [soft, wide]
    articulation: sustain
    register: mid
    prominence: support
    pattern: "x....... | ....x..."
sections:
  - id: intro
    title: cooler hum
    duration: 35s
    harmony: "Fm9 Bb13 | Ebmaj9 C7 | Fm9 Dbmaj9 | Bb13 C7"
    scene: "intro glow"
    variation: "establish"
    profile:
      density: light
      brightness: warm
      motion: still
    roles:
      lead:
        active: false
      hat:
        pattern: "x...x... | x...x..."
  - id: verse-a
    title: neon aisle
    duration: 55s
    harmony: "Fm9 Bb13 | Ebmaj9 C7 | Abmaj9 G7 | Cm9 F7"
    scene: "head pulse"
    variation: "introduce-hook"
    roles:
      lead:
        motif: "9 . b9 . 7 . 5 . | 6 . 5 . 3 . 2 1"
  - id: chorus
    title: street reflection
    duration: 55s
    harmony: "Abmaj9 G7 | Cm9 F7 | Fm9 Bb13 | Ebmaj9 C7"
    scene: "chorus lift"
    variation: "open-register"
    profile:
      density: busy
      brightness: bright
      motion: moving
    roles:
      lead:
        motif: "11 . 9 . 7 . 9 11 | 7 . 5 . 3 . 2 1"
      pad:
        active: true
      hat:
        pattern: "xxxxxxxx | x.xxx.xx"
  - id: break
    title: cash drawer
    duration: 40s
    harmony: "Dbmaj9 C7 | Fm9 Fm9 | Abmaj9 G7 | Cm9 F7"
    scene: "break thin"
    variation: "subtract"
    profile:
      density: sparse
      brightness: warm
      motion: still
    roles:
      lead:
        motif: "5 . . . 3 . . . | 1 . . . . . . ."
      snare:
        pattern: "........ | ....x..."
      pad:
        active: false
  - id: outro
    title: shutters down
    duration: 40s
    harmony: "Fm9 Bb13 | Ebmaj9 C7 | Fm9 Dbmaj9 | Bb13 C7"
    scene: "outro settle"
    variation: "cadence"
    roles:
      lead:
        motif: "3 . 2 . 1 . . . | . . . . . . . ."

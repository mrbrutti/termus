title: Corner Store Neon
description: Flute-hook lofi with a steadier low end, brighter chorus, and more obvious storefront pulse.
style: lofi
listen_mode: album-side
seed: 88119
tags: [lofi, neon, storefront, flute, pulse]
key: Fmin
tempo: 82
globals:
  density: steady
  brightness: balanced
  motion: moving
  reverb: room
  phrase: long
roles:
  ep:
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
  flute:
    family: woodwind
    tone: [soft, breathy]
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
  vibes:
    family: mallet
    tone: [glass, delicate]
    articulation: halo
    register: air
    prominence: air
    pattern: "....x... | ..x....."
sections:
  - id: intro
    title: cooler hum
    duration: 12s
    harmony: "Fm9 Bb13 | Ebmaj9 C7 | Fm9 Dbmaj9 | Bb13 C7"
    scene: "intro glow"
    variation: "establish"
    profile:
      density: light
      brightness: warm
      motion: still
    roles:
      flute:
        active: true
        motif: "9 . . . 7 . . . | 5 . . 3 . . 2 1"
      pad:
        active: false
      ep:
        pattern: "x....... | x..x...."
  - id: verse-a
    title: neon aisle
    duration: 55s
    harmony: "Fm9 Bb13 | Ebmaj9 C7 | Abmaj9 G7 | Cm9 F7"
    scene: "head pulse"
    variation: "introduce-hook"
    roles:
      flute:
        active: true
        motif: "9 . 7 . 5 . 6 5 | 3 . 2 . 1 . . ."
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
      flute:
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
      flute:
        motif: "5 . . . 3 . . . | 1 . . . . . . ."
      pad:
        active: false
      hat:
        pattern: "x...x... | ....x..."
  - id: outro
    title: shutters down
    duration: 40s
    harmony: "Fm9 Bb13 | Ebmaj9 C7 | Fm9 Dbmaj9 | Bb13 C7"
    scene: "outro settle"
    variation: "cadence"
    roles:
      flute:
        motif: "3 . . 2 . . 1 . | . . . . . . . ."

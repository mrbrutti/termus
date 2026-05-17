title: Soft Tape / Rain Bus
description: Rhodes-led late bus study with a patient hook, lifted bridge, and a quiet curbside return.
style: lofi
listen_mode: album-side
seed: 42017
tags: [lofi, rain, bus, dusty, late-night]
key: Dmin
tempo: 78
globals:
  density: steady
  brightness: warm
  motion: gentle
  reverb: room
  phrase: long
roles:
  rhodes:
    family: electric_piano
    tone: [warm, dusty, soft]
    articulation: stab
    register: mid
    prominence: support
    pattern: "x..x.x.. | x.x...x."
  bass:
    family: bass
    tone: [round, woody]
    articulation: legato
    register: low
    prominence: anchor
    pattern: "x...x..x | x.x...x."
  kick:
    family: drums
    tone: [dusty, soft]
    articulation: pocket
    prominence: anchor
    pattern: "x...x..x | x...x..."
  snare:
    family: drums
    tone: [dusty, soft]
    articulation: pocket
    prominence: anchor
    pattern: "....x... | ....x..."
  hat:
    family: drums
    tone: [dry, tight]
    articulation: pocket
    prominence: support
    pattern: "x.x.x.x. | x.xxx.x."
  vibes:
    family: mallet
    tone: [glass, soft]
    articulation: halo
    register: high
    prominence: air
    pattern: "..x..... | ....x..."
  guitar:
    family: guitar
    tone: [soft, warm]
    articulation: answer
    register: mid
    prominence: support
    pattern: ".x..x..x | ..x...x."
  lead:
    family: reed_lead
    tone: [breathy, intimate]
    articulation: lyrical
    register: mid-high
    prominence: lead
    motif: "5 . 6 . 7 . 9 7 | 5 . 3 . 2 . 1 ."
  pad:
    family: pad
    tone: [soft, wide]
    articulation: sustain
    register: mid
    prominence: air
    pattern: "x....... | ....x..."
sections:
  - id: intro
    title: windshield beads
    duration: 16s
    harmony: "Dm9 G13 | Cmaj9 A7 | Dm9 Bbmaj9 | G13 A7"
    scene: "intro hush"
    variation: "establish"
    profile:
      density: light
      brightness: warm
      motion: still
    roles:
      lead:
        active: true
        motif: "5 . . . 7 . . 5 | . . 3 . 2 . 1 ."
      guitar:
        active: true
        pattern: ".x...... | ....x..."
      snare:
        pattern: "........ | ....x..."
      hat:
        pattern: "x...x... | x...x..."
  - id: verse-a
    title: aisle sway
    duration: 60s
    harmony: "Dm9 G13 | Cmaj9 A7 | Bbmaj9 A7 | Dm9 G13"
    scene: "head glide"
    variation: "introduce-hook"
    roles:
      lead:
        motif: "9 . b9 . 7 . 5 . | 6 . 5 . 3 . 2 1"
      vibes:
        pattern: "..x...x. | ....x..."
  - id: verse-b
    title: tunnel orange
    duration: 60s
    harmony: "Fm9 Bb13 | Ebmaj9 C7 | Dm9 G13 | Cmaj9 A7"
    scene: "answer brighter"
    variation: "sequence-up"
    profile:
      density: busy
      brightness: balanced
      motion: moving
    roles:
      lead:
        motif: "5 . 6 7 9 . b9 7 | 5 . 3 . 2 . 1 ."
      guitar:
        active: true
        pattern: ".x..x..x | ..x.x..."
      kick:
        pattern: "x..xx..x | x...x..x"
  - id: bridge
    title: transfer glow
    duration: 50s
    harmony: "Bbmaj9 A7 | Dm9 G13 | Fmaj9 Em7b5 | A7 Dm9"
    scene: "bridge lift"
    variation: "open-register"
    profile:
      density: busy
      brightness: bright
      motion: moving
    roles:
      lead:
        motif: "11 . 9 . 7 . 5 3 | 5 . 6 . 7 . 9 ."
      rhodes:
        pattern: "x.x.x..x | x..x.x.."
      hat:
        pattern: "xxxxxxxx | x.xxx.xx"
      pad:
        active: true
  - id: breakdown
    title: red light idle
    duration: 40s
    harmony: "Dm9 Dm9 | Bbmaj9 A7 | Dm9 G13 | A7 A7"
    scene: "breakdown thin"
    variation: "subtract"
    profile:
      density: sparse
      brightness: warm
      motion: still
      reverb: halo
    roles:
      lead:
        motif: "3 . . . 2 . . . | 1 . . . . . . ."
      guitar:
        active: false
      kick:
        pattern: "x....... | x...x..."
      hat:
        pattern: "x...x... | ....x..."
      vibes:
        pattern: "....x... | ........"
  - id: outro
    title: curbside return
    duration: 45s
    harmony: "Dm9 G13 | Cmaj9 A7 | Dm9 Bbmaj9 | G13 A7"
    scene: "outro home"
    variation: "cadence"
    roles:
      lead:
        motif: "3 . 2 . 1 . . . | . . . . . . . ."
      guitar:
        active: false
      pad:
        active: false

title: Soft Tape / Walkman Streetlights
description: Heavier kick pocket, brighter answer phrases, and a night-drive bridge.
style: lofi
listen_mode: album-side
seed: 64211
tags: [lofi, walkman, night-drive, tape, beat]
key: Emin
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
    tone: [warm, dusty, soft]
    articulation: stab
    register: mid
    prominence: support
    pattern: "x..x .x.. | x.x. .x.x"
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
    motif: "5 . . 7 | 9 . 7 5 | 3 . . 1 | . . . ."
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
    pattern: "x..x x..x | ..x. ..x."
sections:
  - id: intro
    title: tape click
    duration: 60s
    harmony: "Em9 A13 Dmaj9 B7 | Em9 Cmaj9 A13 B7"
    scene: "intro pulse shadow"
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
    title: crosswalk loop
    duration: 135s
    harmony: "Em9 A13 Dmaj9 B7 | Gmaj9 F#7 Bm9 E7"
    scene: "head drive"
    variation: "introduce-hook"
    roles:
      lead:
        active: true
        motif: "9 . b9 7 | 5 . 6 5 | 3 . 2 1 | . 9 7 5"
      drums:
        pattern: "x..x x..x | ..x. ..x. | x.x.x.x."
  - id: bridge
    title: underpass chorus
    duration: 150s
    harmony: "Bm9 E13 Amaj9 F#7 | Gmaj9 A13 Dmaj9 B7"
    scene: "bridge brighter push"
    variation: "lift-register"
    profile:
      density: busy
      brightness: bright
      motion: moving
    roles:
      lead:
        motif: "11 . 9 7 | #9 7 5 3 | 5 . 6 7 | 9 . 7 3"
      guitar:
        active: true
      keys:
        pattern: "x.x. .x.x | x..x .x.."
  - id: breakdown
    title: parking-lot air
    duration: 90s
    harmony: "Cmaj9 B7 Em9 Em9 | Gmaj9 F#7 Bm9 E7"
    scene: "breakdown thin nightair"
    variation: "subtract"
    profile:
      density: light
      brightness: warm
      motion: gentle
      reverb: halo
    roles:
      lead:
        motif: "9 . 7 5 | 3 . . 1 | . . . . | . . . ."
      guitar:
        active: false
  - id: outro
    title: headphones off
    duration: 90s
    harmony: "Em9 A13 Dmaj9 B7 | Em9 Cmaj9 A13 B7"
    scene: "outro home fade"
    variation: "cadence"
    roles:
      lead:
        motif: "3 . 2 1 | . 9 7 5 | 3 . . 1 | . . . ."


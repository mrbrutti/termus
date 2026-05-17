title: Soft Tape / Walkman Streetlights
description: Night-drive tape beat with a stronger low end, sharper hooks, and a brighter freeway chorus.
style: lofi
listen_mode: album-side
seed: 64211
tags: [lofi, walkman, streetlights, drive, tape]
key: Emin
tempo: 84
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
    pattern: "x.x...x. | x..x.x.."
  bass:
    family: bass
    tone: [round, woody]
    articulation: legato
    register: low
    prominence: anchor
    pattern: "x..xx..x | x...x..x"
  kick:
    family: drums
    tone: [tight, direct]
    articulation: pocket
    prominence: anchor
    pattern: "x..xx..x | x..xx..."
  snare:
    family: drums
    tone: [tight, dry]
    articulation: pocket
    prominence: anchor
    pattern: "....x... | ....x..."
  hat:
    family: drums
    tone: [dry, tight]
    articulation: pocket
    prominence: support
    pattern: "x.xxx.x. | x.xxxxxx"
  guitar:
    family: guitar
    tone: [soft, warm]
    articulation: answer
    register: mid
    prominence: support
    pattern: ".x..x..x | ..x..x.."
  lead:
    family: reed_lead
    tone: [present, intimate]
    articulation: lyrical
    register: high
    prominence: lead
    motif: "5 . 7 . 9 . 11 9 | 7 . 5 . 3 . 2 1"
  pad:
    family: pad
    tone: [wide, soft]
    articulation: sustain
    register: mid
    prominence: support
    pattern: "x....... | ....x..."
sections:
  - id: intro
    title: tape click
    duration: 35s
    harmony: "Em9 A13 | Dmaj9 B7 | Em9 Cmaj9 | A13 B7"
    scene: "intro pulse"
    variation: "establish"
    profile:
      density: light
      brightness: warm
      motion: still
    roles:
      lead:
        active: false
      guitar:
        active: false
  - id: verse-a
    title: crosswalk loop
    duration: 50s
    harmony: "Em9 A13 | Dmaj9 B7 | Gmaj9 F#7 | Bm9 E7"
    scene: "head drive"
    variation: "introduce-hook"
    roles:
      lead:
        motif: "9 . 7 . 5 . 6 5 | 3 . 2 . 1 . . ."
  - id: chorus
    title: freeway lift
    duration: 55s
    harmony: "Bm9 E13 | Amaj9 F#7 | Gmaj9 A13 | Dmaj9 B7"
    scene: "chorus brighter"
    variation: "lift-register"
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
    title: overpass air
    duration: 40s
    harmony: "Cmaj9 B7 | Em9 Em9 | Gmaj9 F#7 | Bm9 E7"
    scene: "breakdown thin"
    variation: "subtract"
    profile:
      density: sparse
      brightness: warm
      motion: gentle
      reverb: halo
    roles:
      lead:
        motif: "5 . . . 3 . . . | 1 . . . . . . ."
      guitar:
        active: false
      pad:
        active: false
      hat:
        pattern: "x...x... | ....x..."
  - id: outro
    title: headphones off
    duration: 40s
    harmony: "Em9 A13 | Dmaj9 B7 | Em9 Cmaj9 | A13 B7"
    scene: "outro settle"
    variation: "cadence"
    roles:
      lead:
        motif: "3 . 2 . 1 . . . | . . . . . . . ."

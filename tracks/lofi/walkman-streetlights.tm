title: Walkman Streetlights
description: More driving cassette beat with guitar lead, synth bass, and a wider freeway chorus than the rest of the pack.
style: lofi
listen_mode: album-side
seed: 64211
tags: [lofi, walkman, streetlights, guitar, drive]
key: Emin
tempo: 84
globals:
  density: steady
  brightness: balanced
  motion: moving
  reverb: room
  phrase: long
roles:
  ep:
    family: electric_piano
    tone: [warm, dusty, soft]
    articulation: stab
    register: mid
    prominence: support
    pattern: "x.x...x. | x..x.x.."
  sub:
    family: synth_bass
    tone: [direct, warm]
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
  guitar-lead:
    family: guitar
    tone: [soft, warm]
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
  vibes:
    family: mallet
    tone: [soft, glass]
    articulation: halo
    register: air
    prominence: air
    pattern: "..x..... | ....x..."
sections:
  - id: intro
    title: tape click
    duration: 12s
    harmony: "Em9 A13 | Dmaj9 B7 | Em9 Cmaj9 | A13 B7"
    scene: "intro pulse"
    variation: "establish"
    profile:
      density: light
      brightness: warm
      motion: still
    roles:
      guitar-lead:
        active: true
        motif: "5 . . 7 . 9 . 7 | 5 . . 3 . 2 . 1"
      pad:
        active: false
      ep:
        pattern: "x.x..... | x...x..."
  - id: verse-a
    title: crosswalk loop
    duration: 50s
    harmony: "Em9 A13 | Dmaj9 B7 | Gmaj9 F#7 | Bm9 E7"
    scene: "head drive"
    variation: "introduce-hook"
    roles:
      guitar-lead:
        active: true
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
      guitar-lead:
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
      guitar-lead:
        motif: "5 . . . 3 . . . | 1 . . . . . . ."
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
      guitar-lead:
        motif: "3 . . 2 . . 1 . | . . . . . . . ."

title: Chamber Loop / Lantern Room
description: Piano-led chamber miniature with winds in the answer phrase, a dim middle chamber, and a resolved final room.
style: classical
listen_mode: album-side
seed: 26031
tags: [classical, chamber, piano, lantern, room]
key: Gminor
tempo: 90
globals:
  density: steady
  brightness: balanced
  motion: gentle
  phrase: long
roles:
  piano:
    family: acoustic_piano
    tone: [clear, present]
    articulation: legato
    register: mid
    prominence: lead
    motif: "5 . 6 . 5 . 3 2 | 1 . . . 2 . 3 ."
  strings:
    family: strings
    tone: [lush, soft]
    articulation: sustain
    register: mid-high
    prominence: support
    pattern: "x....... | ....x..."
  winds:
    family: woodwind
    tone: [soft, airy]
    articulation: answer
    register: high
    prominence: support
    pattern: ".x....x. | ..x....."
  brass:
    family: brass
    tone: [rich, soft]
    articulation: swell
    register: mid
    prominence: support
    pattern: "........ | ....x..."
  harp:
    family: strings
    tone: [soft, airy]
    articulation: answer
    register: high
    prominence: air
    pattern: "..x..... | ....x..."
  choir:
    family: choir
    tone: [airy, soft]
    articulation: sustain
    register: high
    prominence: air
    pattern: "........ | x......."
sections:
  - id: threshold
    title: threshold
    duration: 40s
    harmony: "Gm9 Ebmaj9 | Fmaj9 D7 | Gm9 Cm9 | D7 Gm9"
    scene: "intro chamber"
    variation: "establish"
    roles:
      brass:
        active: false
      choir:
        active: false
  - id: theme
    title: lantern theme
    duration: 55s
    harmony: "Gm9 Ebmaj9 | Fmaj9 D7 | Gm9 Cm9 | D7 Gm9"
    scene: "theme full-room"
    variation: "statement"
    roles:
      piano:
        motif: "5 . 6 7 9 . 7 5 | 3 . 2 . 1 . 2 3"
      winds:
        pattern: ".x..x... | ..x....."
    events:
      - kind: pickup
        bar: 8
        roles: [piano]
        motif: "3 5 6 9"
      - kind: stop
        bar: 10
        bars: 1
        roles: [piano, strings, harp]
  - id: interior
    title: interior room
    duration: 45s
    harmony: "Cm9 Gm9 | Ebmaj9 D7 | Gm9 Cm9 | D7 D7"
    scene: "interior thin"
    variation: "subtract"
    profile:
      density: light
      brightness: warm
    roles:
      piano:
        motif: "5 . . . 3 . . . | 1 . . . . . . ."
      brass:
        active: false
      choir:
        active: false
      winds:
        pattern: "........ | ..x....."
    events:
      - kind: drop
        bar: 6
        bars: 1
        roles: [strings]
  - id: gallery
    title: upper gallery
    duration: 50s
    harmony: "Ebmaj9 D7 | Gm9 Cm9 | Fmaj9 D7 | Gm9 Gm9"
    scene: "gallery lift"
    variation: "lift"
    profile:
      density: busy
      brightness: bright
    roles:
      brass:
        active: true
      choir:
        active: true
      piano:
        motif: "11 . 9 7 5 . 6 7 | 9 . 7 5 3 . 2 1"
    events:
      - kind: pickup
        bar: 9
        roles: [piano]
        motif: "5 6 7 11"
  - id: close
    title: lamp out
    duration: 40s
    harmony: "Gm9 Ebmaj9 | Fmaj9 D7 | Gm9 D7 | Gm9 Gm9"
    scene: "outro resolve"
    variation: "cadence"
    roles:
      piano:
        motif: "3 . 2 . 1 . . . | . . . . . . . ."

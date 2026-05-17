title: Chamber Loop / Stairwell Etude
description: Quicker chamber etude with harp filigree, suspended inner section, and a brass-touched coda.
style: classical
listen_mode: album-side
seed: 18290
tags: [classical, etude, stairwell, harp, chamber]
key: Dmajor
tempo: 96
globals:
  density: steady
  brightness: balanced
  motion: moving
  phrase: long
roles:
  piano:
    family: acoustic_piano
    tone: [clear, present]
    articulation: legato
    register: mid
    prominence: lead
    motif: "5 . 6 . 7 . 9 7 | 5 . 3 . 2 . 1 ."
  strings:
    family: strings
    tone: [lush, soft]
    articulation: sustain
    register: high
    prominence: support
    pattern: "x....... | ....x..."
  winds:
    family: woodwind
    tone: [soft, airy]
    articulation: answer
    register: high
    prominence: support
    pattern: "..x..... | ....x..."
  brass:
    family: brass
    tone: [rich, warm]
    articulation: swell
    register: mid
    prominence: support
    pattern: "........ | ....x..."
  harp:
    family: strings
    tone: [airy, soft]
    articulation: answer
    register: high
    prominence: air
    pattern: ".x..x... | ..x....."
sections:
  - id: opening
    title: stairwell opening
    duration: 35s
    harmony: "Dmaj9 A/C# | Bm9 Gmaj9 | Dmaj9 F#m7 | Em9 A7"
    scene: "entry bright"
    variation: "establish"
    roles:
      brass:
        active: false
  - id: running-theme
    title: running theme
    duration: 50s
    harmony: "Dmaj9 A/C# | Bm9 Gmaj9 | Em9 A7 | Dmaj9 Dmaj9"
    scene: "theme active"
    variation: "statement"
    roles:
      piano:
        motif: "9 . 7 5 6 . 5 3 | 5 . 2 . 1 . 2 3"
  - id: middle
    title: suspended landing
    duration: 45s
    harmony: "Bm9 Gmaj9 | Dmaj9 F#m7 | Em9 A7 | Em9 A7"
    scene: "middle suspended"
    variation: "thin"
    profile:
      density: light
      brightness: warm
    roles:
      piano:
        motif: "5 . . . 3 . . . | 1 . . . . . . ."
      strings:
        pattern: "x....... | x......."
      harp:
        pattern: "x....... | ....x..."
  - id: coda
    title: brass coda
    duration: 40s
    harmony: "Dmaj9 A/C# | Bm9 Gmaj9 | Em9 A7 | Dmaj9 Dmaj9"
    scene: "coda swell"
    variation: "cadence"
    roles:
      brass:
        active: true
      piano:
        motif: "3 . 2 . 1 . . . | . . . . . . . ."

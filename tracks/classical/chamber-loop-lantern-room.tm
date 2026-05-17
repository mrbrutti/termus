title: Chamber Loop / Lantern Room
description: Piano-led chamber study with a denser second statement, a quiet interior room, and a resolving close.
style: classical
listen_mode: album-side
seed: 26031
tags: [classical, chamber, piano, lantern]
key: Gminor
tempo: 92
globals:
  density: steady
  brightness: balanced
  motion: gentle
  phrase: long
roles:
  piano:
    family: acoustic_piano
    tone: [clear]
    articulation: legato
    register: mid
    prominence: lead
    motif: "5 . 6 5 | 3 . 2 1"
  strings:
    family: strings
    tone: [lush]
    articulation: sustain
    register: mid-high
    prominence: support
    pattern: "x... .... | x... ...."
  winds:
    family: woodwind
    tone: [soft]
    articulation: answer
    register: mid-high
    prominence: support
    pattern: ".x.. .... | ..x. ...."
  brass:
    family: brass
    tone: [rich]
    articulation: swell
    register: mid
    prominence: support
    pattern: ".... x... | .... ...."
sections:
  - id: intro
    title: threshold
    duration: 120s
    harmony: "Gm9 Ebmaj9 | Fmaj9 D7"
    scene: "intro chamber"
    variation: "establish"
    roles:
      brass:
        active: false
  - id: a
    title: lantern theme
    duration: 180s
    harmony: "Gm9 Ebmaj9 | Fmaj9 D7 | Gm9 Cm9 | D7 Gm9"
    scene: "theme full-room"
    variation: "statement"
    roles:
      piano:
        motif: "5 . 6 5 | 3 . 2 1 | 5 . 6 7 | 9 . 7 3"
  - id: interior
    title: interior room
    duration: 135s
    harmony: "Cm9 Gm9 | Ebmaj9 D7"
    scene: "interior thin"
    variation: "subtract"
    profile:
      density: light
      brightness: warm
    roles:
      strings:
        pattern: "x... .... | .... x..."
      brass:
        active: false
  - id: outro
    title: lamp out
    duration: 120s
    harmony: "Gm9 Ebmaj9 | Fmaj9 D7 | Gm9 D7 | Gm9 Gm9"
    scene: "outro resolve"
    variation: "cadence"
    roles:
      piano:
        motif: "3 . 2 1 | 1 . . ."
      brass:
        active: false


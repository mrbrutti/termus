title: Soft Tape / Walkman Streetlights
description: Heavier kick pocket, brighter answer phrases, and a night-drive bridge.
listen_mode: album-side
seed: 64211
tags: [lofi, walkman, night-drive, tape, beat]
globals:
  density: steady
  brightness: balanced
  motion: moving
  reverb: room
  phrase: long
sections:
  - title: tape click
    algo: lofi
    duration: 60s
    profile:
      density: light
      brightness: warm
      motion: still
    audit:
      form: "intro:8"
      harmony: "Em9 A13 Dmaj9 B7 | Em9 Cmaj9 A13 B7"
      lead: "5 . . 7 | 9 . 7 5 | 3 . . 1 | . . . ."
      comp: "x . . x | . x . x | x . . x | . x . ."
      drums: "bd: x... x..x | sd: ..x. ..x. | hh: x.x. x.x."
      arrange: "bass drums comp texture"
  - title: crosswalk loop
    algo: lofi
    duration: 2m15s
    profile:
      density: steady
      brightness: balanced
      motion: moving
    audit:
      form: "a:16"
      harmony: "Em9 A13 Dmaj9 B7 | Gmaj9 F#7 Bm9 E7"
      lead: "9 . b9 7 | 5 . 6 5 | 3 . 2 1 | . 9 7 5"
      comp: "x . x . | . x . x | x . x . | . x x ."
      drums: "bd: x..x x..x | sd: ..x. ..x. | hh: x.x.x.x. | fill: .... ...x"
      arrange: "bass drums comp texture +lead"
  - title: underpass chorus
    algo: lofi
    duration: 2m30s
    profile:
      density: busy
      brightness: bright
      motion: moving
      phrase: long
    audit:
      form: "b:16"
      harmony: "Bm9 E13 Amaj9 F#7 | Gmaj9 A13 Dmaj9 B7"
      lead: "11 . 9 7 | #9 7 5 3 | 5 . 6 7 | 9 . 7 3"
      comp: "x . . x | . x x . | x . x . | . x . x"
      drums: "bd: x..x x..x | sd: ..x. ..x. | hh: x.x.x.x. | fill: ..x. ..xx"
      arrange: "bass drums comp texture +lead"
  - title: parking-lot air
    algo: lofi
    duration: 90s
    profile:
      density: light
      brightness: warm
      motion: gentle
      reverb: halo
    audit:
      form: "breakdown:8"
      harmony: "Cmaj9 B7 Em9 Em9 | Gmaj9 F#7 Bm9 E7"
      lead: "9 . 7 5 | 3 . . 1 | . . . . | . . . ."
      comp: "x . . . | . x . . | x . . x | . . . x"
      drums: "bd: x... x... | sd: ..x. ..x. | hh: x... x..."
      arrange: "bass drums texture"
  - title: headphones off
    algo: lofi
    duration: 90s
    profile:
      density: steady
      brightness: warm
      motion: gentle
    audit:
      form: "cadence:8 outro:8"
      harmony: "Em9 A13 Dmaj9 B7 | Em9 Cmaj9 A13 B7"
      lead: "3 . 2 1 | . 9 7 5 | 3 . . 1 | . . . ."
      comp: "x . . x | . x . . | x . . x | . x . ."
      drums: "bd: x... x... | sd: ..x. ..x. | hh: x.x. x.x."
      arrange: "bass drums comp texture"

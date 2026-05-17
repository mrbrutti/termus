title: Soft Tape / Rain Bus
description: Late-night lo-fi ride with a suspended bridge, denser pocket, and a quiet platform return.
listen_mode: album-side
seed: 42017
tags: [lofi, warm, late-night, dusty, ride]
globals:
  density: steady
  brightness: warm
  motion: gentle
  reverb: room
  phrase: long
sections:
  - title: curbside intro
    algo: lofi
    duration: 90s
    profile:
      density: sparse
      brightness: warm
      motion: still
      phrase: natural
    audit:
      form: "intro:8"
      harmony: "Dm9 G13 Cmaj9 A7 | Dm9 Bbmaj9 G13 A7"
      lead: "5 . . 9 | b9 7 . 5 | 3 . 2 1 | . . . ."
      comp: "x . . x | . x . x | x . x . | . x . ."
      drums: "bd: x... x..x | sd: ..x. ..x. | hh: x.x. x.x."
      arrange: "bass drums comp texture"
  - title: aisle sway
    algo: lofi
    duration: 2m30s
    profile:
      density: steady
      brightness: warm
      motion: gentle
      phrase: long
    audit:
      form: "a:16"
      harmony: "Dm9 G13 Cmaj9 A7 | Bbmaj9 A7 Dm9 G13"
      lead: "9 . b9 7 | 5 . 6 5 | 3 . 2 1 | . 9 7 5"
      comp: "x . x . | . x . x | x . . x | . x x ."
      drums: "bd: x... x..x | sd: ..x. ..x. | hh: x.x.x.x. | fill: .... ...x"
      arrange: "bass drums comp texture +lead"
  - title: tunnel answer
    algo: lofi
    duration: 2m
    profile:
      density: busy
      brightness: balanced
      motion: moving
      phrase: long
    audit:
      form: "a_prime:16"
      harmony: "Fm9 Bb13 Ebmaj9 C7 | Dm9 G13 Cmaj9 A7"
      lead: "5 . 6 7 | 9 . b9 7 | 5 - 3 1 | . 9 7 3"
      comp: "x . x . | x . . x | . x . x | x . x ."
      drums: "bd: x..x x..x | sd: ..x. ..x. | hh: x.x.x.x. | fill: ..x. ...x"
      arrange: "bass drums comp texture +lead"
  - title: orange transfer
    algo: lofi
    duration: 2m
    profile:
      density: busy
      brightness: balanced
      motion: moving
      reverb: room
      phrase: long
    audit:
      form: "b:16"
      harmony: "Bbmaj9 A7 Dm9 G13 | Fmaj9 Em7b5 A7 Dm9"
      lead: "11 . 9 7 | b9 7 5 3 | 5 . 6 7 | 9 . 7 3"
      comp: "x . . x | . x x . | x . x . | . x . x"
      drums: "bd: x..x x..x | sd: ..x. ..x. | hh: x.x.x.x. | fill: ..x. ..xx"
      arrange: "bass drums comp texture +lead"
  - title: platform return
    algo: lofi
    duration: 90s
    profile:
      density: light
      brightness: warm
      motion: gentle
      phrase: natural
    audit:
      form: "cadence:8 outro:8"
      harmony: "Dm9 G13 Cmaj9 A7 | Dm9 Bbmaj9 G13 A7"
      lead: "3 . 2 1 | . 9 7 5 | 3 . . 1 | . . . ."
      comp: "x . . x | . x . . | x . . x | . x . ."
      drums: "bd: x... x... | sd: ..x. ..x. | hh: x.x. x.x."
      arrange: "bass drums comp texture"

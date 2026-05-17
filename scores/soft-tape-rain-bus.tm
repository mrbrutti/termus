title: Soft Tape / Rain Bus
description: Late-night lo-fi study with a brighter bridge and a sparse return.
listen_mode: album-side
seed: 42017
tags: [lofi, warm, late-night, dusty]
globals:
  density: steady
  brightness: warm
  motion: gentle
  reverb: room
  phrase: natural
sections:
  - title: curbside intro
    algo: lofi
    duration: 2m
    profile:
      density: sparse
      brightness: warm
      motion: still
    audit:
      form: "intro:8"
      harmony: "Dm9 G13 | Cmaj9 Am7"
      lead: "5 . 6 5 | 3 . 2 1"
      comp: "x . . x | . x . ."
      drums: "bd: x... x..x | sd: ..x. ..x."
      arrange: "bass drums comp"
  - title: rain loop
    algo: lofi
    duration: 3m30s
    profile:
      density: steady
      motion: gentle
      phrase: long
    audit:
      form: "a:16"
      harmony: "Dm9 G13 | Cmaj9 Am7"
      lead: "9 . 7 5 | 3 . 2 1"
      comp: "x . x . | . x . x"
      drums: "bd: x... x..x | sd: ..x. ..x. | hh: x.x. x.x."
      arrange: "bass drums comp +lead"
  - title: orange signal
    algo: lofi
    duration: 2m30s
    profile:
      density: busy
      brightness: balanced
      motion: moving
      phrase: long
    audit:
      form: "b:16"
      harmony: "Fm9 Bb13 | Ebmaj9 C7"
      lead: "5 . 6 7 | 9 . 7 3"
      comp: "x . x . | x . . x"
      drums: "bd: x..x x..x | sd: ..x. ..x. | hh: x.x.x.x."
      arrange: "bass drums comp +lead +texture"
  - title: platform return
    algo: lofi
    duration: 2m
    profile:
      density: light
      brightness: warm
      motion: gentle
    audit:
      form: "outro:8"
      harmony: "Dm9 G13 | Cmaj9 Am7"
      lead: "3 . 2 1 | 1 . . ."
      comp: "x . . x | . x . ."
      drums: "bd: x... x... | sd: ..x. ..x."
      arrange: "bass drums comp"


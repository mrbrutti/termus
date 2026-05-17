title: Glass Chapel / Vespers
description: Sparse bell-lit movements with a restrained middle and luminous cadence.
listen_mode: album-side
seed: 8801
tags: [bells, glass, sacred, sparse]
globals:
  density: sparse
  brightness: bright
  motion: gentle
  reverb: halo
  phrase: long
sections:
  - title: first light
    algo: bells
    duration: 2m30s
    profile:
      density: sparse
      brightness: bright
      motion: still
    audit:
      form: "intro:8 a:8"
      harmony: "i7 bVIImaj7 | iv7 i7"
      lead: "5 . . 7 | 9 . 7 5"
      comp: "x . . . | . x . ."
      arrange: "texture bass"
  - title: nave echo
    algo: bells
    duration: 3m
    profile:
      density: light
      brightness: balanced
      motion: gentle
    audit:
      form: "a_prime:16"
      harmony: "i7 bVIImaj7 | iv7 i7"
      lead: "9 . 7 5 | 3 . 2 1"
      comp: "x . . x | . x . ."
      arrange: "texture +lead bass"
  - title: stone halo
    algo: bells
    duration: 2m
    profile:
      density: sparse
      brightness: warm
      motion: still
    audit:
      form: "breakdown:8"
      harmony: "iv7 i7 | bVIImaj7 i7"
      lead: "5 . . . | 3 . . ."
      comp: "x . . . | . . . x"
      arrange: "texture"
  - title: vesper close
    algo: bells
    duration: 2m30s
    profile:
      density: light
      brightness: bright
      motion: gentle
    audit:
      form: "cadence:8 outro:8"
      harmony: "i7 bVIImaj7 | iv7 i7"
      lead: "3 . 2 1 | 1 . . ."
      comp: "x . . x | . x . ."
      arrange: "texture +lead bass"


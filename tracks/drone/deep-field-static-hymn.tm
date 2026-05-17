title: Deep Field / Static Hymn
description: Long drone arc that thins into a darker center and comes back wider.
style: drone
listen_mode: hour-stream
seed: 64003
tags: [drone, slow, field, hymn]
key: Dminor
tempo: 46
globals:
  density: light
  brightness: warm
  motion: still
  reverb: halo
  phrase: long
roles:
  bed:
    family: pad
    tone: [wide, soft]
    articulation: sustain
    register: mid
    prominence: support
    pattern: "x... .... | x... ...."
  strings:
    family: strings
    tone: [soft, floating]
    articulation: sustain
    register: mid-high
    prominence: support
    pattern: ".... x... | .... x..."
  choir:
    family: choir
    tone: [airy]
    articulation: sustain
    register: high
    prominence: air
    pattern: "x... .... | .... x..."
  shimmer:
    family: lead
    tone: [icy, shimmer]
    articulation: bloom
    register: air
    prominence: air
    pattern: ".... ..x. | .... x..."
  bass:
    family: synth_bass
    tone: [warm]
    articulation: sustain
    register: sub
    prominence: anchor
    pattern: "x... .... | x... ...."
sections:
  - id: establish
    title: antenna glow
    duration: 210s
    harmony: "Dm11 Bbmaj9 | Fmaj9 Cmaj9"
    scene: "establish wide"
    variation: "settle"
  - id: shadow
    title: cloud cover
    duration: 180s
    harmony: "Dm11 Cmaj9 | Bbmaj9 Fmaj9"
    scene: "shadow darker"
    variation: "thin"
    roles:
      choir:
        active: false
      shimmer:
        active: false
  - id: return
    title: tower horizon
    duration: 240s
    harmony: "Dm11 Bbmaj9 | Fmaj9 Cmaj9"
    scene: "return wider"
    variation: "lift"
    roles:
      choir:
        active: true
      shimmer:
        active: true


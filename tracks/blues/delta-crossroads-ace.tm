title: Delta Crossroads (ACE)
description: SP25 ACE-Step blues. Emaj at 96 BPM, slide guitar Delta blues.
  acestep engine.
style: blues
substyle: delta
listen_mode: album-side
render_engine: acestep
seed: 25050

key: Emaj
tempo: 96
total_duration: 3m
tags: [blues, delta, slide-guitar, harmonica, shuffle, sp25]

acestep:
  style: >
    Delta blues: slide guitar on resonator with bottleneck slide, finger-
    picked acoustic guitar steady on the low strings, harmonica wailing
    over the top, foot stomp on the downbeats. Shuffle feel, dusty
    crossroads at dusk, cicadas. Instrumental, no vocals.
  tags: [blues, delta, slide guitar, harmonica, acoustic, shuffle]
  scale: major
  time_signature: 12/8
  motif: a slide guitar phrase that bends up to the flat third and down to the tonic
  inference_steps: 8
  sections:
    - id: intro
      bars: 4
      description: slide guitar alone, foot stomp begins
      harmony: "E7"
      dynamic: soft
    - id: head
      bars: 12
      description: full 12-bar form, harmonica states the head
      harmony: "E7 A7 E7 E7 A7 A7 E7 E7 B7 A7 E7 B7"
      dynamic: building
    - id: solo
      bars: 24
      description: harmonica and slide guitar trade solos over two choruses
      harmony: "E7 A7 E7 E7 A7 A7 E7 E7 B7 A7 E7 B7"
      dynamic: peak
    - id: outro
      bars: 8
      description: slide guitar restates the head, foot stomp fades
      harmony: "E7 A7 E7 B7 E7"
      dynamic: fade

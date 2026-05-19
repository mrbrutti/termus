title: Midnight Pool Blue (ACE)
description: SP25 ACE-Step chill. Dmin at 94 BPM, late-night sub bass and
  glittering chimes. acestep engine.
style: chill
substyle: late-night
listen_mode: album-side
render_engine: acestep
seed: 25033

key: Dmin
tempo: 94
total_duration: 3m
tags: [chill, late-night, sub-bass, chimes, pool, sp25]

acestep:
  style: >
    late-night chill: deep sub bass holds the low end, sparse rhodes chords
    spread wide in stereo, brushed half-time kit with very soft snare,
    glittering chime motif that floats above. Midnight swimming pool, water
    light flickering on a courtyard wall, no vocals, instrumental.
  tags: [chill, late-night, sub bass, rhodes, chimes, pool, atmospheric]
  scale: minor
  time_signature: 4/4
  motif: a sparkling chime arpeggio that outlines the minor ninth chord
  inference_steps: 8
  sections:
    - id: intro
      bars: 8
      description: sub bass pulses, chimes ring out, no drums
      harmony: "Dm9"
      dynamic: soft
    - id: head
      bars: 16
      description: half-time kit enters, rhodes spreads wide
      harmony: "Dm9 Am9 Gm9 A7sus"
      dynamic: building
    - id: bridge
      bars: 8
      description: brief relative major lift, chimes glint brighter
      harmony: "Bbmaj9 Fmaj9 Gm9 A7sus"
      dynamic: peak
    - id: outro
      bars: 8
      description: drums drop out, chimes and sub bass remain
      harmony: "Dm9 A7sus Dm9"
      dynamic: fade

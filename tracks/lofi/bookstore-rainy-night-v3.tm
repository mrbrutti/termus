title: Bookstore Rainy Night (v3)
description: SP21 ACE-Step v3 reference track. Lofi in D minor, slow rainy
  evening. The acestep engine is asked for a long-form lofi piece with
  rich natural-language guidance and per-section descriptions.
style: lofi
substyle: bookstore
listen_mode: hour-stream
render_engine: acestep
seed: 21070

key: Dmin
tempo: 86
total_duration: 3m
tags: [lofi, rhodes, rainy, bookstore, nocturne, sp21]

acestep:
  style: >
    warm lo-fi instrumental jazz in a quiet bookstore on a rainy night.
    Tape-saturated Fender Rhodes carries a wandering minor melody over
    upright bass and brushed drums. Vinyl crackle and distant rain
    soften the edges. No vocals, no percussion lead, sparse and patient.
  tags: [lofi, rhodes, jazz, instrumental, rainy, nocturne]
  scale: minor
  time_signature: 4/4
  motif: stepwise minor descent from the fifth to the root, with one held suspended ninth per phrase
  inference_steps: 8
  sections:
    - id: intro
      bars: 8
      description: rain texture in, soft Rhodes states the motif over a single chord, no drums yet
      harmony: "Dm9"
      dynamic: soft
    - id: head
      bars: 16
      description: brushed kick and walking bass enter, Rhodes plays the motif with light variations
      harmony: "Dm9 Gm7 Bbmaj7 A7sus"
      dynamic: building
    - id: bridge
      bars: 16
      description: shift to relative major colour, dynamic peaks briefly, then pulls back
      harmony: "Bbmaj7 Fmaj7 Em7b5 A7"
      dynamic: peak
    - id: outro
      bars: 8
      description: drums drop, Rhodes alone, motif fragments and fades into rain
      harmony: "Dm9 A7sus Dm9"
      dynamic: fade

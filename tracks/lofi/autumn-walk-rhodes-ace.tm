title: Autumn Walk Rhodes (ACE)
description: SP25 ACE-Step lofi. Amin at 88 BPM, felt piano with wandering
  saxophone over dilla-feel kit. acestep engine.
style: lofi
substyle: rhodes-sax
listen_mode: album-side
render_engine: acestep
seed: 25012

key: Amin
tempo: 88
total_duration: 3m
tags: [lofi, rhodes, saxophone, autumn, dilla, sp25]

acestep:
  style: >
    felt-piano lofi with a wandering tenor saxophone melody over warm minor
    ninth chords. Dilla-feel drum kit with ghost snares and a slightly
    off-grid swing pocket. Autumn evening city walk, late golden light,
    leaves underfoot. Instrumental, no vocals, gentle saturation and a hint
    of vinyl crackle.
  tags: [lofi, felt piano, rhodes, saxophone, dilla beats, autumn]
  scale: minor
  time_signature: 4/4
  motif: a wistful saxophone phrase that descends from the ninth to the third, with a held suspension
  inference_steps: 8
  sections:
    - id: intro
      bars: 8
      description: felt piano alone over a sustained chord, sax breathes in
      harmony: "Am9"
      dynamic: soft
    - id: head
      bars: 16
      description: kit enters with ghost snares, sax states the motif
      harmony: "Am9 Dm9 Fmaj9 E7sus"
      dynamic: building
    - id: bridge
      bars: 12
      description: brief modulation to relative major, sax improvises briefly
      harmony: "Fmaj9 Cmaj9 Bm7b5 E7"
      dynamic: peak
    - id: outro
      bars: 8
      description: drums fade, sax and rhodes alone, motif resolves
      harmony: "Am9 E7sus Am9"
      dynamic: fade

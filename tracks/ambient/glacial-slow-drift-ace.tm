title: Glacial Slow Drift (ACE)
description: SP25 ACE-Step ambient. Amin at 55 BPM, Stars-of-the-Lid-style
  drone clusters. acestep engine.
style: ambient
substyle: drone-cluster
listen_mode: hour-stream
render_engine: acestep
seed: 25042

key: Amin
tempo: 55
total_duration: 5m
tags: [ambient, drone, stars-of-the-lid, glacial, arctic, sp25]

acestep:
  style: >
    Stars of the Lid style ambient: very slow harmonic motion, drone-cluster
    pads that overlap and crossfade, no rhythm at all, evolving across
    minutes rather than bars. Arctic stillness, frost forming on glass,
    suspended time. Instrumental, no vocals, deeply slow.
  tags: [ambient, drone, stars of the lid, beatless, glacial, slow]
  scale: minor
  time_signature: 4/4
  motif: a single sustained low pitch on the tonic that swells in waves
  inference_steps: 8
  sections:
    - id: drift-1
      bars: 24
      description: tonic drone fades in from silence over many seconds
      harmony: "Am"
      dynamic: soft
    - id: drift-2
      bars: 32
      description: drone splits into a triad, gentle beating between notes
      harmony: "Am Em"
      dynamic: building
    - id: drift-3
      bars: 32
      description: full pad cluster, evolving harmonic shadows
      harmony: "Am Em Fmaj7"
      dynamic: peak
    - id: drift-4
      bars: 24
      description: cluster recedes, tonic drone alone, fades to silence
      harmony: "Am"
      dynamic: fade

title: Forest After Rain (ACE)
description: SP25 ACE-Step ambient. Emin at 65 BPM, choir pad over rain
  field recording. acestep engine.
style: ambient
substyle: field-recording
listen_mode: hour-stream
render_engine: acestep
seed: 25041

key: Emin
tempo: 65
total_duration: 4m
tags: [ambient, field-recording, rain, choir, forest, sp25]

acestep:
  style: >
    ambient with field recording: subtle rain texture in the background,
    sustained choir pad layered with soft synth, soft chime arpeggios that
    sparkle out of the texture, slow harmonic motion that drifts between
    minor and relative major. Forest after a storm, water dripping from
    branches, no vocals as a lead, instrumental.
  tags: [ambient, choir pad, rain, chimes, field recording, forest]
  scale: minor
  time_signature: 4/4
  motif: a falling chime arpeggio that outlines the minor seventh chord
  inference_steps: 8
  sections:
    - id: intro
      bars: 16
      description: rain texture, choir fades in
      harmony: "Em9"
      dynamic: soft
    - id: middle
      bars: 24
      description: chimes begin, choir thickens, rain settles
      harmony: "Em9 Cmaj9"
      dynamic: building
    - id: bloom
      bars: 24
      description: full choir, chimes sparkle, brief relative major lift
      harmony: "Em9 Cmaj9 Am9 Bm7"
      dynamic: peak
    - id: outro
      bars: 16
      description: choir thins, rain dominates, chimes fade
      harmony: "Em9"
      dynamic: fade

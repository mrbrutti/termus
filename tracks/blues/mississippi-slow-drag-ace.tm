title: Mississippi Slow Drag (ACE)
description: SP25 ACE-Step blues. Bbmaj at 72 BPM, dobro slide with lazy
  shuffle pocket. acestep engine.
style: blues
substyle: mississippi
listen_mode: album-side
render_engine: acestep
seed: 25053

key: Bbmaj
tempo: 72
total_duration: 3m
tags: [blues, mississippi, dobro, slide, slow, sp25]

acestep:
  style: >
    very slow Mississippi blues: dobro resonator with bottleneck slide,
    upright bass walking a lazy shuffle pattern, harmonica in cross harp
    position, drums with brushes on a relaxed pocket, hot afternoon sound.
    Porch swing, cicadas, no vocals, instrumental.
  tags: [blues, mississippi, dobro, slide guitar, upright bass, harmonica, slow shuffle]
  scale: major
  time_signature: 12/8
  motif: a dobro slide phrase that swoops up to the flat third and back to the root
  inference_steps: 8
  sections:
    - id: intro
      bars: 4
      description: dobro slide alone, upright bass enters with a held tonic
      harmony: "Bb7"
      dynamic: soft
    - id: head
      bars: 12
      description: full 12-bar form, harmonica states the head
      harmony: "Bb7 Eb7 Bb7 Bb7 Eb7 Eb7 Bb7 Bb7 F7 Eb7 Bb7 F7"
      dynamic: building
    - id: solo
      bars: 24
      description: dobro and harmonica trade solos, brushed drums steady
      harmony: "Bb7 Eb7 Bb7 Bb7 Eb7 Eb7 Bb7 Bb7 F7 Eb7 Bb7 F7"
      dynamic: peak
    - id: outro
      bars: 8
      description: dobro restates the head, band winds down on the tonic
      harmony: "Bb7 Eb7 Bb7 F7 Bb7"
      dynamic: fade

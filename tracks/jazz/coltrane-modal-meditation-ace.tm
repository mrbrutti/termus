title: Coltrane Modal Meditation (ACE)
description: SP25 ACE-Step jazz. D dorian at 110 BPM, modal jazz with McCoy
  Tyner-style intensity. acestep engine.
style: jazz
substyle: modal
listen_mode: album-side
render_engine: acestep
seed: 25023

key: Dmin
tempo: 110
total_duration: 3m
tags: [jazz, modal, coltrane, mccoy-tyner, dorian, sp25]

acestep:
  style: >
    modal jazz: piano plays sus4 voicings on D dorian for long stretches,
    walking bass holds the modal center, ride cymbal spang-a-lang, tenor
    saxophone explores modal lines with breath and overtones. McCoy
    Tyner-style left-hand intensity build. Spiritual, searching, no vocals,
    instrumental jazz.
  tags: [jazz, modal, dorian, tenor saxophone, mccoy tyner, spiritual]
  scale: dorian
  time_signature: 4/4
  motif: a stepwise dorian melody that uses the sixth and fourth as anchors
  inference_steps: 8
  sections:
    - id: intro
      bars: 8
      description: piano sus4 voicings alone, bass enters with a pedal D
      harmony: "Dm11"
      dynamic: soft
    - id: head
      bars: 16
      description: drums enter on ride, sax states the dorian melody
      harmony: "Dm11 Em11"
      dynamic: building
    - id: solo
      bars: 32
      description: sax improvises, piano comps with quartal voicings
      harmony: "Dm11 Em11 Dm11 Em11"
      dynamic: peak
    - id: outro
      bars: 8
      description: piano returns to sus4 voicings, sax exits, bass alone
      harmony: "Dm11"
      dynamic: fade

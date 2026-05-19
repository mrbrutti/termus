title: Bossa Nova Rooftop (ACE)
description: SP25 ACE-Step jazz. Cmin bossa at 124 BPM, nylon-string guitar
  over warm Rhodes pad, sunset in Rio. acestep engine.
style: jazz
substyle: bossa-nova
listen_mode: album-side
render_engine: acestep
seed: 25022

key: Cmin
tempo: 124
total_duration: 3m
tags: [jazz, bossa-nova, brazil, nylon-guitar, sunset, sp25]

acestep:
  style: >
    bossa nova: nylon-string acoustic guitar with light fingerstyle pattern,
    soft brushed drums on a relaxed samba pattern, walking upright bass,
    warm Rhodes pad sustaining underneath. Sunset rooftop in Rio, distant
    ocean, evening breeze. Instrumental, no vocals, gentle and warm.
  tags: [bossa nova, jazz, nylon guitar, brushes, rhodes, brazil]
  scale: minor
  time_signature: 4/4
  motif: a lilting guitar fingerstyle pattern on the i to IVmaj7 of the key
  inference_steps: 8
  sections:
    - id: intro
      bars: 8
      description: nylon guitar alone, fingerstyle pattern, no rhythm section yet
      harmony: "Cm9 Fmaj7"
      dynamic: soft
    - id: head
      bars: 16
      description: brushed samba kit and bass enter, rhodes pad joins
      harmony: "Cm9 Fmaj7 Bbmaj7 Eb6 Abmaj7 Dm7b5 G7 Cm9"
      dynamic: building
    - id: bridge
      bars: 16
      description: harmonic shift toward the relative major, guitar improvises
      harmony: "Ebmaj7 Abmaj7 Dm7b5 G7 Cm9 Fmaj7 Bbmaj7 G7"
      dynamic: peak
    - id: outro
      bars: 8
      description: rhythm section thins, nylon guitar resolves the motif
      harmony: "Cm9 Fmaj7 Cm9"
      dynamic: fade

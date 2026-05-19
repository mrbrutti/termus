title: Stellar Aurora (ACE)
description: SP25 ACE-Step ambient. F#min at 62 BPM, arpeggiated bell motifs
  over sustained drone. acestep engine.
style: ambient
substyle: cinematic
listen_mode: hour-stream
render_engine: acestep
seed: 25043

key: F#min
tempo: 62
total_duration: 4m
tags: [ambient, cinematic, aurora, bells, vast, sp25]

acestep:
  style: >
    cinematic ambient: arpeggiated bell motifs ring out over a sustained
    drone, occasional swells of strings rise and recede, vast cosmic
    scale. Aurora over a frozen plain, stars in motion, no rhythm, no
    vocals, instrumental.
  tags: [ambient, cinematic, bells, drone, strings, aurora, cosmic]
  scale: minor
  time_signature: 4/4
  motif: a rising bell arpeggio that traces the minor pentatonic
  inference_steps: 8
  sections:
    - id: intro
      bars: 16
      description: drone fades in, single bell tone rings
      harmony: "F#m"
      dynamic: soft
    - id: middle
      bars: 24
      description: bell arpeggio begins, distant strings swell underneath
      harmony: "F#m Amaj7"
      dynamic: building
    - id: bloom
      bars: 24
      description: full bell arpeggios, strings at peak, drone shimmers
      harmony: "F#m Amaj7 Dmaj7 C#m"
      dynamic: peak
    - id: outro
      bars: 16
      description: strings recede, bells slow, drone alone fades
      harmony: "F#m"
      dynamic: fade

title: Chicago After Midnight (ACE)
description: SP25 ACE-Step blues. Gmaj at 120 BPM, overdriven Chicago electric
  blues. acestep engine.
style: blues
substyle: chicago-electric
listen_mode: album-side
render_engine: acestep
seed: 25051

key: Gmaj
tempo: 120
total_duration: 3m
tags: [blues, chicago, electric, harmonica, sp25]

acestep:
  style: >
    Chicago electric blues: overdriven electric guitar with bend-heavy
    phrasing, harmonica wailing through a microphone, walking electric
    bass, drums with a backbeat on 2 and 4, light hi-hat. Smoky bar, neon
    sign reflections in a rain-wet street. Instrumental, no vocals.
  tags: [blues, chicago, electric guitar, harmonica, walking bass, backbeat]
  scale: major
  time_signature: 4/4
  motif: a guitar phrase that bends the flat seventh up to the tonic
  inference_steps: 8
  sections:
    - id: intro
      bars: 4
      description: guitar alone over a held G7, drum fill cues the band
      harmony: "G7"
      dynamic: soft
    - id: head
      bars: 12
      description: full 12-bar form, harmonica plays the head
      harmony: "G7 C7 G7 G7 C7 C7 G7 G7 D7 C7 G7 D7"
      dynamic: building
    - id: solo
      bars: 24
      description: guitar takes a chorus, harmonica takes a chorus
      harmony: "G7 C7 G7 G7 C7 C7 G7 G7 D7 C7 G7 D7"
      dynamic: peak
    - id: outro
      bars: 8
      description: band drops to a stop-time turnaround, fades on the tonic
      harmony: "G7 C7 G7 D7 G7"
      dynamic: fade

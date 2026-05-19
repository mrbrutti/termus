title: Bourbon Street Blues (ACE)
description: SP25 ACE-Step jazz. F major up-tempo blues at 138 BPM with
  smoky New Orleans club feel. acestep engine.
style: jazz
substyle: blues-uptempo
listen_mode: album-side
render_engine: acestep
seed: 25021

key: Fmaj
tempo: 138
total_duration: 3m
tags: [jazz, blues, new-orleans, tenor-sax, uptempo, sp25]

acestep:
  style: >
    up-tempo jazz blues in F major: walking upright bass on every quarter,
    hi-hat on 2 and 4 with light ride cymbal, comping piano playing rootless
    voicings, tenor saxophone improvises over the changes. Smoky New
    Orleans club at midnight, brick walls, cigarette glow. Instrumental,
    no vocals.
  tags: [jazz, blues, new orleans, tenor saxophone, walking bass, uptempo]
  scale: major
  time_signature: 4/4
  motif: a bluesy bend-and-release tenor sax phrase that sits on the flat third
  inference_steps: 8
  sections:
    - id: head
      bars: 12
      description: full band on the 12-bar blues form, sax states the head
      harmony: "F7 Bb7 F7 F7 Bb7 Bb7 F7 D7 Gm7 C7 F7 C7"
      dynamic: building
    - id: solo-1
      bars: 24
      description: tenor sax improvises through two choruses
      harmony: "F7 Bb7 F7 F7 Bb7 Bb7 F7 D7 Gm7 C7 F7 C7"
      dynamic: peak
    - id: solo-2
      bars: 12
      description: piano takes a chorus, bass walks twice as busy
      harmony: "F7 Bb7 F7 F7 Bb7 Bb7 F7 D7 Gm7 C7 F7 C7"
      dynamic: peak
    - id: outro
      bars: 8
      description: sax restates the head, band ends on the tonic with a turnaround
      harmony: "F7 Bb7 F7 C7 F6"
      dynamic: fade

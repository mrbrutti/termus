title: Twelve Bar Rain (ACE)
description: SP25 ACE-Step blues. Amaj at 88 BPM, slow 12-bar with B3 organ
  and tenor sax. acestep engine.
style: blues
substyle: slow-12-bar
listen_mode: album-side
render_engine: acestep
seed: 25052

key: Amaj
tempo: 88
total_duration: 3m
tags: [blues, slow, organ, tenor-sax, sp25]

acestep:
  style: >
    slow 12-bar blues: warm clean electric guitar with light tremolo, B3
    Hammond organ pad with leslie chorale, walking electric bass, brushed
    drums in a slow shuffle, tenor saxophone solo over the changes. Soulful
    and patient, rainy afternoon. Instrumental, no vocals.
  tags: [blues, slow blues, b3 organ, tenor saxophone, brushes, soulful]
  scale: major
  time_signature: 4/4
  motif: a soulful tenor sax phrase that bends up to the flat fifth and resolves to the tonic
  inference_steps: 8
  sections:
    - id: intro
      bars: 4
      description: organ pad alone, guitar joins with a sustained chord
      harmony: "A7"
      dynamic: soft
    - id: head
      bars: 12
      description: full 12-bar form, sax states the head
      harmony: "A7 D7 A7 A7 D7 D7 A7 A7 E7 D7 A7 E7"
      dynamic: building
    - id: solo
      bars: 24
      description: tenor sax improvises two slow choruses
      harmony: "A7 D7 A7 A7 D7 D7 A7 A7 E7 D7 A7 E7"
      dynamic: peak
    - id: outro
      bars: 8
      description: organ and sax wind down, guitar holds the tonic chord
      harmony: "A7 D7 A7 E7 A7"
      dynamic: fade

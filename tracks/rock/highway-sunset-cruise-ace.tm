title: Highway Sunset Cruise (ACE)
description: SP25 ACE-Step rock. Emaj at 124 BPM, classic-rock instrumental
  cruise. acestep engine.
style: rock
substyle: classic-rock
listen_mode: album-side
render_engine: acestep
seed: 25061

key: Emaj
tempo: 124
total_duration: 3m
tags: [rock, classic-rock, instrumental, highway, organ, sp25]

acestep:
  style: >
    mid-tempo classic rock instrumental: clean electric guitar arpeggios
    weaving against an organ pad, warm electric bass on root-fifth pattern,
    steady drums with hi-hat on the offbeats, open-road feeling. Sunset on
    a long highway, no vocals, instrumental.
  tags: [rock, classic rock, clean guitar, hammond organ, electric bass, highway]
  scale: major
  time_signature: 4/4
  motif: an arpeggiated guitar phrase that ascends a major triad and falls to the seventh
  inference_steps: 8
  sections:
    - id: intro
      bars: 8
      description: clean guitar arpeggio alone, organ pad swells in
      harmony: "Emaj7"
      dynamic: soft
    - id: verse
      bars: 16
      description: bass and drums enter, full band on the main progression
      harmony: "Emaj7 C#m7 Amaj7 B7sus Emaj7 C#m7 Amaj7 B7sus"
      dynamic: building
    - id: bridge
      bars: 16
      description: organ takes a lead line, guitar drops to chord stabs
      harmony: "Amaj7 Bmaj7 C#m7 Amaj7 Bmaj7 C#m7 Amaj7 B7sus"
      dynamic: peak
    - id: outro
      bars: 8
      description: drums thin out, organ and guitar resolve to tonic
      harmony: "Emaj7 B7sus Emaj7"
      dynamic: fade

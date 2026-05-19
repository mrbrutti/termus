title: Basement Jam Session (ACE)
description: SP25 ACE-Step rock. Dmin at 116 BPM, blues-rock instrumental jam.
  acestep engine.
style: rock
substyle: blues-rock
listen_mode: album-side
render_engine: acestep
seed: 25062

key: Dmin
tempo: 116
total_duration: 3m
tags: [rock, blues-rock, jam, hammond, distorted-guitar, sp25]

acestep:
  style: >
    blues-rock instrumental jam: distorted electric guitar with an extended
    pentatonic solo, Hammond B3 organ comping behind, driving electric bass
    on a sixteenth-note pattern, classic rock drum kit with snare on 2 and
    4 and tom fills. Basement studio, amps cranked, no vocals, instrumental.
  tags: [rock, blues rock, distorted guitar, hammond organ, electric bass, jam]
  scale: minor
  time_signature: 4/4
  motif: a Hendrix-style pentatonic guitar phrase that hammers on the flat third
  inference_steps: 8
  sections:
    - id: intro
      bars: 4
      description: drums alone count off, guitar feedback swells in
      harmony: "Dm7"
      dynamic: soft
    - id: verse
      bars: 16
      description: full band on the main groove, organ pad enters
      harmony: "Dm7 G7 Dm7 G7 Dm7 G7 Bb7 A7"
      dynamic: building
    - id: solo
      bars: 32
      description: guitar takes an extended solo, organ comps with stabs
      harmony: "Dm7 G7 Dm7 G7 Bb7 A7 Dm7 A7"
      dynamic: peak
    - id: outro
      bars: 8
      description: guitar lands on the tonic, drums fade, organ holds
      harmony: "Dm7 A7 Dm7"
      dynamic: fade

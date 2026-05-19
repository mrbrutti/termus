title: Garage Saturday Night (ACE)
description: SP25 ACE-Step rock. Amin at 132 BPM, garage rock power chords.
  acestep engine.
style: rock
substyle: garage
listen_mode: album-side
render_engine: acestep
seed: 25060

key: Amin
tempo: 132
total_duration: 3m
tags: [rock, garage, power-chords, sp25]

acestep:
  style: >
    garage rock: overdriven electric guitar power chords, driving bass on
    eighth notes locking with the kick, punchy drums with snare on 2 and
    4, no vocals, raw and unpolished. Garage band energy, single-mic
    room, instrumental.
  tags: [rock, garage rock, power chords, distorted guitar, driving bass]
  scale: minor
  time_signature: 4/4
  motif: a four-chord guitar progression that rocks between Am and the relative major's IV
  inference_steps: 8
  sections:
    - id: intro
      bars: 4
      description: drums alone count off, guitar feedback swells
      harmony: "Am"
      dynamic: soft
    - id: verse
      bars: 16
      description: full band on the main four-chord progression
      harmony: "Am F C G Am F C G"
      dynamic: building
    - id: chorus
      bars: 16
      description: guitar opens to fifths and sixths, drums hit harder
      harmony: "F G Am C F G Am C"
      dynamic: peak
    - id: outro
      bars: 8
      description: band locks on the tonic, guitar feedback fades
      harmony: "Am F C G Am"
      dynamic: fade

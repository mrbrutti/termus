title: Mountain Fog Drift (ACE)
description: SP25 ACE-Step chill. Amin at 88 BPM, half-time chillhop with
  plate reverb and distant flute. acestep engine.
style: chill
substyle: atmospheric-chillhop
listen_mode: album-side
render_engine: acestep
seed: 25031

key: Amin
tempo: 88
total_duration: 3m
tags: [chill, chillhop, atmospheric, mountain, flute, sp25]

acestep:
  style: >
    atmospheric chillhop: half-time drums with soft kick on 1 and snare on
    3, plate-reverb pad chords, sparse rhodes stabs on the offbeats, a
    distant flute melody that fades in and out of focus. Foggy mountain
    stillness, slow drifting clouds, no vocals, instrumental.
  tags: [chillhop, chill, atmospheric, half-time, plate reverb, flute, mountain]
  scale: minor
  time_signature: 4/4
  motif: a slow falling flute phrase that traces the minor pentatonic
  inference_steps: 8
  sections:
    - id: intro
      bars: 8
      description: pad and wind texture, flute breath, no drums
      harmony: "Am9"
      dynamic: soft
    - id: head
      bars: 16
      description: half-time kit enters, rhodes stabs, flute states the motif
      harmony: "Am9 Em9 Fmaj9 Dm9"
      dynamic: building
    - id: bridge
      bars: 8
      description: brief major lift, flute climbs higher
      harmony: "Fmaj9 Cmaj9 Dm9 E7sus"
      dynamic: peak
    - id: outro
      bars: 8
      description: drums fade, flute alone with pad, motif dissolves
      harmony: "Am9 Em9 Am9"
      dynamic: fade

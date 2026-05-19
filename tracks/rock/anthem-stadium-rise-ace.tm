title: Anthem Stadium Rise (ACE)
description: SP25 ACE-Step rock. Gmaj at 128 BPM, anthemic instrumental
  build. acestep engine.
style: rock
substyle: anthemic
listen_mode: album-side
render_engine: acestep
seed: 25063

key: Gmaj
tempo: 128
total_duration: 3m
tags: [rock, anthemic, stadium, climactic, sp25]

acestep:
  style: >
    anthemic rock instrumental: clean electric guitar arpeggios that
    transform into distorted power chords across the arc, big drums with
    open hi-hats and snare on 2 and 4, sustaining electric bass with light
    overdrive, climactic build through to a peak chorus, stadium-sized
    sound. No vocals, instrumental.
  tags: [rock, anthemic, stadium, clean to distorted guitar, big drums, climactic]
  scale: major
  time_signature: 4/4
  motif: a four-note guitar arpeggio that climbs the major triad and lands on the fifth
  inference_steps: 8
  sections:
    - id: intro
      bars: 8
      description: clean guitar arpeggios alone, drums tease the snare
      harmony: "Gmaj7"
      dynamic: soft
    - id: verse
      bars: 16
      description: bass and full kit enter, guitar still clean
      harmony: "Gmaj7 Em7 Cmaj7 D7sus Gmaj7 Em7 Cmaj7 D7sus"
      dynamic: building
    - id: chorus
      bars: 16
      description: guitar switches to distorted power chords, drums hit huge
      harmony: "G D Em C G D Em C"
      dynamic: peak
    - id: outro
      bars: 8
      description: distortion fades, clean guitar arpeggio returns
      harmony: "Gmaj7 D7sus Gmaj7"
      dynamic: fade

title: Sunset Balcony Loop (ACE)
description: SP25 ACE-Step chill. Bbmaj at 102 BPM, warm chillhop boom-bap.
  acestep engine.
style: chill
substyle: chillhop
listen_mode: album-side
render_engine: acestep
seed: 25032

key: Bbmaj
tempo: 102
total_duration: 3m
tags: [chill, chillhop, boom-bap, sunset, rooftop, sp25]

acestep:
  style: >
    warm chillhop: rhodes comping warm major-seventh chords over a relaxed
    boom-bap drum pattern with a slightly dusty snare, walking upright bass
    rounds out the bottom, sustained pad sits behind. Sunset rooftop loop,
    golden hour, no vocals, instrumental.
  tags: [chillhop, boom-bap, rhodes, upright bass, warm, sunset]
  scale: major
  time_signature: 4/4
  motif: a syncopated rhodes phrase that lands on the major seventh
  inference_steps: 8
  sections:
    - id: intro
      bars: 8
      description: rhodes alone, pad slowly fades in
      harmony: "Bbmaj9"
      dynamic: soft
    - id: head
      bars: 16
      description: boom-bap drums and bass enter, rhodes states the motif
      harmony: "Bbmaj9 Gm9 Ebmaj9 F7sus"
      dynamic: building
    - id: bridge
      bars: 8
      description: relative minor lift, rhodes voicings open up
      harmony: "Gm9 Ebmaj9 Dm9 F7sus"
      dynamic: peak
    - id: outro
      bars: 8
      description: drums thin, rhodes and pad close the loop
      harmony: "Bbmaj9 F7sus Bbmaj9"
      dynamic: fade

title: Autumn Leaves After Hours (ACE)
description: SP25 ACE-Step jazz piano trio. Emin to G major bridge, 128 BPM.
  Late-night Bill Evans-style rootless voicings with tenor sax. acestep
  engine.
style: jazz
substyle: piano-trio
listen_mode: album-side
render_engine: acestep
seed: 25020

key: Emin
tempo: 128
total_duration: 3m
tags: [jazz, piano-trio, bill-evans, tenor-sax, after-hours, sp25]

acestep:
  style: >
    late-night jazz piano trio with a tenor saxophone guest. Walking upright
    bass and brushed ride cymbal spang-a-lang behind Bill Evans-style
    rootless rhodes / acoustic piano voicings. Tenor sax states the melody
    in a statement-and-answer call with the piano. Intimate club room, soft
    air conditioning hum. Instrumental, no vocals.
  tags: [jazz, piano trio, walking bass, brushed drums, tenor saxophone, after hours]
  scale: minor
  time_signature: 4/4
  motif: a four-bar tenor sax phrase that traces the ii-V-i of the home key
  inference_steps: 8
  sections:
    - id: intro
      bars: 8
      description: piano alone with rootless voicings, no rhythm section yet
      harmony: "Em9 A7"
      dynamic: soft
    - id: head
      bars: 32
      description: walking bass and brushes enter, sax states the head
      harmony: "Em7 A7 Dm7 G7 Cmaj7 Fmaj7 Bm7b5 E7"
      dynamic: building
    - id: bridge
      bars: 16
      description: shift to G major, piano takes a short solo, sax responds
      harmony: "G7 Cmaj7 Am7 D7 Bm7 E7 Am7 D7"
      dynamic: peak
    - id: outro
      bars: 8
      description: sax restates the head, trio fades on tonic
      harmony: "Em7 A7 Em9"
      dynamic: fade

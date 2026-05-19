title: Cassette Dream Loop (ACE)
description: SP25 ACE-Step lofi reference. Dmin at 82 BPM, dusty cassette
  loop with felt-hammer Rhodes and walking upright bass. Built for the
  acestep engine; SF2 comparison is in the v2 lofi corpus.
style: lofi
substyle: cassette
listen_mode: album-side
render_engine: acestep
seed: 25010

key: Dmin
tempo: 82
total_duration: 3m
tags: [lofi, cassette, rhodes, bass, sp25]

acestep:
  style: >
    warm fender rhodes with felt hammer noise carries a slow minor motif
    over a walking upright bass with finger squeak and brushed snare in a
    shuffle pocket. Sustained low pad hums underneath. Tape-saturated
    cassette dust, subtle vinyl crackle, gentle wow-and-flutter pitch drift.
    Instrumental, no vocals, late-night patience.
  tags: [lofi, cassette, rhodes, upright bass, brushed drums, instrumental]
  scale: minor
  time_signature: 4/4
  motif: descending stepwise minor figure from the fifth, lingering on the seventh
  inference_steps: 8
  sections:
    - id: intro
      bars: 8
      description: tape hiss in, rhodes alone with the motif, no rhythm section
      harmony: "Dm9"
      dynamic: soft
    - id: head
      bars: 16
      description: walking bass and brushed snare enter, rhodes states the motif twice
      harmony: "Dm9 Gm7 Bbmaj7 A7sus"
      dynamic: building
    - id: bridge
      bars: 12
      description: relative major lift, rhodes plays a question-and-answer phrase
      harmony: "Bbmaj7 Fmaj7 Em7b5 A7"
      dynamic: peak
    - id: outro
      bars: 8
      description: rhythm section drops out, rhodes alone, motif fragments into tape noise
      harmony: "Dm9 A7sus Dm9"
      dynamic: fade

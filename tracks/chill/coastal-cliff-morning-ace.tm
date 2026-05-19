title: Coastal Cliff Morning (ACE)
description: SP25 ACE-Step chill. Fmaj at 96 BPM, chillhop with soft EP and
  syncopated bass. acestep engine.
style: chill
substyle: chillhop
listen_mode: album-side
render_engine: acestep
seed: 25030

key: Fmaj
tempo: 96
total_duration: 3m
tags: [chill, chillhop, ocean, morning, sp25]

acestep:
  style: >
    chillhop: 4/4 mid-tempo groove, soft electric piano with chorus,
    syncopated bass line that locks in with the kick, brushed snare,
    sustained warm pad, light bell sparkle on phrase endings. Ocean
    morning serenity, light breeze, gulls in the far distance.
    Instrumental, no vocals.
  tags: [chillhop, chill, electric piano, soft drums, ocean, morning]
  scale: major
  time_signature: 4/4
  motif: a five-note electric piano phrase that arcs up to the sixth and falls to the third
  inference_steps: 8
  sections:
    - id: intro
      bars: 8
      description: pad and ocean texture, no rhythm section yet
      harmony: "Fmaj9"
      dynamic: soft
    - id: head
      bars: 16
      description: drums and bass enter, EP states the motif
      harmony: "Fmaj9 Dm9 Bbmaj9 C7sus"
      dynamic: building
    - id: bridge
      bars: 8
      description: minor lift, bells sparkle on phrase endings
      harmony: "Dm9 Bbmaj9 Am9 C7sus"
      dynamic: peak
    - id: outro
      bars: 8
      description: drums thin out, EP and pad fade
      harmony: "Fmaj9 C7sus Fmaj9"
      dynamic: fade

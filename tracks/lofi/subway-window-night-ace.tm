title: Subway Window Night (ACE)
description: SP25 ACE-Step lofi nocturne. Emin at 76 BPM, muted trumpet
  over upright bass and brushed kit, late-night transit. acestep engine.
style: lofi
substyle: nocturne
listen_mode: album-side
render_engine: acestep
seed: 25011

key: Emin
tempo: 76
total_duration: 3m
tags: [lofi, jazz, trumpet, nocturne, subway, sp25]

acestep:
  style: >
    lo-fi instrumental jazz: muted trumpet states a wandering minor melody
    over walking upright bass and brushed snare kit. Distant subway hum and
    far-away traffic rumble sit deep in the mix. Late-night nocturne, dim
    train window, no vocals, sparse and patient.
  tags: [lofi, jazz, muted trumpet, upright bass, brushed drums, nocturne]
  scale: minor
  time_signature: 4/4
  motif: a four-note trumpet phrase climbing the minor pentatonic then falling to the root
  inference_steps: 8
  sections:
    - id: intro
      bars: 8
      description: subway hum in, muted trumpet alone, no rhythm yet
      harmony: "Em9"
      dynamic: soft
    - id: head
      bars: 16
      description: walking bass and brushes enter, trumpet plays the motif
      harmony: "Em9 Am11 Cmaj7 B7sus"
      dynamic: building
    - id: bridge
      bars: 12
      description: minor-to-major lift, trumpet pushes higher with breath noise
      harmony: "Cmaj7 G/B Am9 B7"
      dynamic: peak
    - id: outro
      bars: 8
      description: drums drop, trumpet alone, fades back into subway hum
      harmony: "Em9 B7sus Em9"
      dynamic: fade

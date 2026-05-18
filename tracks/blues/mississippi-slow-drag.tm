title: Mississippi Slow Drag
description: SP19 blues in Bb major — very slow, sparse. Built on style=jazz, mix_bus=jazz. Emphasizes the between-notes feel.
style: blues
substyle: slow-drag
mix_bus: jazz
listen_mode: hour-stream
seed: 75004
tags: [blues, slow, sparse, bb-major, sp19]
key: Bbmaj
tempo: 72
globals: {density: sparse, brightness: warm, motion: gentle, phrase: long}

form: jazz_blues_12bar
total_duration: 7m

motif_library:
  slow_drag:
    pattern: "1 . . . b3 . . . | 4 . . . . . . . | b7 . . . 5 . . . | b3 . . . 1 . . ."
    description: "very slow blues — between the beats"
    bars: 4

roles:
  bass:
    family: bass
    voice: jazz_upright_bass
    auto_voice: walking_bass
    register: low
    prominence: anchor
    humanize: {timing_ms: 8, velocity: 10}
    chain: {reverb_send: 0.25, compress: gentle, pan_offset: -0.04}

  piano:
    family: acoustic_piano
    voice: jazz_grand_piano
    auto_voice: shell_voicing
    register: mid
    prominence: support
    humanize: {timing_ms: 12, velocity: 12, phrase_shape: arc}
    chain: {reverb_send: 0.45, compress: gentle}

  tenor:
    family: reed_lead
    voice: jazz_tenor_sax
    auto_phrase: slow_ballad
    register: mid-high
    prominence: lead
    humanize: {timing_ms: 18, velocity: 16, accent: phrase_arc}
    chain: {reverb_send: 0.58, compress: gentle}

  kick:
    family: drums
    prominence: anchor
    humanize: {timing_ms: 8, velocity: 7}
    chain: {reverb_send: 0.10, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 5.0, pitch: "", dur: 0.25, vel: 86}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 8, velocity: 8}
    chain: {reverb_send: 0.34, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 3.0, pitch: "", dur: 0.30, vel: 80}
      - {beat: 7.0, pitch: "", dur: 0.30, vel: 78}

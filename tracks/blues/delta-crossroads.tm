title: Delta Crossroads
description: SP19 blues in E major (classic key). No-code-change test — built on form=jazz_blues_12bar, mix_bus=jazz. Walking bass + minimal piano + tenor lead. Slow shuffle feel.
style: blues
substyle: delta-blues
mix_bus: jazz
listen_mode: hour-stream
seed: 75001
tags: [blues, walking, slow-shuffle, e-major, sp19]
key: Emaj
tempo: 96
globals: {density: full, brightness: warm, motion: moving, phrase: long}

form: jazz_blues_12bar
total_duration: 6m

motif_library:
  delta_lick:
    pattern: "1 b3 4 b5 5 . . . | 4 b3 1 . . . . . | b7 . 5 b3 1 . . ."
    description: "delta blues turnaround lick"
    bars: 4

roles:
  bass:
    family: bass
    voice: jazz_upright_bass
    auto_voice: walking_bass
    register: low
    prominence: anchor
    humanize: {timing_ms: 6, velocity: 8}
    chain: {reverb_send: 0.18, compress: gentle, pan_offset: -0.05}

  piano:
    family: acoustic_piano
    voice: jazz_grand_piano
    auto_voice: shell_voicing
    register: mid
    prominence: support
    humanize: {timing_ms: 8, velocity: 10}
    chain: {reverb_send: 0.40, compress: gentle}

  tenor:
    family: reed_lead
    voice: jazz_tenor_sax
    auto_phrase: blues_lick
    register: high
    prominence: lead
    humanize: {timing_ms: 14, velocity: 14, accent: phrase_arc}
    chain: {reverb_send: 0.55, compress: gentle}

  kick:
    family: drums
    prominence: anchor
    humanize: {timing_ms: 5, velocity: 8}
    chain: {reverb_send: 0.08, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.0, pitch: "", dur: 0.25, vel: 96}
      - {beat: 3.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 5.0, pitch: "", dur: 0.25, vel: 96}
      - {beat: 7.0, pitch: "", dur: 0.25, vel: 88}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 5, velocity: 7}
    chain: {reverb_send: 0.30, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 2.0, pitch: "", dur: 0.30, vel: 88}
      - {beat: 4.0, pitch: "", dur: 0.30, vel: 88}
      - {beat: 6.0, pitch: "", dur: 0.30, vel: 88}
      - {beat: 8.0, pitch: "", dur: 0.30, vel: 88}

  hat:
    family: drums
    prominence: support
    humanize: {timing_ms: 3, velocity: 5}
    chain: {reverb_send: 0.16, compress: "off", pan_offset: 0.20}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.10, vel: 70}
      - {beat: 1.67, pitch: "", dur: 0.10, vel: 56}
      - {beat: 2.00, pitch: "", dur: 0.10, vel: 64}
      - {beat: 2.67, pitch: "", dur: 0.10, vel: 56}
      - {beat: 3.00, pitch: "", dur: 0.10, vel: 70}
      - {beat: 3.67, pitch: "", dur: 0.10, vel: 56}
      - {beat: 4.00, pitch: "", dur: 0.10, vel: 64}
      - {beat: 4.67, pitch: "", dur: 0.10, vel: 56}

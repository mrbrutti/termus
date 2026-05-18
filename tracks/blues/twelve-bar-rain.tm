title: Twelve-Bar Rain
description: SP19 blues in A major. Slower, soulful. Built on style=jazz, mix_bus=jazz; heavy use of dominants.
style: blues
substyle: slow-blues
mix_bus: jazz
listen_mode: hour-stream
seed: 75002
tags: [blues, slow, soulful, a-major, sp19]
key: Amaj
tempo: 88
globals: {density: full, brightness: warm, motion: gentle, phrase: long}

textures:
  - {name: rain, level_db: -44}

form: jazz_blues_12bar
total_duration: 6m

motif_library:
  rain_blues:
    pattern: "1 . b3 . 4 . b5 . | 5 . . . b3 . 1 . | b7 . 5 . b3 . 1 ."
    description: "soulful slow blues — pentatonic with b5"
    bars: 4

roles:
  bass:
    family: bass
    voice: jazz_upright_bass
    auto_voice: walking_bass
    register: low
    prominence: anchor
    humanize: {timing_ms: 6, velocity: 8}
    chain: {reverb_send: 0.20, compress: gentle, pan_offset: -0.04}

  piano:
    family: acoustic_piano
    voice: jazz_grand_piano
    auto_voice: shell_voicing
    register: mid
    prominence: support
    humanize: {timing_ms: 8, velocity: 10}
    chain: {reverb_send: 0.42, compress: gentle}

  tenor:
    family: reed_lead
    voice: jazz_tenor_sax
    auto_phrase: blues_lick
    register: mid-high
    prominence: lead
    humanize: {timing_ms: 14, velocity: 14, accent: phrase_arc}
    chain: {reverb_send: 0.55, compress: gentle}

  kick:
    family: drums
    prominence: anchor
    humanize: {timing_ms: 5, velocity: 7}
    chain: {reverb_send: 0.08, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.0, pitch: "", dur: 0.25, vel: 92}
      - {beat: 5.0, pitch: "", dur: 0.25, vel: 90}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 6, velocity: 7}
    chain: {reverb_send: 0.34, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 3.0, pitch: "", dur: 0.30, vel: 88}
      - {beat: 7.0, pitch: "", dur: 0.30, vel: 86}

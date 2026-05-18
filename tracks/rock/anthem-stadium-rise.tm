title: Anthem Stadium Rise
description: SP19 rock ballad in G major — big-form attempt using jazz_aaba_32bar. style=lofi, mix_bus=lofi.
style: rock
substyle: rock-ballad
mix_bus: lofi
listen_mode: hour-stream
seed: 76004
tags: [rock, anthem, ballad, g-major, sp19]
key: Gmaj
tempo: 128
globals: {density: full, brightness: bright, motion: moving, phrase: long}

form: jazz_aaba_32bar
total_duration: 7m

motif_library:
  anthem_theme:
    pattern: "1 . 3 . 5 . 6 . | >2 . 6 . 5 . 3 . | 5 . 3 . 1 . . . | . . . . . . . ."
    description: "anthemic rising arc"
    bars: 8

roles:
  piano:
    family: acoustic_piano
    voice: jazz_grand_piano
    auto_voice: rhodes_comp
    register: mid-high
    prominence: support
    humanize: {timing_ms: 4, velocity: 12}
    chain: {reverb_send: 0.30, compress: punchy, tape_drive_db: 2.0}

  bass:
    family: bass
    voice: lofi_round_bass
    auto_voice: walking_bass
    register: low
    prominence: anchor
    humanize: {timing_ms: 4, velocity: 10}
    chain: {reverb_send: 0.14, compress: punchy, pan_offset: -0.04}

  kick:
    family: drums
    voice: lofi_dusty_kick
    prominence: anchor
    humanize: {timing_ms: 3, velocity: 8}
    chain: {reverb_send: 0.06, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.0, pitch: "", dur: 0.25, vel: 116}
      - {beat: 3.0, pitch: "", dur: 0.25, vel: 100}
      - {beat: 5.0, pitch: "", dur: 0.25, vel: 116}
      - {beat: 7.0, pitch: "", dur: 0.25, vel: 100}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 4, velocity: 8}
    chain: {reverb_send: 0.30, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 2.0, pitch: "", dur: 0.30, vel: 108}
      - {beat: 4.0, pitch: "", dur: 0.30, vel: 106}
      - {beat: 6.0, pitch: "", dur: 0.30, vel: 108}
      - {beat: 8.0, pitch: "", dur: 0.30, vel: 106}

  hat:
    family: drums
    prominence: support
    humanize: {timing_ms: 2, velocity: 6}
    chain: {reverb_send: 0.10, compress: "off", pan_offset: 0.20}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.08, vel: 82}
      - {beat: 1.50, pitch: "", dur: 0.08, vel: 70}
      - {beat: 2.00, pitch: "", dur: 0.08, vel: 82}
      - {beat: 2.50, pitch: "", dur: 0.08, vel: 70}
      - {beat: 3.00, pitch: "", dur: 0.08, vel: 82}
      - {beat: 3.50, pitch: "", dur: 0.08, vel: 70}
      - {beat: 4.00, pitch: "", dur: 0.08, vel: 82}
      - {beat: 4.50, pitch: "", dur: 0.08, vel: 70}

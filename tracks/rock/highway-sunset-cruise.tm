title: Highway Sunset Cruise
description: SP19 rock in E major — no-code-change test. Major-key rock feel, simpler harmony. Built on style=lofi.
style: rock
substyle: rock-cruise
mix_bus: lofi
listen_mode: hour-stream
seed: 76002
tags: [rock, major, cruising, e-major, sp19]
key: Emaj
tempo: 124
globals: {density: full, brightness: bright, motion: moving, phrase: long}

form: chill_ababcb
total_duration: 6m

motif_library:
  cruise_theme:
    pattern: "1 . 3 5 | 5 . 3 1 | 6 . 5 3 | 5 . 1 ."
    description: "open major rock — anthemic"
    bars: 4

roles:
  piano:
    family: acoustic_piano
    voice: jazz_grand_piano
    auto_voice: rhodes_comp
    register: mid-high
    prominence: support
    humanize: {timing_ms: 4, velocity: 10}
    chain: {reverb_send: 0.22, compress: punchy, tape_drive_db: 2.5}

  bass:
    family: bass
    voice: lofi_round_bass
    auto_voice: walking_bass
    register: low
    prominence: lead
    humanize: {timing_ms: 4, velocity: 10}
    chain: {reverb_send: 0.12, compress: punchy, pan_offset: -0.04}

  kick:
    family: drums
    voice: lofi_dusty_kick
    prominence: anchor
    humanize: {timing_ms: 2, velocity: 8}
    chain: {reverb_send: 0.06, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.0, pitch: "", dur: 0.25, vel: 114}
      - {beat: 3.5, pitch: "", dur: 0.25, vel: 100}
      - {beat: 5.0, pitch: "", dur: 0.25, vel: 114}
      - {beat: 7.5, pitch: "", dur: 0.25, vel: 100}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 3, velocity: 8}
    chain: {reverb_send: 0.30, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 2.0, pitch: "", dur: 0.30, vel: 108}
      - {beat: 4.0, pitch: "", dur: 0.30, vel: 108}
      - {beat: 6.0, pitch: "", dur: 0.30, vel: 108}
      - {beat: 8.0, pitch: "", dur: 0.30, vel: 108}

  hat:
    family: drums
    prominence: support
    humanize: {timing_ms: 2, velocity: 6}
    chain: {reverb_send: 0.10, compress: "off", pan_offset: 0.20}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.08, vel: 80}
      - {beat: 1.50, pitch: "", dur: 0.08, vel: 70}
      - {beat: 2.00, pitch: "", dur: 0.08, vel: 80}
      - {beat: 2.50, pitch: "", dur: 0.08, vel: 70}
      - {beat: 3.00, pitch: "", dur: 0.08, vel: 80}
      - {beat: 3.50, pitch: "", dur: 0.08, vel: 70}
      - {beat: 4.00, pitch: "", dur: 0.08, vel: 80}
      - {beat: 4.50, pitch: "", dur: 0.08, vel: 70}

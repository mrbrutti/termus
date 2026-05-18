title: Basement Jam Session
description: SP19 rock in D minor — blues-rock crossover. Slower groove. style=lofi, mix_bus=lofi.
style: rock
substyle: rock-blues
mix_bus: lofi
listen_mode: hour-stream
seed: 76003
tags: [rock, blues-rock, basement, d-minor, sp19]
key: Dmin
tempo: 116
globals: {density: full, brightness: balanced, motion: moving, phrase: long}

form: chill_journey
total_duration: 7m

motif_library:
  basement_riff:
    pattern: "1 . b3 . 4 . . 1 | b3 . . 1 b7 . 5 . | 1 . . . . . . ."
    description: "blues-rock minor pentatonic"
    bars: 4

roles:
  piano:
    family: acoustic_piano
    voice: jazz_grand_piano
    auto_voice: shell_voicing
    register: mid
    prominence: support
    humanize: {timing_ms: 6, velocity: 12}
    chain: {reverb_send: 0.22, compress: punchy, tape_drive_db: 2.5}

  bass:
    family: bass
    voice: lofi_round_bass
    auto_voice: walking_bass
    register: low
    prominence: lead
    humanize: {timing_ms: 5, velocity: 12}
    chain: {reverb_send: 0.14, compress: punchy, pan_offset: -0.06}

  kick:
    family: drums
    voice: lofi_dusty_kick
    prominence: anchor
    humanize: {timing_ms: 3, velocity: 8}
    chain: {reverb_send: 0.06, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.0, pitch: "", dur: 0.25, vel: 112}
      - {beat: 3.5, pitch: "", dur: 0.25, vel: 96}
      - {beat: 5.0, pitch: "", dur: 0.25, vel: 112}
      - {beat: 7.5, pitch: "", dur: 0.25, vel: 96}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 4, velocity: 8}
    chain: {reverb_send: 0.30, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 2.0, pitch: "", dur: 0.30, vel: 106}
      - {beat: 4.0, pitch: "", dur: 0.30, vel: 104}
      - {beat: 6.0, pitch: "", dur: 0.30, vel: 106}
      - {beat: 8.0, pitch: "", dur: 0.30, vel: 104}

  hat:
    family: drums
    prominence: support
    humanize: {timing_ms: 2, velocity: 6}
    chain: {reverb_send: 0.10, compress: "off", pan_offset: 0.20}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.08, vel: 78}
      - {beat: 1.50, pitch: "", dur: 0.08, vel: 66}
      - {beat: 2.00, pitch: "", dur: 0.08, vel: 78}
      - {beat: 2.50, pitch: "", dur: 0.08, vel: 66}
      - {beat: 3.00, pitch: "", dur: 0.08, vel: 78}
      - {beat: 3.50, pitch: "", dur: 0.08, vel: 66}
      - {beat: 4.00, pitch: "", dur: 0.08, vel: 78}
      - {beat: 4.50, pitch: "", dur: 0.08, vel: 66}

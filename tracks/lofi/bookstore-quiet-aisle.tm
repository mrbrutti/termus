title: Bookstore Quiet Aisle
description: SP19 lofi in C minor on chill_journey form. Slower, more harmonic motion. No pad — Rhodes + bass + minimal drums. Room-tone texture.
style: lofi
substyle: bookstore
listen_mode: hour-stream
seed: 71003
tags: [lofi, rhodes, quiet, bookstore, sp19]
key: Cmin
tempo: 86
mix_bus: lofi
globals: {density: sparse, brightness: warm, motion: gentle, reverb: room}

textures:
  - {name: room_tone, level_db: -42}
  - {name: vinyl, level_db: -46}

form: chill_journey
total_duration: 7m

motif_library:
  page_turn:
    pattern: "1 . 3 5 | 7 . 5 3 | b6 . b3 1 | 3 5 b3 1"
    description: "wandering minor melody"
    bars: 4

roles:
  rhodes:
    family: piano
    voice: lofi_rhodes_warm
    auto_voice: shell_voicing
    register: mid
    prominence: support
    humanize: {timing_ms: 8, velocity: 10, accent: dilla, phrase_shape: arc}
    chain: {reverb_send: 0.30, compress: gentle, tape_drive_db: 0.8}

  bass:
    family: bass
    voice: lofi_round_bass
    auto_voice: walking_bass
    register: low
    prominence: anchor
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.10, compress: gentle, pan_offset: -0.06}

  kick:
    family: drums
    voice: lofi_dusty_kick
    prominence: anchor
    humanize: {timing_ms: 4, velocity: 7, accent: dilla}
    chain: {reverb_send: 0.06, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.00, pitch: "", dur: 0.25, vel: 96}
      - {beat: 5.00, pitch: "", dur: 0.25, vel: 92}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 5, velocity: 7, accent: dilla}
    chain: {reverb_send: 0.40, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 3.00, pitch: "", dur: 0.30, vel: 88}
      - {beat: 7.00, pitch: "", dur: 0.30, vel: 86}

  hat:
    family: drums
    prominence: support
    humanize: {timing_ms: 2, velocity: 4}
    chain: {reverb_send: 0.16, compress: "off", pan_offset: 0.22}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.10, vel: 60}
      - {beat: 2.50, pitch: "", dur: 0.10, vel: 52}
      - {beat: 4.00, pitch: "", dur: 0.10, vel: 60}

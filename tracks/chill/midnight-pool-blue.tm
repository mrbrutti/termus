title: Midnight Pool Blue
description: SP19 chill ABABCB in D minor. Sparse, bass-forward, late-night feel.
style: chill
substyle: late-night
mix_bus: chill
listen_mode: hour-stream
seed: 73004
tags: [chill, sparse, bass-forward, late, sp19]
key: Dmin
tempo: 94
globals: {density: sparse, brightness: warm, motion: gentle, reverb: room}

form: chill_ababcb
total_duration: 6m

motif_library:
  pool_theme:
    pattern: "1 . . . | 3 . 5 . | b7 . 5 . | 3 1 . ."
    description: "minor melody — sparse, between-notes"
    bars: 4

roles:
  pad:
    family: pad
    voice: chill_pad_warm
    auto_voice: pad_sustain
    register: low
    prominence: air
    humanize: {timing_ms: 2, velocity: 4}
    chain: {reverb_send: 0.50, compress: glue}

  keys:
    family: piano
    voice: lofi_rhodes_warm
    auto_voice: shell_voicing
    register: mid
    prominence: support
    humanize: {timing_ms: 7, velocity: 9, phrase_shape: arc}
    chain: {reverb_send: 0.42, compress: gentle, tape_drive_db: 0.4}

  bass:
    family: bass
    voice: lofi_round_bass
    auto_voice: walking_bass
    register: low
    prominence: lead
    humanize: {timing_ms: 6, velocity: 9}
    chain: {reverb_send: 0.18, compress: punchy, pan_offset: -0.04}

  kick:
    family: drums
    prominence: anchor
    humanize: {timing_ms: 4, velocity: 7}
    chain: {reverb_send: 0.08, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.0, pitch: "", dur: 0.25, vel: 96}
      - {beat: 5.5, pitch: "", dur: 0.25, vel: 88}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.34, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 3.0, pitch: "", dur: 0.30, vel: 80}
      - {beat: 7.0, pitch: "", dur: 0.30, vel: 78}

title: Sunset Balcony Loop
description: SP19 chill journey in Bb major. Warm Rhodes, half-time kit, sustained pad.
style: chill
substyle: half-time-chill
mix_bus: chill
listen_mode: hour-stream
seed: 73003
tags: [chill, pad, halftime, warm, sp19]
key: Bbmaj
tempo: 102
globals: {density: full, brightness: warm, motion: moving, reverb: halo}

form: chill_journey
total_duration: 7m

motif_library:
  sunset_theme:
    pattern: "5 . 3 1 | 3 . 5 6 | >2 . 6 5 | 3 1 . ."
    description: "warm major melody — sunset glow"
    bars: 4

roles:
  pad:
    family: pad
    voice: chill_pad_warm
    auto_voice: pad_sustain
    register: low
    prominence: air
    humanize: {timing_ms: 2, velocity: 4}
    chain: {reverb_send: 0.55, compress: glue}

  keys:
    family: piano
    voice: lofi_rhodes_warm
    auto_voice: rhodes_comp
    register: mid
    prominence: support
    humanize: {timing_ms: 6, velocity: 8}
    chain: {reverb_send: 0.42, compress: gentle, tape_drive_db: 0.6}

  bass:
    family: bass
    voice: lofi_round_bass
    auto_voice: walking_bass
    register: low
    prominence: anchor
    humanize: {timing_ms: 5, velocity: 6}
    chain: {reverb_send: 0.10, compress: gentle, pan_offset: -0.05}

  kick:
    family: drums
    prominence: anchor
    humanize: {timing_ms: 4, velocity: 7}
    chain: {reverb_send: 0.06, compress: punchy}
    loop_bars: 4
    events:
      - {beat: 1.0,  pitch: "", dur: 0.25, vel: 102}
      - {beat: 9.0,  pitch: "", dur: 0.25, vel: 100}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.30, compress: punchy}
    loop_bars: 4
    events:
      - {beat: 5.0, pitch: "", dur: 0.30, vel: 92}
      - {beat: 13.0, pitch: "", dur: 0.30, vel: 90}

  hat:
    family: drums
    prominence: support
    humanize: {timing_ms: 2, velocity: 4}
    chain: {reverb_send: 0.14, compress: "off", pan_offset: 0.22}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.08, vel: 64}
      - {beat: 2.00, pitch: "", dur: 0.08, vel: 56}
      - {beat: 3.00, pitch: "", dur: 0.08, vel: 64}
      - {beat: 4.00, pitch: "", dur: 0.08, vel: 56}

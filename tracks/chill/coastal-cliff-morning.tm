title: Coastal Cliff Morning
description: SP19 chill journey in F major. Open pads + bright Rhodes + soft kit.
style: chill
substyle: open-air
mix_bus: chill
listen_mode: hour-stream
seed: 73001
tags: [chill, pad, bright, coastal, sp19]
key: Fmaj
tempo: 96
globals: {density: full, brightness: bright, motion: moving, reverb: halo}

form: chill_journey
total_duration: 7m

motif_library:
  cliff_theme:
    pattern: "3 . 5 6 | >2 . 6 5 | 3 . 1 5 | 6 5 3 1"
    description: "open major melody"
    bars: 4

roles:
  pad:
    family: pad
    voice: chill_pad_warm
    auto_voice: pad_crossfade
    register: mid
    prominence: air
    humanize: {timing_ms: 2, velocity: 4}
    chain: {reverb_send: 0.55, compress: glue}

  keys:
    family: piano
    voice: lofi_rhodes_warm
    auto_voice: rhodes_comp
    register: mid-high
    prominence: support
    humanize: {timing_ms: 5, velocity: 7}
    chain: {reverb_send: 0.45, compress: gentle, tape_drive_db: 0.4}

  bass:
    family: bass
    voice: lofi_round_bass
    auto_voice: pedal_root
    register: low
    prominence: anchor
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.12, compress: gentle}

  kick:
    family: drums
    prominence: anchor
    humanize: {timing_ms: 3, velocity: 7}
    chain: {reverb_send: 0.08, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.0, pitch: "", dur: 0.25, vel: 100}
      - {beat: 5.0, pitch: "", dur: 0.25, vel: 96}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.32, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 3.0, pitch: "", dur: 0.30, vel: 92}
      - {beat: 7.0, pitch: "", dur: 0.30, vel: 88}

  hat:
    family: drums
    prominence: support
    humanize: {timing_ms: 2, velocity: 4}
    chain: {reverb_send: 0.14, compress: "off", pan_offset: 0.20}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.08, vel: 68, art: accent}
      - {beat: 1.50, pitch: "", dur: 0.08, vel: 52}
      - {beat: 2.00, pitch: "", dur: 0.08, vel: 64}
      - {beat: 2.50, pitch: "", dur: 0.08, vel: 52}
      - {beat: 3.00, pitch: "", dur: 0.08, vel: 68, art: accent}
      - {beat: 3.50, pitch: "", dur: 0.08, vel: 52}
      - {beat: 4.00, pitch: "", dur: 0.08, vel: 64}
      - {beat: 4.50, pitch: "", dur: 0.08, vel: 52}

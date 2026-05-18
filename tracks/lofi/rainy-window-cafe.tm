title: Rainy Window Cafe
description: "SP19 lofi loop in D minor with actual rain ambience (texture: rain). Rhodes + walking sub-bass + brushed drums."
style: lofi
substyle: rainy-cafe
listen_mode: hour-stream
seed: 71001
tags: [lofi, rhodes, rain, brushes, sp19]
key: Dmin
tempo: 84
mix_bus: lofi
globals: {density: full, brightness: warm, motion: gentle, reverb: warm}

# SP19-D: real rain ambience replacing the "rain-like vinyl noise" misperception.
textures:
  - {name: rain, level_db: -36}
  - {name: vinyl, level_db: -44}

# SP18 form: lofi_loop_form. SP19-A: phrase dynamics auto-on for AABA sections.
# Pickup beats on loop_b lead the listener into the bridge.
form: lofi_loop_form
total_duration: 6m

motif_library:
  rhodes_theme:
    pattern: "5 . 7 5 | 3 . 5 3 | 7 . >2 7 | 5 . 3 1"
    description: "main 4-bar Rhodes motif — sigh-fall contour"
    bars: 4

roles:
  rhodes:
    family: piano
    voice: lofi_rhodes_warm
    auto_voice: rhodes_comp
    register: mid
    prominence: support
    humanize: {timing_ms: 6, velocity: 8, accent: dilla}
    chain: {reverb_send: 0.35, compress: gentle, tape_drive_db: 1.0}

  bass:
    family: bass
    voice: lofi_round_bass
    auto_voice: walking_bass
    register: low
    prominence: anchor
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.12, compress: gentle, pan_offset: -0.08}

  sub:
    family: synth_bass
    voice: lofi_round_bass
    auto_voice: pedal_root
    register: sub
    prominence: anchor
    humanize: {timing_ms: 3, velocity: 4}
    chain: {reverb_send: 0.04, compress: punchy, pan_offset: 0}

  kick:
    family: drums
    voice: lofi_dusty_kick
    prominence: anchor
    humanize: {timing_ms: 3, velocity: 8, accent: dilla}
    chain: {reverb_send: 0.06, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.00, pitch: "", dur: 0.25, vel: 112}
      - {beat: 1.75, pitch: "", dur: 0.25, vel: 96}
      - {beat: 3.00, pitch: "", dur: 0.25, vel: 104}
      - {beat: 3.75, pitch: "", dur: 0.25, vel: 88}
      - {beat: 5.00, pitch: "", dur: 0.25, vel: 110}
      - {beat: 5.75, pitch: "", dur: 0.25, vel: 96}
      - {beat: 7.00, pitch: "", dur: 0.25, vel: 100}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 4, velocity: 6, accent: dilla}
    chain: {reverb_send: 0.34, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 2.00, pitch: "", dur: 0.25, vel: 96}
      - {beat: 2.50, pitch: "", dur: 0.20, vel: 44, art: ghost}
      - {beat: 3.50, pitch: "", dur: 0.20, vel: 40, art: ghost}
      - {beat: 4.00, pitch: "", dur: 0.25, vel: 94}
      - {beat: 6.00, pitch: "", dur: 0.25, vel: 96}
      - {beat: 6.50, pitch: "", dur: 0.20, vel: 44, art: ghost}
      - {beat: 7.50, pitch: "", dur: 0.20, vel: 40, art: ghost}
      - {beat: 8.00, pitch: "", dur: 0.25, vel: 94}

  hat:
    family: drums
    prominence: support
    humanize: {timing_ms: 2, velocity: 4}
    chain: {reverb_send: 0.12, compress: "off", pan_offset: 0.25}
    loop_bars: 1
    events:
      - {beat: 1.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 1.5, pitch: "", dur: 0.1, vel: 52}
      - {beat: 2.0, pitch: "", dur: 0.1, vel: 68}
      - {beat: 2.5, pitch: "", dur: 0.1, vel: 76, art: accent}
      - {beat: 3.0, pitch: "", dur: 0.1, vel: 68}
      - {beat: 3.5, pitch: "", dur: 0.1, vel: 52}
      - {beat: 4.0, pitch: "", dur: 0.1, vel: 68}
      - {beat: 4.5, pitch: "", dur: 0.1, vel: 76, art: accent}

  pad:
    family: pad
    voice: lofi_pad_warm
    register: mid
    humanize: {timing_ms: 0, velocity: 0}
    chain: {reverb_send: 0.45, compress: "off"}

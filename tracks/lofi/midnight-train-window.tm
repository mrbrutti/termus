title: Midnight Train Window
description: SP19 lofi in E minor. Descending Rhodes motif, Dilla-late groove, vinyl ambience only.
style: lofi
substyle: dilla-late
listen_mode: hour-stream
seed: 71002
tags: [lofi, rhodes, train, dilla, descending, sp19]
key: Emin
tempo: 78
mix_bus: lofi
globals: {density: full, brightness: warm, motion: gentle, reverb: warm}

textures:
  - {name: vinyl, level_db: -38}

form: lofi_loop_form
total_duration: 5m

motif_library:
  descend_theme:
    pattern: ">2 . 7 5 | 3 . 1 7 | >2 . 7 5 | 3 1 7 5"
    description: "descending arc — train pulling away"
    bars: 4

roles:
  rhodes:
    family: piano
    voice: lofi_rhodes_warm
    auto_voice: rhodes_comp
    register: mid
    prominence: support
    humanize: {timing_ms: 8, velocity: 10, accent: dilla}
    chain: {reverb_send: 0.42, compress: gentle, tape_drive_db: 1.5}

  bass:
    family: bass
    voice: lofi_round_bass
    auto_voice: pedal_root
    register: low
    prominence: anchor
    humanize: {timing_ms: 5, velocity: 8, accent: dilla}
    chain: {reverb_send: 0.10, compress: gentle, pan_offset: -0.05}

  kick:
    family: drums
    voice: lofi_dusty_kick
    prominence: anchor
    humanize: {timing_ms: 6, velocity: 10, accent: dilla}
    chain: {reverb_send: 0.08, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.00, pitch: "", dur: 0.25, vel: 110}
      - {beat: 2.50, pitch: "", dur: 0.25, vel: 80, art: ghost}
      - {beat: 4.75, pitch: "", dur: 0.25, vel: 88}
      - {beat: 5.00, pitch: "", dur: 0.25, vel: 104}
      - {beat: 7.25, pitch: "", dur: 0.25, vel: 80}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 8, velocity: 10, accent: dilla}
    chain: {reverb_send: 0.38, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 3.00, pitch: "", dur: 0.30, vel: 100}
      - {beat: 7.00, pitch: "", dur: 0.30, vel: 98}
      - {beat: 7.75, pitch: "", dur: 0.20, vel: 50, art: ghost}

  hat:
    family: drums
    prominence: support
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.18, compress: "off", pan_offset: 0.20}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.10, vel: 72}
      - {beat: 1.75, pitch: "", dur: 0.10, vel: 55, art: ghost}
      - {beat: 2.50, pitch: "", dur: 0.10, vel: 68}
      - {beat: 3.25, pitch: "", dur: 0.10, vel: 60}
      - {beat: 4.00, pitch: "", dur: 0.10, vel: 76, art: accent}

  pad:
    family: pad
    voice: lofi_pad_warm
    register: low
    humanize: {timing_ms: 0, velocity: 0}
    chain: {reverb_send: 0.50, compress: "off"}

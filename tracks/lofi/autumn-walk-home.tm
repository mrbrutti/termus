title: Autumn Walk Home
description: SP19 lofi in A minor — brighter feel with auto-phrased saxophone melody floating above Rhodes & bass.
style: lofi
substyle: autumn
listen_mode: hour-stream
seed: 71004
tags: [lofi, sax, autumn, melody, sp19]
key: Amin
tempo: 90
mix_bus: lofi
globals: {density: full, brightness: balanced, motion: moving, reverb: warm}

textures:
  - {name: vinyl, level_db: -42}

form: lofi_loop_form
total_duration: 6m

motif_library:
  autumn_theme:
    pattern: "1 . 3 5 | 7 . >2 1 | 5 . 3 1 | 7 5 3 1"
    description: "questioning ascending arc"
    bars: 4

roles:
  sax:
    family: reed_lead
    voice: jazz_tenor_sax
    auto_phrase: question_answer
    register: mid-high
    prominence: lead
    humanize: {timing_ms: 12, velocity: 12, accent: phrase_arc, phrase_shape: arc}
    chain: {reverb_send: 0.45, compress: gentle}

  rhodes:
    family: piano
    voice: lofi_rhodes_warm
    auto_voice: rhodes_comp
    register: mid
    prominence: support
    humanize: {timing_ms: 6, velocity: 8, accent: dilla}
    chain: {reverb_send: 0.32, compress: gentle, tape_drive_db: 0.8}

  bass:
    family: bass
    voice: lofi_round_bass
    auto_voice: walking_bass
    register: low
    prominence: anchor
    humanize: {timing_ms: 5, velocity: 6}
    chain: {reverb_send: 0.10, compress: gentle, pan_offset: -0.06}

  kick:
    family: drums
    voice: lofi_dusty_kick
    prominence: anchor
    humanize: {timing_ms: 3, velocity: 8, accent: dilla}
    chain: {reverb_send: 0.05, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.00, pitch: "", dur: 0.25, vel: 110}
      - {beat: 3.00, pitch: "", dur: 0.25, vel: 100}
      - {beat: 5.00, pitch: "", dur: 0.25, vel: 108}
      - {beat: 6.50, pitch: "", dur: 0.25, vel: 90}
      - {beat: 7.50, pitch: "", dur: 0.25, vel: 92}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 4, velocity: 6, accent: dilla}
    chain: {reverb_send: 0.30, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 2.00, pitch: "", dur: 0.25, vel: 100}
      - {beat: 4.00, pitch: "", dur: 0.25, vel: 98}
      - {beat: 6.00, pitch: "", dur: 0.25, vel: 100}
      - {beat: 8.00, pitch: "", dur: 0.25, vel: 98}

  hat:
    family: drums
    prominence: support
    humanize: {timing_ms: 2, velocity: 5}
    chain: {reverb_send: 0.14, compress: "off", pan_offset: 0.25}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.08, vel: 76, art: accent}
      - {beat: 1.50, pitch: "", dur: 0.08, vel: 56}
      - {beat: 2.00, pitch: "", dur: 0.08, vel: 70}
      - {beat: 2.50, pitch: "", dur: 0.08, vel: 56}
      - {beat: 3.00, pitch: "", dur: 0.08, vel: 76, art: accent}
      - {beat: 3.50, pitch: "", dur: 0.08, vel: 56}
      - {beat: 4.00, pitch: "", dur: 0.08, vel: 70}
      - {beat: 4.50, pitch: "", dur: 0.08, vel: 56}

  pad:
    family: pad
    voice: lofi_pad_warm
    register: mid
    humanize: {timing_ms: 0, velocity: 0}
    chain: {reverb_send: 0.40, compress: "off"}

title: Midnight Ballad Slow
description: SP18 form-driven jazz ballad — AABA 32-bar with bridge, two solos and head out. ~5-6 minutes per pass.
style: jazz
mix_bus: jazz
listen_mode: hour-stream
seed: 62738
tags: [jazz, ballad, slow, piano, brushes, walking, sp18]
key: Cmaj
tempo: 96
globals: {density: heavy, brightness: warm, motion: moving, phrase: long}

# SP18 form: jazz_aaba_32bar — intro + 2 head choruses (AABA) + bridge + 2 solo choruses + head out + outro.
# Total: 8 + 8*4 + 8*2 + 8 + 8 = 72 bars @ 96 BPM = ~3m per pass; hour-stream repeats.
form: jazz_aaba_32bar
total_duration: 6m

motif_library:
  ballad_theme:
    pattern: "5 . 7 5 | 3 . 5 3 | 7 . >2 7 | 5 . 3 1"
    description: "main ballad melody — slow falling contour"
    bars: 4
  bridge_motif:
    pattern: "1 . 3 5 | 7 . 5 3"
    description: "bridge contrast — rising answer"
    bars: 2

roles:
  piano:
    family: piano
    voice: jazz_piano_warm
    auto_voice: rhodes_comp
    register: mid
    prominence: lead
    humanize: {timing_ms: 8, velocity: 8}
    chain: {reverb_send: 0.30, compress: gentle}

  bass:
    family: bass
    voice: jazz_upright_bass
    auto_voice: walking_bass
    register: low
    prominence: anchor
    humanize: {timing_ms: 5, velocity: 6}
    chain: {reverb_send: 0.18, compress: gentle, pan_offset: -0.08}

  kick:
    family: drums
    voice: jazz_kit_kick
    prominence: support
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.15, compress: gentle}
    loop_bars: 2
    events:
      - {beat: 1.00, pitch: "", dur: 0.25, vel: 60}
      - {beat: 3.00, pitch: "", dur: 0.25, vel: 56}
      - {beat: 5.00, pitch: "", dur: 0.25, vel: 60}
      - {beat: 7.00, pitch: "", dur: 0.25, vel: 56}

  snare:
    family: drums
    voice: jazz_kit_snare
    prominence: support
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.30, compress: gentle}
    loop_bars: 2
    events:
      - {beat: 2.50, pitch: "", dur: 0.2, vel: 36, art: ghost}
      - {beat: 4.50, pitch: "", dur: 0.2, vel: 36, art: ghost}
      - {beat: 6.50, pitch: "", dur: 0.2, vel: 36, art: ghost}
      - {beat: 8.50, pitch: "", dur: 0.2, vel: 36, art: ghost}

  hat:
    family: drums
    voice: jazz_kit_hat
    prominence: support
    humanize: {timing_ms: 3, velocity: 4}
    chain: {reverb_send: 0.20, compress: "off"}
    loop_bars: 1
    events:
      - {beat: 2.0, pitch: "", dur: 0.1, vel: 64}
      - {beat: 4.0, pitch: "", dur: 0.1, vel: 64}

  pad:
    family: pad
    voice: jazz_string_pad
    register: mid
    humanize: {timing_ms: 0, velocity: 0}
    chain: {reverb_send: 0.55, compress: "off"}

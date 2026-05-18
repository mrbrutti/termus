title: Bookstore After Rain
description: SP18 form-driven lofi — D minor, lofi_loop_form. Multi-section motif development, instrument arrangement arc, explicit section transitions.
style: lofi
substyle: piano-ballad
listen_mode: hour-stream
seed: 28011
tags: [lofi, piano, rain, dilla, sp18]
key: Dmin
tempo: 84
mix_bus: lofi
globals: {density: full, brightness: warm, motion: gentle, reverb: warm}

# SP18 multi-scale form: lofi_loop_form expands into intro/loop_a/loop_b/bridge/loop_c/outro.
# Total: 8 + 16 + 16 + 8 + 16 + 8 = 72 bars @ 84 BPM = ~3.4m per pass; hour-stream loops the
# pass to fill the hour.
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
    chain: {reverb_send: 0.28, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 2.00, pitch: "", dur: 0.25, vel: 102}
      - {beat: 2.75, pitch: "", dur: 0.25, vel: 44, art: ghost}
      - {beat: 3.50, pitch: "", dur: 0.25, vel: 40, art: ghost}
      - {beat: 4.00, pitch: "", dur: 0.25, vel: 100}
      - {beat: 6.00, pitch: "", dur: 0.25, vel: 102}
      - {beat: 6.50, pitch: "", dur: 0.25, vel: 44, art: ghost}
      - {beat: 7.50, pitch: "", dur: 0.25, vel: 40, art: ghost}
      - {beat: 8.00, pitch: "", dur: 0.25, vel: 100}

  hat:
    family: drums
    prominence: support
    humanize: {timing_ms: 2, velocity: 4}
    chain: {reverb_send: 0.10, compress: "off", pan_offset: 0.25}
    loop_bars: 1
    events:
      - {beat: 1.0, pitch: "", dur: 0.1, vel: 76}
      - {beat: 1.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 2.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 2.5, pitch: "", dur: 0.1, vel: 78, art: accent}
      - {beat: 3.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 3.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 4.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 4.5, pitch: "", dur: 0.1, vel: 78, art: accent}

  pad:
    family: pad
    voice: lofi_pad_warm
    register: mid
    humanize: {timing_ms: 0, velocity: 0}
    chain: {reverb_send: 0.45, compress: "off"}

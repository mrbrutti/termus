title: Bookstore After Rain
description: Felt-piano lofi in Dmin — intent-driven rhodes and walking bass, Dilla kit with ghost snares. SP16 reference.
style: lofi
substyle: piano-ballad
listen_mode: hour-stream
seed: 28011
tags: [lofi, piano, rain, dilla, sp16]
key: Dmin
tempo: 86
mix_bus: lofi
globals: {density: full, brightness: warm, motion: gentle, reverb: warm}

# SP16 intent-driven authoring: the engine generates idiomatic rhodes
# voicings and walking bass from the harmony; explicit drum events keep
# rhythmic clarity. Humanization applies to all events.

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

  hat_closed:
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

sections:
  - id: intro
    title: rain-on-glass
    duration: 12s
    harmony: "Dm9 Gm7 | Bb6 A7"
    scene: "intro hush"
    variation: "establish"
    groove: dilla_late
    intensity: 0.4
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.25}
          - {at: 100, value: 0.55}

  - id: verse
    title: paperback turn
    duration: 32s
    harmony: "Dm9 Gm7 | Bb6 A7"
    scene: "head glide"
    variation: "statement"
    groove: dilla_late
    intensity: 0.7
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.55}
          - {at: 50, value: 0.85}
          - {at: 100, value: 0.65}

  - id: bridge
    title: light through curtain
    duration: 18s
    harmony: "Bb6 A7 | Dm9 Gm7"
    scene: "bridge lift"
    variation: "open-register"
    groove: dilla_late
    intensity: 0.8

  - id: outro
    title: shelf-closing
    duration: 20s
    harmony: "Dm9 Gm7 | Bb6 A7"
    scene: "outro hush"
    variation: "cadence"
    groove: dilla_late
    intensity: 0.5
    fill_at_end: true
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 1.0}

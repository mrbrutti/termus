title: Dusty Swing / After Hours
description: F-major bop blues — intent-driven rootless piano, walking bass, spang-a-lang ride. SP16 reference.
style: jazz
mix_bus: jazz
listen_mode: hour-stream
seed: 31440
tags: [jazz, swing, bop, walking, sp16]
key: Fmaj
tempo: 138
globals: {density: full, brightness: bright, motion: restless, phrase: long}

# 12-bar F bop blues. Bass, piano, tenor are intent-driven; the drum kit is
# explicit because rhythm wants exact placement.

roles:
  bass:
    family: bass
    voice: jazz_upright_bass
    auto_voice: walking_with_anticipation
    register: low
    prominence: anchor
    humanize: {timing_ms: 5, velocity: 6}
    chain: {reverb_send: 0.15, compress: gentle, pan_offset: -0.05}

  piano:
    family: acoustic_piano
    voice: jazz_grand_piano
    auto_voice: drop2
    register: high
    prominence: support
    humanize: {timing_ms: 6, velocity: 8}
    chain: {reverb_send: 0.40, compress: gentle}

  tenor:
    family: reed_lead
    voice: jazz_tenor_sax
    auto_voice: bell_arpeggio
    register: high
    prominence: lead
    humanize: {timing_ms: 10, velocity: 10, accent: phrase_arc}
    chain: {reverb_send: 0.50, compress: gentle}

  ride:
    family: drums
    voice: jazz_ride_cymbal
    prominence: support
    humanize: {timing_ms: 4, velocity: 7, accent: swing_accent}
    chain: {reverb_send: 0.30, compress: "off", pan_offset: 0.15}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 1.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 2.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 2.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 3.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 3.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 4.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 4.67, pitch: "", dur: 0.15, vel: 70}

  kick:
    family: drums
    prominence: anchor
    humanize: {timing_ms: 3, velocity: 6}
    chain: {reverb_send: 0.05, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.0,  pitch: "", dur: 0.25, vel: 72}
      - {beat: 3.0,  pitch: "", dur: 0.25, vel: 64}
      - {beat: 5.0,  pitch: "", dur: 0.25, vel: 70}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.30, compress: punchy}
    loop_bars: 1
    events:
      - {beat: 2.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 4.0, pitch: "", dur: 0.25, vel: 84}
      - {beat: 3.67, pitch: "", dur: 0.25, vel: 42, art: ghost}

sections:
  - id: head-a
    title: opening chorus
    duration: 16s
    harmony: "F7 | Bb7 | F7 | Cm7 F7"
    scene: "head statement"
    variation: "statement"
    groove: swing_56
    intensity: 0.6

  - id: head-a2
    title: second chorus
    duration: 16s
    harmony: "Bb7 | Bdim7 | F7 D7 | Gm7 C7"
    scene: "head glide"
    variation: "sequence-up"
    groove: swing_56
    intensity: 0.75

  - id: bridge
    title: D7 lift
    duration: 8s
    harmony: "D7 | G7 C7"
    scene: "bridge lift"
    variation: "open-register"
    groove: swing_56
    intensity: 0.85
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 0.6}

  - id: out-head
    title: bar stools empty
    duration: 16s
    harmony: "F7 | Bb7 | F7 | Cm7 F7"
    scene: "outro cadence"
    variation: "cadence"
    groove: swing_56
    intensity: 0.55
    fill_at_end: true
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 0.6}

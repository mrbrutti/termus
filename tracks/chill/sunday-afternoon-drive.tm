title: Sunday Afternoon Drive
description: C-major chill — intent-driven pad sustain, Rhodes comp, half-time kit. SP16 reference.
style: chill
substyle: half-time-chill
mix_bus: chill
listen_mode: album-side
seed: 19334
tags: [chill, pad, halftime, sp16]
key: Cmaj
tempo: 100
globals: {density: full, brightness: balanced, motion: moving, reverb: halo}

# 4-bar harmonic loop, half-time feel. Pad and Rhodes are intent-driven;
# drum kit explicit for groove precision.

roles:
  pad:
    family: pad
    voice: chill_pad_warm
    auto_voice: pad_crossfade
    register: low
    prominence: air
    humanize: {timing_ms: 2, velocity: 4}
    chain: {reverb_send: 0.55, compress: glue}

  rhodes:
    family: electric_piano
    voice: lofi_rhodes_warm
    auto_voice: rhodes_comp
    register: mid
    prominence: support
    humanize: {timing_ms: 5, velocity: 7}
    chain: {reverb_send: 0.40, compress: gentle, tape_drive_db: 0.5}

  bass:
    family: bass
    voice: lofi_round_bass
    auto_voice: pedal_root
    register: low
    prominence: anchor
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.10, compress: gentle}

  kick:
    family: drums
    prominence: anchor
    humanize: {timing_ms: 3, velocity: 7}
    chain: {reverb_send: 0.06, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.0,  pitch: "", dur: 0.25, vel: 104}
      - {beat: 5.0,  pitch: "", dur: 0.25, vel: 96}
      - {beat: 6.5,  pitch: "", dur: 0.25, vel: 80}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.30, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 3.0,  pitch: "", dur: 0.3, vel: 96}
      - {beat: 7.0,  pitch: "", dur: 0.3, vel: 92}

  hat_closed:
    family: drums
    prominence: support
    humanize: {timing_ms: 2, velocity: 4}
    chain: {reverb_send: 0.12, compress: "off", pan_offset: 0.2}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.05, vel: 66, art: accent}
      - {beat: 1.50, pitch: "", dur: 0.05, vel: 50}
      - {beat: 2.00, pitch: "", dur: 0.05, vel: 60}
      - {beat: 2.50, pitch: "", dur: 0.05, vel: 50}
      - {beat: 3.00, pitch: "", dur: 0.05, vel: 66, art: accent}
      - {beat: 3.50, pitch: "", dur: 0.05, vel: 50}
      - {beat: 4.00, pitch: "", dur: 0.05, vel: 60}
      - {beat: 4.50, pitch: "", dur: 0.05, vel: 50}

  shaker:
    family: drums
    prominence: air
    humanize: {timing_ms: 2, velocity: 3}
    chain: {reverb_send: 0.15, compress: "off"}
    loop_bars: 1
    events:
      - {beat: 1.0, pitch: "", dur: 0.1, vel: 56}
      - {beat: 1.5, pitch: "", dur: 0.1, vel: 48}
      - {beat: 2.0, pitch: "", dur: 0.1, vel: 52}
      - {beat: 2.5, pitch: "", dur: 0.1, vel: 48}
      - {beat: 3.0, pitch: "", dur: 0.1, vel: 56}
      - {beat: 3.5, pitch: "", dur: 0.1, vel: 48}
      - {beat: 4.0, pitch: "", dur: 0.1, vel: 52}
      - {beat: 4.5, pitch: "", dur: 0.1, vel: 48}

sections:
  - id: intro
    title: window-open
    duration: 16s
    harmony: "Cmaj7 G/B | Am7 F | Dm7 G7 | Cmaj7 Am7"
    scene: "intro establish"
    variation: "establish"
    intensity: 0.4
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.3}
          - {at: 100, value: 0.6}

  - id: a-section
    title: lane glide
    duration: 32s
    harmony: "Cmaj7 G/B | Am7 F | Dm7 G7 | Cmaj7 Am7"
    scene: "head glide"
    variation: "statement"
    intensity: 0.7

  - id: b-section
    title: bridge lift
    duration: 16s
    harmony: "Fmaj7 Em7 | Dm7 G7 | Cmaj7 Am7 | Dm7 G7"
    scene: "bridge lift"
    variation: "sequence-up"
    intensity: 0.85

  - id: a-out
    title: cruise out
    duration: 32s
    harmony: "Cmaj7 G/B | Am7 F | Dm7 G7 | Cmaj7 Cmaj7"
    scene: "outro cadence"
    variation: "cadence"
    intensity: 0.55
    fill_at_end: true

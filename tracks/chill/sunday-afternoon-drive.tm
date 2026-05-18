title: Sunday Afternoon Drive
description: SP18 form-driven chill journey — C major, chill_ababcb. Multi-section motif development, instrument curve, transitions.
style: chill
substyle: half-time-chill
mix_bus: chill
listen_mode: hour-stream
seed: 19334
tags: [chill, pad, halftime, sp18]
key: Cmaj
tempo: 92
globals: {density: full, brightness: balanced, motion: moving, reverb: halo}

# SP18 form: chill_ababcb — emerge + verse B1 + verse A2 + verse B2 + bridge C + return B.
# Total: 8 + 16 + 16 + 16 + 12 + 16 = 84 bars @ 92 BPM = ~3.7m per pass; hour-stream loops the pass.
form: chill_ababcb
total_duration: 7m

motif_library:
  drive_theme:
    pattern: "3 . 5 3 | 1 . 3 5 | 7 . 5 3 | 1 . 3 1"
    description: "main drifting melody — gentle arc"
    bars: 4

roles:
  pad:
    family: pad
    voice: chill_pad_warm
    auto_voice: pad_crossfade
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

  hat:
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

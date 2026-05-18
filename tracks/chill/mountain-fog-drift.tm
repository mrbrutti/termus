title: Mountain Fog Drift
description: SP19 chill ABABCB in A minor. Cooler tonality showcases plate reverb.
style: chill
substyle: cool
mix_bus: chill
listen_mode: hour-stream
seed: 73002
tags: [chill, pad, plate-reverb, fog, sp19]
key: Amin
tempo: 88
globals: {density: full, brightness: balanced, motion: gentle, reverb: halo}

form: chill_ababcb
total_duration: 7m

motif_library:
  fog_theme:
    pattern: "1 . b3 5 | b6 . 5 b3 | b7 . 5 b3 | 1 . . ."
    description: "minor descending — slow fog"
    bars: 4

roles:
  pad:
    family: pad
    voice: chill_pad_warm
    auto_voice: pad_crossfade
    register: low
    prominence: air
    humanize: {timing_ms: 2, velocity: 4}
    chain: {reverb_send: 0.60, compress: glue}

  keys:
    family: piano
    voice: lofi_rhodes_warm
    auto_voice: shell_voicing
    register: mid
    prominence: support
    humanize: {timing_ms: 6, velocity: 8}
    chain: {reverb_send: 0.50, compress: gentle}

  bass:
    family: bass
    voice: lofi_round_bass
    auto_voice: pedal_root
    register: low
    prominence: anchor
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.14, compress: gentle, pan_offset: -0.04}

  kick:
    family: drums
    prominence: anchor
    humanize: {timing_ms: 3, velocity: 6}
    chain: {reverb_send: 0.10, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.0, pitch: "", dur: 0.25, vel: 92}
      - {beat: 5.0, pitch: "", dur: 0.25, vel: 88}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.36, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 3.0, pitch: "", dur: 0.30, vel: 84}
      - {beat: 7.0, pitch: "", dur: 0.30, vel: 80}

  hat:
    family: drums
    prominence: support
    humanize: {timing_ms: 2, velocity: 4}
    chain: {reverb_send: 0.18, compress: "off", pan_offset: 0.22}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.08, vel: 60}
      - {beat: 2.50, pitch: "", dur: 0.08, vel: 50}
      - {beat: 4.00, pitch: "", dur: 0.08, vel: 60}

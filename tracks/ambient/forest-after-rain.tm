title: Forest After Rain
description: SP19 ambient palindrome in E minor. Rain texture at low level, choir pad, soft chimes.
style: ambient
mix_bus: ambient
listen_mode: hour-stream
seed: 74002
tags: [ambient, palindrome, rain, forest, choir, sp19]
key: Emin
tempo: 65
globals: {density: full, brightness: balanced, motion: slow, reverb: cathedral}

textures:
  - {name: rain, level_db: -42}

form: ambient_palindrome
total_duration: 10m

motif_library:
  forest_chime:
    pattern: "5 . . . . . 3 . | . . 1 . . . . . | b7 . . . . . . . | 5 . . . . . . ."
    description: "scattered chime — leaves dripping"
    bars: 8

roles:
  pad:
    family: pad
    voice: ambient_pad_dark
    auto_voice: pad_sustain
    register: low
    prominence: anchor
    humanize: {timing_ms: 0, velocity: 3}
    chain: {reverb_send: 0.60, compress: glue}

  choir:
    family: choir
    voice: ambient_drone_choir
    auto_voice: pad_crossfade
    register: mid
    prominence: support
    humanize: {timing_ms: 0, velocity: 3}
    chain: {reverb_send: 0.65, compress: "off"}

  bells:
    family: bells
    voice: bell_celesta
    auto_voice: bell_arpeggio
    register: high
    prominence: lead
    humanize: {timing_ms: 0, velocity: 6}
    chain: {reverb_send: 0.55, compress: "off"}

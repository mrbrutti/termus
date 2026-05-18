title: Deep Sea Cathedral
description: SP19 ambient long-form in C minor. Cathedral reverb, 4 drone layers, bell motif.
style: ambient
mix_bus: ambient
listen_mode: hour-stream
seed: 74001
tags: [ambient, drone, cathedral, deep, bells, sp19]
key: Cmin
tempo: 60
globals: {density: full, brightness: balanced, motion: slow, reverb: cathedral}

form: ambient_emerge_drift_recede
total_duration: 12m

motif_library:
  cathedral_bell:
    pattern: "1 . . . . . . . | 5 . . . . . . . | b3 . . . . . . . | 1 . . . . . . ."
    description: "sparse cathedral bell strikes"
    bars: 8

roles:
  pad:
    family: pad
    voice: ambient_pad_dark
    auto_voice: pad_sustain
    register: low
    prominence: anchor
    humanize: {timing_ms: 0, velocity: 4}
    chain: {reverb_send: 0.70, compress: glue}

  drone:
    family: pad
    voice: ambient_drone_choir
    auto_voice: pad_crossfade
    register: low
    prominence: support
    humanize: {timing_ms: 0, velocity: 3}
    chain: {reverb_send: 0.70, compress: "off"}

  strings:
    family: strings
    voice: ambient_strings_soft
    auto_voice: pad_crossfade
    register: mid
    prominence: support
    humanize: {timing_ms: 0, velocity: 3}
    chain: {reverb_send: 0.65, compress: "off"}

  bells:
    family: bells
    voice: bell_struck_bright
    auto_voice: bell_arpeggio
    register: high
    prominence: lead
    personality: bell_struck
    room: cathedral_large
    reverb_send_db: -8
    humanize: {timing_ms: 0, velocity: 6}
    chain: {reverb_send: 0.70, compress: "off"}

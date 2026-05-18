title: Glacial Slow Drift
description: SP19 ambient in A minor — slowest tempo, dark drone, very slow chord motion.
style: ambient
mix_bus: ambient
listen_mode: hour-stream
seed: 74003
tags: [ambient, drone, glacial, slow, sp19]
key: Amin
tempo: 55
globals: {density: sparse, brightness: balanced, motion: slow, reverb: cathedral}

form: ambient_emerge_drift_recede
total_duration: 14m

motif_library:
  glacial_hint:
    pattern: "1 . . . . . . . | . . . . . . . . | b3 . . . . . . . | . . . . . . . ."
    description: "very sparse hint — like ice cracking once"
    bars: 8

roles:
  drone:
    family: pad
    voice: ambient_pad_dark
    auto_voice: pad_sustain
    register: sub
    prominence: anchor
    humanize: {timing_ms: 0, velocity: 3}
    chain: {reverb_send: 0.70, compress: glue}

  strings:
    family: strings
    voice: ambient_strings_soft
    auto_voice: pad_crossfade
    register: low
    prominence: support
    humanize: {timing_ms: 0, velocity: 3}
    chain: {reverb_send: 0.65, compress: "off"}

  pad:
    family: pad
    voice: ambient_pad_dark
    auto_voice: pad_crossfade
    register: low
    prominence: support
    humanize: {timing_ms: 0, velocity: 3}
    chain: {reverb_send: 0.70, compress: "off"}

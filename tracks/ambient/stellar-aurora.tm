title: Stellar Aurora
description: SP19 ambient palindrome in F# minor — bell_arpeggio prominent over slow pads.
style: ambient
mix_bus: ambient
listen_mode: hour-stream
seed: 74004
tags: [ambient, palindrome, bells, aurora, sp19]
key: F#min
tempo: 62
globals: {density: full, brightness: bright, motion: slow, reverb: cathedral}

form: ambient_palindrome
total_duration: 10m

motif_library:
  aurora_bell:
    pattern: "1 . 5 . . 3 . . | . 5 . . . . 7 . | 1 . 3 . 5 . . . | . . . . . . . ."
    description: "shimmering bell cascade"
    bars: 8

roles:
  pad:
    family: pad
    voice: ambient_pad_dark
    auto_voice: pad_sustain
    register: mid
    prominence: support
    humanize: {timing_ms: 0, velocity: 3}
    chain: {reverb_send: 0.65, compress: glue}

  bells:
    family: bells
    voice: bell_struck_bright
    auto_voice: bell_arpeggio
    register: high
    prominence: lead
    personality: bell_struck
    room: cathedral_large
    reverb_send_db: -6
    humanize: {timing_ms: 0, velocity: 8}
    chain: {reverb_send: 0.70, compress: "off"}

  celesta:
    family: bells
    voice: bell_celesta
    auto_voice: bell_arpeggio
    register: high
    prominence: support
    humanize: {timing_ms: 0, velocity: 6}
    chain: {reverb_send: 0.65, compress: "off"}

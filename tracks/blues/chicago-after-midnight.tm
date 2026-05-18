title: Chicago After Midnight
description: SP19 blues in G major. Faster, energetic. Built on style=jazz, mix_bus=jazz. Tenor solos.
style: blues
substyle: chicago-blues
mix_bus: jazz
listen_mode: hour-stream
seed: 75003
tags: [blues, chicago, fast, g-major, sp19]
key: Gmaj
tempo: 120
globals: {density: full, brightness: bright, motion: restless, phrase: long}

form: jazz_blues_12bar
total_duration: 6m

motif_library:
  chicago_lick:
    pattern: "1 b3 5 b7 | >2 b3 . 1 | b7 5 b3 1 | . . . ."
    description: "chicago-style call & response — punchy"
    bars: 4

roles:
  bass:
    family: bass
    voice: jazz_upright_bass
    auto_voice: walking_with_anticipation
    register: low
    prominence: anchor
    humanize: {timing_ms: 4, velocity: 8}
    chain: {reverb_send: 0.18, compress: gentle, pan_offset: -0.05}

  piano:
    family: acoustic_piano
    voice: jazz_grand_piano
    auto_voice: shell_voicing
    register: mid-high
    prominence: support
    humanize: {timing_ms: 6, velocity: 10}
    chain: {reverb_send: 0.38, compress: gentle}

  tenor:
    family: reed_lead
    voice: jazz_tenor_sax
    auto_phrase: call_response
    register: high
    prominence: lead
    humanize: {timing_ms: 12, velocity: 14, accent: phrase_arc}
    chain: {reverb_send: 0.50, compress: gentle}

  ride:
    family: drums
    voice: jazz_ride_cymbal
    prominence: support
    humanize: {timing_ms: 5, velocity: 7, accent: swing_accent}
    chain: {reverb_send: 0.30, compress: "off", pan_offset: 0.15}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.25, vel: 92}
      - {beat: 1.67, pitch: "", dur: 0.15, vel: 76}
      - {beat: 2.00, pitch: "", dur: 0.25, vel: 84}
      - {beat: 2.67, pitch: "", dur: 0.15, vel: 76}
      - {beat: 3.00, pitch: "", dur: 0.25, vel: 92}
      - {beat: 3.67, pitch: "", dur: 0.15, vel: 76}
      - {beat: 4.00, pitch: "", dur: 0.25, vel: 84}
      - {beat: 4.67, pitch: "", dur: 0.15, vel: 76}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 5, velocity: 7, accent: swing_accent}
    chain: {reverb_send: 0.30, compress: punchy}
    loop_bars: 1
    events:
      - {beat: 2.0, pitch: "", dur: 0.25, vel: 90}
      - {beat: 4.0, pitch: "", dur: 0.25, vel: 88}

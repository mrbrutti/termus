title: Bourbon Street Blues
description: SP19 jazz 12-bar blues in F major. Tenor + piano + walking bass + ride.
style: jazz
substyle: blues-jazz
mix_bus: jazz
listen_mode: hour-stream
seed: 72003
tags: [jazz, blues, walking, bourbon, sp19]
key: Fmaj
tempo: 138
globals: {density: full, brightness: bright, motion: restless, phrase: long}

form: jazz_blues_12bar
total_duration: 5m

motif_library:
  blues_lick:
    pattern: "1 b3 4 b5 | 5 . . 4 | b3 . 1 . | . . . ."
    description: "classic blues lick — minor pentatonic"
    bars: 4

roles:
  bass:
    family: bass
    voice: jazz_upright_bass
    auto_voice: walking_with_anticipation
    register: low
    prominence: anchor
    humanize: {timing_ms: 5, velocity: 6}
    chain: {reverb_send: 0.20, compress: gentle, pan_offset: -0.05}

  piano:
    family: acoustic_piano
    voice: jazz_grand_piano
    auto_voice: shell_voicing
    register: mid-high
    prominence: support
    humanize: {timing_ms: 7, velocity: 8}
    chain: {reverb_send: 0.40, compress: gentle}

  tenor:
    family: reed_lead
    voice: jazz_tenor_sax
    auto_phrase: blues_lick
    register: high
    prominence: lead
    humanize: {timing_ms: 12, velocity: 12, accent: phrase_arc}
    chain: {reverb_send: 0.48, compress: gentle}

  ride:
    family: drums
    voice: jazz_ride_cymbal
    prominence: support
    humanize: {timing_ms: 5, velocity: 7, accent: swing_accent}
    chain: {reverb_send: 0.32, compress: "off", pan_offset: 0.15}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.25, vel: 90}
      - {beat: 1.67, pitch: "", dur: 0.15, vel: 74}
      - {beat: 2.00, pitch: "", dur: 0.25, vel: 84}
      - {beat: 2.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 3.00, pitch: "", dur: 0.25, vel: 90}
      - {beat: 3.67, pitch: "", dur: 0.15, vel: 74}
      - {beat: 4.00, pitch: "", dur: 0.25, vel: 84}
      - {beat: 4.67, pitch: "", dur: 0.15, vel: 72}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 5, velocity: 6, accent: swing_accent}
    chain: {reverb_send: 0.30, compress: punchy}
    loop_bars: 1
    events:
      - {beat: 2.0, pitch: "", dur: 0.25, vel: 86}
      - {beat: 3.67, pitch: "", dur: 0.20, vel: 44, art: ghost}
      - {beat: 4.0, pitch: "", dur: 0.25, vel: 84}

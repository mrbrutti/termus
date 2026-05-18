title: Autumn Leaves / After Hours
description: SP19 jazz AABA in E minor → G major bridge. Walking bass + spang-a-lang ride + comping piano + tenor lead.
style: jazz
substyle: aaba-ballad
mix_bus: jazz
listen_mode: hour-stream
seed: 72001
tags: [jazz, aaba, walking, sp19]
key: Emin
tempo: 130
globals: {density: full, brightness: bright, motion: restless, phrase: long}

form: jazz_aaba_32bar
total_duration: 6m

motif_library:
  leaves_theme:
    pattern: "5 . 7 5 | 3 . 5 3 | b7 . 5 3 | 1 . . ."
    description: "stepwise descending arc"
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
    auto_voice: drop2
    register: mid-high
    prominence: support
    humanize: {timing_ms: 7, velocity: 8}
    chain: {reverb_send: 0.42, compress: gentle}

  tenor:
    family: reed_lead
    voice: jazz_tenor_sax
    auto_phrase: question_answer
    register: high
    prominence: lead
    humanize: {timing_ms: 10, velocity: 10, accent: phrase_arc}
    chain: {reverb_send: 0.50, compress: gentle}

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
      - {beat: 2.00, pitch: "", dur: 0.25, vel: 82}
      - {beat: 2.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 3.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 3.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 4.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 4.67, pitch: "", dur: 0.15, vel: 70}

  kick:
    family: drums
    prominence: anchor
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.05, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.0, pitch: "", dur: 0.25, vel: 70}
      - {beat: 3.0, pitch: "", dur: 0.25, vel: 64}
      - {beat: 5.0, pitch: "", dur: 0.25, vel: 70}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 5, velocity: 6}
    chain: {reverb_send: 0.30, compress: punchy}
    loop_bars: 1
    events:
      - {beat: 2.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 4.0, pitch: "", dur: 0.25, vel: 84}
      - {beat: 3.67, pitch: "", dur: 0.20, vel: 42, art: ghost}

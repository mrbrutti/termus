title: Blue Bossa / Late Set
description: SP19 jazz with bossa feel in C minor. Lighter drums, Rhodes comping, tenor lead.
style: jazz
substyle: bossa
mix_bus: jazz
listen_mode: hour-stream
seed: 72002
tags: [jazz, bossa, latin, sp19]
key: Cmin
tempo: 124
globals: {density: full, brightness: balanced, motion: moving, reverb: warm}

form: chill_ababcb
total_duration: 6m

motif_library:
  bossa_theme:
    pattern: "5 . 7 . | b3 . 5 . | b7 . . 5 | 3 . 1 ."
    description: "bossa melody — relaxed minor"
    bars: 4

roles:
  rhodes:
    family: piano
    voice: lofi_rhodes_warm
    auto_voice: rhodes_comp
    register: mid
    prominence: support
    humanize: {timing_ms: 6, velocity: 8}
    chain: {reverb_send: 0.40, compress: gentle, tape_drive_db: 0.5}

  bass:
    family: bass
    voice: jazz_upright_bass
    auto_voice: pedal_root
    register: low
    prominence: anchor
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.15, compress: gentle, pan_offset: -0.05}

  tenor:
    family: reed_lead
    voice: jazz_tenor_sax
    auto_phrase: descending_arc
    register: mid-high
    prominence: lead
    humanize: {timing_ms: 10, velocity: 10, accent: phrase_arc}
    chain: {reverb_send: 0.42, compress: gentle}

  kick:
    family: drums
    prominence: anchor
    humanize: {timing_ms: 3, velocity: 6}
    chain: {reverb_send: 0.05, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.00, pitch: "", dur: 0.25, vel: 92}
      - {beat: 2.50, pitch: "", dur: 0.25, vel: 80}
      - {beat: 5.00, pitch: "", dur: 0.25, vel: 90}
      - {beat: 6.50, pitch: "", dur: 0.25, vel: 78}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.34, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.75, pitch: "", dur: 0.20, vel: 56, art: ghost}
      - {beat: 3.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 4.25, pitch: "", dur: 0.20, vel: 52, art: ghost}
      - {beat: 7.00, pitch: "", dur: 0.25, vel: 80}

  hat:
    family: drums
    prominence: support
    humanize: {timing_ms: 2, velocity: 4}
    chain: {reverb_send: 0.16, compress: "off", pan_offset: 0.25}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.10, vel: 72}
      - {beat: 1.50, pitch: "", dur: 0.10, vel: 56}
      - {beat: 2.00, pitch: "", dur: 0.10, vel: 68}
      - {beat: 2.50, pitch: "", dur: 0.10, vel: 56}
      - {beat: 3.00, pitch: "", dur: 0.10, vel: 72}
      - {beat: 3.50, pitch: "", dur: 0.10, vel: 56}
      - {beat: 4.00, pitch: "", dur: 0.10, vel: 68}
      - {beat: 4.50, pitch: "", dur: 0.10, vel: 56}

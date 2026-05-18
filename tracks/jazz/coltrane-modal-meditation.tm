title: Coltrane Modal Meditation
description: SP19 jazz modal — D dorian, minimal chord changes, big tonal center. Head-solo-head form.
style: jazz
substyle: modal
mix_bus: jazz
listen_mode: hour-stream
seed: 72004
tags: [jazz, modal, coltrane, dorian, sp19]
key: Dmin
tempo: 110
globals: {density: full, brightness: balanced, motion: moving, reverb: warm}

form: jazz_head_solo_head
total_duration: 8m

motif_library:
  modal_theme:
    pattern: "1 . 3 5 | b7 5 3 1 | 5 . b7 1 | 3 1 . ."
    description: "modal D dorian — repeated cell"
    bars: 4

roles:
  bass:
    family: bass
    voice: jazz_upright_bass
    auto_voice: pedal_root
    register: low
    prominence: anchor
    humanize: {timing_ms: 5, velocity: 6}
    chain: {reverb_send: 0.20, compress: gentle, pan_offset: -0.05}

  piano:
    family: acoustic_piano
    voice: jazz_grand_piano
    auto_voice: rhodes_comp
    register: mid
    prominence: support
    humanize: {timing_ms: 7, velocity: 8}
    chain: {reverb_send: 0.45, compress: gentle}

  tenor:
    family: reed_lead
    voice: jazz_tenor_sax
    auto_phrase: modal_drift
    register: high
    prominence: lead
    humanize: {timing_ms: 12, velocity: 12, accent: phrase_arc}
    chain: {reverb_send: 0.55, compress: gentle}

  ride:
    family: drums
    voice: jazz_ride_cymbal
    prominence: support
    humanize: {timing_ms: 4, velocity: 6, accent: swing_accent}
    chain: {reverb_send: 0.32, compress: "off", pan_offset: 0.15}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 1.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 2.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 2.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 3.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 3.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 4.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 4.67, pitch: "", dur: 0.15, vel: 70}

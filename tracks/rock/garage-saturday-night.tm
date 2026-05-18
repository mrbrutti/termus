title: Garage Saturday Night
description: SP19 rock in A minor — no-code-change test. style=lofi (only available kit), mix_bus=lofi (tape sat for grit). Driving bass + piano stabs (guitar substitute) + punchy drums.
style: rock
substyle: rock-garage
mix_bus: lofi
listen_mode: hour-stream
seed: 76001
tags: [rock, garage, driving, sp19]
key: Amin
tempo: 132
globals: {density: full, brightness: bright, motion: restless, phrase: long}

# Rock-specific gaps documented in SP19 report:
# - No distorted guitar voice in voice_library — using jazz_grand_piano as power-chord stand-in
# - No rock_form template — using chill_journey as closest available
# - No power_chord voicing — shell_voicing approximates root+3+7
# - Drums are lofi/dusty rather than punchy rock kit

form: chill_journey
total_duration: 6m

motif_library:
  rock_riff:
    pattern: "1 . b3 5 | 1 . b7 5 | 1 . b3 5 | 1 . . ."
    description: "driving minor rock riff"
    bars: 4

roles:
  piano:
    family: acoustic_piano
    voice: jazz_grand_piano
    auto_voice: shell_voicing
    register: mid
    prominence: support
    humanize: {timing_ms: 5, velocity: 12}
    chain: {reverb_send: 0.18, compress: punchy, tape_drive_db: 3.0}

  bass:
    family: bass
    voice: lofi_round_bass
    auto_voice: pedal_root
    register: low
    prominence: lead
    humanize: {timing_ms: 4, velocity: 10}
    chain: {reverb_send: 0.10, compress: punchy, pan_offset: -0.05}

  kick:
    family: drums
    voice: lofi_dusty_kick
    prominence: anchor
    humanize: {timing_ms: 2, velocity: 8}
    chain: {reverb_send: 0.06, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.0, pitch: "", dur: 0.25, vel: 116}
      - {beat: 3.0, pitch: "", dur: 0.25, vel: 108}
      - {beat: 5.0, pitch: "", dur: 0.25, vel: 116}
      - {beat: 7.0, pitch: "", dur: 0.25, vel: 108}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 3, velocity: 8}
    chain: {reverb_send: 0.28, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 2.0, pitch: "", dur: 0.30, vel: 110}
      - {beat: 4.0, pitch: "", dur: 0.30, vel: 108}
      - {beat: 6.0, pitch: "", dur: 0.30, vel: 110}
      - {beat: 8.0, pitch: "", dur: 0.30, vel: 108}

  hat:
    family: drums
    prominence: support
    humanize: {timing_ms: 2, velocity: 6}
    chain: {reverb_send: 0.10, compress: "off", pan_offset: 0.20}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.08, vel: 84}
      - {beat: 1.50, pitch: "", dur: 0.08, vel: 70}
      - {beat: 2.00, pitch: "", dur: 0.08, vel: 84}
      - {beat: 2.50, pitch: "", dur: 0.08, vel: 70}
      - {beat: 3.00, pitch: "", dur: 0.08, vel: 84}
      - {beat: 3.50, pitch: "", dur: 0.08, vel: 70}
      - {beat: 4.00, pitch: "", dur: 0.08, vel: 84}
      - {beat: 4.50, pitch: "", dur: 0.08, vel: 70}

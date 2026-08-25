title: Slow Burn Sunday
description: |
  Chill/pop composition built to commercial hit-formula discipline:
  108 BPM, hook lands at 26.7s, bridge at ~61% through, total 2:40.
  I–V–vi–IV chorus progression; vi–IV–I–V verse for tension; bridge
  modulates to IV. Verse-quiet to chorus-loud dynamic arc.
style: chill
substyle: pop-discipline
listen_mode: hour-stream
seed: 108001
tags: [chill, pop, hit-formula, demonstration]
key: Cmaj
tempo: 108
mix_bus: chill
globals: {density: full, brightness: balanced, motion: gentle, reverb: warm}

textures:
  - {name: room_tone, level_db: -44}

total_duration: 2m40s

# Single 4-bar hook motif that becomes the chorus theme.
# Treatment differs per section (hint -> introduce -> develop -> fragment -> return)
# to create the repetition+variation the persona prescribes.
motif_library:
  hook_main:
    pattern: "5 . 3 5 | 1> . 7 5 | 3 . 5 1> | 7 5 3 1"
    description: "ascending-then-resolving hook landing on tonic"
    bars: 4

roles:
  rhodes:
    family: piano
    voice: lofi_rhodes_warm
    auto_voice: rhodes_comp
    register: mid
    prominence: lead
    humanize: {timing_ms: 6, velocity: 8, accent: phrase_arc, phrase_shape: arc}
    chain: {reverb_send: 0.30, compress: gentle, tape_drive_db: 0.6}

  pad:
    family: pad
    voice: chill_pad_warm
    auto_voice: pad_crossfade
    register: mid
    prominence: support
    humanize: {timing_ms: 0, velocity: 0}
    chain: {reverb_send: 0.45, compress: "off"}

  bass:
    family: bass
    voice: lofi_round_bass
    auto_voice: walking_with_anticipation
    register: low
    prominence: anchor
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.12, compress: gentle, pan_offset: -0.05}

  sax:
    family: reed_lead
    voice: jazz_tenor_sax
    auto_phrase: ascending_arc
    register: mid-high
    prominence: lead
    humanize: {timing_ms: 12, velocity: 12, accent: phrase_arc, phrase_shape: arc}
    chain: {reverb_send: 0.50, compress: gentle}

  kick:
    family: drums
    voice: lofi_dusty_kick
    prominence: anchor
    humanize: {timing_ms: 3, velocity: 8, accent: dilla}
    chain: {reverb_send: 0.05, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 1.00, pitch: "", dur: 0.25, vel: 112}
      - {beat: 3.00, pitch: "", dur: 0.25, vel: 100}
      - {beat: 4.75, pitch: "", dur: 0.25, vel: 88}
      - {beat: 5.00, pitch: "", dur: 0.25, vel: 110}
      - {beat: 7.00, pitch: "", dur: 0.25, vel: 100}

  snare:
    family: drums
    prominence: support
    humanize: {timing_ms: 4, velocity: 6}
    chain: {reverb_send: 0.32, compress: punchy}
    loop_bars: 2
    events:
      - {beat: 2.00, pitch: "", dur: 0.25, vel: 100}
      - {beat: 4.00, pitch: "", dur: 0.25, vel: 100}
      - {beat: 4.50, pitch: "", dur: 0.20, vel: 42, art: ghost}
      - {beat: 6.00, pitch: "", dur: 0.25, vel: 100}
      - {beat: 8.00, pitch: "", dur: 0.25, vel: 100}

  hat:
    family: drums
    prominence: support
    humanize: {timing_ms: 2, velocity: 5}
    chain: {reverb_send: 0.14, compress: "off", pan_offset: 0.22}
    loop_bars: 1
    events:
      - {beat: 1.00, pitch: "", dur: 0.08, vel: 78, art: accent}
      - {beat: 1.50, pitch: "", dur: 0.08, vel: 56}
      - {beat: 2.00, pitch: "", dur: 0.08, vel: 72}
      - {beat: 2.50, pitch: "", dur: 0.08, vel: 56}
      - {beat: 3.00, pitch: "", dur: 0.08, vel: 78, art: accent}
      - {beat: 3.50, pitch: "", dur: 0.08, vel: 56}
      - {beat: 4.00, pitch: "", dur: 0.08, vel: 72}
      - {beat: 4.50, pitch: "", dur: 0.08, vel: 56}

sections:
  # 4 bars × 2.22s/bar = 8.89s. Pad + bass only; no drums, no Rhodes. Hook absent.
  - id: intro
    role: emerge
    bars: 4
    arrangement:
      pad: {enter_bar: 1}
      bass: {enter_bar: 3}
    dynamic_curve: crescendo
    harmony: "Cmaj7 | Cmaj7 | Am7 | Am7"
    transition_to_next: pickup

  # 8 bars × 2.22s = 17.78s. Verse (low intensity). Rhodes hints at hook with
  # the "hint" motif treatment. Drums enter quietly.
  - id: verse1
    role: head_statement
    bars: 8
    arrangement:
      rhodes: {enter_bar: 1, fade_in_bars: 2}
      kick: {enter_bar: 3}
      hat: {enter_bar: 5}
    motif: hook_main
    motif_treatment: hint
    dynamic_curve: arc
    # vi-IV-I-V — minor-flavored opening for tension before the chorus release
    harmony: "Am7 Fmaj7 | Cmaj7 G | Am7 Fmaj7 | Cmaj7 G"
    transition_to_next: turnaround

  # CHORUS — hits at t ≈ 26.7s (just under the 30s persona rule).
  # 12 bars × 2.22s = 26.67s. Full ensemble. Hook stated explicitly.
  # I-V-vi-IV — the proven "Axis of Awesome" pop progression.
  - id: chorus1
    role: chorus
    bars: 12
    arrangement:
      sax: {enter_bar: 1, prominent: true}
      snare: {enter_bar: 1}
    motif: hook_main
    motif_treatment: introduce
    dynamic_curve: arc
    harmony: "Cmaj7 G | Am7 Fmaj7 | Cmaj7 G | Am7 Fmaj7 | Cmaj7 G | Am7 Fmaj7"
    transition_to_next: pickup

  # Verse 2 — same as verse 1 structurally but motif moves to "develop"
  - id: verse2
    role: head_variation
    bars: 8
    motif: hook_main
    motif_treatment: develop
    dynamic_curve: arc
    harmony: "Am7 Fmaj7 | Cmaj7 G | Am7 Fmaj7 | Cmaj7 G"
    transition_to_next: turnaround

  # Chorus 2 — repeat of chorus 1 with slight variation via "develop"
  - id: chorus2
    role: chorus
    bars: 12
    motif: hook_main
    motif_treatment: develop
    dynamic_curve: arc
    harmony: "Cmaj7 G | Am7 Fmaj7 | Cmaj7 G | Am7 Fmaj7 | Cmaj7 G | Am7 Fmaj7"
    transition_to_next: breakdown

  # BRIDGE at t = 8.89 + 17.78 + 26.67 + 17.78 + 26.67 = 97.79s of total ~160s = 61%
  # Close to the 67% rule. Modulates to IV (F major) for contrast.
  # Dynamics drop: drums out, sax + bass + pad only.
  - id: bridge
    role: contrast
    bars: 8
    arrangement:
      kick: {exit_bar: 1}
      snare: {exit_bar: 1}
      hat: {exit_bar: 1}
      rhodes: {exit_bar: 1, fade_out_bars: 2}
    motif: hook_main
    motif_treatment: fragment
    dynamic_curve: decrescendo
    # IV-key feel: F-C/E-Dm-Bb (relative to home C major)
    harmony: "Fmaj7 | C/E | Dm7 | Bb"
    transition_to_next: swell

  # Final chorus — extended (16 bars) with full elaboration. The "return" treatment.
  - id: chorus3
    role: chorus_climax
    bars: 16
    arrangement:
      rhodes: {enter_bar: 1}
      kick: {enter_bar: 1}
      snare: {enter_bar: 1}
      hat: {enter_bar: 1}
      sax: {enter_bar: 1, prominent: true}
    motif: hook_main
    motif_treatment: return
    dynamic_curve: arc
    harmony: "Cmaj7 G | Am7 Fmaj7 | Cmaj7 G | Am7 Fmaj7 | Cmaj7 G | Am7 Fmaj7 | Cmaj7 G | Am7 Fmaj7"
    transition_to_next: fade

  # Outro — decrescendo
  - id: outro
    role: recede
    bars: 4
    arrangement:
      sax: {exit_bar: 1, fade_out_bars: 2}
      kick: {exit_bar: 3}
      snare: {exit_bar: 3}
      hat: {exit_bar: 3}
      rhodes: {exit_bar: 5, fade_out_bars: 4}
      bass: {exit_bar: 5, fade_out_bars: 3}
      pad: {exit_bar: 5, fade_out_bars: 4}
    motif: hook_main
    motif_treatment: fragment
    dynamic_curve: decrescendo
    harmony: "Cmaj7 G | Am7 Fmaj7"

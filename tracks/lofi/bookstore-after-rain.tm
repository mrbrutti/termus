title: Bookstore After Rain
description: Late-night felt piano with boom-bap kick, ghost snares, 16th-note hats, Rhodes chops, and sub-bass.
style: lofi
mix_bus: lofi
listen_mode: album-side
seed: 28011
tags: [lofi, piano, rain, neon, boom-bap, ghost]
key: Dmin
tempo: 81
globals: {density: full, brightness: warm, motion: moving, reverb: halo}
roles:
  keys:
    family: electric_piano
    tone: [warm, dusty, soft]
    register: mid
    prominence: support
    personality: piano_felt
    room: bedroom_small
    reverb_send_db: -12
    pattern: "x.x..x.x x.x..x.x"
  bass:
    family: bass
    tone: [round, woody]
    register: low
    prominence: anchor
    pattern: "x..x x..x x..x x.x."
  sub:
    family: synth_bass
    tone: [deep, round]
    register: low
    prominence: anchor
    pattern: "x... x... x... x..."
  kick:
    family: drums
    tone: [dusty, deep]
    prominence: anchor
    pattern: "x..x ..x. x..x ...."
  snare:
    family: drums
    tone: [dusty, soft]
    prominence: support
    pattern: "....x...x..xx..."
  hat:
    family: drums
    tone: [dry, tight]
    prominence: support
    pattern: "x.x.x.x.x.x.x.x."
  ride:
    family: drums
    tone: [live, soft]
    prominence: support
    pattern: "....x.......x..."
  lead:
    family: reed_lead
    tone: [breathy, intimate]
    register: mid-high
    prominence: lead
    motif: "5 . 6 . 7 . 9 7 | 5 . 3 . 2 . 1 ."
  pad:
    family: pad
    tone: [soft, wide]
    register: mid
    prominence: air
    pattern: "x..............."
sections:
  - id: intro
    title: rain-on-glass
    duration: 14s
    harmony: "Dm9 Gm7 | Bbmaj9 Am7"
    scene: "intro hush"
    variation: "establish"
    groove: dilla_late
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.2}
          - {at: 100, value: 0.55}
  - id: verse
    title: paperback turn
    duration: 36s
    harmony: "Dm9 Gm7 | Cmaj9 Bbmaj7"
    scene: "head glide"
    variation: "statement"
    groove: dilla_late
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 0.7}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.55}
          - {at: 60, value: 0.85}
          - {at: 100, value: 0.65}
  - id: bridge
    title: stack-shadows
    duration: 22s
    harmony: "Bb6 Am7 | Dm9 Gm7"
    scene: "bridge tilt"
    variation: "open-register"
    groove: dilla_late
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 0.8}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.7}
          - {at: 60, value: 0.9}
          - {at: 100, value: 0.75}
  - id: verse2
    title: second read
    duration: 28s
    harmony: "Dm9 Gm7 | Cmaj9 Am7"
    scene: "head glide"
    variation: "sequence-up"
    groove: dilla_late
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 1.0}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.65}
          - {at: 100, value: 0.75}
  - id: outro
    title: shelf-closing
    duration: 20s
    harmony: "Dm9 A7 | Dm6"
    scene: "outro hush"
    variation: "cadence"
    groove: dilla_late
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 1.0}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.75}
          - {at: 100, value: 0.25}

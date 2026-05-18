title: Bookstore After Rain
description: Late-night neon piano with felt hammers and tape drift.
style: lofi
mix_bus: lofi
listen_mode: album-side
seed: 28011
tags: [lofi, piano, rain, neon]
key: Dmin
tempo: 81
globals: {density: steady, brightness: balanced, motion: gentle, reverb: warm}
roles:
  keys:
    family: electric_piano
    tone: [warm, soft]
    register: mid
    prominence: support
    personality: piano_felt
    room: bedroom_small
    reverb_send_db: -14
    pattern: "x..x ..x. x..x ..x."
  bass:
    family: synth_bass
    tone: [round]
    register: low
    prominence: anchor
    pattern: "x... x... x... x..."
  kick:
    family: drums
    tone: [dusty]
    prominence: anchor
    pattern: "x... ...x x... ...."
  snare:
    family: drums
    tone: [dusty]
    prominence: support
    pattern: ".... x... .... x..."
  hat:
    family: drums
    tone: [dry]
    prominence: support
    pattern: "x.x.x.x.x.x.x.x."
  lead:
    family: reed_lead
    tone: [breathy, intimate]
    register: mid-high
    prominence: lead
    motif: "5 . 6 . 7 . 9 7 | 5 . 3 . 2 . 1 ."
sections:
  - id: intro
    title: rain-on-glass
    duration: 12s
    harmony: "Dm9 Gm7"
    scene: "intro hush"
    variation: "establish"
    groove: dilla_late
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.2}
          - {at: 100, value: 0.5}
  - id: verse
    title: paperback turn
    duration: 40s
    harmony: "Dm9 Gm7 | Cmaj9 Bb7"
    scene: "head glide"
    variation: "statement"
    groove: dilla_late
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.5}
          - {at: 60, value: 0.85}
          - {at: 100, value: 0.6}
  - id: bridge
    title: stack-shadows
    duration: 24s
    harmony: "Bb6 Am7 | Dm9 Gm7"
    scene: "bridge tilt"
    variation: "open-register"
    groove: dilla_late
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
          - {at: 0, value: 0.7}
          - {at: 100, value: 0.3}

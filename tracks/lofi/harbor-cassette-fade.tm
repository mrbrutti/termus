title: Harbor Cassette Fade
description: Slow-motion tape Rhodes with dense boom-bap, open-hat accents, walking sub, vibes texture, and sax motif.
style: lofi
mix_bus: lofi
listen_mode: album-side
seed: 51908
tags: [lofi, harbor, cassette, rhodes, tape, sub, vibes]
key: Emin
tempo: 76
globals: {density: full, brightness: warm, motion: moving, reverb: halo}
roles:
  keys:
    family: electric_piano
    tone: [warm, dusty]
    register: mid
    prominence: support
    personality: piano_felt
    room: bedroom_small
    reverb_send_db: -14
    pattern: "x.x..x.. x.x..x.."
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
    pattern: "x... x... x... .x.."
  kick:
    family: drums
    tone: [dusty, soft]
    prominence: anchor
    pattern: "x..x ...x x..x ...."
  snare:
    family: drums
    tone: [dusty]
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
  vibes:
    family: mallet
    tone: [glass, soft, warm]
    register: mid-high
    prominence: air
    pattern: "..x. .... ..x. ...."
  lead:
    family: reed_lead
    tone: [breathy, intimate]
    register: mid-high
    prominence: lead
    motif: "5 . . 7 9 . 7 5 | 3 . 2 . 1 . . ."
  pad:
    family: pad
    tone: [soft, wide]
    register: mid
    prominence: air
    pattern: "x..............."
sections:
  - id: intro
    title: tide-on-tape
    duration: 14s
    harmony: "Em9 Am7 | Cmaj9 Bm7"
    scene: "intro hush"
    variation: "establish"
    groove: dilla_late
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.15}
          - {at: 100, value: 0.45}
  - id: verse
    title: harbor lamp
    duration: 34s
    harmony: "Em9 Am7 | Dmaj9 C#7"
    scene: "head glide"
    variation: "statement"
    groove: dilla_late
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 0.8}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.45}
          - {at: 50, value: 0.7}
          - {at: 100, value: 0.6}
  - id: bridge
    title: fog-bank turn
    duration: 22s
    harmony: "Cmaj9 B7 | Em9 Am7"
    scene: "bridge tilt"
    variation: "open-register"
    groove: dilla_late
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 1.0}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.6}
          - {at: 100, value: 0.82}
  - id: verse2
    title: foghorn echo
    duration: 28s
    harmony: "Em9 Am7 | Cmaj9 Dsus4"
    scene: "head glide"
    variation: "sequence-up"
    groove: dilla_late
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 0.8}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.7}
          - {at: 100, value: 0.75}
  - id: outro
    title: rope-light closing
    duration: 22s
    harmony: "Em9 B7 | Em6"
    scene: "outro hush"
    variation: "cadence"
    groove: dilla_late
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 1.0}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.8}
          - {at: 100, value: 0.25}

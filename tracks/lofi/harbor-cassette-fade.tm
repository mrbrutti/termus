title: Harbor Cassette Fade
description: Slow-motion tape Rhodes with tritone sub on V and long cutoff sweeps.
style: lofi
mix_bus: lofi
listen_mode: album-side
seed: 51908
tags: [lofi, harbor, cassette, rhodes, tape]
key: Emin
tempo: 76
globals: {density: steady, brightness: warm, motion: gentle, reverb: halo}
roles:
  keys:
    family: electric_piano
    tone: [warm, dusty]
    register: mid
    prominence: support
    personality: piano_felt
    room: bedroom_small
    reverb_send_db: -16
    pattern: "x..x .... x..x ...."
  bass:
    family: synth_bass
    tone: [round, soft]
    register: low
    prominence: anchor
    pattern: "x... .... x... .x.."
  kick:
    family: drums
    tone: [dusty, soft]
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
    pattern: "x... x.x. x... x.x."
  pad:
    family: pad
    tone: [soft, wide]
    register: mid
    prominence: air
    pattern: "x....... ........"
sections:
  - id: intro
    title: tide-on-tape
    duration: 16s
    harmony: "Em9 Am7"
    scene: "intro hush"
    variation: "establish"
    groove: straight
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.15}
          - {at: 100, value: 0.4}
  - id: verse
    title: harbor lamp
    duration: 42s
    harmony: "Em9 Am7 | Dmaj9 C#7"
    scene: "head glide"
    variation: "statement"
    groove: straight
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.4}
          - {at: 50, value: 0.65}
          - {at: 100, value: 0.55}
  - id: bridge
    title: fog-bank turn
    duration: 20s
    harmony: "Cmaj9 B7 | Em9 Am7"
    scene: "bridge tilt"
    variation: "open-register"
    groove: straight
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 0.9}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.55}
          - {at: 100, value: 0.8}
  - id: outro
    title: rope-light closing
    duration: 22s
    harmony: "Em9 B7 | Em6"
    scene: "outro hush"
    variation: "cadence"
    groove: straight
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 1.0}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.8}
          - {at: 100, value: 0.25}

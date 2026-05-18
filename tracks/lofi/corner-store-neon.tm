title: Corner Store Neon
description: Buzzing fluorescent guitar loop with dense boom-bap, ghost hats, sub-bass stabs, and Rhodes chops.
style: lofi
mix_bus: lofi
listen_mode: album-side
seed: 37220
tags: [lofi, guitar, neon, swing, boom-bap, ghost]
key: Fmin
tempo: 92
globals: {density: full, brightness: balanced, motion: moving, reverb: halo}
roles:
  guitar:
    family: guitar
    tone: [warm, soft]
    register: mid
    prominence: support
    personality: guitar_nylon
    room: bedroom_small
    reverb_send_db: -10
  keys:
    family: electric_piano
    tone: [warm, dusty]
    register: mid
    prominence: air
  bass:
    family: bass
    tone: [round, woody]
    register: low
    prominence: anchor
  sub:
    family: synth_bass
    tone: [deep, round]
    register: low
    prominence: anchor
  kick:
    family: drums
    tone: [dusty, soft]
    prominence: anchor
  snare:
    family: drums
    tone: [dusty, soft]
    prominence: support
  hat:
    family: drums
    tone: [dry, tight]
    prominence: support
  ride:
    family: drums
    tone: [live, dry]
    prominence: support
  lead:
    family: reed_lead
    tone: [breathy, intimate]
    register: mid-high
    prominence: lead
    motif: "5 . 7 . b7 . 5 3 | 5 . . . 3 . 1 ."
  pad:
    family: pad
    tone: [soft, wide]
    register: mid
    prominence: air
sections:
  - id: intro
    title: store-front hiss
    duration: 12s
    harmony: "Fm9 Bbm7 | Dbmaj9 Cm7"
    scene: "intro hush"
    variation: "establish"
    groove: swing_56
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.45}
          - {at: 100, value: 0.7}
  - id: verse
    title: register glow
    duration: 34s
    harmony: "Fm9 Bbm7 | Ebmaj9 Db7"
    scene: "head glide"
    variation: "statement"
    groove: swing_56
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 0.9}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.7}
          - {at: 50, value: 0.65}
          - {at: 100, value: 0.65}
  - id: bridge
    title: cooler hum
    duration: 22s
    harmony: "Dbmaj9 C7 | Fm9 Bbm7"
    scene: "bridge tilt"
    variation: "open-register"
    groove: swing_56
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 1.0}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.55}
          - {at: 100, value: 0.8}
  - id: verse2
    title: back-shelf blink
    duration: 26s
    harmony: "Fm9 Bbm7 | Ebmaj9 Cm7"
    scene: "head glide"
    variation: "sequence-up"
    groove: swing_56
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 0.8}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.65}
          - {at: 100, value: 0.7}
  - id: outro
    title: closed sign
    duration: 18s
    harmony: "Fm9 Db7 | Fm6"
    scene: "outro hush"
    variation: "cadence"
    groove: swing_56
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 1.0}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.65}
          - {at: 100, value: 0.2}

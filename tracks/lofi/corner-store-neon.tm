title: Corner Store Neon
description: Buzzing fluorescent guitar loop with swing pocket and ii-V chain inserts.
style: lofi
mix_bus: lofi
listen_mode: album-side
seed: 37220
tags: [lofi, guitar, neon, swing]
key: Fmin
tempo: 90
globals: {density: steady, brightness: balanced, motion: gentle, reverb: room}
roles:
  guitar:
    family: guitar
    tone: [warm, soft]
    register: mid
    prominence: support
    personality: guitar_nylon
    room: bedroom_small
    reverb_send_db: -12
    pattern: "x..x ..x. x..x ..x."
  bass:
    family: bass
    tone: [round, woody]
    register: low
    prominence: anchor
    pattern: "x... x..x x... x..."
  kick:
    family: drums
    tone: [dusty, soft]
    prominence: anchor
    pattern: "x... x... x... x..."
  snare:
    family: drums
    tone: [dusty, soft]
    prominence: support
    pattern: ".... x... .... x..."
  hat:
    family: drums
    tone: [dry, tight]
    prominence: support
    pattern: "x.xx x.xx x.xx x.xx"
  keys:
    family: electric_piano
    tone: [warm]
    register: mid
    prominence: air
    pattern: "..x. .... ..x. ...."
sections:
  - id: intro
    title: store-front hiss
    duration: 14s
    harmony: "Fm9 Bbm7"
    scene: "intro hush"
    variation: "establish"
    groove: swing_56
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.5}
          - {at: 100, value: 0.75}
  - id: verse
    title: register glow
    duration: 44s
    harmony: "Fm9 Bbm7 | Ebmaj9 Db7"
    scene: "head glide"
    variation: "statement"
    groove: swing_56
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 0.8}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.75}
          - {at: 50, value: 0.6}
          - {at: 100, value: 0.6}
  - id: bridge
    title: cooler hum
    duration: 24s
    harmony: "Dbmaj9 C7 | Fm9 Bbm7"
    scene: "bridge tilt"
    variation: "open-register"
    groove: swing_56
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 1.0}
  - id: outro
    title: closed sign
    duration: 18s
    harmony: "Fm9 Db7 | Fm6"
    scene: "outro hush"
    variation: "cadence"
    groove: swing_56
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.6}
          - {at: 100, value: 0.2}

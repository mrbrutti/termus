title: Concourse Pulse / Window
description: Glass-steps track with brighter repeats and a clipped upper bridge.
style: phase
substyle: glass-steps
listen_mode: endless
seed: 57104
tags: [phase, concourse, pulse, window]
key: Gmaj
tempo: 76
globals: {density: light, brightness: balanced, motion: moving, reverb: room}
roles:
  mallet-a: {family: mallet, tone: [glass], register: high, prominence: lead, pattern: "x... x..."}
  mallet-b: {family: mallet, tone: [soft], register: high, prominence: answer, pattern: ".... x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: air, pattern: "x....... | ....x..."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
sections:
  - {id: intro, title: window count, duration: 22s, harmony: "Gmaj9 D/F# | Em9 Cmaj9", scene: "glass intro", variation: "establish"}
  - {id: body, title: pulse lane, duration: 54s, harmony: "Gmaj9 Bm7 | Cmaj9 Dadd9", scene: "interlock answer", variation: "statement"}
  - {id: bridge, title: overhead pane, duration: 36s, harmony: "Am9 Cmaj9 | Gmaj9 D/F#", scene: "bridge lift", variation: "sequence"}
  - {id: outro, title: final pulse, duration: 26s, harmony: "Gmaj9 D/F# | Cmaj9 Gmaj9", scene: "outro cadence", variation: "cadence"}

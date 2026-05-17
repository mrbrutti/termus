title: Escalator Counterlight
description: Glass-steps phasing study with brighter mallet answers and a shorter close.
style: phase
substyle: glass-steps
listen_mode: endless
seed: 57101
tags: [phase, escalator, counterlight, glass]
key: Cmaj
tempo: 74
globals: {density: light, brightness: balanced, motion: moving, reverb: room}
roles:
  mallet-a: {family: mallet, tone: [glass], register: high, prominence: lead, pattern: "x... x..."}
  mallet-b: {family: mallet, tone: [soft], register: high, prominence: answer, pattern: ".... x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: air, pattern: "x....... | ....x..."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
sections:
  - {id: intro, title: first teeth, duration: 22s, harmony: "Cmaj9 Am7 | Fmaj9 Gsus4", scene: "glass intro", variation: "establish"}
  - {id: body, title: mirrored rise, duration: 56s, harmony: "Cmaj9 Em7 | Fmaj9 Gsus4", scene: "interlock answer", variation: "statement"}
  - {id: bridge, title: landing blur, duration: 38s, harmony: "Am9 Fmaj9 | Cmaj9 Gsus4", scene: "bridge lift", variation: "sequence"}
  - {id: outro, title: last plate, duration: 26s, harmony: "Cmaj9 Am7 | Fmaj9 Cmaj9", scene: "outro cadence", variation: "cadence"}

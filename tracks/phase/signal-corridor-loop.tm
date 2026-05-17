title: Signal Corridor / Loop
description: Warm-interlock corridor pattern with softer bass and longer pad glue.
style: phase
substyle: warm-interlock
listen_mode: endless
seed: 57105
tags: [phase, signal, corridor, warm]
key: Amin
tempo: 71
globals: {density: light, brightness: warm, motion: moving, reverb: room}
roles:
  mallet-a: {family: mallet, tone: [soft], register: high, prominence: lead, pattern: "x... x..."}
  mallet-b: {family: mallet, tone: [soft], register: high, prominence: answer, pattern: ".... x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
sections:
  - {id: intro, title: signal latch, duration: 20s, harmony: "Am9 E/G# | Fmaj9 Dm9", scene: "warm intro", variation: "establish"}
  - {id: body, title: corridor loop, duration: 52s, harmony: "Am9 Cmaj9 | Fmaj9 E7", scene: "interlock answer", variation: "statement"}
  - {id: bridge, title: handrail turn, duration: 34s, harmony: "Dm9 Fmaj9 | Am9 E/G#", scene: "bridge warm", variation: "glide"}
  - {id: outro, title: latch dim, duration: 24s, harmony: "Am9 E/G# | Fmaj9 Am9", scene: "outro cadence", variation: "cadence"}

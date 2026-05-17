title: Static Lantern / Field
description: Soft-static field piece with bell dust, slower bridge motion, and dim exit.
style: drone
substyle: soft-static
listen_mode: endless
seed: 53108
tags: [drone, static, lantern, field]
key: Cmaj
tempo: 46
globals: {density: sparse, brightness: balanced, motion: still, reverb: room}
roles:
  bed: {family: pad, tone: [soft], register: mid, prominence: air, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
  bells: {family: bells, tone: [soft], register: high, prominence: answer, pattern: "........ | ....x..."}
  lead: {family: woodwind, tone: [thin], register: high, prominence: lead, motif: "5 . . . | 3 . . ."}
sections:
  - {id: intro, title: field lantern, duration: 30s, harmony: "Cmaj9 Am7 | Fmaj9 Gsus4", scene: "intro static", variation: "establish"}
  - {id: body, title: long field, duration: 82s, harmony: "Cmaj9 Em7 | Fmaj9 Gsus4", scene: "drift still", variation: "statement"}
  - {id: bridge, title: dim marker, duration: 56s, harmony: "Am9 Fmaj9 | Cmaj9 Gsus4", scene: "bridge drift", variation: "sequence"}
  - {id: outro, title: quiet lamp, duration: 42s, harmony: "Cmaj9 Am7 | Fmaj9 Cmaj9", scene: "outro home", variation: "cadence"}

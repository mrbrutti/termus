title: Long Wire / Matron
description: Soft-static signal bed with thin upper whistle and steady low pulse.
style: drone
substyle: soft-static
listen_mode: endless
seed: 53102
tags: [drone, wire, static, long]
key: Dmin
tempo: 46
globals: {density: sparse, brightness: balanced, motion: still, reverb: room}
roles:
  bed: {family: pad, tone: [soft, wide], register: mid, prominence: air, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
  lead: {family: woodwind, tone: [thin], register: high, prominence: lead, motif: "5 . . . | 3 . . ."}
  bells: {family: bells, tone: [soft], register: high, prominence: answer, pattern: "........ | ....x..."}
sections:
  - {id: intro, title: wire wake, duration: 32s, harmony: "Dm9 Bbmaj9 | Cadd9 Am7", scene: "intro static", variation: "establish"}
  - {id: body, title: long carrier, duration: 84s, harmony: "Dm9 Gm9 | Bbmaj9 A7", scene: "drift still", variation: "statement"}
  - {id: bridge, title: relay fade, duration: 58s, harmony: "Fmaj9 C/E | Dm9 Gsus4", scene: "bridge drift", variation: "sequence"}
  - {id: outro, title: matron hush, duration: 44s, harmony: "Dm9 Bbmaj9 | A7 Dm9", scene: "outro home", variation: "cadence"}

title: Last Train / Halo
description: Choir-fog nocturne with brighter middle lift and a long held close.
style: ambient
substyle: choir-fog
listen_mode: endless
seed: 50103
tags: [ambient, halo, train, choir]
key: Dmin
tempo: 54
globals: {density: light, brightness: warm, motion: gentle, reverb: halo}
roles:
  choir: {family: choir, tone: [soft, wide], register: mid, prominence: air, pattern: "x....... | ....x..."}
  pad: {family: pad, tone: [soft], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [round], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
  lead: {family: woodwind, tone: [breathy], register: high, prominence: lead, motif: "5 . . . | 3 . 2 ."}
sections:
  - {id: intro, title: distant board, duration: 30s, harmony: "Dm9 Bbmaj9 | Cadd9 Am7", scene: "intro haze", variation: "establish"}
  - {id: verse, title: sodium halo, duration: 64s, harmony: "Dm9 Gm9 | Bbmaj9 A7", scene: "head drift", variation: "statement"}
  - {id: bridge, title: signal bloom, duration: 56s, harmony: "Fmaj9 C/E | Dm9 Gsus4", scene: "bridge lift", variation: "sequence"}
  - {id: outro, title: shutter close, duration: 42s, harmony: "Dm9 Bbmaj9 | A7 Dm9", scene: "outro home", variation: "cadence"}

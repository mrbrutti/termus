title: River Station / Mist
description: Thin station-haze suite with bell answers and a quieter descent.
style: ambient
substyle: station-haze
listen_mode: endless
seed: 50104
tags: [ambient, river, mist, station]
key: Fmaj
tempo: 62
globals: {density: sparse, brightness: balanced, motion: gentle, reverb: room}
roles:
  pad: {family: pad, tone: [wide], register: mid, prominence: air, pattern: "x....... | ....x..."}
  bells: {family: bells, tone: [glass, soft], register: high, prominence: answer, pattern: "....x... | ..x....."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
  lead: {family: woodwind, tone: [hollow], register: mid-high, prominence: lead, motif: "5 . 4 . | 3 . 1 ."}
sections:
  - {id: intro, title: quay lights, duration: 26s, harmony: "Fmaj9 C/E | Dm9 Gsus4", scene: "intro haze", variation: "establish"}
  - {id: body, title: river glass, duration: 70s, harmony: "Fmaj9 Am7 | Dm9 Bbmaj9", scene: "drift answer", variation: "statement"}
  - {id: bridge, title: underpass glow, duration: 48s, harmony: "Gm9 Dm9 | Bbmaj9 Cadd9", scene: "bridge drift", variation: "glide"}
  - {id: outro, title: final platform, duration: 36s, harmony: "Fmaj9 C/E | Dm9 Fmaj9", scene: "outro release", variation: "cadence"}

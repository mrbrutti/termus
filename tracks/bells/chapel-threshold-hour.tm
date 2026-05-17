title: Chapel Threshold / Hour
description: Cloister-rain threshold piece with darker low end and open bridge air.
style: bells
substyle: cloister-rain
listen_mode: album-side
seed: 51104
tags: [bells, chapel, threshold, hour]
key: Dmin
tempo: 50
globals: {density: light, brightness: balanced, motion: gentle, reverb: halo}
roles:
  bells: {family: bells, tone: [soft], register: high, prominence: lead, motif: "5 . 6 . | 7 . 5 ."}
  box: {family: music_box, tone: [soft], register: high, prominence: answer, pattern: "....x... | ..x....."}
  choir: {family: choir, tone: [wide], register: mid, prominence: air, pattern: "x....... | ....x..."}
  bass: {family: bass, tone: [round], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
sections:
  - {id: intro, title: door seam, duration: 24s, harmony: "Dm9 Bbmaj9 | Cadd9 A7", scene: "chapel intro", variation: "establish"}
  - {id: body, title: inner hour, duration: 54s, harmony: "Dm9 Gm9 | Bbmaj9 A7", scene: "threshold answer", variation: "statement"}
  - {id: bridge, title: low transept, duration: 44s, harmony: "Fmaj9 C/E | Dm9 Gsus4", scene: "release air", variation: "open-register"}
  - {id: outro, title: stone return, duration: 30s, harmony: "Dm9 Bbmaj9 | A7 Dm9", scene: "outro cadence", variation: "cadence"}

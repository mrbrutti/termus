title: Bellwell Rain / Nave
description: Cloister-rain chapel study with music-box replies and a held final round.
style: bells
substyle: cloister-rain
listen_mode: album-side
seed: 51102
tags: [bells, rain, nave, cloister]
key: Emin
tempo: 52
globals: {density: light, brightness: balanced, motion: gentle, reverb: halo}
roles:
  bells: {family: bells, tone: [soft, glass], register: high, prominence: lead, motif: "5 . 6 . | 7 . 5 ."}
  box: {family: music_box, tone: [soft], register: high, prominence: answer, pattern: "....x... | ..x....."}
  pad: {family: pad, tone: [wide], register: mid, prominence: air, pattern: "x....... | ....x..."}
  bass: {family: bass, tone: [round], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
sections:
  - {id: intro, title: rain nave, duration: 26s, harmony: "Em9 Cmaj9 | Dsus4 B7", scene: "cloister intro", variation: "establish"}
  - {id: body, title: gutter prayer, duration: 58s, harmony: "Em9 Am9 | Cmaj9 B7", scene: "devotional answer", variation: "statement"}
  - {id: bridge, title: wet stone, duration: 42s, harmony: "Gmaj9 D/F# | Em9 Cmaj9", scene: "threshold release", variation: "glide"}
  - {id: outro, title: low candle, duration: 32s, harmony: "Em9 Cmaj9 | B7 Em9", scene: "outro cadence", variation: "cadence"}

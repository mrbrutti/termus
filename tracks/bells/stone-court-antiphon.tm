title: Stone Court / Antiphon
description: Rain cloister round with antiphonal bell replies and a low held cadence.
style: bells
substyle: cloister-rain
listen_mode: album-side
seed: 51107
tags: [bells, stone, court, antiphon]
key: Fmin
tempo: 50
globals: {density: light, brightness: warm, motion: gentle, reverb: halo}
roles:
  bells: {family: bells, tone: [soft, glass], register: high, prominence: lead, motif: "5 . 6 . | 7 . 5 ."}
  box: {family: music_box, tone: [soft], register: high, prominence: answer, pattern: "....x... | ..x....."}
  pad: {family: pad, tone: [wide], register: mid, prominence: air, pattern: "x....... | ....x..."}
  bass: {family: bass, tone: [round], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
sections:
  - {id: intro, title: cloister path, duration: 22s, harmony: "Fm9 Dbmaj9 | Ebadd9 C7", scene: "cloister intro", variation: "establish"}
  - {id: body, title: antiphon walk, duration: 54s, harmony: "Fm9 Bbm9 | Dbmaj9 C7", scene: "devotional answer", variation: "statement"}
  - {id: bridge, title: wet court, duration: 42s, harmony: "Abmaj9 Eb/G | Fm9 Dbmaj9", scene: "release threshold", variation: "glide"}
  - {id: outro, title: stone cadence, duration: 30s, harmony: "Fm9 Dbmaj9 | C7 Fm9", scene: "outro cadence", variation: "cadence"}

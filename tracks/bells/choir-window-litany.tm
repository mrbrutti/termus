title: Choir Window / Litany
description: Brighter vespers-glass setting with choir breath and slower cadential tolls.
style: bells
substyle: vespers-glass
listen_mode: album-side
seed: 51103
tags: [bells, choir, window, litany]
key: Amaj
tempo: 55
globals: {density: light, brightness: warm, motion: still, reverb: halo}
roles:
  bells: {family: bells, tone: [glass], register: high, prominence: lead, motif: "5 . . 7 | 9 . 7 5"}
  choir: {family: choir, tone: [soft], register: mid, prominence: air, pattern: "x....... | ....x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
sections:
  - {id: intro, title: glass lintel, duration: 22s, harmony: "Amaj9 E/G# | F#m9 Dmaj9", scene: "vespers intro", variation: "establish"}
  - {id: body, title: litany air, duration: 56s, harmony: "Amaj9 C#m7 | Dmaj9 Eadd9", scene: "devotional answer", variation: "statement"}
  - {id: bridge, title: white nave, duration: 40s, harmony: "Bm9 Dmaj9 | Amaj9 E/G#", scene: "chapel release", variation: "sequence"}
  - {id: outro, title: final toll, duration: 28s, harmony: "Amaj9 E/G# | Dmaj9 Amaj9", scene: "outro cadence", variation: "cadence"}

title: Winter Cloister / Round
description: Bright glass-chapel round with wider choir bed and a slower last toll.
style: bells
substyle: vespers-glass
listen_mode: album-side
seed: 51108
tags: [bells, winter, cloister, round]
key: Emaj
tempo: 53
globals: {density: light, brightness: balanced, motion: still, reverb: halo}
roles:
  bells: {family: bells, tone: [glass], register: high, prominence: lead, motif: "5 . . 7 | 9 . 7 5"}
  choir: {family: choir, tone: [soft], register: mid, prominence: air, pattern: "x....... | ....x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
sections:
  - {id: intro, title: frost aisle, duration: 20s, harmony: "Emaj9 B/D# | C#m9 Amaj9", scene: "vespers intro", variation: "establish"}
  - {id: body, title: snow round, duration: 56s, harmony: "Emaj9 G#m7 | Amaj9 Badd9", scene: "devotional answer", variation: "statement"}
  - {id: release, title: choir frost, duration: 40s, harmony: "F#m9 Amaj9 | Emaj9 B/D#", scene: "release prayer", variation: "subtract"}
  - {id: outro, title: last toll, duration: 28s, harmony: "Emaj9 B/D# | Amaj9 Emaj9", scene: "outro cadence", variation: "cadence"}

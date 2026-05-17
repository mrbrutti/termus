title: Remote Power / Bloom
description: Cathedral-bed substation bloom with devotional choir weight and long decay.
style: drone
substyle: cathedral-bed
listen_mode: endless
seed: 53107
tags: [drone, remote, power, bloom]
key: Bbmin
tempo: 41
globals: {density: sparse, brightness: warm, motion: still, reverb: cathedral}
roles:
  bed: {family: choir, tone: [wide, soft], register: mid, prominence: air, pattern: "x....... | x......."}
  bass: {family: synth_bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
  bells: {family: bells, tone: [dark], register: high, prominence: answer, pattern: "....x... | ........"}
  lead: {family: woodwind, tone: [hollow], register: high, prominence: lead, motif: "5 . . . | 2 . . ."}
sections:
  - {id: intro, title: low switch, duration: 38s, harmony: "Bbm9 Gbmaj9 | Abadd9 F7", scene: "intro field", variation: "establish"}
  - {id: body, title: power bloom, duration: 92s, harmony: "Bbm9 Ebm9 | Gbmaj9 F7", scene: "drift still", variation: "statement"}
  - {id: bridge, title: tower sheen, duration: 62s, harmony: "Dbmaj9 Ab/C | Bbm9 Gbmaj9", scene: "bridge lift", variation: "glide"}
  - {id: outro, title: cold relay, duration: 48s, harmony: "Bbm9 Gbmaj9 | F7 Bbm9", scene: "outro release", variation: "cadence"}

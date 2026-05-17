title: Weather Band / Procession
description: Cathedral-bed procession with larger choir floor and slower bass pulses.
style: drone
substyle: cathedral-bed
listen_mode: endless
seed: 53105
tags: [drone, weather, procession, choir]
key: Gmin
tempo: 43
globals: {density: sparse, brightness: warm, motion: still, reverb: cathedral}
roles:
  bed: {family: choir, tone: [soft, devotional], register: mid, prominence: air, pattern: "x....... | x......."}
  bass: {family: synth_bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
  bells: {family: bells, tone: [dark], register: high, prominence: answer, pattern: "....x... | ........"}
  lead: {family: woodwind, tone: [hollow], register: high, prominence: lead, motif: "5 . . . | 1 . . ."}
sections:
  - {id: intro, title: weather seal, duration: 36s, harmony: "Gm9 Ebmaj9 | Fsus4 Dm9", scene: "intro field", variation: "establish"}
  - {id: body, title: procession floor, duration: 90s, harmony: "Gm9 Cm9 | Ebmaj9 D7", scene: "drift still", variation: "statement"}
  - {id: bridge, title: wet banner, duration: 60s, harmony: "Bbmaj9 F/A | Gm9 Ebmaj9", scene: "bridge lift", variation: "glide"}
  - {id: outro, title: final aisle, duration: 46s, harmony: "Gm9 Ebmaj9 | D7 Gm9", scene: "outro release", variation: "cadence"}

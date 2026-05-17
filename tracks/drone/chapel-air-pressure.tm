title: Chapel Air / Pressure
description: Cathedral-bed drone with low bell undertow and slow held release.
style: drone
substyle: cathedral-bed
listen_mode: endless
seed: 53101
tags: [drone, chapel, air, pressure]
key: Cmin
tempo: 44
globals: {density: sparse, brightness: warm, motion: still, reverb: cathedral}
roles:
  bed: {family: choir, tone: [soft, wide], register: mid, prominence: air, pattern: "x....... | x......."}
  bass: {family: synth_bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
  bells: {family: bells, tone: [dark], register: high, prominence: answer, pattern: "....x... | ........"}
  lead: {family: woodwind, tone: [hollow], register: high, prominence: lead, motif: "5 . . . | 1 . . ."}
sections:
  - {id: intro, title: pressure rise, duration: 34s, harmony: "Cm9 Abmaj9 | Bbadd9 G7", scene: "intro field", variation: "establish"}
  - {id: body, title: nave pressure, duration: 88s, harmony: "Cm9 Fm9 | Abmaj9 G7", scene: "drift still", variation: "statement"}
  - {id: bridge, title: hollow crossing, duration: 62s, harmony: "Ebmaj9 Bb/D | Cm9 G7", scene: "bridge lift", variation: "glide"}
  - {id: outro, title: last chamber, duration: 46s, harmony: "Cm9 Abmaj9 | G7 Cm9", scene: "outro release", variation: "cadence"}

title: Grid Sleep / Litany
description: Soft-static grid hymn with thin response bells and a longer fade.
style: drone
substyle: soft-static
listen_mode: endless
seed: 53104
tags: [drone, grid, sleep, litany]
key: Amin
tempo: 45
globals: {density: sparse, brightness: balanced, motion: still, reverb: room}
roles:
  bed: {family: pad, tone: [soft], register: mid, prominence: air, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
  bells: {family: bells, tone: [soft], register: high, prominence: answer, pattern: "........ | ....x..."}
  lead: {family: woodwind, tone: [thin], register: high, prominence: lead, motif: "5 . . . | 3 . . ."}
sections:
  - {id: intro, title: substation hush, duration: 34s, harmony: "Am9 Fmaj9 | Cadd9 Gsus4", scene: "intro static", variation: "establish"}
  - {id: body, title: sleep grid, duration: 82s, harmony: "Am9 Dm9 | Fmaj9 Gsus4", scene: "drift still", variation: "statement"}
  - {id: bridge, title: maintenance glow, duration: 56s, harmony: "Cmaj9 G/B | Am9 Em7", scene: "bridge drift", variation: "sequence"}
  - {id: outro, title: relay dark, duration: 44s, harmony: "Am9 Fmaj9 | Gsus4 Am9", scene: "outro home", variation: "cadence"}

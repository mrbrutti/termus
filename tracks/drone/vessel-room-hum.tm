title: Vessel Room / Hum
description: Soft-static engine-room drone with hollow lead breaths and low restraint.
style: drone
substyle: soft-static
listen_mode: endless
seed: 53106
tags: [drone, vessel, hum, room]
key: Emaj
tempo: 47
globals: {density: sparse, brightness: balanced, motion: still, reverb: room}
roles:
  bed: {family: pad, tone: [soft, wide], register: mid, prominence: air, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
  bells: {family: bells, tone: [soft], register: high, prominence: answer, pattern: "........ | ....x..."}
  lead: {family: woodwind, tone: [thin], register: high, prominence: lead, motif: "5 . . . | 3 . . ."}
sections:
  - {id: intro, title: hull light, duration: 30s, harmony: "Emaj9 B/D# | C#m9 Amaj9", scene: "intro static", variation: "establish"}
  - {id: body, title: vessel hum, duration: 84s, harmony: "Emaj9 G#m7 | Amaj9 Badd9", scene: "drift still", variation: "statement"}
  - {id: bridge, title: ballast shift, duration: 54s, harmony: "F#m9 Amaj9 | Emaj9 B/D#", scene: "bridge drift", variation: "sequence"}
  - {id: outro, title: stern quiet, duration: 42s, harmony: "Emaj9 B/D# | Amaj9 Emaj9", scene: "outro home", variation: "cadence"}

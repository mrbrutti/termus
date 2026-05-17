title: Tunnel Air / Stillness
description: Darker station-haze piece with bell dust and a held release.
style: ambient
substyle: station-haze
listen_mode: endless
seed: 50102
tags: [ambient, tunnel, stillness, station]
key: Cmaj
tempo: 60
globals: {density: sparse, brightness: balanced, motion: still, reverb: room}
roles:
  pad: {family: pad, tone: [soft, wide], register: mid, prominence: air, pattern: "x....... | ....x..."}
  bells: {family: bells, tone: [glass, light], register: high, prominence: air, pattern: "..x..... | ....x..."}
  bass: {family: synth_bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
  lead: {family: woodwind, tone: [hollow], register: mid-high, prominence: lead, motif: "5 . . . | 3 . 1 ."}
sections:
  - {id: intro, title: rail hum, duration: 28s, harmony: "Cmaj9 Am7 | Fmaj9 Gsus4", scene: "intro field", variation: "establish"}
  - {id: body, title: platform air, duration: 68s, harmony: "Cmaj9 Em7 | Fmaj9 Gsus4", scene: "drift still", variation: "statement"}
  - {id: release, title: late transfer, duration: 52s, harmony: "Am7 Fmaj9 | Cmaj9 Gsus4", scene: "release settle", variation: "subtract"}
  - {id: outro, title: empty stairs, duration: 38s, harmony: "Cmaj9 Am7 | Fmaj9 Cmaj9", scene: "outro snow", variation: "cadence"}

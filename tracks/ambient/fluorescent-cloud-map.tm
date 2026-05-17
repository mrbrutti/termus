title: Fluorescent Cloud / Map
description: Brighter station-haze study with high bell dust and a late hold.
style: ambient
substyle: station-haze
listen_mode: endless
seed: 50106
tags: [ambient, fluorescent, cloud, map]
key: Ebmaj
tempo: 58
globals: {density: sparse, brightness: bright, motion: gentle, reverb: room}
roles:
  pad: {family: pad, tone: [wide], register: mid, prominence: air, pattern: "x....... | ....x..."}
  bells: {family: bells, tone: [soft, glass], register: high, prominence: air, pattern: "...x.... | ....x..."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
  lead: {family: woodwind, tone: [hollow], register: high, prominence: lead, motif: "5 . 6 . | 3 . 2 1"}
sections:
  - {id: intro, title: map glow, duration: 24s, harmony: "Ebmaj9 Cm7 | Abmaj9 Bbadd9", scene: "intro haze", variation: "establish"}
  - {id: drift, title: stair reflections, duration: 66s, harmony: "Ebmaj9 Gm7 | Abmaj9 Fm9", scene: "head answer", variation: "statement"}
  - {id: bridge, title: switchback air, duration: 50s, harmony: "Fm9 Bbadd9 | Gm7 Abmaj9", scene: "bridge lift", variation: "glide"}
  - {id: outro, title: afterimage, duration: 34s, harmony: "Ebmaj9 Cm7 | Bbadd9 Ebmaj9", scene: "outro settle", variation: "cadence"}

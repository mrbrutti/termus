title: Bookstore After Rain
description: Dusty-rhodes shelf-side study with thinner drums and a warm cadence.
style: lofi
substyle: dusty-rhodes
listen_mode: album-side
seed: 55103
tags: [lofi, bookstore, rain, warm]
key: Ebmaj
tempo: 74
globals: {density: light, brightness: warm, motion: gentle, reverb: room}
roles:
  rhodes: {family: electric_piano, tone: [warm, dusty], register: mid, prominence: support, pattern: "x..x .x.."}
  bass: {family: bass, tone: [round], register: low, prominence: anchor, pattern: "x... x..x"}
  kick: {family: drums, tone: [dusty], prominence: anchor, pattern: "x... x..."}
  snare: {family: drums, tone: [dusty], prominence: support, pattern: ".... x..."}
  lead: {family: woodwind, tone: [soft], register: mid-high, prominence: lead, motif: "5 . . 7 | 9 . 7 5"}
sections:
  - {id: intro, title: damp jacket, duration: 14s, harmony: "Ebmaj9 Cm7 | Abmaj9 Bbadd9", scene: "intro hush", variation: "establish"}
  - {id: verse, title: aisle dust, duration: 54s, harmony: "Ebmaj9 Gm7 | Abmaj9 Bbadd9", scene: "head glide", variation: "statement"}
  - {id: breakdown, title: receipt fold, duration: 30s, harmony: "Cm9 Abmaj9 | Ebmaj9 Bbadd9", scene: "breakdown thin", variation: "subtract"}
  - {id: outro, title: window steam, duration: 26s, harmony: "Ebmaj9 Cm7 | Abmaj9 Ebmaj9", scene: "outro home", variation: "cadence"}

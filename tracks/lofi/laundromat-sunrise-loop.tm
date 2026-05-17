title: Laundromat Sunrise / Loop
description: Dusty-rhodes morning study with wider keys and a softer breakdown.
style: lofi
substyle: dusty-rhodes
listen_mode: album-side
seed: 55101
tags: [lofi, laundromat, sunrise, rhodes]
key: Cmaj
tempo: 76
globals: {density: steady, brightness: warm, motion: gentle, reverb: room}
roles:
  rhodes: {family: electric_piano, tone: [warm, dusty], register: mid, prominence: support, pattern: "x..x .x.."}
  bass: {family: bass, tone: [round], register: low, prominence: anchor, pattern: "x... x..x"}
  kick: {family: drums, tone: [dusty], prominence: anchor, pattern: "x... x..."}
  snare: {family: drums, tone: [dusty], prominence: support, pattern: ".... x..."}
  hat: {family: drums, tone: [dry], prominence: support, pattern: "x.x.x.x."}
  lead: {family: reed_lead, tone: [breathy], register: mid-high, prominence: lead, motif: "5 . . 7 | 9 . 7 5"}
sections:
  - {id: intro, title: soap hum, duration: 14s, harmony: "Cmaj9 Am7 | Fmaj9 Gsus4", scene: "intro hush", variation: "establish"}
  - {id: verse, title: warm drums, duration: 52s, harmony: "Cmaj9 Em7 | Fmaj9 Gsus4", scene: "head glide", variation: "statement"}
  - {id: breakdown, title: rinse wait, duration: 34s, harmony: "Am9 Fmaj9 | Cmaj9 Gsus4", scene: "breakdown thin", variation: "subtract"}
  - {id: outro, title: first light, duration: 28s, harmony: "Cmaj9 Am7 | Fmaj9 Cmaj9", scene: "outro home", variation: "cadence"}

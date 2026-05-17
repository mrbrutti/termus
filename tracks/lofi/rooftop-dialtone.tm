title: Rooftop Dialtone
description: Guitar-neon rooftop loop with higher lead answers and a clipped outro.
style: lofi
substyle: guitar-neon
listen_mode: album-side
seed: 55102
tags: [lofi, rooftop, dialtone, neon]
key: Amin
tempo: 84
globals: {density: steady, brightness: balanced, motion: gentle, reverb: room}
roles:
  guitar: {family: guitar, tone: [warm], register: mid, prominence: support, pattern: "x..x ..x."}
  bass: {family: synth_bass, tone: [soft], register: low, prominence: anchor, pattern: "x... x..x"}
  kick: {family: drums, tone: [dusty], prominence: anchor, pattern: "x... x..."}
  snare: {family: drums, tone: [dusty], prominence: support, pattern: ".... x..."}
  hat: {family: drums, tone: [dry], prominence: support, pattern: "x.x.x.x."}
  lead: {family: woodwind, tone: [airy], register: high, prominence: lead, motif: "5 . 7 . | 9 . 5 ."}
sections:
  - {id: intro, title: roof static, duration: 12s, harmony: "Am9 Fmaj9 | Cadd9 Gsus4", scene: "intro hush", variation: "establish"}
  - {id: verse, title: skyline answer, duration: 50s, harmony: "Am9 Cmaj9 | Fmaj9 Gsus4", scene: "head glide", variation: "statement"}
  - {id: bridge, title: antenna bloom, duration: 36s, harmony: "Dm9 Fmaj9 | Am9 Gsus4", scene: "bridge lift", variation: "open-register"}
  - {id: outro, title: hangup glow, duration: 26s, harmony: "Am9 Fmaj9 | Gsus4 Am9", scene: "outro home", variation: "cadence"}

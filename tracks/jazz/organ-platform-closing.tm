title: Organ Platform / Closing
description: Greasier organ-combo with faster shout release and darker cadence.
style: jazz
substyle: organ-combo
listen_mode: album-side
seed: 54102
tags: [jazz, organ, platform, closing]
key: Fmaj
tempo: 116
globals: {density: steady, brightness: warm, swing: groove}
roles:
  lead: {family: brass, tone: [present], register: high, prominence: lead, motif: "5 . 6 7 | 9 . 7 3"}
  organ: {family: organ, tone: [warm], register: mid, prominence: support, pattern: "x... x..."}
  bass: {family: bass, tone: [woody], register: low, prominence: anchor, pattern: "x... x..."}
  kick: {family: drums, tone: [live], prominence: anchor, pattern: "x...x... | ....x..."}
  snare: {family: drums, tone: [live], prominence: support, pattern: "....x... | ..x....."}
sections:
  - {id: intro, title: gate close, duration: 10s, harmony: "Gm7 C7 | Fmaj7 D7", scene: "intro lean", variation: "establish"}
  - {id: head, title: departure call, duration: 44s, harmony: "Gm7 C7 | Fmaj7 D7 | Bbmaj7 A7 | Gm7 C7", scene: "head clipped", variation: "statement"}
  - {id: shout, title: platform bark, duration: 38s, harmony: "Gm7 Gb7 | Fmaj7 D7 | Bbmaj7 A7 | Gm7 C7", scene: "shout bridge", variation: "lift-register"}
  - {id: outro, title: dark rails, duration: 26s, harmony: "Gm7 C7 | Fmaj7 Fmaj7", scene: "outro cadence", variation: "cadence"}

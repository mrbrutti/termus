title: Trumpet Stairwell / Newsstand
description: Organ-combo airport chart with clipped head and a tight bridge stop.
style: jazz
substyle: organ-combo
listen_mode: album-side
seed: 54101
tags: [jazz, trumpet, stairwell, organ]
key: Bbmaj
tempo: 118
globals: {density: steady, brightness: balanced, swing: groove}
roles:
  trumpet: {family: brass, tone: [present], register: high, prominence: lead, motif: "5 . 7 . | 9 . 7 5"}
  organ: {family: organ, tone: [warm], register: mid, prominence: support, pattern: "x... x..."}
  bass: {family: bass, tone: [woody], register: low, prominence: anchor, pattern: "x... x..."}
  ride: {family: drums, tone: [live], prominence: support, pattern: "x..x.x.. | x..x.xx."}
  snare: {family: drums, tone: [live], prominence: support, pattern: "....x... | ....x..."}
sections:
  - {id: intro, title: stairs open, duration: 12s, harmony: "Cm7 F7 | Bbmaj7 G7", scene: "intro lean", variation: "establish"}
  - {id: head, title: paper stand, duration: 46s, harmony: "Cm7 F7 | Bbmaj7 G7 | Ebmaj7 D7 | Cm7 F7", scene: "head clipped", variation: "statement"}
  - {id: bridge, title: landing light, duration: 40s, harmony: "Cm7 B7 | Bbmaj7 G7 | Ebmaj7 D7 | Cm7 F7", scene: "bridge reharm", variation: "sequence"}
  - {id: outro, title: final stairs, duration: 28s, harmony: "Cm7 F7 | Bbmaj7 Bbmaj7", scene: "outro runway", variation: "cadence"}

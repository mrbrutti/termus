title: Monorail Shout / Figure
description: Organ-combo shout chart with heavier drums and a clipped runway ending.
style: jazz
substyle: organ-combo
listen_mode: album-side
seed: 54106
tags: [jazz, monorail, shout, organ]
key: Gmaj
tempo: 120
globals: {density: busy, brightness: balanced, swing: heavy}
roles:
  lead: {family: brass, tone: [bright], register: high, prominence: lead, motif: "5 . 7 9 | 6 . 5 3"}
  organ: {family: organ, tone: [warm], register: mid, prominence: support, pattern: "x... x..."}
  bass: {family: bass, tone: [woody], register: low, prominence: anchor, pattern: "x... x..."}
  kick: {family: drums, tone: [live], prominence: anchor, pattern: "x...x... | ....x..."}
  snare: {family: drums, tone: [live], prominence: support, pattern: "....x... | ..x.xx.."}
sections:
  - {id: intro, title: overhead ping, duration: 10s, harmony: "Am7 D7 | Gmaj7 E7", scene: "intro lean", variation: "establish"}
  - {id: head, title: rail figure, duration: 42s, harmony: "Am7 D7 | Gmaj7 E7 | Cmaj7 B7 | Am7 D7", scene: "head clipped", variation: "statement"}
  - {id: shout, title: station stack, duration: 38s, harmony: "Am7 Ab7 | Gmaj7 E7 | Cmaj7 B7 | Am7 D7", scene: "shout chorus", variation: "lift-register"}
  - {id: outro, title: brake sparks, duration: 24s, harmony: "Am7 D7 | Gmaj7 Gmaj7", scene: "outro cadence", variation: "cadence"}

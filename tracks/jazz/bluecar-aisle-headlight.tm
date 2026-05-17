title: Bluecar Aisle / Headlight
description: Traditional trio-after-hours chart with longer head and a dim bridge.
style: jazz
substyle: trio-after-hours
listen_mode: album-side
seed: 54104
tags: [jazz, trio, bluecar, headlight]
key: Dmaj
tempo: 124
globals: {density: steady, brightness: warm, swing: groove}
roles:
  piano: {family: acoustic_piano, tone: [warm], register: mid, prominence: support, pattern: "x..x .x.."}
  lead: {family: reed_lead, tone: [breathy], register: mid-high, prominence: lead, motif: "5 . 6 7 | 9 . 7 3"}
  bass: {family: bass, tone: [woody], register: low, prominence: anchor, pattern: "x... x..."}
  ride: {family: drums, tone: [live], prominence: support, pattern: "x..x.x.. | x..x.xx."}
  snare: {family: drums, tone: [live], prominence: support, pattern: "....x... | ..x....."}
sections:
  - {id: intro, title: door slide, duration: 12s, harmony: "Em7 A7 | Dmaj7 B7", scene: "intro lean", variation: "establish"}
  - {id: head, title: aisle blue, duration: 50s, harmony: "Em7 A7 | Dmaj7 B7 | Gmaj7 F#7 | Em7 A7", scene: "head clipped", variation: "statement"}
  - {id: bridge, title: passing lamps, duration: 42s, harmony: "Em7 Eb7 | Dmaj7 B7 | Gmaj7 F#7 | Em7 A7", scene: "bridge reharm", variation: "sequence"}
  - {id: outro, title: last car, duration: 28s, harmony: "Em7 A7 | Dmaj7 Dmaj7", scene: "outro cadence", variation: "cadence"}

title: Lampglass Variations
description: Chamber-lantern variation study with wider strings and slower bass entries.
style: classical
substyle: chamber-lantern
listen_mode: album-side
seed: 52107
tags: [classical, lampglass, variations, chamber]
key: Bmin
tempo: 92
globals: {density: light, brightness: balanced, motion: gentle, reverb: room}
roles:
  piano: {family: acoustic_piano, tone: [clear], register: mid, prominence: support, pattern: "x..x .x.."}
  strings: {family: strings, tone: [soft], register: high, prominence: lead, pattern: "x....... | x......."}
  lead: {family: woodwind, tone: [lyrical], register: high, prominence: answer, motif: "5 . 4 . | 3 . 1 ."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... x..."}
sections:
  - {id: intro, title: shade open, duration: 18s, harmony: "Bm9 F#m7 | Gmaj9 Em9", scene: "chamber intro", variation: "establish"}
  - {id: body, title: glass turn, duration: 48s, harmony: "Bm9 Dmaj9 | Gmaj9 F#7", scene: "head answer", variation: "statement"}
  - {id: bridge, title: waltz hinge, duration: 34s, harmony: "Em9 Gmaj9 | Bm9 F#m7", scene: "bridge chamber", variation: "sequence"}
  - {id: outro, title: shade close, duration: 26s, harmony: "Bm9 F#m7 | Gmaj9 Bm9", scene: "outro cadence", variation: "cadence"}

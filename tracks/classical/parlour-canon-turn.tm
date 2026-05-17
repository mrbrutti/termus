title: Parlour Canon / Turn
description: Chamber-lantern canon with brighter strings and a clipped final cadence.
style: classical
substyle: chamber-lantern
listen_mode: album-side
seed: 52102
tags: [classical, parlour, canon, chamber]
key: Gmaj
tempo: 94
globals: {density: steady, brightness: balanced, motion: gentle, reverb: room}
roles:
  piano: {family: acoustic_piano, tone: [clear], register: mid, prominence: support, pattern: "x..x .x.."}
  strings: {family: strings, tone: [soft], register: mid-high, prominence: lead, pattern: "x....... | x......."}
  lead: {family: woodwind, tone: [lyrical], register: high, prominence: lead, motif: "5 . 4 . | 3 . 1 ."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... x..."}
sections:
  - {id: intro, title: lantern dust, duration: 16s, harmony: "Gmaj9 D/F# | Em9 Cmaj9", scene: "chamber intro", variation: "establish"}
  - {id: body, title: parlour turn, duration: 50s, harmony: "Gmaj9 Bm7 | Cmaj9 Dadd9", scene: "head statement", variation: "statement"}
  - {id: bridge, title: canon corner, duration: 36s, harmony: "Am9 Cmaj9 | Gmaj9 D/F#", scene: "bridge chamber", variation: "sequence"}
  - {id: outro, title: final lamp, duration: 26s, harmony: "Gmaj9 D/F# | Cmaj9 Gmaj9", scene: "outro cadence", variation: "cadence"}

title: Balcony Lesson Piece
description: Chamber-lantern study with brighter top line and clipped chamber tags.
style: classical
substyle: chamber-lantern
listen_mode: album-side
seed: 52106
tags: [classical, balcony, lesson, chamber]
key: Cmaj
tempo: 96
globals: {density: steady, brightness: balanced, motion: gentle, reverb: room}
roles:
  piano: {family: acoustic_piano, tone: [clear], register: mid, prominence: support, pattern: "x..x .x.."}
  strings: {family: strings, tone: [soft], register: mid-high, prominence: lead, pattern: "x....... | x......."}
  lead: {family: woodwind, tone: [lyrical], register: high, prominence: lead, motif: "5 . 6 . | 3 . 2 1"}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... x..."}
sections:
  - {id: intro, title: rail count, duration: 16s, harmony: "Cmaj9 G/B | Am9 Fmaj9", scene: "chamber intro", variation: "establish"}
  - {id: body, title: lesson phrase, duration: 52s, harmony: "Cmaj9 Em7 | Fmaj9 Gadd9", scene: "head statement", variation: "statement"}
  - {id: bridge, title: balcony sweep, duration: 36s, harmony: "Dm9 Fmaj9 | Cmaj9 G/B", scene: "bridge chamber", variation: "glide"}
  - {id: outro, title: page turn, duration: 24s, harmony: "Cmaj9 G/B | Fmaj9 Cmaj9", scene: "outro cadence", variation: "cadence"}

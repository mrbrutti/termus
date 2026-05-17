title: Midnight Study / Nocturne
description: Piano-led nocturne with a shorter development and quiet return.
style: classical
substyle: nocturne-room
listen_mode: album-side
seed: 52101
tags: [classical, midnight, study, nocturne]
key: Dmaj
tempo: 88
globals: {density: light, brightness: warm, motion: gentle, reverb: room}
roles:
  piano: {family: acoustic_piano, tone: [warm], register: mid, prominence: lead, pattern: "x..x .x.."}
  strings: {family: strings, tone: [soft], register: mid, prominence: air, pattern: "x....... | ....x..."}
  lead: {family: woodwind, tone: [lyrical], register: high, prominence: answer, motif: "5 . 6 7 | 3 . 2 1"}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... x..."}
sections:
  - {id: intro, title: lamp key, duration: 18s, harmony: "Dmaj9 A/C# | Bm9 Gmaj9", scene: "nocturne intro", variation: "establish"}
  - {id: body, title: study turn, duration: 52s, harmony: "Dmaj9 F#m7 | Gmaj9 Asus4", scene: "head statement", variation: "statement"}
  - {id: bridge, title: half-open book, duration: 34s, harmony: "Em9 Gmaj9 | Dmaj9 A/C#", scene: "bridge development", variation: "sequence"}
  - {id: outro, title: desk close, duration: 28s, harmony: "Dmaj9 A/C# | Gmaj9 Dmaj9", scene: "outro cadence", variation: "cadence"}

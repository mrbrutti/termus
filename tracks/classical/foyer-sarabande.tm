title: Foyer Sarabande
description: Small-room nocturne with slower bass steps and a suspended bridge.
style: classical
substyle: nocturne-room
listen_mode: album-side
seed: 52104
tags: [classical, foyer, sarabande, nocturne]
key: Fmaj
tempo: 86
globals: {density: light, brightness: warm, motion: still, reverb: room}
roles:
  piano: {family: acoustic_piano, tone: [warm], register: mid, prominence: lead, pattern: "x..x .x.."}
  strings: {family: strings, tone: [soft], register: mid, prominence: air, pattern: "x....... | ....x..."}
  lead: {family: woodwind, tone: [lyrical], register: high, prominence: answer, motif: "5 . 3 . | 2 . 1 ."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... x..."}
sections:
  - {id: intro, title: rug hush, duration: 18s, harmony: "Fmaj9 C/E | Dm9 Bbmaj9", scene: "nocturne intro", variation: "establish"}
  - {id: body, title: foyer slow turn, duration: 54s, harmony: "Fmaj9 Am7 | Bbmaj9 Cadd9", scene: "head statement", variation: "statement"}
  - {id: bridge, title: doorway pause, duration: 32s, harmony: "Gm9 Bbmaj9 | Fmaj9 C/E", scene: "bridge suspended", variation: "sequence"}
  - {id: outro, title: lamp dim, duration: 28s, harmony: "Fmaj9 C/E | Bbmaj9 Fmaj9", scene: "outro cadence", variation: "cadence"}

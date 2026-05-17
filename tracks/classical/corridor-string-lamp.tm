title: Corridor String / Lamp
description: String-forward chamber study with thinner piano answers and a held last bar.
style: classical
substyle: chamber-lantern
listen_mode: album-side
seed: 52103
tags: [classical, corridor, strings, lamp]
key: Amin
tempo: 90
globals: {density: light, brightness: warm, motion: still, reverb: room}
roles:
  piano: {family: acoustic_piano, tone: [warm], register: mid, prominence: support, pattern: "x..x ...."}
  strings: {family: strings, tone: [soft], register: mid-high, prominence: lead, pattern: "x....... | x......."}
  lead: {family: woodwind, tone: [lyrical], register: high, prominence: answer, motif: "5 . . 3 | 2 . 1 ."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... x..."}
sections:
  - {id: intro, title: lamplit hall, duration: 18s, harmony: "Am9 E/G# | Fmaj9 Dm9", scene: "chamber intro", variation: "establish"}
  - {id: body, title: corridor turn, duration: 48s, harmony: "Am9 Cmaj9 | Fmaj9 E7", scene: "head answer", variation: "statement"}
  - {id: bridge, title: stair landing, duration: 34s, harmony: "Dm9 Fmaj9 | Am9 E/G#", scene: "bridge development", variation: "glide"}
  - {id: outro, title: last lamp, duration: 26s, harmony: "Am9 E/G# | Fmaj9 Am9", scene: "outro cadence", variation: "cadence"}

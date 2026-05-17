title: Hush Waltz Etude
description: Small nocturne-room waltz with softer bridge air and a held final line.
style: classical
substyle: nocturne-room
listen_mode: album-side
seed: 52108
tags: [classical, hush, waltz, etude]
key: Gmin
tempo: 82
globals: {density: light, brightness: warm, motion: still, reverb: room}
roles:
  piano: {family: acoustic_piano, tone: [warm], register: mid, prominence: lead, pattern: "x..x .x.."}
  strings: {family: strings, tone: [soft], register: mid, prominence: air, pattern: "x....... | ....x..."}
  lead: {family: woodwind, tone: [lyrical], register: high, prominence: answer, motif: "5 . . 3 | 2 . 1 ."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... x..."}
sections:
  - {id: intro, title: hush lamp, duration: 18s, harmony: "Gm9 D/F# | Ebmaj9 Cm9", scene: "nocturne intro", variation: "establish"}
  - {id: body, title: narrow turn, duration: 50s, harmony: "Gm9 Bbmaj9 | Ebmaj9 D7", scene: "head statement", variation: "statement"}
  - {id: bridge, title: upper hush, duration: 34s, harmony: "Cm9 Ebmaj9 | Gm9 D/F#", scene: "bridge suspended", variation: "glide"}
  - {id: outro, title: final curtsey, duration: 28s, harmony: "Gm9 D/F# | Ebmaj9 Gm9", scene: "outro cadence", variation: "cadence"}

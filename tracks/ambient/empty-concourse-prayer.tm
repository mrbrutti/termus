title: Empty Concourse / Prayer
description: Warm choir-fog meditation with a suspended bridge and soft floor bass.
style: ambient
substyle: choir-fog
listen_mode: endless
seed: 50105
tags: [ambient, concourse, prayer, choir]
key: Amin
tempo: 52
globals: {density: light, brightness: warm, motion: still, reverb: halo}
roles:
  choir: {family: choir, tone: [soft, devotional], register: mid, prominence: air, pattern: "x....... | ....x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: synth_bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
  lead: {family: woodwind, tone: [breathy], register: high, prominence: lead, motif: "5 . . . | 2 . 1 ."}
sections:
  - {id: intro, title: escalator hush, duration: 34s, harmony: "Am9 Fmaj9 | Cadd9 Gsus4", scene: "intro fog", variation: "establish"}
  - {id: body, title: prayer strip, duration: 74s, harmony: "Am9 Dm9 | Fmaj9 Gsus4", scene: "drift still", variation: "statement"}
  - {id: bridge, title: vaulted bend, duration: 54s, harmony: "Cmaj9 G/B | Am9 Em7", scene: "bridge suspended", variation: "sequence"}
  - {id: outro, title: tiled release, duration: 42s, harmony: "Am9 Fmaj9 | Gsus4 Am9", scene: "outro home", variation: "cadence"}

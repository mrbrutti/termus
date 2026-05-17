title: Rainroom Prelude
description: Warmer nocturne-room prelude with a denser middle phrase and softer release.
style: classical
substyle: nocturne-room
listen_mode: album-side
seed: 52105
tags: [classical, rainroom, prelude, warm]
key: Ebmaj
tempo: 84
globals: {density: light, brightness: warm, motion: gentle, reverb: room}
roles:
  piano: {family: acoustic_piano, tone: [warm], register: mid, prominence: lead, pattern: "x..x .x.."}
  strings: {family: strings, tone: [soft], register: mid, prominence: air, pattern: "x....... | ....x..."}
  lead: {family: woodwind, tone: [lyrical], register: high, prominence: answer, motif: "5 . 4 . | 3 . 1 ."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... x..."}
sections:
  - {id: intro, title: rain pane, duration: 16s, harmony: "Ebmaj9 Bb/D | Cm9 Abmaj9", scene: "nocturne intro", variation: "establish"}
  - {id: body, title: desk prelude, duration: 50s, harmony: "Ebmaj9 Gm7 | Abmaj9 Bbadd9", scene: "head answer", variation: "statement"}
  - {id: bridge, title: wet page, duration: 34s, harmony: "Fm9 Abmaj9 | Ebmaj9 Bb/D", scene: "bridge development", variation: "sequence"}
  - {id: outro, title: curtain close, duration: 26s, harmony: "Ebmaj9 Bb/D | Abmaj9 Ebmaj9", scene: "outro cadence", variation: "cadence"}

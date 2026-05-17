title: Frosted Panes / Engine
description: Glass-steps frozen-window study with brisker repeats and a held finish.
style: phase
substyle: glass-steps
listen_mode: endless
seed: 57106
tags: [phase, frosted, panes, engine]
key: Ebmaj
tempo: 75
globals: {density: light, brightness: balanced, motion: moving, reverb: room}
roles:
  mallet-a: {family: mallet, tone: [glass], register: high, prominence: lead, pattern: "x... x..."}
  mallet-b: {family: mallet, tone: [soft], register: high, prominence: answer, pattern: ".... x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: air, pattern: "x....... | ....x..."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
sections:
  - {id: intro, title: pane frost, duration: 22s, harmony: "Ebmaj9 Bb/D | Cm9 Abmaj9", scene: "glass intro", variation: "establish"}
  - {id: body, title: engine loop, duration: 56s, harmony: "Ebmaj9 Gm7 | Abmaj9 Bbadd9", scene: "interlock answer", variation: "statement"}
  - {id: bridge, title: cold mirror, duration: 38s, harmony: "Fm9 Abmaj9 | Ebmaj9 Bb/D", scene: "bridge lift", variation: "sequence"}
  - {id: outro, title: thaw line, duration: 26s, harmony: "Ebmaj9 Bb/D | Abmaj9 Ebmaj9", scene: "outro cadence", variation: "cadence"}

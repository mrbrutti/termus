title: Dim Platform / Lattice
description: Glass-steps lattice with higher bell sheen and a clipped station close.
style: phase
substyle: glass-steps
listen_mode: endless
seed: 57108
tags: [phase, platform, lattice, glass]
key: Fmaj
tempo: 77
globals: {density: light, brightness: balanced, motion: moving, reverb: room}
roles:
  mallet-a: {family: mallet, tone: [glass], register: high, prominence: lead, pattern: "x... x..."}
  mallet-b: {family: mallet, tone: [soft], register: high, prominence: answer, pattern: ".... x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: air, pattern: "x....... | ....x..."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
sections:
  - {id: intro, title: lattice glow, duration: 22s, harmony: "Fmaj9 C/E | Dm9 Bbmaj9", scene: "glass intro", variation: "establish"}
  - {id: body, title: dim platform, duration: 56s, harmony: "Fmaj9 Am7 | Bbmaj9 Cadd9", scene: "interlock answer", variation: "statement"}
  - {id: bridge, title: tiled echo, duration: 38s, harmony: "Gm9 Bbmaj9 | Fmaj9 C/E", scene: "bridge lift", variation: "sequence"}
  - {id: outro, title: final lattice, duration: 26s, harmony: "Fmaj9 C/E | Bbmaj9 Fmaj9", scene: "outro cadence", variation: "cadence"}

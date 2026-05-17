title: Warm Hall / Interlock
description: Pad-forward interlock with softer mallet tails and a held release.
style: phase
substyle: warm-interlock
listen_mode: endless
seed: 57103
tags: [phase, hall, interlock, warm]
key: Fmaj
tempo: 70
globals: {density: light, brightness: warm, motion: moving, reverb: room}
roles:
  mallet-a: {family: mallet, tone: [soft], register: high, prominence: lead, pattern: "x... x..."}
  mallet-b: {family: mallet, tone: [soft], register: high, prominence: answer, pattern: ".... x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: air, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
sections:
  - {id: intro, title: warm tile, duration: 20s, harmony: "Fmaj9 C/E | Dm9 Bbmaj9", scene: "warm intro", variation: "establish"}
  - {id: body, title: hall turn, duration: 52s, harmony: "Fmaj9 Am7 | Bbmaj9 Cadd9", scene: "interlock answer", variation: "statement"}
  - {id: bridge, title: side corridor, duration: 34s, harmony: "Gm9 Bbmaj9 | Fmaj9 C/E", scene: "bridge warm", variation: "glide"}
  - {id: outro, title: close tile, duration: 24s, harmony: "Fmaj9 C/E | Bbmaj9 Fmaj9", scene: "outro cadence", variation: "cadence"}

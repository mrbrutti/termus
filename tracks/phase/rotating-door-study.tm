title: Rotating Door / Study
description: Warm-interlock study with slower top cells and a narrow close.
style: phase
substyle: warm-interlock
listen_mode: endless
seed: 57107
tags: [phase, rotating, door, study]
key: Bmin
tempo: 69
globals: {density: light, brightness: warm, motion: moving, reverb: room}
roles:
  mallet-a: {family: mallet, tone: [soft], register: high, prominence: lead, pattern: "x... x..."}
  mallet-b: {family: mallet, tone: [soft], register: high, prominence: answer, pattern: ".... x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
sections:
  - {id: intro, title: hinge count, duration: 20s, harmony: "Bm9 F#m7 | Gmaj9 Em9", scene: "warm intro", variation: "establish"}
  - {id: body, title: door study, duration: 50s, harmony: "Bm9 Dmaj9 | Gmaj9 F#7", scene: "interlock answer", variation: "statement"}
  - {id: bridge, title: brass edge, duration: 34s, harmony: "Em9 Gmaj9 | Bm9 F#m7", scene: "bridge warm", variation: "glide"}
  - {id: outro, title: final hinge, duration: 24s, harmony: "Bm9 F#m7 | Gmaj9 Bm9", scene: "outro cadence", variation: "cadence"}

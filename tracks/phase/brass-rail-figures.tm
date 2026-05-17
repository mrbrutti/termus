title: Brass Rail / Figures
description: Warm-interlock study with longer pad glue and narrower top answers.
style: phase
substyle: warm-interlock
listen_mode: endless
seed: 57102
tags: [phase, brass, rail, warm]
key: Dmin
tempo: 72
globals: {density: light, brightness: warm, motion: moving, reverb: room}
roles:
  mallet-a: {family: mallet, tone: [soft], register: high, prominence: lead, pattern: "x... x..."}
  mallet-b: {family: mallet, tone: [soft], register: high, prominence: answer, pattern: ".... x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
sections:
  - {id: intro, title: rail light, duration: 20s, harmony: "Dm9 Bbmaj9 | Cadd9 A7", scene: "warm intro", variation: "establish"}
  - {id: body, title: brass repeats, duration: 54s, harmony: "Dm9 Gm9 | Bbmaj9 A7", scene: "interlock answer", variation: "statement"}
  - {id: bridge, title: handrail glow, duration: 36s, harmony: "Fmaj9 C/E | Dm9 Gsus4", scene: "bridge warm", variation: "glide"}
  - {id: outro, title: lower floor, duration: 24s, harmony: "Dm9 Bbmaj9 | A7 Dm9", scene: "outro cadence", variation: "cadence"}

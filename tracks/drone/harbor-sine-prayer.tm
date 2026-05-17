title: Harbor Sine / Prayer
description: Cathedral-bed harbor drone with devotional upper bell calls.
style: drone
substyle: cathedral-bed
listen_mode: endless
seed: 53103
tags: [drone, harbor, sine, prayer]
key: Fmin
tempo: 42
globals: {density: sparse, brightness: warm, motion: still, reverb: cathedral}
roles:
  bed: {family: choir, tone: [soft, wide], register: mid, prominence: air, pattern: "x....... | x......."}
  bass: {family: synth_bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
  bells: {family: bells, tone: [dark], register: high, prominence: answer, pattern: "....x... | ........"}
  lead: {family: woodwind, tone: [hollow], register: high, prominence: lead, motif: "5 . . . | 2 . . ."}
sections:
  - {id: intro, title: harbor lowlight, duration: 36s, harmony: "Fm9 Dbmaj9 | Ebadd9 C7", scene: "intro field", variation: "establish"}
  - {id: body, title: quay prayer, duration: 86s, harmony: "Fm9 Bbm9 | Dbmaj9 C7", scene: "drift still", variation: "statement"}
  - {id: bridge, title: rope shadow, duration: 60s, harmony: "Abmaj9 Eb/G | Fm9 Dbmaj9", scene: "bridge lift", variation: "glide"}
  - {id: outro, title: wet beam, duration: 48s, harmony: "Fm9 Dbmaj9 | C7 Fm9", scene: "outro release", variation: "cadence"}

title: Rain Escalator / Vigil
description: Choir-fog vigil with a suspended upper register bridge and still close.
style: ambient
substyle: choir-fog
listen_mode: endless
seed: 50107
tags: [ambient, rain, escalator, vigil]
key: Bmin
tempo: 55
globals: {density: light, brightness: warm, motion: gentle, reverb: halo}
roles:
  choir: {family: choir, tone: [soft], register: mid, prominence: air, pattern: "x....... | ....x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [round], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
  lead: {family: woodwind, tone: [breathy], register: high, prominence: lead, motif: "5 . . . | 3 . 2 ."}
sections:
  - {id: intro, title: wet comb, duration: 30s, harmony: "Bm9 Gmaj9 | Asus4 F#m7", scene: "intro prayer", variation: "establish"}
  - {id: body, title: landing hum, duration: 70s, harmony: "Bm9 Em9 | Gmaj9 Asus4", scene: "head drift", variation: "statement"}
  - {id: bridge, title: upper rail, duration: 52s, harmony: "Dmaj9 A/C# | Gmaj9 F#m7", scene: "bridge lift", variation: "open-register"}
  - {id: outro, title: bottom floor, duration: 38s, harmony: "Bm9 Gmaj9 | Asus4 Bm9", scene: "outro home", variation: "cadence"}

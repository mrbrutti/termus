title: Choir Fog / Overpass
description: Layered choir haze with slow bridge lift and a snowed-out return.
style: ambient
substyle: choir-fog
listen_mode: endless
seed: 50101
tags: [ambient, choir, fog, overpass]
key: Gmin
tempo: 56
globals: {density: light, brightness: warm, motion: gentle, reverb: halo}
roles:
  choir: {family: choir, tone: [soft, mist], register: mid, prominence: air, pattern: "x....... | ....x..."}
  pad: {family: pad, tone: [wide, soft], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [round], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
  lead: {family: woodwind, tone: [breathy], register: high, prominence: lead, motif: "5 . 3 . | 1 . . ."}
sections:
  - {id: intro, title: sodium veil, duration: 32s, harmony: "Gm9 Ebmaj9 | Fsus4 Dm9", scene: "intro haze", variation: "establish"}
  - {id: drift, title: service road blur, duration: 72s, harmony: "Gm9 Dm9 | Ebmaj9 Fsus4", scene: "drift answer", variation: "glide"}
  - {id: bridge, title: overpass blue, duration: 58s, harmony: "Cm9 Bbmaj9 | Gm9 Dm9", scene: "bridge lift", variation: "open-register"}
  - {id: outro, title: snow return, duration: 40s, harmony: "Gm9 Ebmaj9 | Fsus4 Gm9", scene: "outro home", variation: "cadence"}

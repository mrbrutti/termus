title: Dusk Vesper / Ladder
description: Glass-lit vespers piece with shorter bridge steps and a restrained tag.
style: bells
substyle: vespers-glass
listen_mode: album-side
seed: 51105
tags: [bells, dusk, vesper, ladder]
key: Gmaj
tempo: 56
globals: {density: light, brightness: warm, motion: gentle, reverb: halo}
roles:
  bells: {family: bells, tone: [glass, soft], register: high, prominence: lead, motif: "5 . . 7 | 9 . 7 5"}
  pad: {family: pad, tone: [wide], register: mid, prominence: air, pattern: "x....... | ....x..."}
  choir: {family: choir, tone: [soft], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
sections:
  - {id: intro, title: rung glass, duration: 20s, harmony: "Gmaj9 D/F# | Em9 Cmaj9", scene: "vespers intro", variation: "establish"}
  - {id: body, title: fading rung, duration: 52s, harmony: "Gmaj9 Bm7 | Cmaj9 Dadd9", scene: "devotional answer", variation: "statement"}
  - {id: bridge, title: upper chapel, duration: 40s, harmony: "Am9 Cmaj9 | Gmaj9 D/F#", scene: "bridge release", variation: "glide"}
  - {id: outro, title: dusk tag, duration: 28s, harmony: "Gmaj9 D/F# | Cmaj9 Gmaj9", scene: "outro cadence", variation: "cadence"}

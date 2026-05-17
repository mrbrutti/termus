title: Porcelain Night / Doll
description: Staircase-song nursery piece with thinner release and a brighter bridge.
style: lullaby
substyle: staircase-song
listen_mode: album-side
seed: 56104
tags: [lullaby, porcelain, doll, staircase]
key: Amin
tempo: 69
globals: {density: light, brightness: balanced, motion: gentle, reverb: room}
roles:
  lead: {family: music_box, tone: [soft], register: high, prominence: lead, motif: "5 . . 3 | 2 . 1 ."}
  choir: {family: choir, tone: [soft], register: mid, prominence: air, pattern: "x....... | ....x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
sections:
  - {id: intro, title: shelf hinge, duration: 18s, harmony: "Am9 E/G# | Fmaj9 Dm9", scene: "intro staircase", variation: "establish"}
  - {id: verse, title: doll turn, duration: 50s, harmony: "Am9 Cmaj9 | Fmaj9 E7", scene: "verse settle", variation: "statement"}
  - {id: bridge, title: moon stair, duration: 30s, harmony: "Dm9 Fmaj9 | Am9 E/G#", scene: "bridge hush", variation: "glide"}
  - {id: outro, title: porcelain sleep, duration: 24s, harmony: "Am9 E/G# | Fmaj9 Am9", scene: "outro sleep", variation: "cadence"}

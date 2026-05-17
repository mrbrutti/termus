title: Soft Drawer / Lantern
description: Staircase-song lullaby with higher box register and a quiet turning bridge.
style: lullaby
substyle: staircase-song
listen_mode: album-side
seed: 56106
tags: [lullaby, drawer, lantern, staircase]
key: Ebmaj
tempo: 67
globals: {density: light, brightness: balanced, motion: gentle, reverb: room}
roles:
  lead: {family: music_box, tone: [soft], register: high, prominence: lead, motif: "5 . . 3 | 2 . 1 ."}
  choir: {family: choir, tone: [soft], register: mid, prominence: air, pattern: "x....... | ....x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
sections:
  - {id: intro, title: drawer light, duration: 18s, harmony: "Ebmaj9 Bb/D | Cm9 Abmaj9", scene: "intro staircase", variation: "establish"}
  - {id: verse, title: lantern fold, duration: 48s, harmony: "Ebmaj9 Gm7 | Abmaj9 Bbadd9", scene: "verse settle", variation: "statement"}
  - {id: bridge, title: stair breath, duration: 28s, harmony: "Fm9 Abmaj9 | Ebmaj9 Bb/D", scene: "bridge hush", variation: "glide"}
  - {id: outro, title: drawer close, duration: 24s, harmony: "Ebmaj9 Bb/D | Abmaj9 Ebmaj9", scene: "outro sleep", variation: "cadence"}

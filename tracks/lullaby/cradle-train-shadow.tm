title: Cradle Train / Shadow
description: Paper-box nursery study with lower bass shadow and quicker release.
style: lullaby
substyle: paper-box
listen_mode: album-side
seed: 56107
tags: [lullaby, cradle, train, shadow]
key: Bmin
tempo: 63
globals: {density: light, brightness: warm, motion: still, reverb: room}
roles:
  lead: {family: music_box, tone: [soft], register: high, prominence: lead, motif: "5 . 3 . | 2 . 1 ."}
  choir: {family: choir, tone: [soft], register: mid, prominence: air, pattern: "x....... | ....x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
sections:
  - {id: intro, title: carriage hush, duration: 16s, harmony: "Bm9 F#m7 | Gmaj9 Em9", scene: "intro paper", variation: "establish"}
  - {id: verse, title: blanket rail, duration: 44s, harmony: "Bm9 Dmaj9 | Gmaj9 F#7", scene: "verse settle", variation: "statement"}
  - {id: release, title: shadow floor, duration: 26s, harmony: "Em9 Gmaj9 | Bm9 F#m7", scene: "release sleep", variation: "subtract"}
  - {id: outro, title: cradle dark, duration: 22s, harmony: "Bm9 F#m7 | Gmaj9 Bm9", scene: "outro sleep", variation: "cadence"}

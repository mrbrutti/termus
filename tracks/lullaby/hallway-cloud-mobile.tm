title: Hallway Cloud / Mobile
description: Paper-box cradle song with softer box tones and a slower final hush.
style: lullaby
substyle: paper-box
listen_mode: album-side
seed: 56101
tags: [lullaby, hallway, cloud, mobile]
key: Cmaj
tempo: 66
globals: {density: light, brightness: warm, motion: still, reverb: room}
roles:
  lead: {family: music_box, tone: [soft], register: high, prominence: lead, motif: "5 . 3 . | 2 . 1 ."}
  choir: {family: choir, tone: [soft], register: mid, prominence: air, pattern: "x....... | ....x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
sections:
  - {id: intro, title: doorframe hush, duration: 18s, harmony: "Cmaj9 Am7 | Fmaj9 Gsus4", scene: "intro paper", variation: "establish"}
  - {id: verse, title: mobile turn, duration: 48s, harmony: "Cmaj9 Em7 | Fmaj9 Gsus4", scene: "verse settle", variation: "statement"}
  - {id: release, title: floorboard dim, duration: 30s, harmony: "Am9 Fmaj9 | Cmaj9 Gsus4", scene: "release sleep", variation: "subtract"}
  - {id: outro, title: cloud close, duration: 24s, harmony: "Cmaj9 Am7 | Fmaj9 Cmaj9", scene: "outro sleep", variation: "cadence"}

title: Wool Lamp / Whisper
description: Paper-box room lullaby with deeper pad support and a short held close.
style: lullaby
substyle: paper-box
listen_mode: album-side
seed: 56105
tags: [lullaby, wool, lamp, whisper]
key: Fmaj
tempo: 65
globals: {density: light, brightness: warm, motion: still, reverb: room}
roles:
  lead: {family: music_box, tone: [soft], register: high, prominence: lead, motif: "5 . 3 . | 2 . 1 ."}
  choir: {family: choir, tone: [soft], register: mid, prominence: air, pattern: "x....... | ....x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
sections:
  - {id: intro, title: wool shade, duration: 16s, harmony: "Fmaj9 C/E | Dm9 Bbmaj9", scene: "intro paper", variation: "establish"}
  - {id: verse, title: dim whisper, duration: 46s, harmony: "Fmaj9 Am7 | Bbmaj9 Cadd9", scene: "verse settle", variation: "statement"}
  - {id: release, title: floor hush, duration: 28s, harmony: "Dm9 Bbmaj9 | Fmaj9 C/E", scene: "release sleep", variation: "subtract"}
  - {id: outro, title: lamp off, duration: 22s, harmony: "Fmaj9 C/E | Bbmaj9 Fmaj9", scene: "outro sleep", variation: "cadence"}

title: Snow Window / Hum
description: Paper-box winter lullaby with slower choir bed and a short release.
style: lullaby
substyle: paper-box
listen_mode: album-side
seed: 56103
tags: [lullaby, snow, window, winter]
key: Gmaj
tempo: 64
globals: {density: light, brightness: warm, motion: still, reverb: room}
roles:
  lead: {family: music_box, tone: [soft], register: high, prominence: lead, motif: "5 . 3 . | 2 . 1 ."}
  choir: {family: choir, tone: [soft], register: mid, prominence: air, pattern: "x....... | ....x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
sections:
  - {id: intro, title: sill frost, duration: 16s, harmony: "Gmaj9 D/F# | Em9 Cmaj9", scene: "intro paper", variation: "establish"}
  - {id: verse, title: glass breath, duration: 46s, harmony: "Gmaj9 Bm7 | Cmaj9 Dadd9", scene: "verse settle", variation: "statement"}
  - {id: release, title: mitten hush, duration: 28s, harmony: "Em9 Cmaj9 | Gmaj9 D/F#", scene: "release sleep", variation: "subtract"}
  - {id: outro, title: quiet pane, duration: 22s, harmony: "Gmaj9 D/F# | Cmaj9 Gmaj9", scene: "outro sleep", variation: "cadence"}

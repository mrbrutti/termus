title: Attic Teacup / Song
description: Staircase-song lullaby with higher box figures and a narrow bridge.
style: lullaby
substyle: staircase-song
listen_mode: album-side
seed: 56102
tags: [lullaby, attic, teacup, staircase]
key: Dmaj
tempo: 68
globals: {density: light, brightness: balanced, motion: gentle, reverb: room}
roles:
  lead: {family: music_box, tone: [soft], register: high, prominence: lead, motif: "5 . . 3 | 2 . 1 ."}
  choir: {family: choir, tone: [soft], register: mid, prominence: air, pattern: "x....... | ....x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
sections:
  - {id: intro, title: attic hinge, duration: 18s, harmony: "Dmaj9 A/C# | Bm9 Gmaj9", scene: "intro staircase", variation: "establish"}
  - {id: verse, title: teacup sway, duration: 50s, harmony: "Dmaj9 F#m7 | Gmaj9 Asus4", scene: "verse settle", variation: "statement"}
  - {id: bridge, title: rafters dim, duration: 30s, harmony: "Em9 Gmaj9 | Dmaj9 A/C#", scene: "bridge hush", variation: "glide"}
  - {id: outro, title: cup sleep, duration: 24s, harmony: "Dmaj9 A/C# | Gmaj9 Dmaj9", scene: "outro sleep", variation: "cadence"}

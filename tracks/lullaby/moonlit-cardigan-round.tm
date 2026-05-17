title: Moonlit Cardigan / Round
description: Staircase-song round with softer choir bed and a narrow late cadence.
style: lullaby
substyle: staircase-song
listen_mode: album-side
seed: 56108
tags: [lullaby, moonlit, cardigan, round]
key: Gmin
tempo: 68
globals: {density: light, brightness: balanced, motion: gentle, reverb: room}
roles:
  lead: {family: music_box, tone: [soft], register: high, prominence: lead, motif: "5 . . 3 | 2 . 1 ."}
  choir: {family: choir, tone: [soft], register: mid, prominence: air, pattern: "x....... | ....x..."}
  pad: {family: pad, tone: [wide], register: mid, prominence: support, pattern: "x....... | x......."}
  bass: {family: bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
sections:
  - {id: intro, title: knit hush, duration: 18s, harmony: "Gm9 D/F# | Ebmaj9 Cm9", scene: "intro staircase", variation: "establish"}
  - {id: verse, title: cardigan round, duration: 48s, harmony: "Gm9 Bbmaj9 | Ebmaj9 D7", scene: "verse settle", variation: "statement"}
  - {id: bridge, title: moon stair, duration: 28s, harmony: "Cm9 Ebmaj9 | Gm9 D/F#", scene: "bridge hush", variation: "glide"}
  - {id: outro, title: round sleep, duration: 24s, harmony: "Gm9 D/F# | Ebmaj9 Gm9", scene: "outro sleep", variation: "cadence"}

title: Side Altar / Resonance
description: Cloister-rain side-room setting with lower box figures and a hushed close.
style: bells
substyle: cloister-rain
listen_mode: album-side
seed: 51106
tags: [bells, altar, resonance, cloister]
key: Cmin
tempo: 51
globals: {density: light, brightness: balanced, motion: still, reverb: halo}
roles:
  bells: {family: bells, tone: [soft], register: high, prominence: lead, motif: "5 . 6 . | 7 . 5 ."}
  box: {family: music_box, tone: [soft], register: mid-high, prominence: answer, pattern: "..x..... | ....x..."}
  choir: {family: choir, tone: [wide], register: mid, prominence: air, pattern: "x....... | ....x..."}
  bass: {family: bass, tone: [round], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
sections:
  - {id: intro, title: incense seam, duration: 24s, harmony: "Cm9 Abmaj9 | Bbadd9 G7", scene: "chapel intro", variation: "establish"}
  - {id: body, title: stone answer, duration: 52s, harmony: "Cm9 Fm9 | Abmaj9 G7", scene: "threshold answer", variation: "statement"}
  - {id: release, title: arch hush, duration: 38s, harmony: "Ebmaj9 Bb/D | Cm9 G7", scene: "release prayer", variation: "subtract"}
  - {id: outro, title: altar dim, duration: 30s, harmony: "Cm9 Abmaj9 | G7 Cm9", scene: "outro cadence", variation: "cadence"}

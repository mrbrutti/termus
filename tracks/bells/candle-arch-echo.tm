title: Candle Arch / Echo
description: Glass-led devotional with slow answer phrases and a reflective close.
style: bells
substyle: vespers-glass
listen_mode: album-side
seed: 51101
tags: [bells, candle, arch, vespers]
key: Bbmin
tempo: 54
globals: {density: light, brightness: warm, motion: still, reverb: halo}
roles:
  bells: {family: bells, tone: [glass, soft], register: high, prominence: lead, motif: "5 . . 7 | 9 . 7 5"}
  box: {family: music_box, tone: [soft], register: high, prominence: answer, pattern: "..x..... | ....x..."}
  choir: {family: choir, tone: [wide], register: mid, prominence: air, pattern: "x....... | ....x..."}
  bass: {family: bass, tone: [round], register: low, prominence: anchor, pattern: "x... .... | ....x..."}
sections:
  - {id: intro, title: wick room, duration: 24s, harmony: "Bbm9 Fsus4 | Gbmaj9 F7", scene: "vespers intro", variation: "establish"}
  - {id: answer, title: arch answer, duration: 54s, harmony: "Bbm9 Ebm9 | Gbmaj9 F7", scene: "devotional answer", variation: "statement"}
  - {id: release, title: stone hush, duration: 42s, harmony: "Dbmaj9 Ab/C | Bbm9 F7", scene: "release prayer", variation: "subtract"}
  - {id: outro, title: spent wax, duration: 30s, harmony: "Bbm9 Fsus4 | Gbmaj9 Bbm9", scene: "outro cadence", variation: "cadence"}

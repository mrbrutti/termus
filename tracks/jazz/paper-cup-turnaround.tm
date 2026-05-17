title: Paper Cup / Turnaround
description: Vibes-cellar turnover chart with shorter phrases and a dry late tag.
style: jazz
substyle: vibes-cellar
listen_mode: album-side
seed: 54107
tags: [jazz, paper, cup, turnaround]
key: Amin
tempo: 130
globals: {density: steady, brightness: balanced, swing: heavy}
roles:
  vibes: {family: mallet, tone: [soft], register: high, prominence: support, pattern: "x..x .x.."}
  lead: {family: reed_lead, tone: [breathy], register: mid-high, prominence: lead, motif: "5 . 7 9 | 6 . 5 3"}
  bass: {family: bass, tone: [woody], register: low, prominence: anchor, pattern: "x... x..."}
  ride: {family: drums, tone: [live], prominence: support, pattern: "x..x.x.. | x..x.xx."}
  rim: {family: drums, tone: [dry], prominence: support, pattern: "...x.... | ....x.x."}
sections:
  - {id: intro, title: cup stain, duration: 10s, harmony: "Bm7 E7 | Am7 F#7", scene: "intro lean", variation: "establish"}
  - {id: head, title: curb turnaround, duration: 44s, harmony: "Bm7 E7 | Am7 F#7 | Dmaj7 C#7 | Bm7 E7", scene: "head clipped", variation: "statement"}
  - {id: bridge, title: alley lamps, duration: 36s, harmony: "Bm7 Bb7 | Am7 F#7 | Dmaj7 C#7 | Bm7 E7", scene: "bridge reharm", variation: "sequence"}
  - {id: outro, title: dry tag, duration: 24s, harmony: "Bm7 E7 | Am7 Am7", scene: "outro cadence", variation: "cadence"}

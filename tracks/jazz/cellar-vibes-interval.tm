title: Cellar Vibes / Interval
description: Vibes-cellar chart with mallet comp and a narrower tenor release.
style: jazz
substyle: vibes-cellar
listen_mode: album-side
seed: 54103
tags: [jazz, vibes, cellar, interval]
key: Ebmaj
tempo: 132
globals: {density: steady, brightness: balanced, swing: heavy}
roles:
  vibes: {family: mallet, tone: [soft, glass], register: high, prominence: support, pattern: "x..x .x.."}
  lead: {family: reed_lead, tone: [breathy], register: mid-high, prominence: lead, motif: "5 . 7 9 | 6 . 5 3"}
  bass: {family: bass, tone: [woody], register: low, prominence: anchor, pattern: "x... x..."}
  ride: {family: drums, tone: [live], prominence: support, pattern: "x..x.x.. | x..x.xx."}
  rim: {family: drums, tone: [dry], prominence: support, pattern: "...x.... | ....x.x."}
sections:
  - {id: intro, title: basement count, duration: 12s, harmony: "Fm7 Bb7 | Ebmaj7 C7", scene: "intro lean", variation: "establish"}
  - {id: head, title: cellar interval, duration: 48s, harmony: "Fm7 Bb7 | Ebmaj7 C7 | Abmaj7 G7 | Fm7 Bb7", scene: "head clipped", variation: "statement"}
  - {id: bridge, title: stone table, duration: 40s, harmony: "Fm7 E7 | Ebmaj7 C7 | Abmaj7 G7 | Fm7 Bb7", scene: "bridge reharm", variation: "sequence"}
  - {id: outro, title: quiet stairs, duration: 26s, harmony: "Fm7 Bb7 | Ebmaj7 Ebmaj7", scene: "outro cadence", variation: "cadence"}

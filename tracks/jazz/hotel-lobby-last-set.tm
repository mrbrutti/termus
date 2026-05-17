title: Hotel Lobby / Last Set
description: Trio-after-hours room chart with softer top line and slower release.
style: jazz
substyle: trio-after-hours
listen_mode: album-side
seed: 54105
tags: [jazz, hotel, lobby, trio]
key: Cmaj
tempo: 122
globals: {density: steady, brightness: warm, swing: groove}
roles:
  piano: {family: acoustic_piano, tone: [warm], register: mid, prominence: support, pattern: "x..x .x.."}
  lead: {family: reed_lead, tone: [breathy], register: mid-high, prominence: lead, motif: "5 . . 7 | 9 . 7 5"}
  bass: {family: bass, tone: [woody], register: low, prominence: anchor, pattern: "x... x..."}
  ride: {family: drums, tone: [live], prominence: support, pattern: "x..x.x.. | x..x.xx."}
  rim: {family: drums, tone: [dry], prominence: support, pattern: "...x.... | ....x.x."}
sections:
  - {id: intro, title: carpet hush, duration: 10s, harmony: "Dm7 G7 | Cmaj7 A7", scene: "intro lean", variation: "establish"}
  - {id: head, title: empty barcloth, duration: 46s, harmony: "Dm7 G7 | Cmaj7 A7 | Fmaj7 E7 | Dm7 G7", scene: "head clipped", variation: "statement"}
  - {id: release, title: lamped booth, duration: 36s, harmony: "Dm7 Db7 | Cmaj7 A7 | Fmaj7 E7 | Dm7 G7", scene: "release answer", variation: "subtract"}
  - {id: outro, title: last set, duration: 26s, harmony: "Dm7 G7 | Cmaj7 Cmaj7", scene: "outro cadence", variation: "cadence"}

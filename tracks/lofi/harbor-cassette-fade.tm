title: Harbor Cassette / Fade
description: Guitar-neon tape piece with synth bass bloom and a clipped floor return.
style: lofi
substyle: guitar-neon
listen_mode: album-side
seed: 55105
tags: [lofi, harbor, cassette, neon]
key: Gmin
tempo: 82
globals: {density: steady, brightness: balanced, motion: gentle, reverb: room}
roles:
  guitar: {family: guitar, tone: [warm], register: mid, prominence: support, pattern: "x..x ..x."}
  bass: {family: synth_bass, tone: [soft], register: low, prominence: anchor, pattern: "x... x..x"}
  kick: {family: drums, tone: [dusty], prominence: anchor, pattern: "x... x..."}
  snare: {family: drums, tone: [dusty], prominence: support, pattern: ".... x..."}
  hat: {family: drums, tone: [dry], prominence: support, pattern: "x.x.x.x."}
  lead: {family: reed_lead, tone: [breathy], register: mid-high, prominence: lead, motif: "5 . 7 . | 9 . 5 ."}
sections:
  - {id: intro, title: tape wake, duration: 12s, harmony: "Gm9 Ebmaj9 | Fsus4 Dm9", scene: "intro hush", variation: "establish"}
  - {id: verse, title: harbor lane, duration: 52s, harmony: "Gm9 Bbmaj9 | Ebmaj9 Fsus4", scene: "head glide", variation: "statement"}
  - {id: bridge, title: dockline glow, duration: 36s, harmony: "Cm9 Ebmaj9 | Gm9 Fsus4", scene: "bridge lift", variation: "open-register"}
  - {id: outro, title: cassette drag, duration: 26s, harmony: "Gm9 Ebmaj9 | Fsus4 Gm9", scene: "outro home", variation: "cadence"}

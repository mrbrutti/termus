title: Signal Snow / Foyer
description: Thin station-haze foyer piece with shorter phrases and an icy release.
style: ambient
substyle: station-haze
listen_mode: endless
seed: 50108
tags: [ambient, signal, snow, foyer]
key: Dmaj
tempo: 59
globals: {density: sparse, brightness: balanced, motion: still, reverb: room}
roles:
  pad: {family: pad, tone: [wide], register: mid, prominence: air, pattern: "x....... | ....x..."}
  bells: {family: bells, tone: [glass], register: high, prominence: answer, pattern: "....x... | ..x....."}
  bass: {family: synth_bass, tone: [soft], register: low, prominence: anchor, pattern: "x... .... | x... ...."}
  lead: {family: woodwind, tone: [hollow], register: high, prominence: lead, motif: "5 . 4 . | 3 . 1 ."}
sections:
  - {id: intro, title: front mat, duration: 24s, harmony: "Dmaj9 A/C# | Bm9 Gmaj9", scene: "intro haze", variation: "establish"}
  - {id: drift, title: snowfall cue, duration: 62s, harmony: "Dmaj9 F#m7 | Gmaj9 Asus4", scene: "head answer", variation: "statement"}
  - {id: release, title: dim concierge, duration: 44s, harmony: "Bm9 Gmaj9 | Dmaj9 Asus4", scene: "release thin", variation: "subtract"}
  - {id: outro, title: signal white, duration: 34s, harmony: "Dmaj9 A/C# | Gmaj9 Dmaj9", scene: "outro snow", variation: "cadence"}

title: Midnight Ballad Slow
description: Slow ballad trio with tritone subs on V, expression swells, and a brushed feel.
style: jazz
mix_bus: jazz
listen_mode: album-side
seed: 62738
tags: [jazz, ballad, slow, piano, brushes]
key: Dbmaj
tempo: 58
globals: {density: sparse, brightness: warm, phrase: long}
roles:
  piano:
    family: acoustic_piano
    tone: [clear, warm]
    articulation: lyrical
    register: mid
    prominence: lead
    motif: "9 . 7 . 5 . 3 . | 1 . . . . . . ."
  bass:
    family: bass
    tone: [woody, round]
    articulation: walk
    register: low
    prominence: anchor
    pattern: "x... x... | x... x..."
  brushes:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: support
    pattern: "x.x. x.x. | x.x. x.x."
  strings:
    family: strings
    tone: [soft, warm]
    register: mid-high
    prominence: air
    pattern: "x....... | ........"
sections:
  - id: intro
    title: late-hour hush
    duration: 14s
    harmony: "Dbmaj7 Bbm7"
    scene: "intro hush"
    variation: "establish"
    groove: straight
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.3}
          - {at: 100, value: 0.6}
  - id: head
    title: slow confession
    duration: 36s
    harmony: "Dbmaj7 Bbm7 | Ebm7 Ab7 | Dbmaj7 Bbm7 | Ebm7 Ab7"
    scene: "head lyrical"
    variation: "statement"
    groove: straight
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 0.7}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.6}
          - {at: 50, value: 0.9}
          - {at: 100, value: 0.75}
  - id: bridge
    title: tender pivot
    duration: 22s
    harmony: "Gbmaj7 F7 | Bbm7 Eb7 | Ebm7 Ab7 | Dbmaj7 Bbm7"
    scene: "bridge lift"
    variation: "open-register"
    groove: straight
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 1.0}
  - id: outro
    title: last note rings
    duration: 18s
    harmony: "Ebm7 Ab7 | Dbmaj7 Dbmaj7"
    scene: "outro cadence"
    variation: "cadence"
    groove: straight
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.75}
          - {at: 100, value: 0.25}

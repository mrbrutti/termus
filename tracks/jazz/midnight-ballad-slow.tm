title: Midnight Ballad Slow
description: Slow ballad with dense piano lead, walking bass on every quarter, brushes texture, strings halo, tritone subs.
style: jazz
mix_bus: jazz
listen_mode: album-side
seed: 62738
tags: [jazz, ballad, slow, piano, brushes, walking, strings]
key: Dbmaj
tempo: 78
globals: {density: heavy, brightness: warm, motion: moving, phrase: long}
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
  brushes:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: support
  ride:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: support
  snare:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: support
  strings:
    family: strings
    tone: [soft, warm]
    register: mid-high
    prominence: air
  comp:
    family: acoustic_piano
    tone: [soft, warm]
    articulation: comp
    register: mid
    prominence: support
sections:
  - id: intro
    title: late-hour hush
    duration: 14s
    harmony: "Dbmaj9 Bbm11 | Ebm9 Ab9sus4"
    scene: "intro hush"
    variation: "establish"
    groove: straight
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.35}
          - {at: 100, value: 0.65}
  - id: head
    title: slow confession
    duration: 32s
    harmony: "Dbmaj9 Bbm11 | Ebm9 Ab9sus4 | Dbmaj9 Bbm11 | Ebm9 Ab9sus4"
    scene: "head lyrical"
    variation: "statement"
    groove: straight
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 0.8}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.65}
          - {at: 50, value: 0.95}
          - {at: 100, value: 0.8}
  - id: bridge
    title: tender pivot
    duration: 24s
    harmony: "Gbmaj9 F9 | Bbm11 Eb9sus4 | Ebm9 Ab9sus4 | Dbmaj9 Bbm11"
    scene: "bridge lift"
    variation: "open-register"
    groove: straight
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 1.0}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.8}
          - {at: 60, value: 1.0}
          - {at: 100, value: 0.85}
  - id: head2
    title: recapitulation
    duration: 24s
    harmony: "Dbmaj9 Bbm11 | Ebm9 Ab9sus4 | Dbmaj9 Bbm11 | Ebm9 Ab9sus4"
    scene: "head lyrical"
    variation: "sequence-up"
    groove: straight
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 0.9}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.85}
          - {at: 100, value: 0.8}
  - id: outro
    title: last note rings
    duration: 18s
    harmony: "Ebm9 Ab9sus4 | Dbmaj9 Dbmaj9"
    scene: "outro cadence"
    variation: "cadence"
    groove: straight
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 1.0}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.8}
          - {at: 100, value: 0.2}

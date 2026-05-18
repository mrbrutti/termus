title: Bossa Nova Rooftop
description: Bossa jazz trio with dense guitar comping, quarter walking bass, ride pattern, samba kick, and flute lead.
style: jazz
mix_bus: jazz
listen_mode: hour-stream
seed: 44190
tags: [jazz, bossa, guitar, rooftop, trio, walking, ride]
key: Gmaj
tempo: 124
globals: {density: full, brightness: balanced, motion: restless, phrase: long}
roles:
  guitar:
    family: guitar
    tone: [warm, soft]
    articulation: comp
    register: mid
    prominence: support
  piano:
    family: acoustic_piano
    tone: [clear, warm]
    articulation: comp
    register: mid
    prominence: support
  bass:
    family: bass
    tone: [woody, round]
    articulation: walk
    register: low
    prominence: anchor
  ride:
    family: drums
    tone: [live, bright]
    articulation: swing
    prominence: support
  kick:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: anchor
  snare:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: support
  flute:
    family: woodwind
    tone: [airy, soft]
    register: high
    prominence: lead
    motif: "5 . 7 9 . 7 5 3 | 2 . 1 . . . . ."
sections:
  - id: intro
    title: warm-up breeze
    duration: 12s
    harmony: "Gmaj7 Em7 | Am7 D7"
    scene: "intro hush"
    variation: "establish"
    groove: bossa_loose
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.45}
          - {at: 100, value: 0.7}
  - id: verse
    title: open terrace
    duration: 32s
    harmony: "Gmaj7 Em7 | Am7 D7"
    scene: "head glide"
    variation: "statement"
    groove: bossa_loose
    substitutions:
      - {rule: secondary_dominant, of: ii, probability: 0.8}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.7}
          - {at: 100, value: 0.8}
  - id: bridge
    title: moon-rise shift
    duration: 24s
    harmony: "Cmaj7 B7 | Em7 A7 | Am7 D7 | Gmaj7 E7"
    scene: "bridge lift"
    variation: "open-register"
    groove: bossa_loose
    substitutions:
      - {rule: secondary_dominant, of: ii, probability: 1.0}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.8}
          - {at: 60, value: 1.0}
          - {at: 100, value: 0.85}
  - id: verse2
    title: second chorus
    duration: 26s
    harmony: "Gmaj7 Em7 | Am7 D7"
    scene: "head glide"
    variation: "sequence-up"
    groove: bossa_loose
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 0.9}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.85}
          - {at: 100, value: 0.8}
  - id: outro
    title: last balcony
    duration: 18s
    harmony: "Am7 D7 | Gmaj7 Gmaj7"
    scene: "outro cadence"
    variation: "cadence"
    groove: bossa_loose
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 0.8}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.8}
          - {at: 100, value: 0.3}

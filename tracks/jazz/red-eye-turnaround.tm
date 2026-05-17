title: Dusty Swing / Red Eye Turnaround
description: Leaner trumpet chart with a clipped head, darker bridge reharm, and a runway-close coda.
style: jazz
listen_mode: album-side
seed: 18421
tags: [jazz, trumpet, turnaround, midnight, quartet]
key: Fmaj
tempo: 122
globals:
  density: steady
  brightness: balanced
  swing: groove
  phrase: long
roles:
  bass:
    family: bass
    tone: [woody, round]
    articulation: walk
    register: low
    prominence: anchor
    pattern: "x...x..x | x.x...x."
  ride:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: support
    pattern: "x..x.x.. | x..x.xx."
  kick:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: anchor
    pattern: "x...x... | ....x..."
  snare:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: anchor
    pattern: "........ | ....x..."
  rim:
    family: drums
    tone: [dry, tight]
    articulation: swing
    prominence: support
    pattern: "...x.... | ....x.x."
  trumpet:
    family: brass
    tone: [present, bright]
    articulation: lyrical
    register: high
    prominence: lead
    motif: "5 . 7 . 9 . 7 5 | 3 . 2 . 1 . . ."
  organ:
    family: organ
    tone: [soft, warm]
    articulation: sustain
    register: mid
    prominence: support
    pattern: "x..x.... | ....x..x"
sections:
  - id: intro
    title: terminal lights
    duration: 12s
    harmony: "Gm7 C7 | Fmaj7 D7 | Gm7 C7 | Fmaj7 Fmaj7"
    scene: "intro lean"
    variation: "establish"
    profile:
      density: light
      brightness: warm
    roles:
      trumpet:
        active: true
        motif: "5 . . 7 . . 5 . | . . 3 . 2 . 1 ."
      ride:
        pattern: "x...x... | x..x...."
  - id: head
    title: aisle melody
    duration: 50s
    harmony: "Gm7 C7 | Fmaj7 D7 | Gm7 C7 | Fmaj7 Fmaj7 | Bbmaj7 A7 | Gm7 C7 | Fmaj7 D7 | Gm7 C7"
    scene: "head clipped"
    variation: "statement"
    roles:
      trumpet:
        active: true
        motif: "9 . 7 5 6 . 5 3 | 5 . 2 . 1 . . ."
      organ:
        active: true
    events:
      - kind: pickup
        bar: 8
        roles: [trumpet]
        motif: "5 6 7 9"
      - kind: fill
        bar: 8
        roles: [snare, ride, kick]
  - id: bridge
    title: gate change
    duration: 55s
    harmony: "Gm7 Gb7 | Fmaj7 D7 | Bbmaj7 A7 | Gm7 C7 | Ebmaj7 D7 | Gm7 C7 | Fmaj7 D7 | Gm7 C7"
    scene: "bridge reharm"
    variation: "reharm"
    profile:
      density: busy
      brightness: bright
      swing: heavy
    roles:
      trumpet:
        motif: "11 . 9 7 5 . 3 1 | 9 . b9 7 5 . 2 1"
      organ:
        active: true
        pattern: "x..x.x.. | .x..x..."
      snare:
        pattern: "....x... | ..x.xx.."
    events:
      - kind: stop
        bar: 5
        bars: 1
        roles: [trumpet, organ, bass, kick]
      - kind: fill
        bar: 8
        roles: [snare, ride]
  - id: release
    title: luggage belt
    duration: 40s
    harmony: "Bbmaj7 A7 | Gm7 C7 | Fmaj7 D7 | Gm7 C7"
    scene: "release answer"
    variation: "answer"
    roles:
      trumpet:
        motif: "9 . 7 . 5 . 3 2 | 1 . . . . . . ."
      organ:
        active: false
    events:
      - kind: drop
        bar: 2
        bars: 1
        roles: [ride, rim]
  - id: outro
    title: wheels down
    duration: 35s
    harmony: "Gm7 C7 | Fmaj7 D7 | Gm7 C7 | Fmaj7 Fmaj7"
    scene: "outro runway"
    variation: "cadence"
    profile:
      density: light
      brightness: warm
    roles:
      trumpet:
        motif: "3 . 2 . 1 . . . | . . . . . . . ."
      ride:
        pattern: "x...x... | x..x...."
    events:
      - kind: stab
        bar: 3
        roles: [organ]
        pattern: "x... ...."

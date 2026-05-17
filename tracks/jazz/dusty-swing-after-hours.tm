title: Dusty Swing / After Hours
description: Piano-trio head that opens into an alto-led bridge, then settles back into a brushed last call.
style: jazz
listen_mode: album-side
seed: 7319
tags: [jazz, swing, trio, late-set, after-hours]
key: Cmaj
tempo: 126
globals:
  density: steady
  brightness: balanced
  swing: groove
  phrase: long
roles:
  piano:
    family: acoustic_piano
    tone: [clear, present]
    articulation: comp
    register: mid
    prominence: support
    pattern: "x..x.x.. | .x..x..x"
  bass:
    family: bass
    tone: [woody, round]
    articulation: walk
    register: low
    prominence: anchor
    pattern: "x.x.x.x. | x.x.x.x."
  ride:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: support
    pattern: "x.x.x.x. | x.x.xx.x"
  kick:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: anchor
    pattern: "x....... | x...x..."
  snare:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: anchor
    pattern: "....x... | ..x.x..."
  alto:
    family: reed_lead
    tone: [present, live]
    articulation: lyrical
    register: high
    prominence: lead
    motif: "5 . 6 7 9 . 7 5 | 3 . 2 . 1 . . ."
sections:
  - id: intro
    title: count-in lamps
    duration: 35s
    harmony: "Dm7 G7 | Cmaj7 A7 | Dm7 G7 | Cmaj7 Cmaj7"
    scene: "intro trio"
    variation: "establish"
    profile:
      density: light
      brightness: warm
    roles:
      alto:
        active: false
      snare:
        pattern: "........ | ..x....."
  - id: head
    title: booth melody
    duration: 55s
    harmony: "Dm7 G7 | Cmaj7 A7 | Dm7 G7 | Cmaj7 Cmaj7 | Em7 A7 | Dm7 G7 | Em7 A7 | Dm7 G7"
    scene: "head statement"
    variation: "head"
    roles:
      alto:
        active: true
        motif: "9 . 7 5 6 . 5 3 | 5 . 2 . 1 . . ."
  - id: solo
    title: back booth lift
    duration: 60s
    harmony: "Dm7 Db7 | Cmaj7 A7 | Fmaj7 E7 | Dm7 G7 | Em7 A7 | Dm7 G7 | Cmaj7 A7 | Dm7 G7"
    scene: "solo brighter"
    variation: "turnaround-lift"
    profile:
      density: busy
      brightness: bright
      swing: heavy
    roles:
      alto:
        motif: "11 . 9 7 5 . 3 1 | 9 . b9 7 5 . 2 1"
      piano:
        pattern: "x.x..x.. | .x.x.x.."
      snare:
        pattern: "....x... | ..x.xx.."
  - id: shout
    title: room answer
    duration: 45s
    harmony: "Fmaj7 E7 | Dm7 G7 | Em7 A7 | Dm7 G7"
    scene: "shout compact"
    variation: "answer"
    profile:
      density: busy
      brightness: balanced
    roles:
      alto:
        motif: "9 . 11 9 7 . 5 3 | 5 . 6 7 9 . 7 5"
      kick:
        pattern: "x...x..x | x...x..."
  - id: outro
    title: last call
    duration: 35s
    harmony: "Dm7 G7 | Cmaj7 A7 | Dm7 G7 | Cmaj7 Cmaj7"
    scene: "outro soft"
    variation: "cadence"
    profile:
      density: light
      brightness: warm
    roles:
      alto:
        motif: "3 . 2 . 1 . . . | . . . . . . . ."
      snare:
        pattern: "........ | ..x....."

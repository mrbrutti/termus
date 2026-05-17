title: Dusty Swing / After Hours
description: Small-group late set with a quieter intro, brighter middle, and coda.
listen_mode: album-side
seed: 7319
tags: [jazz, swing, small-group, night]
globals:
  density: steady
  brightness: balanced
  swing: groove
  phrase: long
sections:
  - title: count-in
    algo: jazz
    duration: 90s
    profile:
      density: light
      brightness: warm
      swing: groove
    audit:
      form: "intro:8"
      harmony: "ii7 V7 | Imaj7 VI7"
      lead: "9 . 7 5 | 3 . 2 1"
      comp: "x . . x | . x . x"
      drums: "ride: x.x. x.x. | hh: .x.. .x.."
      arrange: "bass drums comp"
  - title: house theme
    algo: jazz
    duration: 3m
    profile:
      density: steady
      brightness: balanced
      swing: groove
      phrase: long
    audit:
      form: "a:16"
      harmony: "ii7 V7 | Imaj7 VI7"
      lead: "5 . 6 7 | 9 . 7 3"
      comp: "x . x . | . x . x"
      drums: "ride: x.x. x.x. | hh: .x.. .x.. | sn: ...x ...x"
      arrange: "bass drums comp +lead"
  - title: back booth solo
    algo: jazz
    duration: 3m30s
    profile:
      density: busy
      brightness: bright
      swing: heavy
      phrase: long
    audit:
      form: "b:16"
      harmony: "ii7 bII7 | Imaj7 VI7"
      lead: "9 . b9 7 | 5 . 3 1"
      comp: "x . x . | x . . x"
      drums: "ride: x.x. x.x. | hh: .x.. .x.. | fill: .... ...x"
      arrange: "bass drums comp +lead"
  - title: last call
    algo: jazz
    duration: 2m
    profile:
      density: light
      brightness: warm
      swing: groove
    audit:
      form: "cadence:8 outro:8"
      harmony: "ii7 V7 | Imaj7 VI7"
      lead: "3 . 2 1 | 1 . . ."
      comp: "x . . x | . x . ."
      drums: "ride: x.x. x.x. | hh: .x.. .x.."
      arrange: "bass drums comp"


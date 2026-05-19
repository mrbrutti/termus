title: Deep Sea Cathedral (ACE)
description: SP25 ACE-Step ambient drone. Cmin at 60 BPM, beatless three-pad
  layered drone. acestep engine.
style: ambient
substyle: drone
listen_mode: hour-stream
render_engine: acestep
seed: 25040

key: Cmin
tempo: 60
total_duration: 4m
tags: [ambient, drone, beatless, cathedral, ocean, sp25]

acestep:
  style: >
    beatless ambient drone: three sustained pad layers stacked in low, mid,
    and high registers. A distant bell motif rings out every few bars,
    soaked in cathedral-scale reverb. Slowly evolving harmonic motion,
    oceanic depth, no rhythm, no vocals, instrumental. Long, patient,
    immersive.
  tags: [ambient, drone, beatless, pad, bells, cathedral reverb, ocean]
  scale: minor
  time_signature: 4/4
  motif: a single bell tone on the fifth that rings out and slowly decays
  inference_steps: 8
  sections:
    - id: intro
      bars: 16
      description: low pad fades in alone, very slow attack
      harmony: "Cm"
      dynamic: soft
    - id: middle
      bars: 32
      description: mid and high pads join, bell motif begins ringing out
      harmony: "Cm Abmaj7"
      dynamic: building
    - id: bloom
      bars: 24
      description: all three pads at full, bell rings repeat with longer reverb tails
      harmony: "Cm Abmaj7 Ebmaj7 Fm9"
      dynamic: peak
    - id: outro
      bars: 16
      description: high pad fades, mid pad recedes, low drone alone
      harmony: "Cm"
      dynamic: fade

title: Slow Drone Fragments
description: Four sustained drone layers + sparse bell motif — no drums, fully event-authored.
style: ambient
mix_bus: ambient
listen_mode: album-side
seed: 11023
tags: [ambient, drone, modal, slow, pad, bell]
key: Amin
tempo: 60
globals: {density: sparse, brightness: warm, motion: slow, reverb: cathedral}

roles:
  # Low drone: root A, holds through entire section
  drone_low:
    family: pad
    tone: [soft, wide, deep]
    register: low
    prominence: anchor
    events:
      # Each event spans 16 beats (at 60 BPM = 16 seconds)
      - {beat: 1.0, pitch: A1, dur: 16.0, vel: 52}

  # Mid drone: fifth above (E), slightly brighter
  drone_mid:
    family: pad
    tone: [soft, warm]
    register: mid
    prominence: support
    events:
      - {beat: 1.0, pitch: E3, dur: 16.0, vel: 46}

  # High drone: ninth above (B), very soft
  drone_high:
    family: pad
    tone: [soft, airy]
    register: high
    prominence: air
    events:
      - {beat: 1.0, pitch: B4, dur: 16.0, vel: 38}

  # Bell motif: sparse 8 events per pass, drifting melody
  bell_motif:
    family: bells
    tone: [glass, sparkle]
    register: high
    prominence: air
    events:
      - {beat: 1.0,  pitch: A5, dur: 1.5, vel: 62}
      - {beat: 3.0,  pitch: E5, dur: 1.5, vel: 55}
      - {beat: 5.5,  pitch: B5, dur: 1.0, vel: 58}
      - {beat: 7.0,  pitch: G5, dur: 2.0, vel: 50}
      - {beat: 9.5,  pitch: A5, dur: 1.5, vel: 60}
      - {beat: 11.0, pitch: C6, dur: 1.0, vel: 54}
      - {beat: 13.0, pitch: E5, dur: 2.0, vel: 52}
      - {beat: 15.5, pitch: A4, dur: 0.5, vel: 48}

sections:
  - id: open
    title: still water
    duration: 22s
    harmony: "Am9 Am9"
    scene: "intro still"
    variation: "establish"
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.15}
          - {at: 100, value: 0.45}

  - id: drift
    title: motif rising
    duration: 32s
    harmony: "Am9 Fmaj9 | Am9 G6"
    scene: "head drift"
    variation: "statement"
    # Section override: shift the drone_mid up to the seventh (G) for colour
    role_events:
      drone_low:
        - {beat: 1.0, pitch: A1, dur: 8.0, vel: 54}
        - {beat: 9.0, pitch: F1, dur: 8.0, vel: 50}
      drone_mid:
        - {beat: 1.0, pitch: E3, dur: 8.0, vel: 48}
        - {beat: 9.0, pitch: C3, dur: 8.0, vel: 48}
      drone_high:
        - {beat: 1.0,  pitch: B4, dur: 8.0, vel: 40}
        - {beat: 9.0,  pitch: G4, dur: 8.0, vel: 40}
      bell_motif:
        # Sparser, more plaintive line over the drift section
        - {beat: 2.0,  pitch: A5, dur: 2.0, vel: 58}
        - {beat: 5.0,  pitch: G5, dur: 1.5, vel: 52}
        - {beat: 8.0,  pitch: E5, dur: 2.5, vel: 54}
        - {beat: 10.5, pitch: F5, dur: 1.5, vel: 50}
        - {beat: 13.0, pitch: C6, dur: 1.0, vel: 56}
        - {beat: 14.5, pitch: A5, dur: 2.0, vel: 50}
        - {beat: 16.0, pitch: G5, dur: 1.5, vel: 48}
        - {beat: 17.5, pitch: E5, dur: 3.0, vel: 46}
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 0.6}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.45}
          - {at: 50, value: 0.65}
          - {at: 100, value: 0.55}

  - id: close
    title: fog return
    duration: 20s
    harmony: "Am9 Am9"
    scene: "outro still"
    variation: "cadence"
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.55}
          - {at: 100, value: 0.1}

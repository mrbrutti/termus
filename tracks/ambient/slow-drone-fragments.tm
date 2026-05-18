title: Slow Drone Fragments
description: A-minor ambient — layered sustained drones (root, fifth, ninth), sparse bell motif, glassy shimmer, no drums — fully event-authored.
style: ambient
mix_bus: ambient
listen_mode: album-side
seed: 11023
tags: [ambient, drone, modal, slow, pad, bell, shimmer]
key: Amin
tempo: 60
globals: {density: sparse, brightness: warm, motion: slow, reverb: cathedral}

# -----------------------------------------------------------------------------
# 16-beat harmonic loop (auto-loop = 4 bars):
#   bar 1: Am9   (A C E G B)
#   bar 2: Am11  (A C E G D)
#   bar 3: Em9   (E G B D F#)
#   bar 4: Am9
# Each drone event spans 8 beats and overlaps the next by 0.5 beats (so the
# old chord's release smears into the new chord's attack).
# -----------------------------------------------------------------------------

roles:
  # drone_root — low pad on the root (A), holds for 8 beats per change.
  drone_root:
    family: pad
    tone: [soft, wide, deep]
    register: low
    prominence: anchor
    events:
      # A1 holds beats 1..8.5 (covers Am9 and Am11)
      - {beat: 1.0, pitch: A1, dur: 7.5, vel: 56}
      # E1 holds beats 8.5..12.5 (Em9, overlapping handover)
      - {beat: 8.5, pitch: E1, dur: 4.0, vel: 54}
      # A1 returns at 12.5..16.5 (Am9 turnaround)
      - {beat: 12.5, pitch: A1, dur: 4.0, vel: 56}

  # drone_fifth — mid pad on the fifth, slightly brighter.
  drone_fifth:
    family: pad
    tone: [soft, warm]
    register: mid
    prominence: support
    events:
      - {beat: 1.0, pitch: E3, dur: 7.5, vel: 50}
      - {beat: 8.5, pitch: B2, dur: 4.0, vel: 48}
      - {beat: 12.5, pitch: E3, dur: 4.0, vel: 50}

  # drone_ninth — high pad on the ninth, very soft, adds shimmer.
  drone_ninth:
    family: pad
    tone: [soft, airy, glass]
    register: high
    prominence: air
    events:
      # B4 (9 of Am9) holds for the first bar
      - {beat: 1.0, pitch: B4, dur: 3.5, vel: 40}
      # D5 (11 of Am11) bar 2
      - {beat: 4.5, pitch: D5, dur: 3.5, vel: 38}
      # F#5 (9 of Em9) bar 3
      - {beat: 8.5, pitch: F#5, dur: 3.5, vel: 40}
      # B4 returns bar 4
      - {beat: 12.5, pitch: B4, dur: 4.0, vel: 40}

  # bell_motif — sparse high-register hits at varying pitches. 8 events per
  # 16-beat phrase, separated by 1.5–3 beats. All in octave 5–6, long sustain.
  bell_motif:
    family: bells
    tone: [glass, sparkle]
    register: high
    prominence: air
    events:
      - {beat: 1.5,  pitch: A5, dur: 4.0, vel: 62}
      - {beat: 4.0,  pitch: E5, dur: 4.0, vel: 56}
      - {beat: 6.0,  pitch: G5, dur: 4.0, vel: 58}
      - {beat: 8.5,  pitch: B5, dur: 4.0, vel: 60}
      - {beat: 10.5, pitch: D6, dur: 4.0, vel: 54}
      - {beat: 12.0, pitch: F#5, dur: 4.0, vel: 56, art: tenuto}
      - {beat: 14.0, pitch: C6, dur: 4.0, vel: 50}
      - {beat: 15.5, pitch: E5, dur: 4.0, vel: 48}

  # shimmer — sparse texture: glassy upper-octave plucks, even sparser than
  # bells. 3 events per 16-beat phrase.
  shimmer:
    family: bells
    tone: [glass, airy]
    register: high
    prominence: air
    events:
      - {beat: 2.5,  pitch: A6, dur: 3.0, vel: 38}
      - {beat: 9.0,  pitch: E6, dur: 3.0, vel: 36}
      - {beat: 13.5, pitch: A6, dur: 3.5, vel: 34}

sections:
  - id: emerge
    title: still water
    duration: 24s
    harmony: "Am9 Am11 | Em9 Am9"
    scene: "intro emerge"
    variation: "establish"
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.15}
          - {at: 100, value: 0.45}

  - id: drift
    title: motif rising
    duration: 32s
    harmony: "Am9 Am11 | Em9 Am9"
    scene: "head drift"
    variation: "statement"
    # Section override: a parallel "drift" colour. Drones move to F (bVI) and
    # the bells trace a more plaintive line.
    role_events:
      drone_root:
        - {beat: 1.0, pitch: A1, dur: 7.5, vel: 56}
        - {beat: 8.5, pitch: F1, dur: 7.5, vel: 54}
        - {beat: 15.5, pitch: A1, dur: 2.0, vel: 54}
      drone_fifth:
        - {beat: 1.0, pitch: E3, dur: 7.5, vel: 50}
        - {beat: 8.5, pitch: C3, dur: 7.5, vel: 50}
        - {beat: 15.5, pitch: E3, dur: 2.0, vel: 48}
      drone_ninth:
        - {beat: 1.0, pitch: B4, dur: 7.5, vel: 42}
        - {beat: 8.5, pitch: G4, dur: 7.5, vel: 42}
        - {beat: 15.5, pitch: B4, dur: 2.0, vel: 40}
      bell_motif:
        - {beat: 2.0,  pitch: A5, dur: 4.5, vel: 58}
        - {beat: 5.0,  pitch: G5, dur: 4.0, vel: 52}
        - {beat: 8.0,  pitch: E5, dur: 4.5, vel: 56}
        - {beat: 10.5, pitch: F5, dur: 3.5, vel: 50}
        - {beat: 13.0, pitch: C6, dur: 3.5, vel: 56}
        - {beat: 14.5, pitch: A5, dur: 4.0, vel: 50}
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 0.6}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.45}
          - {at: 50, value: 0.65}
          - {at: 100, value: 0.55}

  - id: recede
    title: fog return
    duration: 20s
    harmony: "Am9 Am9 | Em9 Am9"
    scene: "outro still"
    variation: "cadence"
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.55}
          - {at: 100, value: 0.1}

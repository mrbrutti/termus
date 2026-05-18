title: Dusty Swing / After Hours
description: Fast bop in Bb — walking bass, spang-a-lang ride, piano comping, tenor head — fully event-authored.
style: jazz
mix_bus: jazz
listen_mode: album-side
seed: 31440
tags: [jazz, swing, bop, tenor, trio, walking, ride]
key: Bbmaj
tempo: 138
globals: {density: full, brightness: bright, motion: restless, phrase: long}

roles:
  # Walking bass: quarter-note walk over Cm7 G7 | Bbmaj7 G7 (2-bar loop = 8 beats)
  bass:
    family: bass
    tone: [woody, round]
    articulation: walk
    register: low
    prominence: anchor
    events:
      # Bar 1: Cm7 — root, third, fifth, seventh
      - {beat: 1.0, pitch: C2, dur: 0.9, vel: 92, art: tenuto}
      - {beat: 2.0, pitch: Eb2, dur: 0.9, vel: 82}
      - {beat: 3.0, pitch: G2, dur: 0.9, vel: 84}
      - {beat: 4.0, pitch: Bb2, dur: 0.9, vel: 80}
      # Bar 2: F7 — root, chromatic approach to 3rd, fifth, chromatic to root
      - {beat: 5.0, pitch: F2, dur: 0.9, vel: 92, art: tenuto}
      - {beat: 6.0, pitch: A2, dur: 0.9, vel: 82}
      - {beat: 7.0, pitch: C3, dur: 0.9, vel: 84}
      - {beat: 8.0, pitch: Eb3, dur: 0.9, vel: 80}
      # Bar 3: Bbmaj7 — root, walk up
      - {beat: 9.0, pitch: Bb1, dur: 0.9, vel: 90, art: tenuto}
      - {beat: 10.0, pitch: D2, dur: 0.9, vel: 80}
      - {beat: 11.0, pitch: F2, dur: 0.9, vel: 82}
      - {beat: 12.0, pitch: A2, dur: 0.9, vel: 78}
      # Bar 4: G7 — chromatic approach to Cm
      - {beat: 13.0, pitch: G2, dur: 0.9, vel: 90, art: tenuto}
      - {beat: 14.0, pitch: B2, dur: 0.9, vel: 80}
      - {beat: 15.0, pitch: D3, dur: 0.9, vel: 82}
      - {beat: 16.0, pitch: F3, dur: 0.9, vel: 78}

  # Ride: spang-a-lang pattern — quarter + 2 swung 8ths per beat
  # At 138 BPM triplet 8ths land at: beat, beat+0.67, beat+1 (next beat)
  # Spang-a-lang = quarter on beats 1,2,3,4 + swung 8th on the "a" of each
  ride:
    family: drums
    tone: [live, bright]
    articulation: swing
    prominence: support
    events:
      # Beat 1 spang-a-lang
      - {beat: 1.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 1.67, pitch: "", dur: 0.15, vel: 72}
      # Beat 2
      - {beat: 2.00, pitch: "", dur: 0.25, vel: 84}
      - {beat: 2.67, pitch: "", dur: 0.15, vel: 70}
      # Beat 3
      - {beat: 3.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 3.67, pitch: "", dur: 0.15, vel: 72}
      # Beat 4
      - {beat: 4.00, pitch: "", dur: 0.25, vel: 84}
      - {beat: 4.67, pitch: "", dur: 0.15, vel: 70}
      # Bar 2
      - {beat: 5.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 5.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 6.00, pitch: "", dur: 0.25, vel: 84}
      - {beat: 6.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 7.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 7.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 8.00, pitch: "", dur: 0.25, vel: 84}
      - {beat: 8.67, pitch: "", dur: 0.15, vel: 70}
      # Bar 3
      - {beat: 9.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 9.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 10.00, pitch: "", dur: 0.25, vel: 84}
      - {beat: 10.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 11.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 11.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 12.00, pitch: "", dur: 0.25, vel: 84}
      - {beat: 12.67, pitch: "", dur: 0.15, vel: 70}
      # Bar 4
      - {beat: 13.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 13.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 14.00, pitch: "", dur: 0.25, vel: 84}
      - {beat: 14.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 15.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 15.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 16.00, pitch: "", dur: 0.25, vel: 84}
      - {beat: 16.67, pitch: "", dur: 0.15, vel: 70}

  # Kick: sparse jazz comping on 1 and 3, occasional 2+
  kick:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: anchor
    events:
      - {beat: 1.0, pitch: "", dur: 0.25, vel: 80}
      - {beat: 3.0, pitch: "", dur: 0.25, vel: 72}
      - {beat: 5.0, pitch: "", dur: 0.25, vel: 78}
      - {beat: 7.0, pitch: "", dur: 0.25, vel: 70}
      - {beat: 9.0, pitch: "", dur: 0.25, vel: 80}
      - {beat: 11.0, pitch: "", dur: 0.25, vel: 72}
      - {beat: 13.0, pitch: "", dur: 0.25, vel: 78}
      - {beat: 14.67, pitch: "", dur: 0.25, vel: 60}
      - {beat: 15.0, pitch: "", dur: 0.25, vel: 72}

  # Snare: jazz comping — sparse hits on 2 and 4, some ghosting
  snare:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: support
    events:
      - {beat: 2.0, pitch: "", dur: 0.25, vel: 90}
      - {beat: 3.67, pitch: "", dur: 0.25, vel: 42, art: ghost}
      - {beat: 4.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 6.0, pitch: "", dur: 0.25, vel: 90}
      - {beat: 7.67, pitch: "", dur: 0.25, vel: 42, art: ghost}
      - {beat: 8.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 10.0, pitch: "", dur: 0.25, vel: 90}
      - {beat: 11.67, pitch: "", dur: 0.25, vel: 42, art: ghost}
      - {beat: 12.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 14.0, pitch: "", dur: 0.25, vel: 90}
      - {beat: 15.67, pitch: "", dur: 0.25, vel: 42, art: ghost}
      - {beat: 16.0, pitch: "", dur: 0.25, vel: 88}

  # Piano: rootless A-voicings (7-9-3-5 above root) — sparse bop comping
  piano:
    family: acoustic_piano
    tone: [clear, present]
    articulation: comp
    register: mid
    prominence: support
    events:
      # Cm7 A-voicing: Bb3-D4-Eb4-G4 (7-9-b3-5)
      - {beat: 1.50, pitch: Bb3, dur: 0.4, vel: 72}
      - {beat: 1.50, pitch: D4, dur: 0.4, vel: 68}
      - {beat: 1.50, pitch: Eb4, dur: 0.4, vel: 70}
      - {beat: 1.50, pitch: G4, dur: 0.4, vel: 74}
      # F7 A-voicing: Eb4-G4-A4-C5 (b7-9-3-5)
      - {beat: 5.50, pitch: Eb4, dur: 0.4, vel: 70}
      - {beat: 5.50, pitch: G4, dur: 0.4, vel: 66}
      - {beat: 5.50, pitch: A4, dur: 0.4, vel: 68}
      - {beat: 5.50, pitch: C5, dur: 0.4, vel: 72}
      # Bbmaj7 A-voicing: A3-D4-F4-A4 (7-3-5-7)
      - {beat: 9.50, pitch: A3, dur: 0.4, vel: 72}
      - {beat: 9.50, pitch: D4, dur: 0.4, vel: 68}
      - {beat: 9.50, pitch: F4, dur: 0.4, vel: 70}
      - {beat: 9.50, pitch: A4, dur: 0.4, vel: 74}
      # G7 A-voicing: F4-A4-B4-D5 (b7-9-3-5)
      - {beat: 13.50, pitch: F4, dur: 0.4, vel: 70}
      - {beat: 13.50, pitch: A4, dur: 0.4, vel: 66}
      - {beat: 13.50, pitch: B4, dur: 0.4, vel: 68}
      - {beat: 13.50, pitch: D5, dur: 0.4, vel: 72}
      # Second stab on off-beats
      - {beat: 3.67, pitch: Bb3, dur: 0.3, vel: 65}
      - {beat: 3.67, pitch: D4, dur: 0.3, vel: 62}
      - {beat: 3.67, pitch: Eb4, dur: 0.3, vel: 63}
      - {beat: 7.67, pitch: Eb4, dur: 0.3, vel: 63}
      - {beat: 7.67, pitch: G4, dur: 0.3, vel: 60}
      - {beat: 7.67, pitch: A4, dur: 0.3, vel: 62}
      - {beat: 11.67, pitch: A3, dur: 0.3, vel: 65}
      - {beat: 11.67, pitch: D4, dur: 0.3, vel: 62}
      - {beat: 11.67, pitch: F4, dur: 0.3, vel: 63}
      - {beat: 15.67, pitch: F4, dur: 0.3, vel: 63}
      - {beat: 15.67, pitch: A4, dur: 0.3, vel: 60}
      - {beat: 15.67, pitch: B4, dur: 0.3, vel: 62}

  # Tenor sax: head melody in bop style over the changes
  tenor:
    family: reed_lead
    tone: [present, round]
    articulation: lyrical
    register: mid-high
    prominence: lead
    events:
      # Bb major scale melodic ideas over Cm7-F7-Bbmaj7-G7
      - {beat: 1.00, pitch: Bb4, dur: 0.5, vel: 90, art: accent}
      - {beat: 1.67, pitch: D5, dur: 0.33, vel: 82}
      - {beat: 2.00, pitch: Eb5, dur: 0.5, vel: 86}
      - {beat: 2.67, pitch: G5, dur: 0.33, vel: 78}
      - {beat: 3.00, pitch: F5, dur: 0.5, vel: 84}
      - {beat: 3.67, pitch: Eb5, dur: 0.33, vel: 76}
      - {beat: 4.00, pitch: D5, dur: 1.0, vel: 88, art: tenuto}
      - {beat: 5.00, pitch: C5, dur: 0.5, vel: 86, art: accent}
      - {beat: 5.67, pitch: Bb4, dur: 0.33, vel: 78}
      - {beat: 6.00, pitch: A4, dur: 0.5, vel: 82}
      - {beat: 6.67, pitch: G4, dur: 0.33, vel: 74}
      - {beat: 7.00, pitch: F4, dur: 0.5, vel: 80}
      - {beat: 7.67, pitch: G4, dur: 0.33, vel: 76}
      - {beat: 8.00, pitch: Bb4, dur: 1.0, vel: 84, art: tenuto}
      - {beat: 9.00, pitch: D5, dur: 0.5, vel: 88, art: accent}
      - {beat: 9.67, pitch: F5, dur: 0.33, vel: 80}
      - {beat: 10.00, pitch: A5, dur: 0.5, vel: 86}
      - {beat: 10.67, pitch: G5, dur: 0.33, vel: 78}
      - {beat: 11.00, pitch: F5, dur: 0.5, vel: 82}
      - {beat: 11.67, pitch: Eb5, dur: 0.33, vel: 74}
      - {beat: 12.00, pitch: D5, dur: 1.0, vel: 86, art: tenuto}
      - {beat: 13.00, pitch: B4, dur: 0.5, vel: 84, art: accent}
      - {beat: 13.67, pitch: D5, dur: 0.33, vel: 76}
      - {beat: 14.00, pitch: F5, dur: 0.5, vel: 80}
      - {beat: 14.67, pitch: D5, dur: 0.33, vel: 74}
      - {beat: 15.00, pitch: Eb5, dur: 0.5, vel: 82}
      - {beat: 15.67, pitch: C5, dur: 0.33, vel: 76}
      - {beat: 16.00, pitch: Bb4, dur: 1.5, vel: 86, art: tenuto}

sections:
  - id: head-a1
    title: first chorus
    duration: 16s
    harmony: "Cm7 F7 | Bbmaj7 G7 | Cm7 F7 | Bbmaj7 Bbmaj7"
    scene: "head relaxed"
    variation: "statement"
    groove: swing_56
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.6}
          - {at: 100, value: 0.85}

  - id: head-a2
    title: second a
    duration: 16s
    harmony: "Cm7 F7 | Bbmaj7 G7 | Dm7 G7 | Cm7 F7"
    scene: "head glide"
    variation: "sequence-up"
    groove: swing_56
    # Section override: give piano some upper-register stabs in bar 3 (Dm7-G7)
    role_events:
      piano:
        # Dm7 A-voicing: C4-F4-G4-A4 (b7-11-5-13)
        - {beat: 1.50, pitch: C4, dur: 0.4, vel: 74}
        - {beat: 1.50, pitch: F4, dur: 0.4, vel: 70}
        - {beat: 1.50, pitch: A4, dur: 0.4, vel: 72}
        - {beat: 3.67, pitch: C4, dur: 0.3, vel: 66}
        - {beat: 3.67, pitch: F4, dur: 0.3, vel: 63}
        # F7 A-voicing
        - {beat: 5.50, pitch: Eb4, dur: 0.4, vel: 72}
        - {beat: 5.50, pitch: G4, dur: 0.4, vel: 68}
        - {beat: 5.50, pitch: A4, dur: 0.4, vel: 70}
        - {beat: 5.50, pitch: C5, dur: 0.4, vel: 74}
        - {beat: 7.67, pitch: Eb4, dur: 0.3, vel: 65}
        - {beat: 7.67, pitch: A4, dur: 0.3, vel: 62}
        # Bbmaj7 A-voicing
        - {beat: 9.50, pitch: A3, dur: 0.4, vel: 72}
        - {beat: 9.50, pitch: D4, dur: 0.4, vel: 68}
        - {beat: 9.50, pitch: F4, dur: 0.4, vel: 70}
        - {beat: 11.67, pitch: A3, dur: 0.3, vel: 66}
        - {beat: 11.67, pitch: D4, dur: 0.3, vel: 63}
        # G7 A-voicing
        - {beat: 13.50, pitch: F4, dur: 0.4, vel: 70}
        - {beat: 13.50, pitch: A4, dur: 0.4, vel: 66}
        - {beat: 13.50, pitch: B4, dur: 0.4, vel: 68}
        - {beat: 15.67, pitch: F4, dur: 0.3, vel: 64}
        - {beat: 15.67, pitch: B4, dur: 0.3, vel: 61}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.8}
          - {at: 100, value: 0.9}

  - id: head-b
    title: bridge climb
    duration: 16s
    harmony: "Ebmaj7 D7 | Dm7 G7 | Cm7 F7 | Bbmaj7 G7"
    scene: "bridge lift"
    variation: "open-register"
    groove: swing_56
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 0.7}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.85}
          - {at: 60, value: 1.0}
          - {at: 100, value: 0.9}

  - id: outro
    title: bar stools empty
    duration: 16s
    harmony: "Cm7 F7 | Bbmaj7 G7 | Cm7 F7 | Bbmaj7 Bbmaj7"
    scene: "outro cadence"
    variation: "cadence"
    groove: swing_56
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 0.7}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.85}
          - {at: 100, value: 0.3}

title: Sunday Afternoon Drive
description: Half-time chill with sustained pad, sparse Rhodes, root-fifth bass, and 8th-note hats — fully event-authored.
style: chill
mix_bus: chill
listen_mode: album-side
seed: 19334
tags: [chill, pop, pad, afternoon, drive, halftime]
key: Amin
tempo: 100
globals: {density: full, brightness: balanced, motion: moving, reverb: halo}

roles:
  # Pad: whole-note chord washes — 4 beats per chord, big and airy
  pad:
    family: pad
    tone: [soft, wide]
    register: mid
    prominence: air
    events:
      # Am9: A-C-E-B (root-m3-5-9)
      - {beat: 1.0, pitch: A3, dur: 3.9, vel: 62}
      - {beat: 1.0, pitch: C4, dur: 3.9, vel: 58}
      - {beat: 1.0, pitch: E4, dur: 3.9, vel: 60}
      - {beat: 1.0, pitch: B4, dur: 3.9, vel: 56}
      # Fmaj9: F-A-C-G (root-3-5-9)
      - {beat: 5.0, pitch: F3, dur: 3.9, vel: 62}
      - {beat: 5.0, pitch: A3, dur: 3.9, vel: 58}
      - {beat: 5.0, pitch: C4, dur: 3.9, vel: 60}
      - {beat: 5.0, pitch: G4, dur: 3.9, vel: 56}
      # Cmaj9: C-E-G-D (root-3-5-9)
      - {beat: 9.0, pitch: C3, dur: 3.9, vel: 62}
      - {beat: 9.0, pitch: E3, dur: 3.9, vel: 58}
      - {beat: 9.0, pitch: G3, dur: 3.9, vel: 60}
      - {beat: 9.0, pitch: D4, dur: 3.9, vel: 56}
      # G7: G-B-D-F (root-3-5-b7)
      - {beat: 13.0, pitch: G3, dur: 3.9, vel: 62}
      - {beat: 13.0, pitch: B3, dur: 3.9, vel: 58}
      - {beat: 13.0, pitch: D4, dur: 3.9, vel: 60}
      - {beat: 13.0, pitch: F4, dur: 3.9, vel: 56}

  # Rhodes: sparse comping on beats 2 and 4, some off-beat fills
  keys:
    family: electric_piano
    tone: [warm, soft]
    register: mid
    prominence: support
    events:
      # Am9 comping
      - {beat: 2.0, pitch: A3, dur: 0.45, vel: 72}
      - {beat: 2.0, pitch: C4, dur: 0.45, vel: 68}
      - {beat: 2.0, pitch: E4, dur: 0.45, vel: 70}
      - {beat: 2.5, pitch: G4, dur: 0.25, vel: 60}
      - {beat: 4.0, pitch: A3, dur: 0.45, vel: 70}
      - {beat: 4.0, pitch: E4, dur: 0.45, vel: 66}
      # Fmaj9 comping
      - {beat: 6.0, pitch: F3, dur: 0.45, vel: 72}
      - {beat: 6.0, pitch: A3, dur: 0.45, vel: 68}
      - {beat: 6.0, pitch: C4, dur: 0.45, vel: 70}
      - {beat: 6.5, pitch: G4, dur: 0.25, vel: 60}
      - {beat: 8.0, pitch: F3, dur: 0.45, vel: 70}
      - {beat: 8.0, pitch: C4, dur: 0.45, vel: 66}
      # Cmaj9 comping
      - {beat: 10.0, pitch: C3, dur: 0.45, vel: 72}
      - {beat: 10.0, pitch: E3, dur: 0.45, vel: 68}
      - {beat: 10.0, pitch: G3, dur: 0.45, vel: 70}
      - {beat: 10.5, pitch: D4, dur: 0.25, vel: 60}
      - {beat: 12.0, pitch: C3, dur: 0.45, vel: 70}
      - {beat: 12.0, pitch: G3, dur: 0.45, vel: 66}
      # G7 comping
      - {beat: 14.0, pitch: G3, dur: 0.45, vel: 72}
      - {beat: 14.0, pitch: B3, dur: 0.45, vel: 68}
      - {beat: 14.0, pitch: D4, dur: 0.45, vel: 70}
      - {beat: 14.5, pitch: F4, dur: 0.25, vel: 60}
      - {beat: 16.0, pitch: G3, dur: 0.45, vel: 70}
      - {beat: 16.0, pitch: D4, dur: 0.45, vel: 66}

  # Bass: root + fifth motion (half-time feel)
  bass:
    family: synth_bass
    tone: [round, soft]
    register: low
    prominence: anchor
    events:
      # Am: root + fifth
      - {beat: 1.0, pitch: A1, dur: 0.9, vel: 90, art: tenuto}
      - {beat: 2.0, pitch: E2, dur: 0.9, vel: 78}
      - {beat: 3.0, pitch: A1, dur: 0.9, vel: 82}
      - {beat: 4.0, pitch: G2, dur: 0.9, vel: 74}
      # Fmaj: root + fifth
      - {beat: 5.0, pitch: F1, dur: 0.9, vel: 88, art: tenuto}
      - {beat: 6.0, pitch: C2, dur: 0.9, vel: 76}
      - {beat: 7.0, pitch: F1, dur: 0.9, vel: 80}
      - {beat: 8.0, pitch: E2, dur: 0.9, vel: 72}
      # Cmaj: root + fifth
      - {beat: 9.0, pitch: C2, dur: 0.9, vel: 88, art: tenuto}
      - {beat: 10.0, pitch: G2, dur: 0.9, vel: 76}
      - {beat: 11.0, pitch: C2, dur: 0.9, vel: 80}
      - {beat: 12.0, pitch: B2, dur: 0.9, vel: 72}
      # G7: root + fifth
      - {beat: 13.0, pitch: G1, dur: 0.9, vel: 88, art: tenuto}
      - {beat: 14.0, pitch: D2, dur: 0.9, vel: 76}
      - {beat: 15.0, pitch: G1, dur: 0.9, vel: 80}
      - {beat: 16.0, pitch: F2, dur: 0.9, vel: 72}

  # Kick: half-time pattern — 1 and 3 of each bar
  kick:
    family: drums
    tone: [soft, deep]
    prominence: anchor
    events:
      - {beat: 1.0, pitch: "", dur: 0.25, vel: 100}
      - {beat: 3.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 5.0, pitch: "", dur: 0.25, vel: 100}
      - {beat: 7.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 9.0, pitch: "", dur: 0.25, vel: 100}
      - {beat: 11.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 13.0, pitch: "", dur: 0.25, vel: 100}
      - {beat: 15.0, pitch: "", dur: 0.25, vel: 88}

  # Clap/snare: backbeat on 3 (half-time feel)
  snare:
    family: drums
    tone: [soft]
    prominence: support
    events:
      - {beat: 3.0, pitch: "", dur: 0.25, vel: 96}
      - {beat: 7.0, pitch: "", dur: 0.25, vel: 96}
      - {beat: 11.0, pitch: "", dur: 0.25, vel: 94}
      - {beat: 15.0, pitch: "", dur: 0.25, vel: 96}
      # Optional ghost: slightly before each snare
      - {beat: 2.75, pitch: "", dur: 0.1, vel: 38, art: ghost}
      - {beat: 6.75, pitch: "", dur: 0.1, vel: 38, art: ghost}
      - {beat: 10.75, pitch: "", dur: 0.1, vel: 38, art: ghost}
      - {beat: 14.75, pitch: "", dur: 0.1, vel: 38, art: ghost}

  # Hat: steady 8ths
  hat:
    family: drums
    tone: [dry, tight]
    prominence: support
    events:
      - {beat: 1.0, pitch: "", dur: 0.1, vel: 74}
      - {beat: 1.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 2.0, pitch: "", dur: 0.1, vel: 70}
      - {beat: 2.5, pitch: "", dur: 0.1, vel: 55}
      - {beat: 3.0, pitch: "", dur: 0.1, vel: 74, art: accent}
      - {beat: 3.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 4.0, pitch: "", dur: 0.1, vel: 70}
      - {beat: 4.5, pitch: "", dur: 0.1, vel: 55}
      - {beat: 5.0, pitch: "", dur: 0.1, vel: 74}
      - {beat: 5.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 6.0, pitch: "", dur: 0.1, vel: 70}
      - {beat: 6.5, pitch: "", dur: 0.1, vel: 55}
      - {beat: 7.0, pitch: "", dur: 0.1, vel: 74, art: accent}
      - {beat: 7.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 8.0, pitch: "", dur: 0.1, vel: 70}
      - {beat: 8.5, pitch: "", dur: 0.1, vel: 55}
      - {beat: 9.0, pitch: "", dur: 0.1, vel: 74}
      - {beat: 9.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 10.0, pitch: "", dur: 0.1, vel: 70}
      - {beat: 10.5, pitch: "", dur: 0.1, vel: 55}
      - {beat: 11.0, pitch: "", dur: 0.1, vel: 74, art: accent}
      - {beat: 11.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 12.0, pitch: "", dur: 0.1, vel: 70}
      - {beat: 12.5, pitch: "", dur: 0.1, vel: 55}
      - {beat: 13.0, pitch: "", dur: 0.1, vel: 74}
      - {beat: 13.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 14.0, pitch: "", dur: 0.1, vel: 70}
      - {beat: 14.5, pitch: "", dur: 0.1, vel: 55}
      - {beat: 15.0, pitch: "", dur: 0.1, vel: 74, art: accent}
      - {beat: 15.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 16.0, pitch: "", dur: 0.1, vel: 70}
      - {beat: 16.5, pitch: "", dur: 0.1, vel: 55}

sections:
  - id: intro
    title: open road
    duration: 14s
    harmony: "Am9 Fmaj9 | Cmaj9 G7"
    scene: "intro hush"
    variation: "establish"
    groove: straight
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.2}
          - {at: 100, value: 0.6}

  - id: verse
    title: window down
    duration: 34s
    harmony: "Am9 Fmaj9 | Cmaj9 Gsus4"
    scene: "head glide"
    variation: "statement"
    groove: straight
    # Section override: verse kicks add a syncopated extra kick on beat 2.5
    role_events:
      kick:
        - {beat: 1.0, pitch: "", dur: 0.25, vel: 102}
        - {beat: 2.5, pitch: "", dur: 0.25, vel: 80}
        - {beat: 3.0, pitch: "", dur: 0.25, vel: 90}
        - {beat: 4.5, pitch: "", dur: 0.25, vel: 76}
        - {beat: 5.0, pitch: "", dur: 0.25, vel: 102}
        - {beat: 6.5, pitch: "", dur: 0.25, vel: 80}
        - {beat: 7.0, pitch: "", dur: 0.25, vel: 90}
        - {beat: 8.5, pitch: "", dur: 0.25, vel: 76}
        - {beat: 9.0, pitch: "", dur: 0.25, vel: 102}
        - {beat: 10.5, pitch: "", dur: 0.25, vel: 80}
        - {beat: 11.0, pitch: "", dur: 0.25, vel: 90}
        - {beat: 12.5, pitch: "", dur: 0.25, vel: 76}
        - {beat: 13.0, pitch: "", dur: 0.25, vel: 102}
        - {beat: 14.5, pitch: "", dur: 0.25, vel: 80}
        - {beat: 15.0, pitch: "", dur: 0.25, vel: 90}
        - {beat: 16.5, pitch: "", dur: 0.25, vel: 76}
    substitutions:
      - {rule: ii_V_chain, apply_to: I, probability: 0.7}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.6}
          - {at: 50, value: 0.78}
          - {at: 100, value: 0.68}

  - id: outro
    title: home at dusk
    duration: 16s
    harmony: "Am9 Fmaj9 | Am6"
    scene: "outro hush"
    variation: "cadence"
    groove: straight
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 1.0}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.75}
          - {at: 100, value: 0.25}

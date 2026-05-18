title: Bookstore After Rain
description: Felt-piano ballad with Dilla-feel kit and walking sub — fully event-authored.
style: lofi
substyle: piano-ballad
listen_mode: album-side
seed: 28011
tags: [lofi, piano, rain, dilla]
key: Dmin
tempo: 86
mix_bus: lofi
globals: {density: full, brightness: warm, motion: gentle, reverb: warm}

roles:
  keys:
    family: piano
    tone: [warm, soft]
    register: mid
    prominence: lead
    personality: piano_felt
    room: bedroom_small
    reverb_send_db: -14
    events:
      # Bar 1 (Dm9): right-hand chord stabs + melody. 4 beats.
      - {beat: 1.00, pitch: D3, dur: 0.5, vel: 78, art: tenuto}
      - {beat: 1.00, pitch: F3, dur: 0.5, vel: 70}
      - {beat: 1.00, pitch: A3, dur: 0.5, vel: 72}
      - {beat: 1.00, pitch: E4, dur: 0.5, vel: 80}
      - {beat: 2.50, pitch: D3, dur: 0.25, vel: 60}
      - {beat: 2.50, pitch: F3, dur: 0.25, vel: 56}
      - {beat: 2.50, pitch: A3, dur: 0.25, vel: 58}
      - {beat: 3.00, pitch: F4, dur: 0.5, vel: 84, art: accent}
      - {beat: 3.50, pitch: E4, dur: 0.25, vel: 68}
      - {beat: 4.00, pitch: D4, dur: 0.75, vel: 72}
      # Bar 2 (Gm7)
      - {beat: 5.00, pitch: G3, dur: 0.5, vel: 78, art: tenuto}
      - {beat: 5.00, pitch: Bb3, dur: 0.5, vel: 70}
      - {beat: 5.00, pitch: D4, dur: 0.5, vel: 72}
      - {beat: 5.00, pitch: F4, dur: 0.5, vel: 80}
      - {beat: 6.50, pitch: G3, dur: 0.25, vel: 60}
      - {beat: 6.50, pitch: Bb3, dur: 0.25, vel: 56}
      - {beat: 7.00, pitch: G4, dur: 0.5, vel: 84, art: accent}
      - {beat: 7.50, pitch: F4, dur: 0.25, vel: 68}
      - {beat: 8.00, pitch: D4, dur: 1.0, vel: 70}
      # Bar 3 (Cmaj9)
      - {beat: 9.00, pitch: C3, dur: 0.5, vel: 76, art: tenuto}
      - {beat: 9.00, pitch: E3, dur: 0.5, vel: 68}
      - {beat: 9.00, pitch: G3, dur: 0.5, vel: 70}
      - {beat: 9.00, pitch: D4, dur: 0.5, vel: 78}
      - {beat: 10.50, pitch: E4, dur: 0.25, vel: 62}
      - {beat: 11.00, pitch: D4, dur: 0.5, vel: 80, art: accent}
      - {beat: 11.50, pitch: C4, dur: 0.25, vel: 65}
      - {beat: 12.00, pitch: B3, dur: 1.0, vel: 70}
      # Bar 4 (Bbmaj7)
      - {beat: 13.00, pitch: Bb2, dur: 0.5, vel: 76, art: tenuto}
      - {beat: 13.00, pitch: D3, dur: 0.5, vel: 68}
      - {beat: 13.00, pitch: F3, dur: 0.5, vel: 70}
      - {beat: 13.00, pitch: A3, dur: 0.5, vel: 78}
      - {beat: 14.50, pitch: C4, dur: 0.25, vel: 62}
      - {beat: 15.00, pitch: D4, dur: 0.5, vel: 82, art: accent}
      - {beat: 15.50, pitch: C4, dur: 0.25, vel: 66}
      - {beat: 16.00, pitch: A3, dur: 1.0, vel: 68}

  bass:
    family: synth_bass
    tone: [round, sub]
    register: low
    prominence: anchor
    events:
      # Walking-ish quarters over 4 bars: Dm9 Gm7 Cmaj9 Bbmaj7
      - {beat: 1.0, pitch: D2, dur: 0.9, vel: 92, art: tenuto}
      - {beat: 2.0, pitch: A2, dur: 0.9, vel: 78}
      - {beat: 3.0, pitch: C3, dur: 0.9, vel: 80}
      - {beat: 4.0, pitch: A2, dur: 0.9, vel: 76}
      - {beat: 5.0, pitch: G2, dur: 0.9, vel: 90, art: tenuto}
      - {beat: 6.0, pitch: D3, dur: 0.9, vel: 78}
      - {beat: 7.0, pitch: F3, dur: 0.9, vel: 82}
      - {beat: 8.0, pitch: D3, dur: 0.9, vel: 76}
      - {beat: 9.0, pitch: C2, dur: 0.9, vel: 88, art: tenuto}
      - {beat: 10.0, pitch: G2, dur: 0.9, vel: 76}
      - {beat: 11.0, pitch: B2, dur: 0.9, vel: 78}
      - {beat: 12.0, pitch: G2, dur: 0.9, vel: 74}
      - {beat: 13.0, pitch: Bb1, dur: 0.9, vel: 88, art: tenuto}
      - {beat: 14.0, pitch: F2, dur: 0.9, vel: 76}
      - {beat: 15.0, pitch: A2, dur: 0.9, vel: 80}
      - {beat: 16.0, pitch: F2, dur: 0.9, vel: 74}

  kick:
    family: drums
    tone: [dusty]
    prominence: anchor
    events:
      # Dilla-ish: late-feel beats over 4 bars
      - {beat: 1.00, pitch: "", dur: 0.25, vel: 112}
      - {beat: 1.75, pitch: "", dur: 0.25, vel: 96}
      - {beat: 2.75, pitch: "", dur: 0.25, vel: 102}
      - {beat: 4.50, pitch: "", dur: 0.25, vel: 88}
      - {beat: 5.00, pitch: "", dur: 0.25, vel: 112}
      - {beat: 5.75, pitch: "", dur: 0.25, vel: 96}
      - {beat: 6.75, pitch: "", dur: 0.25, vel: 100}
      - {beat: 8.50, pitch: "", dur: 0.25, vel: 90}
      - {beat: 9.00, pitch: "", dur: 0.25, vel: 110}
      - {beat: 9.75, pitch: "", dur: 0.25, vel: 94}
      - {beat: 10.75, pitch: "", dur: 0.25, vel: 100}
      - {beat: 12.50, pitch: "", dur: 0.25, vel: 86}
      - {beat: 13.00, pitch: "", dur: 0.25, vel: 112}
      - {beat: 13.75, pitch: "", dur: 0.25, vel: 96}
      - {beat: 14.75, pitch: "", dur: 0.25, vel: 102}
      - {beat: 16.50, pitch: "", dur: 0.25, vel: 88}

  snare:
    family: drums
    tone: [dusty]
    prominence: support
    events:
      # Backbeat + ghost snares over 4 bars
      - {beat: 2.00, pitch: "", dur: 0.25, vel: 102}
      - {beat: 2.75, pitch: "", dur: 0.25, vel: 44, art: ghost}
      - {beat: 3.50, pitch: "", dur: 0.25, vel: 40, art: ghost}
      - {beat: 4.00, pitch: "", dur: 0.25, vel: 100}
      - {beat: 6.00, pitch: "", dur: 0.25, vel: 102}
      - {beat: 6.50, pitch: "", dur: 0.25, vel: 44, art: ghost}
      - {beat: 7.50, pitch: "", dur: 0.25, vel: 40, art: ghost}
      - {beat: 8.00, pitch: "", dur: 0.25, vel: 100}
      - {beat: 10.00, pitch: "", dur: 0.25, vel: 100}
      - {beat: 10.75, pitch: "", dur: 0.25, vel: 44, art: ghost}
      - {beat: 11.50, pitch: "", dur: 0.25, vel: 40, art: ghost}
      - {beat: 12.00, pitch: "", dur: 0.25, vel: 98}
      - {beat: 14.00, pitch: "", dur: 0.25, vel: 102}
      - {beat: 14.75, pitch: "", dur: 0.25, vel: 44, art: ghost}
      - {beat: 15.50, pitch: "", dur: 0.25, vel: 40, art: ghost}
      - {beat: 16.00, pitch: "", dur: 0.25, vel: 100}

  hat:
    family: drums
    tone: [dry, dusty]
    prominence: support
    events:
      # 8th-note hats with shuffle bias over 4 bars (bar 1)
      - {beat: 1.0, pitch: "", dur: 0.1, vel: 76}
      - {beat: 1.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 2.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 2.5, pitch: "", dur: 0.1, vel: 54}
      - {beat: 3.0, pitch: "", dur: 0.1, vel: 76, art: accent}
      - {beat: 3.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 4.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 4.5, pitch: "", dur: 0.1, vel: 54}
      # bar 2
      - {beat: 5.0, pitch: "", dur: 0.1, vel: 76}
      - {beat: 5.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 6.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 6.5, pitch: "", dur: 0.1, vel: 54}
      - {beat: 7.0, pitch: "", dur: 0.1, vel: 76, art: accent}
      - {beat: 7.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 8.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 8.5, pitch: "", dur: 0.1, vel: 54}
      # bar 3
      - {beat: 9.0, pitch: "", dur: 0.1, vel: 76}
      - {beat: 9.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 10.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 10.5, pitch: "", dur: 0.1, vel: 54}
      - {beat: 11.0, pitch: "", dur: 0.1, vel: 76, art: accent}
      - {beat: 11.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 12.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 12.5, pitch: "", dur: 0.1, vel: 54}
      # bar 4
      - {beat: 13.0, pitch: "", dur: 0.1, vel: 76}
      - {beat: 13.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 14.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 14.5, pitch: "", dur: 0.1, vel: 54}
      - {beat: 15.0, pitch: "", dur: 0.1, vel: 76, art: accent}
      - {beat: 15.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 16.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 16.5, pitch: "", dur: 0.1, vel: 54}

sections:
  - id: intro
    title: rain-on-glass
    duration: 12s
    harmony: "Dm9 Gm7"
    scene: "intro hush"
    variation: "establish"
    groove: dilla_late
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.25}
          - {at: 100, value: 0.55}

  - id: verse
    title: paperback turn
    duration: 32s
    harmony: "Dm9 Gm7 | Cmaj9 Bbmaj7"
    scene: "head glide"
    variation: "statement"
    groove: dilla_late
    # Section-level role_events override: verse verse gives bass a busier line
    role_events:
      bass:
        - {beat: 1.0, pitch: D2, dur: 0.45, vel: 94, art: tenuto}
        - {beat: 1.5, pitch: F2, dur: 0.45, vel: 78}
        - {beat: 2.0, pitch: A2, dur: 0.45, vel: 80}
        - {beat: 2.5, pitch: C3, dur: 0.45, vel: 75}
        - {beat: 3.0, pitch: A2, dur: 0.45, vel: 82}
        - {beat: 3.5, pitch: F2, dur: 0.45, vel: 74}
        - {beat: 4.0, pitch: A2, dur: 0.9, vel: 76}
        - {beat: 5.0, pitch: G2, dur: 0.45, vel: 92, art: tenuto}
        - {beat: 5.5, pitch: Bb2, dur: 0.45, vel: 78}
        - {beat: 6.0, pitch: D3, dur: 0.45, vel: 80}
        - {beat: 6.5, pitch: F3, dur: 0.45, vel: 75}
        - {beat: 7.0, pitch: D3, dur: 0.45, vel: 82}
        - {beat: 7.5, pitch: Bb2, dur: 0.45, vel: 74}
        - {beat: 8.0, pitch: D3, dur: 0.9, vel: 76}
        - {beat: 9.0, pitch: C2, dur: 0.45, vel: 90, art: tenuto}
        - {beat: 9.5, pitch: E2, dur: 0.45, vel: 76}
        - {beat: 10.0, pitch: G2, dur: 0.45, vel: 78}
        - {beat: 10.5, pitch: B2, dur: 0.45, vel: 73}
        - {beat: 11.0, pitch: G2, dur: 0.45, vel: 80}
        - {beat: 11.5, pitch: E2, dur: 0.45, vel: 72}
        - {beat: 12.0, pitch: G2, dur: 0.9, vel: 74}
        - {beat: 13.0, pitch: Bb1, dur: 0.45, vel: 90, art: tenuto}
        - {beat: 13.5, pitch: D2, dur: 0.45, vel: 76}
        - {beat: 14.0, pitch: F2, dur: 0.45, vel: 78}
        - {beat: 14.5, pitch: A2, dur: 0.45, vel: 73}
        - {beat: 15.0, pitch: F2, dur: 0.45, vel: 80}
        - {beat: 15.5, pitch: D2, dur: 0.45, vel: 72}
        - {beat: 16.0, pitch: F2, dur: 0.9, vel: 74}
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.55}
          - {at: 50, value: 0.85}
          - {at: 100, value: 0.65}

  - id: outro
    title: shelf-closing
    duration: 20s
    harmony: "Dm9 A7 | Dm6"
    scene: "outro hush"
    variation: "cadence"
    groove: dilla_late
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 1.0}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.7}
          - {at: 100, value: 0.25}

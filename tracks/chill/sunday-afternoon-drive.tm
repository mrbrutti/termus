title: Sunday Afternoon Drive
description: C-major chill — sustained pad, syncopated Rhodes stabs, root-and-octave bass, half-time kick + clap, 16th hat, 8th shaker — fully event-authored.
style: chill
substyle: half-time-chill
mix_bus: chill
listen_mode: album-side
seed: 19334
tags: [chill, pop, pad, afternoon, drive, halftime]
key: Cmaj
tempo: 100
globals: {density: full, brightness: balanced, motion: moving, reverb: halo}

# -----------------------------------------------------------------------------
# 4-bar harmonic loop:
#   bar 1: Cmaj7 | G/B  (bar split into two 2-beat halves)
#   bar 2: Am7   | F
#   bar 3: Dm7   | G7
#   bar 4: Cmaj7 | Am7
# -----------------------------------------------------------------------------

roles:
  # Pad: sustained 2-beat washes (8th-note granularity on chord changes).
  # Each event covers one chord; consecutive events overlap by half-beat so
  # transitions smear briefly into the next chord.
  pad:
    family: pad
    tone: [soft, wide]
    register: mid
    prominence: air
    events:
      # bar 1 first half — Cmaj7 (C E G B): 2 beats, slight overhang into G/B
      - {beat: 1.0, pitch: C4, dur: 2.4, vel: 60}
      - {beat: 1.0, pitch: E4, dur: 2.4, vel: 56}
      - {beat: 1.0, pitch: G4, dur: 2.4, vel: 58}
      - {beat: 1.0, pitch: B4, dur: 2.4, vel: 52}
      # bar 1 second half — G/B (B D G B): 2 beats, overhangs into Am7
      - {beat: 3.0, pitch: B3, dur: 2.4, vel: 60}
      - {beat: 3.0, pitch: D4, dur: 2.4, vel: 56}
      - {beat: 3.0, pitch: G4, dur: 2.4, vel: 58}
      - {beat: 3.0, pitch: B4, dur: 2.4, vel: 52}
      # bar 2 first half — Am7 (A C E G)
      - {beat: 5.0, pitch: A3, dur: 2.4, vel: 60}
      - {beat: 5.0, pitch: C4, dur: 2.4, vel: 56}
      - {beat: 5.0, pitch: E4, dur: 2.4, vel: 58}
      - {beat: 5.0, pitch: G4, dur: 2.4, vel: 54}
      # bar 2 second half — Fmaj7 (F A C E)
      - {beat: 7.0, pitch: F3, dur: 2.4, vel: 60}
      - {beat: 7.0, pitch: A3, dur: 2.4, vel: 56}
      - {beat: 7.0, pitch: C4, dur: 2.4, vel: 58}
      - {beat: 7.0, pitch: E4, dur: 2.4, vel: 54}
      # bar 3 first half — Dm7 (D F A C)
      - {beat: 9.0, pitch: D4, dur: 2.4, vel: 60}
      - {beat: 9.0, pitch: F4, dur: 2.4, vel: 56}
      - {beat: 9.0, pitch: A4, dur: 2.4, vel: 58}
      - {beat: 9.0, pitch: C5, dur: 2.4, vel: 52}
      # bar 3 second half — G7 (G B D F)
      - {beat: 11.0, pitch: G3, dur: 2.4, vel: 60}
      - {beat: 11.0, pitch: B3, dur: 2.4, vel: 56}
      - {beat: 11.0, pitch: D4, dur: 2.4, vel: 58}
      - {beat: 11.0, pitch: F4, dur: 2.4, vel: 52}
      # bar 4 first half — Cmaj7
      - {beat: 13.0, pitch: C4, dur: 2.4, vel: 60}
      - {beat: 13.0, pitch: E4, dur: 2.4, vel: 56}
      - {beat: 13.0, pitch: G4, dur: 2.4, vel: 58}
      - {beat: 13.0, pitch: B4, dur: 2.4, vel: 52}
      # bar 4 second half — Am7 (turnaround)
      - {beat: 15.0, pitch: A3, dur: 2.4, vel: 58}
      - {beat: 15.0, pitch: C4, dur: 2.4, vel: 54}
      - {beat: 15.0, pitch: E4, dur: 2.4, vel: 56}
      - {beat: 15.0, pitch: G4, dur: 2.4, vel: 52}

  # Rhodes comp: syncopated chord stabs on the off-beats. 10 stabs per 4-bar
  # phrase — varies which chord tones appear in each stab.
  rhodes_comp:
    family: electric_piano
    tone: [warm, soft]
    register: mid
    prominence: support
    events:
      # bar 1 Cmaj7 → G/B stab on "and" of 2
      - {beat: 2.50, pitch: E4, dur: 0.4, vel: 72}
      - {beat: 2.50, pitch: G4, dur: 0.4, vel: 70}
      - {beat: 2.50, pitch: B4, dur: 0.4, vel: 74}
      # stab on 4.5
      - {beat: 4.50, pitch: D4, dur: 0.35, vel: 68}
      - {beat: 4.50, pitch: G4, dur: 0.35, vel: 64}
      # bar 2 Am7 → F
      - {beat: 6.50, pitch: C4, dur: 0.4, vel: 72}
      - {beat: 6.50, pitch: E4, dur: 0.4, vel: 68}
      - {beat: 6.50, pitch: A4, dur: 0.4, vel: 70}
      - {beat: 8.50, pitch: A3, dur: 0.35, vel: 66}
      - {beat: 8.50, pitch: C4, dur: 0.35, vel: 64}
      - {beat: 8.50, pitch: F4, dur: 0.35, vel: 68}
      # bar 3 Dm7 → G7
      - {beat: 10.50, pitch: F4, dur: 0.4, vel: 72}
      - {beat: 10.50, pitch: A4, dur: 0.4, vel: 68}
      - {beat: 10.50, pitch: C5, dur: 0.4, vel: 72}
      - {beat: 12.50, pitch: F4, dur: 0.35, vel: 70}
      - {beat: 12.50, pitch: B4, dur: 0.35, vel: 68}
      # bar 4 Cmaj7 → Am7 — resolution
      - {beat: 13.75, pitch: E4, dur: 0.3, vel: 64}
      - {beat: 13.75, pitch: G4, dur: 0.3, vel: 60}
      - {beat: 14.50, pitch: B3, dur: 0.4, vel: 72}
      - {beat: 14.50, pitch: D4, dur: 0.4, vel: 68}
      - {beat: 14.50, pitch: E4, dur: 0.4, vel: 70}
      - {beat: 16.50, pitch: A3, dur: 0.35, vel: 66}
      - {beat: 16.50, pitch: E4, dur: 0.35, vel: 62}

  # Bass: root motion with octave jumps; anticipates chord changes by a
  # half-beat on bars 2 and 4 ("anticipate" technique).
  bass:
    family: synth_bass
    tone: [round, soft]
    register: low
    prominence: anchor
    events:
      # bar 1 Cmaj7 G/B
      - {beat: 1.0, pitch: C2, dur: 1.9, vel: 92, art: tenuto}
      - {beat: 2.5, pitch: G2, dur: 0.45, vel: 76}
      - {beat: 3.0, pitch: B1, dur: 1.9, vel: 84, art: tenuto}
      # bar 2 Am7 F — anticipation on 4.5
      - {beat: 4.5, pitch: A1, dur: 0.5, vel: 78}
      - {beat: 5.0, pitch: A2, dur: 1.5, vel: 88, art: tenuto}
      - {beat: 6.5, pitch: E2, dur: 0.45, vel: 74}
      - {beat: 7.0, pitch: F1, dur: 1.9, vel: 86, art: tenuto}
      # bar 3 Dm7 G7
      - {beat: 9.0, pitch: D2, dur: 1.9, vel: 88, art: tenuto}
      - {beat: 10.5, pitch: A2, dur: 0.45, vel: 74}
      - {beat: 11.0, pitch: G1, dur: 1.9, vel: 90, art: tenuto}
      # bar 4 Cmaj7 Am7 (anticipation back to top)
      - {beat: 12.5, pitch: G2, dur: 0.45, vel: 72}
      - {beat: 13.0, pitch: C2, dur: 1.9, vel: 88, art: tenuto}
      - {beat: 14.5, pitch: G2, dur: 0.45, vel: 72}
      - {beat: 15.0, pitch: A1, dur: 1.5, vel: 84, art: tenuto}
      - {beat: 16.5, pitch: E2, dur: 0.45, vel: 70}

  # Kick: half-time pattern — kick on bar-1-beat-1, bar-2-beat-3.5, bar-3-beat-1,
  # bar-4-beat-3.5. This is the classic "boom .. tap" feel that opens up space.
  kick:
    family: drums
    tone: [round, soft]
    prominence: anchor
    events:
      - {beat: 1.0,  pitch: "", dur: 0.25, vel: 104}
      - {beat: 5.0,  pitch: "", dur: 0.25, vel: 96}
      - {beat: 6.5,  pitch: "", dur: 0.25, vel: 80}
      - {beat: 9.0,  pitch: "", dur: 0.25, vel: 102}
      - {beat: 12.5, pitch: "", dur: 0.25, vel: 84}
      - {beat: 13.0, pitch: "", dur: 0.25, vel: 100}
      - {beat: 15.0, pitch: "", dur: 0.25, vel: 92}
      # bar-4 fill on 4.5
      - {beat: 16.5, pitch: "", dur: 0.25, vel: 78}

  # Snare/clap: hit on beat 3 of bars 1, 2, 4 (the half-time snare). Bar 3
  # gets a slight rim variation.
  snare:
    family: drums
    tone: [tight, clap]
    prominence: support
    events:
      - {beat: 3.0,  pitch: "", dur: 0.3, vel: 96}
      - {beat: 7.0,  pitch: "", dur: 0.3, vel: 92}
      - {beat: 11.0, pitch: "", dur: 0.3, vel: 88, art: ghost}
      - {beat: 11.5, pitch: "", dur: 0.3, vel: 70, art: ghost}
      - {beat: 15.0, pitch: "", dur: 0.3, vel: 96}

  # Closed hat: 16th-note pattern, accent on every 4th hit. Velocity dips
  # on the 16ths between beats for a clean "tickatickaticka" feel.
  hat_closed:
    family: drums
    tone: [dry, tight]
    prominence: support
    events:
      # bar 1 — 16 hits (16ths)
      - {beat: 1.00, pitch: "", dur: 0.05, vel: 78, art: accent}
      - {beat: 1.25, pitch: "", dur: 0.05, vel: 58}
      - {beat: 1.50, pitch: "", dur: 0.05, vel: 62}
      - {beat: 1.75, pitch: "", dur: 0.05, vel: 58}
      - {beat: 2.00, pitch: "", dur: 0.05, vel: 70}
      - {beat: 2.25, pitch: "", dur: 0.05, vel: 58}
      - {beat: 2.50, pitch: "", dur: 0.05, vel: 62}
      - {beat: 2.75, pitch: "", dur: 0.05, vel: 58}
      - {beat: 3.00, pitch: "", dur: 0.05, vel: 78, art: accent}
      - {beat: 3.25, pitch: "", dur: 0.05, vel: 58}
      - {beat: 3.50, pitch: "", dur: 0.05, vel: 62}
      - {beat: 3.75, pitch: "", dur: 0.05, vel: 58}
      - {beat: 4.00, pitch: "", dur: 0.05, vel: 70}
      - {beat: 4.25, pitch: "", dur: 0.05, vel: 58}
      - {beat: 4.50, pitch: "", dur: 0.05, vel: 62}
      - {beat: 4.75, pitch: "", dur: 0.05, vel: 58}
      # bar 2
      - {beat: 5.00, pitch: "", dur: 0.05, vel: 78, art: accent}
      - {beat: 5.25, pitch: "", dur: 0.05, vel: 58}
      - {beat: 5.50, pitch: "", dur: 0.05, vel: 62}
      - {beat: 5.75, pitch: "", dur: 0.05, vel: 58}
      - {beat: 6.00, pitch: "", dur: 0.05, vel: 70}
      - {beat: 6.25, pitch: "", dur: 0.05, vel: 58}
      - {beat: 6.50, pitch: "", dur: 0.05, vel: 62}
      - {beat: 6.75, pitch: "", dur: 0.05, vel: 58}
      - {beat: 7.00, pitch: "", dur: 0.05, vel: 78, art: accent}
      - {beat: 7.25, pitch: "", dur: 0.05, vel: 58}
      - {beat: 7.50, pitch: "", dur: 0.05, vel: 62}
      - {beat: 7.75, pitch: "", dur: 0.05, vel: 58}
      - {beat: 8.00, pitch: "", dur: 0.05, vel: 70}
      - {beat: 8.25, pitch: "", dur: 0.05, vel: 58}
      - {beat: 8.50, pitch: "", dur: 0.05, vel: 62}
      - {beat: 8.75, pitch: "", dur: 0.05, vel: 58}
      # bar 3
      - {beat: 9.00, pitch: "", dur: 0.05, vel: 78, art: accent}
      - {beat: 9.25, pitch: "", dur: 0.05, vel: 58}
      - {beat: 9.50, pitch: "", dur: 0.05, vel: 62}
      - {beat: 9.75, pitch: "", dur: 0.05, vel: 58}
      - {beat: 10.00, pitch: "", dur: 0.05, vel: 70}
      - {beat: 10.25, pitch: "", dur: 0.05, vel: 58}
      - {beat: 10.50, pitch: "", dur: 0.05, vel: 62}
      - {beat: 10.75, pitch: "", dur: 0.05, vel: 58}
      - {beat: 11.00, pitch: "", dur: 0.05, vel: 78, art: accent}
      - {beat: 11.25, pitch: "", dur: 0.05, vel: 58}
      - {beat: 11.50, pitch: "", dur: 0.05, vel: 62}
      - {beat: 11.75, pitch: "", dur: 0.05, vel: 58}
      - {beat: 12.00, pitch: "", dur: 0.05, vel: 70}
      - {beat: 12.25, pitch: "", dur: 0.05, vel: 58}
      - {beat: 12.50, pitch: "", dur: 0.05, vel: 62}
      - {beat: 12.75, pitch: "", dur: 0.05, vel: 58}
      # bar 4 — additional fill: louder accents on 4.25, 4.5, 4.75
      - {beat: 13.00, pitch: "", dur: 0.05, vel: 78, art: accent}
      - {beat: 13.25, pitch: "", dur: 0.05, vel: 58}
      - {beat: 13.50, pitch: "", dur: 0.05, vel: 62}
      - {beat: 13.75, pitch: "", dur: 0.05, vel: 58}
      - {beat: 14.00, pitch: "", dur: 0.05, vel: 70}
      - {beat: 14.25, pitch: "", dur: 0.05, vel: 58}
      - {beat: 14.50, pitch: "", dur: 0.05, vel: 62}
      - {beat: 14.75, pitch: "", dur: 0.05, vel: 58}
      - {beat: 15.00, pitch: "", dur: 0.05, vel: 80, art: accent}
      - {beat: 15.25, pitch: "", dur: 0.05, vel: 64}
      - {beat: 15.50, pitch: "", dur: 0.05, vel: 70}
      - {beat: 15.75, pitch: "", dur: 0.05, vel: 64}
      - {beat: 16.00, pitch: "", dur: 0.05, vel: 78}
      - {beat: 16.25, pitch: "", dur: 0.05, vel: 70}
      - {beat: 16.50, pitch: "", dur: 0.05, vel: 76}
      - {beat: 16.75, pitch: "", dur: 0.05, vel: 70}

  # Shaker: 8th note pulse, very soft, creates a continuous foundation.
  shaker:
    family: drums
    tone: [soft, dry]
    prominence: air
    events:
      - {beat: 1.0,  pitch: "", dur: 0.1, vel: 56}
      - {beat: 1.5,  pitch: "", dur: 0.1, vel: 48}
      - {beat: 2.0,  pitch: "", dur: 0.1, vel: 52}
      - {beat: 2.5,  pitch: "", dur: 0.1, vel: 48}
      - {beat: 3.0,  pitch: "", dur: 0.1, vel: 56}
      - {beat: 3.5,  pitch: "", dur: 0.1, vel: 48}
      - {beat: 4.0,  pitch: "", dur: 0.1, vel: 52}
      - {beat: 4.5,  pitch: "", dur: 0.1, vel: 48}
      - {beat: 5.0,  pitch: "", dur: 0.1, vel: 56}
      - {beat: 5.5,  pitch: "", dur: 0.1, vel: 48}
      - {beat: 6.0,  pitch: "", dur: 0.1, vel: 52}
      - {beat: 6.5,  pitch: "", dur: 0.1, vel: 48}
      - {beat: 7.0,  pitch: "", dur: 0.1, vel: 56}
      - {beat: 7.5,  pitch: "", dur: 0.1, vel: 48}
      - {beat: 8.0,  pitch: "", dur: 0.1, vel: 52}
      - {beat: 8.5,  pitch: "", dur: 0.1, vel: 48}
      - {beat: 9.0,  pitch: "", dur: 0.1, vel: 56}
      - {beat: 9.5,  pitch: "", dur: 0.1, vel: 48}
      - {beat: 10.0, pitch: "", dur: 0.1, vel: 52}
      - {beat: 10.5, pitch: "", dur: 0.1, vel: 48}
      - {beat: 11.0, pitch: "", dur: 0.1, vel: 56}
      - {beat: 11.5, pitch: "", dur: 0.1, vel: 48}
      - {beat: 12.0, pitch: "", dur: 0.1, vel: 52}
      - {beat: 12.5, pitch: "", dur: 0.1, vel: 48}
      - {beat: 13.0, pitch: "", dur: 0.1, vel: 56}
      - {beat: 13.5, pitch: "", dur: 0.1, vel: 48}
      - {beat: 14.0, pitch: "", dur: 0.1, vel: 52}
      - {beat: 14.5, pitch: "", dur: 0.1, vel: 48}
      - {beat: 15.0, pitch: "", dur: 0.1, vel: 56}
      - {beat: 15.5, pitch: "", dur: 0.1, vel: 48}
      - {beat: 16.0, pitch: "", dur: 0.1, vel: 52}
      - {beat: 16.5, pitch: "", dur: 0.1, vel: 48}

sections:
  - id: intro
    title: window-open
    duration: 16s
    harmony: "Cmaj7 G/B | Am7 F | Dm7 G7 | Cmaj7 Am7"
    scene: "intro establish"
    variation: "establish"
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.3}
          - {at: 100, value: 0.6}

  - id: a-section
    title: lane glide
    duration: 32s
    harmony: "Cmaj7 G/B | Am7 F | Dm7 G7 | Cmaj7 Am7"
    scene: "head glide"
    variation: "statement"
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.6}
          - {at: 50, value: 0.85}
          - {at: 100, value: 0.7}

  - id: b-section
    title: bridge lift
    duration: 16s
    harmony: "Fmaj7 Em7 | Dm7 G7 | Cmaj7 Am7 | Dm7 G7"
    scene: "bridge lift"
    variation: "sequence-up"
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.7}
          - {at: 60, value: 0.9}
          - {at: 100, value: 0.75}

  - id: a-out
    title: cruise out
    duration: 32s
    harmony: "Cmaj7 G/B | Am7 F | Dm7 G7 | Cmaj7 Cmaj7"
    scene: "outro cadence"
    variation: "cadence"
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.85}
          - {at: 100, value: 0.35}

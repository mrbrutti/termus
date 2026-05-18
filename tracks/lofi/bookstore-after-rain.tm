title: Bookstore After Rain
description: Felt-piano lofi in Dmin — voice-led rhodes stabs, walking bass with chromatic approach, Dilla kit with ghost snares, sub octave below — fully event-authored.
style: lofi
substyle: piano-ballad
listen_mode: album-side
seed: 28011
tags: [lofi, piano, rain, dilla]
key: Dmin
tempo: 86
mix_bus: lofi
globals: {density: full, brightness: warm, motion: gentle, reverb: warm}

# -----------------------------------------------------------------------------
# Harmonic plan (verse, 4-bar loop, repeats four times per 16-bar verse):
#   bar 1: Dm9  |  bar 2: Gm7  |  bar 3: Bb6  |  bar 4: A7
# -----------------------------------------------------------------------------

roles:
  # Rhodes felt-piano: voice-led chord stabs that vary which 3-4 chord tones
  # appear on each stab. 16 stabs per 4-bar loop.
  rhodes:
    family: piano
    tone: [warm, soft]
    register: mid
    prominence: lead
    personality: piano_felt
    room: bedroom_small
    reverb_send_db: -14
    events:
      # ---- Bar 1: Dm9 (D F A C E) -----------------------------------------
      # Stab 1 on beat 1: full voicing (D-F-A-C)
      - {beat: 1.00, pitch: D3, dur: 0.6, vel: 78, art: tenuto}
      - {beat: 1.00, pitch: F3, dur: 0.6, vel: 70}
      - {beat: 1.00, pitch: A3, dur: 0.6, vel: 72}
      - {beat: 1.00, pitch: C4, dur: 0.6, vel: 80}
      # Stab 2 on "and" of 2 — partial voicing (F-A-E top)
      - {beat: 2.50, pitch: F3, dur: 0.3, vel: 60}
      - {beat: 2.50, pitch: A3, dur: 0.3, vel: 58}
      - {beat: 2.50, pitch: E4, dur: 0.3, vel: 65}
      # Stab 3 on beat 3 — accent, upper voices (F-A-D)
      - {beat: 3.00, pitch: A3, dur: 0.4, vel: 82, art: accent}
      - {beat: 3.00, pitch: C4, dur: 0.4, vel: 76}
      - {beat: 3.00, pitch: E4, dur: 0.4, vel: 80}
      # Stab 4 on "and" of 4 — 9-rooted upper (E-C-A)
      - {beat: 4.50, pitch: A3, dur: 0.3, vel: 64}
      - {beat: 4.50, pitch: C4, dur: 0.3, vel: 60}
      - {beat: 4.50, pitch: E4, dur: 0.3, vel: 68}

      # ---- Bar 2: Gm7 (G Bb D F) ------------------------------------------
      # Smooth voice-led from Dm9: keep F, drop D→D, A→Bb, C→C? Use G+Bb+D+F.
      - {beat: 5.00, pitch: G3, dur: 0.6, vel: 78, art: tenuto}
      - {beat: 5.00, pitch: Bb3, dur: 0.6, vel: 70}
      - {beat: 5.00, pitch: D4, dur: 0.6, vel: 72}
      - {beat: 5.00, pitch: F4, dur: 0.6, vel: 78}
      # Off-beat 6.5: drop2 inversion (Bb-D-F)
      - {beat: 6.50, pitch: Bb3, dur: 0.3, vel: 60}
      - {beat: 6.50, pitch: D4, dur: 0.3, vel: 58}
      - {beat: 6.50, pitch: F4, dur: 0.3, vel: 65}
      # Beat 7 accent (D-F-A upper structure)
      - {beat: 7.00, pitch: D4, dur: 0.4, vel: 82, art: accent}
      - {beat: 7.00, pitch: F4, dur: 0.4, vel: 76}
      - {beat: 7.00, pitch: A4, dur: 0.4, vel: 78}
      # Beat 8.5 anticipation of next chord
      - {beat: 8.50, pitch: Bb3, dur: 0.3, vel: 62}
      - {beat: 8.50, pitch: D4, dur: 0.3, vel: 60}

      # ---- Bar 3: Bb6 (Bb D F G) ------------------------------------------
      - {beat: 9.00, pitch: Bb2, dur: 0.6, vel: 76, art: tenuto}
      - {beat: 9.00, pitch: D3, dur: 0.6, vel: 68}
      - {beat: 9.00, pitch: F3, dur: 0.6, vel: 70}
      - {beat: 9.00, pitch: G3, dur: 0.6, vel: 78}
      # 9.5 ghost upper texture
      - {beat: 9.50, pitch: F4, dur: 0.2, vel: 56}
      - {beat: 9.50, pitch: G4, dur: 0.2, vel: 60}
      # Beat 11 accent — open 6-9 upper structure
      - {beat: 11.00, pitch: G3, dur: 0.4, vel: 80, art: accent}
      - {beat: 11.00, pitch: D4, dur: 0.4, vel: 74}
      - {beat: 11.00, pitch: F4, dur: 0.4, vel: 78}
      # Beat 12.5 walkdown texture
      - {beat: 12.50, pitch: D4, dur: 0.3, vel: 62}
      - {beat: 12.50, pitch: F4, dur: 0.3, vel: 58}

      # ---- Bar 4: A7 (A C# E G — dominant resolving to Dm) ----------------
      - {beat: 13.00, pitch: A2, dur: 0.6, vel: 80, art: tenuto}
      - {beat: 13.00, pitch: C#3, dur: 0.6, vel: 72}
      - {beat: 13.00, pitch: E3, dur: 0.6, vel: 74}
      - {beat: 13.00, pitch: G3, dur: 0.6, vel: 82}
      # 14.5 b9 alt-tension stab
      - {beat: 14.50, pitch: C#4, dur: 0.3, vel: 64}
      - {beat: 14.50, pitch: G4, dur: 0.3, vel: 70}
      # Beat 15 accent — V7 punch
      - {beat: 15.00, pitch: E3, dur: 0.4, vel: 84, art: accent}
      - {beat: 15.00, pitch: G3, dur: 0.4, vel: 80}
      - {beat: 15.00, pitch: C#4, dur: 0.4, vel: 82}
      # Beat 16.5 anticipation back to Dm9 (chromatic from C# to D)
      - {beat: 16.50, pitch: A3, dur: 0.3, vel: 64}
      - {beat: 16.50, pitch: C#4, dur: 0.3, vel: 66}

  # Bass: quarter-note walking with octave jumps + chromatic approach.
  # 4-bar loop matches the rhodes harmony.
  bass:
    family: synth_bass
    tone: [round, sub]
    register: low
    prominence: anchor
    events:
      # Bar 1 Dm9: D – A – C – A (root, fifth, b7, fifth)
      - {beat: 1.0, pitch: D2, dur: 0.9, vel: 92, art: tenuto}
      - {beat: 2.0, pitch: A2, dur: 0.9, vel: 78}
      - {beat: 3.0, pitch: C3, dur: 0.9, vel: 80}
      - {beat: 4.0, pitch: A2, dur: 0.9, vel: 76}
      # Bar 2 Gm7: G – D – F – D (with octave jump)
      - {beat: 5.0, pitch: G2, dur: 0.9, vel: 90, art: tenuto}
      - {beat: 6.0, pitch: D3, dur: 0.9, vel: 78}
      - {beat: 7.0, pitch: F2, dur: 0.9, vel: 82}
      - {beat: 8.0, pitch: D2, dur: 0.9, vel: 76}
      # Bar 3 Bb6: Bb – F – D – F (octave drop on 1, then 5–3–5)
      - {beat: 9.0, pitch: Bb1, dur: 0.9, vel: 88, art: tenuto}
      - {beat: 10.0, pitch: F2, dur: 0.9, vel: 76}
      - {beat: 11.0, pitch: D2, dur: 0.9, vel: 78}
      - {beat: 12.0, pitch: F2, dur: 0.9, vel: 74}
      # Bar 4 A7: A – C# – E – G (chromatic walk, A C#=major3, E=fifth, G=b7 leads to next D)
      - {beat: 13.0, pitch: A1, dur: 0.9, vel: 90, art: tenuto}
      - {beat: 14.0, pitch: C#2, dur: 0.9, vel: 78}
      - {beat: 15.0, pitch: E2, dur: 0.9, vel: 80}
      - {beat: 16.0, pitch: G2, dur: 0.9, vel: 78}

  # Sub: octave below the bass on bar 1 of each chord — adds weight without
  # cluttering the mid-low.
  sub:
    family: synth_bass
    tone: [sub, soft]
    register: low
    prominence: anchor
    events:
      - {beat: 1.0, pitch: D1, dur: 3.5, vel: 76, art: tenuto}
      - {beat: 5.0, pitch: G1, dur: 3.5, vel: 74, art: tenuto}
      - {beat: 9.0, pitch: Bb0, dur: 3.5, vel: 72, art: tenuto}
      - {beat: 13.0, pitch: A1, dur: 3.5, vel: 74, art: tenuto}

  # Kick: Dilla-late hip-hop kit. Beats 1, "and" of 1, 3, "and" of 3 with
  # bar-4 fill variation (additional kick on 4.5 in the A7 bar for tension).
  kick:
    family: drums
    tone: [dusty]
    prominence: anchor
    events:
      # Bar 1 — classic Dilla pattern
      - {beat: 1.00, pitch: "", dur: 0.25, vel: 112}
      - {beat: 1.75, pitch: "", dur: 0.25, vel: 96}
      - {beat: 3.00, pitch: "", dur: 0.25, vel: 104}
      - {beat: 3.75, pitch: "", dur: 0.25, vel: 88}
      # Bar 2 — same feel, slightly different placement
      - {beat: 5.00, pitch: "", dur: 0.25, vel: 110}
      - {beat: 5.75, pitch: "", dur: 0.25, vel: 96}
      - {beat: 7.00, pitch: "", dur: 0.25, vel: 100}
      # Bar 3 — variation: skip 1.75 spot, add 2.5
      - {beat: 9.00, pitch: "", dur: 0.25, vel: 112}
      - {beat: 10.50, pitch: "", dur: 0.25, vel: 88}
      - {beat: 11.00, pitch: "", dur: 0.25, vel: 102}
      - {beat: 11.75, pitch: "", dur: 0.25, vel: 92}
      # Bar 4 — FILL: extra kicks for the V7 turnaround
      - {beat: 13.00, pitch: "", dur: 0.25, vel: 112}
      - {beat: 13.75, pitch: "", dur: 0.25, vel: 98}
      - {beat: 14.50, pitch: "", dur: 0.25, vel: 90}
      - {beat: 15.00, pitch: "", dur: 0.25, vel: 104}
      - {beat: 15.75, pitch: "", dur: 0.25, vel: 96}

  # Snare: backbeat on 2 and 4 with ghost-snare flourishes between.
  snare:
    family: drums
    tone: [dusty]
    prominence: support
    events:
      # Bar 1
      - {beat: 2.00, pitch: "", dur: 0.25, vel: 102}
      - {beat: 2.75, pitch: "", dur: 0.25, vel: 44, art: ghost}
      - {beat: 3.50, pitch: "", dur: 0.25, vel: 40, art: ghost}
      - {beat: 4.00, pitch: "", dur: 0.25, vel: 100}
      # Bar 2
      - {beat: 6.00, pitch: "", dur: 0.25, vel: 102}
      - {beat: 6.50, pitch: "", dur: 0.25, vel: 44, art: ghost}
      - {beat: 7.50, pitch: "", dur: 0.25, vel: 40, art: ghost}
      - {beat: 8.00, pitch: "", dur: 0.25, vel: 100}
      # Bar 3 — bar variation: extra ghost on 2.5
      - {beat: 10.00, pitch: "", dur: 0.25, vel: 100}
      - {beat: 10.50, pitch: "", dur: 0.25, vel: 46, art: ghost}
      - {beat: 10.75, pitch: "", dur: 0.25, vel: 44, art: ghost}
      - {beat: 11.50, pitch: "", dur: 0.25, vel: 40, art: ghost}
      - {beat: 12.00, pitch: "", dur: 0.25, vel: 98}
      # Bar 4 — turnaround flam: snare on 4.5 as bar fill
      - {beat: 14.00, pitch: "", dur: 0.25, vel: 102}
      - {beat: 14.50, pitch: "", dur: 0.25, vel: 50, art: ghost}
      - {beat: 14.75, pitch: "", dur: 0.25, vel: 44, art: ghost}
      - {beat: 15.50, pitch: "", dur: 0.25, vel: 60, art: ghost}
      - {beat: 16.00, pitch: "", dur: 0.25, vel: 100}

  # Closed hat: 8ths with alternating accent on "and" of 2 and "and" of 4.
  # Velocity pattern emulates the "tssk-tssk-TSSK-tssk" lofi hat groove.
  hat_closed:
    family: drums
    tone: [dry, dusty]
    prominence: support
    events:
      # Bar 1 — 8ths
      - {beat: 1.0, pitch: "", dur: 0.1, vel: 76}
      - {beat: 1.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 2.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 2.5, pitch: "", dur: 0.1, vel: 78, art: accent}
      - {beat: 3.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 3.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 4.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 4.5, pitch: "", dur: 0.1, vel: 78, art: accent}
      # Bar 2
      - {beat: 5.0, pitch: "", dur: 0.1, vel: 76}
      - {beat: 5.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 6.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 6.5, pitch: "", dur: 0.1, vel: 78, art: accent}
      - {beat: 7.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 7.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 8.0, pitch: "", dur: 0.1, vel: 72}
      # Bar 3
      - {beat: 9.0, pitch: "", dur: 0.1, vel: 76}
      - {beat: 9.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 10.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 10.5, pitch: "", dur: 0.1, vel: 78, art: accent}
      - {beat: 11.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 11.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 12.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 12.5, pitch: "", dur: 0.1, vel: 78, art: accent}
      # Bar 4
      - {beat: 13.0, pitch: "", dur: 0.1, vel: 76}
      - {beat: 13.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 14.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 14.5, pitch: "", dur: 0.1, vel: 78, art: accent}
      - {beat: 15.0, pitch: "", dur: 0.1, vel: 72}
      - {beat: 15.5, pitch: "", dur: 0.1, vel: 56}
      - {beat: 16.0, pitch: "", dur: 0.1, vel: 72}

  # Open hat: a single hit on the "and" of 4 every OTHER bar (bars 2 and 4 of
  # the 4-bar loop). Adds shoulder-bounce without crowding.
  hat_open:
    family: drums
    tone: [dry, washy]
    prominence: support
    events:
      - {beat: 8.50, pitch: "", dur: 0.5, vel: 70}
      - {beat: 16.50, pitch: "", dur: 0.5, vel: 72, art: accent}

sections:
  - id: intro
    title: rain-on-glass
    duration: 12s
    harmony: "Dm9 Gm7 | Bb6 A7"
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
    harmony: "Dm9 Gm7 | Bb6 A7"
    scene: "head glide"
    variation: "statement"
    groove: dilla_late
    automation:
      - param: cutoff
        breakpoints:
          - {at: 0, value: 0.55}
          - {at: 50, value: 0.85}
          - {at: 100, value: 0.65}

  - id: bridge
    title: light through curtain
    duration: 18s
    harmony: "Bb6 A7 | Dm9 Gm7"
    scene: "bridge lift"
    variation: "open-register"
    groove: dilla_late
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.7}
          - {at: 100, value: 0.85}

  - id: outro
    title: shelf-closing
    duration: 20s
    harmony: "Dm9 Gm7 | Bb6 A7"
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

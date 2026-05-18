title: Dusty Swing / After Hours
description: F-major bop blues — rootless piano voicings, walking bass with chromatic approach, spang-a-lang ride, tenor head — fully event-authored.
style: jazz
mix_bus: jazz
listen_mode: album-side
seed: 31440
tags: [jazz, swing, bop, tenor, trio, walking, ride, blues]
key: Fmaj
tempo: 138
globals: {density: full, brightness: bright, motion: restless, phrase: long}

# -----------------------------------------------------------------------------
# 12-bar bop blues in F (auto-loop = 12 bars = 48 beats). The events below
# walk through the full 12-bar form once; the engine repeats them across the
# section's full duration.
#
#   bar 1: F7   bar 2: Bb7    bar 3: F7      bar 4: Cm7/F7
#   bar 5: Bb7  bar 6: Bdim7  bar 7: F7/D7   bar 8: Gm7/C7
#   bar 9: Bb7  bar 10: F7    bar 11: D7     bar 12: G7/C7  (turnaround)
# -----------------------------------------------------------------------------

roles:
  # Walking upright bass: quarter-note pulse with chromatic approach.
  # Never repeats a pitch within a bar; uses scale walks + chromatic leading
  # tones at chord changes.
  bass:
    family: bass
    tone: [woody, round]
    articulation: walk
    register: low
    prominence: anchor
    loop_bars: 12
    events:
      # bar 1 F7: F – A – C – Eb (root 3 5 b7) — Eb chromatic-approach to D in bar2
      - {beat: 1.0, pitch: F2, dur: 0.9, vel: 92, art: tenuto}
      - {beat: 2.0, pitch: A2, dur: 0.9, vel: 82}
      - {beat: 3.0, pitch: C3, dur: 0.9, vel: 84}
      - {beat: 4.0, pitch: Eb3, dur: 0.9, vel: 80}
      # bar 2 Bb7: Bb – D – F – Ab (root 3 5 b7)
      - {beat: 5.0, pitch: Bb2, dur: 0.9, vel: 92, art: tenuto}
      - {beat: 6.0, pitch: D3, dur: 0.9, vel: 82}
      - {beat: 7.0, pitch: F3, dur: 0.9, vel: 84}
      - {beat: 8.0, pitch: Ab3, dur: 0.9, vel: 80}
      # bar 3 F7: F – C – A – E (root 5 3 7)
      - {beat: 9.0, pitch: F2, dur: 0.9, vel: 90, art: tenuto}
      - {beat: 10.0, pitch: C3, dur: 0.9, vel: 80}
      - {beat: 11.0, pitch: A2, dur: 0.9, vel: 82}
      - {beat: 12.0, pitch: E3, dur: 0.9, vel: 78}
      # bar 4 Cm7 → F7 (ii-V): C – Eb – G – A (chromatic to Bb)
      - {beat: 13.0, pitch: C3, dur: 0.9, vel: 90, art: tenuto}
      - {beat: 14.0, pitch: Eb3, dur: 0.9, vel: 80}
      - {beat: 15.0, pitch: G2, dur: 0.9, vel: 82}
      - {beat: 16.0, pitch: A2, dur: 0.9, vel: 78}
      # bar 5 Bb7: Bb – D – F – Db (root 3 5 b3, chromatic down)
      - {beat: 17.0, pitch: Bb2, dur: 0.9, vel: 92, art: tenuto}
      - {beat: 18.0, pitch: D3, dur: 0.9, vel: 82}
      - {beat: 19.0, pitch: F3, dur: 0.9, vel: 84}
      - {beat: 20.0, pitch: Db3, dur: 0.9, vel: 78}
      # bar 6 Bdim7: B – D – F – Ab (root b3 b5 bb7)
      - {beat: 21.0, pitch: B2, dur: 0.9, vel: 88, art: tenuto}
      - {beat: 22.0, pitch: D3, dur: 0.9, vel: 80}
      - {beat: 23.0, pitch: F3, dur: 0.9, vel: 82}
      - {beat: 24.0, pitch: Ab3, dur: 0.9, vel: 78}
      # bar 7 F7 → D7 (turnaround start): F – A – C – D
      - {beat: 25.0, pitch: F2, dur: 0.9, vel: 90, art: tenuto}
      - {beat: 26.0, pitch: A2, dur: 0.9, vel: 80}
      - {beat: 27.0, pitch: C3, dur: 0.9, vel: 82}
      - {beat: 28.0, pitch: D3, dur: 0.9, vel: 78}
      # bar 8 Gm7 → C7 (ii-V): G – Bb – C – E (C7 root, then 3)
      - {beat: 29.0, pitch: G2, dur: 0.9, vel: 92, art: tenuto}
      - {beat: 30.0, pitch: Bb2, dur: 0.9, vel: 82}
      - {beat: 31.0, pitch: C3, dur: 0.9, vel: 84}
      - {beat: 32.0, pitch: E3, dur: 0.9, vel: 80}
      # bar 9 Bb7: Bb – F – D – Ab
      - {beat: 33.0, pitch: Bb2, dur: 0.9, vel: 90, art: tenuto}
      - {beat: 34.0, pitch: F3, dur: 0.9, vel: 80}
      - {beat: 35.0, pitch: D3, dur: 0.9, vel: 82}
      - {beat: 36.0, pitch: Ab2, dur: 0.9, vel: 78}
      # bar 10 F7: F – A – C – Eb
      - {beat: 37.0, pitch: F2, dur: 0.9, vel: 90, art: tenuto}
      - {beat: 38.0, pitch: A2, dur: 0.9, vel: 80}
      - {beat: 39.0, pitch: C3, dur: 0.9, vel: 82}
      - {beat: 40.0, pitch: Eb3, dur: 0.9, vel: 78}
      # bar 11 D7 (VI7): D – F# – A – C
      - {beat: 41.0, pitch: D3, dur: 0.9, vel: 92, art: tenuto}
      - {beat: 42.0, pitch: F#3, dur: 0.9, vel: 82}
      - {beat: 43.0, pitch: A2, dur: 0.9, vel: 82}
      - {beat: 44.0, pitch: C3, dur: 0.9, vel: 78}
      # bar 12 G7 → C7 turnaround: G – Bb – C – Db (chromatic back to F1)
      - {beat: 45.0, pitch: G2, dur: 0.9, vel: 90, art: tenuto}
      - {beat: 46.0, pitch: Bb2, dur: 0.9, vel: 80}
      - {beat: 47.0, pitch: C3, dur: 0.9, vel: 82}
      - {beat: 48.0, pitch: Db3, dur: 0.9, vel: 78}

  # Ride cymbal — spang-a-lang. Quarter note on the beat + swung-eighth on the
  # "a" of each beat. 8 hits per bar × 12 bars = 96 hits.
  ride:
    family: drums
    tone: [live, bright]
    articulation: swing
    prominence: support
    loop_bars: 12
    events:
      # Generate 12 bars of ride pattern. Each bar = 4 quarters + 4 swung 8ths.
      # bar 1
      - {beat: 1.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 1.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 2.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 2.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 3.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 3.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 4.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 4.67, pitch: "", dur: 0.15, vel: 70}
      # bar 2
      - {beat: 5.00, pitch: "", dur: 0.25, vel: 85}
      - {beat: 5.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 6.00, pitch: "", dur: 0.25, vel: 78}
      - {beat: 6.67, pitch: "", dur: 0.15, vel: 68}
      - {beat: 7.00, pitch: "", dur: 0.25, vel: 85}
      - {beat: 7.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 8.00, pitch: "", dur: 0.25, vel: 78}
      - {beat: 8.67, pitch: "", dur: 0.15, vel: 68}
      # bar 3
      - {beat: 9.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 9.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 10.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 10.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 11.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 11.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 12.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 12.67, pitch: "", dur: 0.15, vel: 70}
      # bar 4
      - {beat: 13.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 13.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 14.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 14.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 15.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 15.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 16.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 16.67, pitch: "", dur: 0.15, vel: 70}
      # bar 5
      - {beat: 17.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 17.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 18.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 18.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 19.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 19.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 20.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 20.67, pitch: "", dur: 0.15, vel: 70}
      # bar 6
      - {beat: 21.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 21.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 22.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 22.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 23.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 23.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 24.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 24.67, pitch: "", dur: 0.15, vel: 70}
      # bar 7
      - {beat: 25.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 25.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 26.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 26.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 27.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 27.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 28.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 28.67, pitch: "", dur: 0.15, vel: 70}
      # bar 8
      - {beat: 29.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 29.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 30.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 30.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 31.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 31.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 32.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 32.67, pitch: "", dur: 0.15, vel: 70}
      # bar 9
      - {beat: 33.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 33.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 34.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 34.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 35.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 35.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 36.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 36.67, pitch: "", dur: 0.15, vel: 70}
      # bar 10
      - {beat: 37.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 37.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 38.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 38.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 39.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 39.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 40.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 40.67, pitch: "", dur: 0.15, vel: 70}
      # bar 11
      - {beat: 41.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 41.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 42.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 42.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 43.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 43.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 44.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 44.67, pitch: "", dur: 0.15, vel: 70}
      # bar 12
      - {beat: 45.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 45.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 46.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 46.67, pitch: "", dur: 0.15, vel: 70}
      - {beat: 47.00, pitch: "", dur: 0.25, vel: 88}
      - {beat: 47.67, pitch: "", dur: 0.15, vel: 72}
      - {beat: 48.00, pitch: "", dur: 0.25, vel: 80}
      - {beat: 48.67, pitch: "", dur: 0.15, vel: 70}

  # Kick: sparse jazz feathering — light hits on beats 1 and 3 of selected
  # bars (skips some bars for breathing room).
  kick:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: anchor
    loop_bars: 12
    events:
      - {beat: 1.0,  pitch: "", dur: 0.25, vel: 72}
      - {beat: 3.0,  pitch: "", dur: 0.25, vel: 64}
      - {beat: 5.0,  pitch: "", dur: 0.25, vel: 70}
      - {beat: 9.0,  pitch: "", dur: 0.25, vel: 72}
      - {beat: 11.0, pitch: "", dur: 0.25, vel: 64}
      - {beat: 13.0, pitch: "", dur: 0.25, vel: 68}
      - {beat: 17.0, pitch: "", dur: 0.25, vel: 72}
      - {beat: 21.0, pitch: "", dur: 0.25, vel: 70}
      - {beat: 25.0, pitch: "", dur: 0.25, vel: 68}
      - {beat: 29.0, pitch: "", dur: 0.25, vel: 70}
      - {beat: 33.0, pitch: "", dur: 0.25, vel: 72}
      - {beat: 37.0, pitch: "", dur: 0.25, vel: 70}
      - {beat: 41.0, pitch: "", dur: 0.25, vel: 72}
      # turnaround fill — 4 kicks on bar 12 closing
      - {beat: 45.0, pitch: "", dur: 0.25, vel: 80}
      - {beat: 46.0, pitch: "", dur: 0.25, vel: 64}
      - {beat: 47.0, pitch: "", dur: 0.25, vel: 70}
      - {beat: 48.5, pitch: "", dur: 0.25, vel: 60}

  # Snare: backbeat on 2 and 4, occasional ghost on swung-and. Skips some
  # bars to leave space.
  snare:
    family: drums
    tone: [live, soft]
    articulation: swing
    prominence: support
    loop_bars: 12
    events:
      # bars 1-4
      - {beat: 2.0,  pitch: "", dur: 0.25, vel: 88}
      - {beat: 4.0,  pitch: "", dur: 0.25, vel: 84}
      - {beat: 3.67, pitch: "", dur: 0.25, vel: 42, art: ghost}
      - {beat: 6.0,  pitch: "", dur: 0.25, vel: 88}
      - {beat: 8.0,  pitch: "", dur: 0.25, vel: 84}
      - {beat: 7.67, pitch: "", dur: 0.25, vel: 42, art: ghost}
      - {beat: 10.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 12.0, pitch: "", dur: 0.25, vel: 84}
      - {beat: 14.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 16.0, pitch: "", dur: 0.25, vel: 86}
      - {beat: 15.67, pitch: "", dur: 0.25, vel: 42, art: ghost}
      # bars 5-8
      - {beat: 18.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 20.0, pitch: "", dur: 0.25, vel: 84}
      - {beat: 22.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 24.0, pitch: "", dur: 0.25, vel: 84}
      - {beat: 23.67, pitch: "", dur: 0.25, vel: 42, art: ghost}
      - {beat: 26.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 28.0, pitch: "", dur: 0.25, vel: 84}
      - {beat: 30.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 32.0, pitch: "", dur: 0.25, vel: 86}
      - {beat: 31.67, pitch: "", dur: 0.25, vel: 42, art: ghost}
      # bars 9-12
      - {beat: 34.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 36.0, pitch: "", dur: 0.25, vel: 84}
      - {beat: 38.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 40.0, pitch: "", dur: 0.25, vel: 84}
      - {beat: 39.67, pitch: "", dur: 0.25, vel: 42, art: ghost}
      - {beat: 42.0, pitch: "", dur: 0.25, vel: 88}
      - {beat: 44.0, pitch: "", dur: 0.25, vel: 84}
      - {beat: 46.0, pitch: "", dur: 0.25, vel: 92, art: accent}
      - {beat: 47.5, pitch: "", dur: 0.25, vel: 70}
      - {beat: 48.0, pitch: "", dur: 0.25, vel: 88}

  # Piano: rootless A/B alternating voicings, comping on the "and" of beats 2
  # and 4 (the syncopation that makes bop swing). Voice-leads smoothly between
  # chord changes — shares notes between adjacent voicings where possible.
  piano:
    family: acoustic_piano
    tone: [clear, present]
    articulation: comp
    register: mid
    prominence: support
    loop_bars: 12
    events:
      # bar 1 F7 — A-voicing: Eb-A-C (b7-3-5)
      - {beat: 2.50, pitch: Eb4, dur: 0.4, vel: 72}
      - {beat: 2.50, pitch: A4, dur: 0.4, vel: 68}
      - {beat: 2.50, pitch: C5, dur: 0.4, vel: 70}
      - {beat: 4.50, pitch: Eb4, dur: 0.35, vel: 66}
      - {beat: 4.50, pitch: A4, dur: 0.35, vel: 62}
      # bar 2 Bb7 — A-voicing: Ab-D-F (b7-3-5)
      - {beat: 6.50, pitch: Ab4, dur: 0.4, vel: 72}
      - {beat: 6.50, pitch: D5, dur: 0.4, vel: 68}
      - {beat: 6.50, pitch: F5, dur: 0.4, vel: 70}
      - {beat: 8.50, pitch: Ab4, dur: 0.35, vel: 66}
      - {beat: 8.50, pitch: D5, dur: 0.35, vel: 62}
      # bar 3 F7 — B-voicing: A-C-Eb-G (3-5-b7-9)
      - {beat: 10.50, pitch: A4, dur: 0.4, vel: 72}
      - {beat: 10.50, pitch: C5, dur: 0.4, vel: 68}
      - {beat: 10.50, pitch: Eb5, dur: 0.4, vel: 70}
      - {beat: 10.50, pitch: G5, dur: 0.4, vel: 74}
      - {beat: 12.50, pitch: A4, dur: 0.35, vel: 64}
      # bar 4 Cm7 → F7 — ii-V: Eb-G-Bb-D / A-C-Eb-G
      - {beat: 13.50, pitch: Eb4, dur: 0.4, vel: 72}
      - {beat: 13.50, pitch: G4, dur: 0.4, vel: 68}
      - {beat: 13.50, pitch: Bb4, dur: 0.4, vel: 70}
      - {beat: 13.50, pitch: D5, dur: 0.4, vel: 72}
      - {beat: 15.50, pitch: Eb4, dur: 0.35, vel: 64}
      - {beat: 15.50, pitch: A4, dur: 0.35, vel: 66}
      # bar 5 Bb7 (IV)
      - {beat: 18.50, pitch: Ab4, dur: 0.4, vel: 70}
      - {beat: 18.50, pitch: D5, dur: 0.4, vel: 68}
      - {beat: 18.50, pitch: F5, dur: 0.4, vel: 70}
      - {beat: 20.50, pitch: Ab4, dur: 0.35, vel: 62}
      # bar 6 Bdim7
      - {beat: 22.50, pitch: F4, dur: 0.4, vel: 72}
      - {beat: 22.50, pitch: Ab4, dur: 0.4, vel: 68}
      - {beat: 22.50, pitch: D5, dur: 0.4, vel: 70}
      - {beat: 24.50, pitch: F4, dur: 0.35, vel: 64}
      # bar 7 F7 → D7
      - {beat: 26.50, pitch: A4, dur: 0.4, vel: 72}
      - {beat: 26.50, pitch: Eb5, dur: 0.4, vel: 68}
      - {beat: 28.50, pitch: F#4, dur: 0.4, vel: 76}
      - {beat: 28.50, pitch: C5, dur: 0.4, vel: 72}
      - {beat: 28.50, pitch: A4, dur: 0.4, vel: 70}
      # bar 8 Gm7 → C7
      - {beat: 30.50, pitch: F4, dur: 0.4, vel: 72}
      - {beat: 30.50, pitch: Bb4, dur: 0.4, vel: 68}
      - {beat: 30.50, pitch: D5, dur: 0.4, vel: 70}
      - {beat: 32.50, pitch: E4, dur: 0.4, vel: 74}
      - {beat: 32.50, pitch: Bb4, dur: 0.4, vel: 70}
      # bar 9 Bb7
      - {beat: 34.50, pitch: Ab4, dur: 0.4, vel: 70}
      - {beat: 34.50, pitch: D5, dur: 0.4, vel: 68}
      - {beat: 34.50, pitch: F5, dur: 0.4, vel: 70}
      - {beat: 36.50, pitch: D5, dur: 0.35, vel: 62}
      # bar 10 F7
      - {beat: 38.50, pitch: Eb4, dur: 0.4, vel: 70}
      - {beat: 38.50, pitch: A4, dur: 0.4, vel: 66}
      - {beat: 38.50, pitch: C5, dur: 0.4, vel: 68}
      - {beat: 40.50, pitch: A4, dur: 0.35, vel: 60}
      # bar 11 D7 (VI7) — tension
      - {beat: 42.50, pitch: F#4, dur: 0.4, vel: 76, art: accent}
      - {beat: 42.50, pitch: C5, dur: 0.4, vel: 72}
      - {beat: 42.50, pitch: A4, dur: 0.4, vel: 70}
      - {beat: 44.50, pitch: F#4, dur: 0.35, vel: 64}
      # bar 12 G7 → C7 turnaround back to F
      - {beat: 46.50, pitch: F4, dur: 0.4, vel: 74}
      - {beat: 46.50, pitch: B4, dur: 0.4, vel: 70}
      - {beat: 46.50, pitch: D5, dur: 0.4, vel: 72}
      - {beat: 48.50, pitch: E4, dur: 0.35, vel: 70}
      - {beat: 48.50, pitch: Bb4, dur: 0.35, vel: 66}
      - {beat: 48.50, pitch: G4, dur: 0.35, vel: 68}

  # Tenor sax: 8-bar phrase with contour (rising motion, peak ~bar 5,
  # descending resolution by bar 8). Repeats the form across the 12 bars.
  tenor:
    family: reed_lead
    tone: [present, round]
    articulation: lyrical
    register: mid-high
    prominence: lead
    loop_bars: 12
    events:
      # bars 1-2: F mixolydian motion (F-A-C-Eb)
      - {beat: 1.00, pitch: F4, dur: 0.5, vel: 90, art: accent}
      - {beat: 1.67, pitch: A4, dur: 0.33, vel: 82}
      - {beat: 2.00, pitch: C5, dur: 0.5, vel: 86}
      - {beat: 2.67, pitch: Eb5, dur: 0.33, vel: 78}
      - {beat: 3.00, pitch: D5, dur: 0.5, vel: 84}
      - {beat: 3.67, pitch: C5, dur: 0.33, vel: 76}
      - {beat: 4.00, pitch: A4, dur: 1.0, vel: 88, art: tenuto}
      # bars 3-4: rising contour
      - {beat: 9.00, pitch: F4, dur: 0.5, vel: 86}
      - {beat: 9.67, pitch: A4, dur: 0.33, vel: 80}
      - {beat: 10.00, pitch: C5, dur: 0.5, vel: 88, art: accent}
      - {beat: 10.67, pitch: Eb5, dur: 0.33, vel: 82}
      - {beat: 11.00, pitch: F5, dur: 0.5, vel: 92, art: accent}
      - {beat: 11.67, pitch: G5, dur: 0.33, vel: 84}
      - {beat: 12.00, pitch: A5, dur: 1.0, vel: 94, art: tenuto}
      # bar 5 PEAK (Bb7): high register
      - {beat: 17.00, pitch: Bb5, dur: 0.5, vel: 96, art: accent}
      - {beat: 17.67, pitch: G5, dur: 0.33, vel: 86}
      - {beat: 18.00, pitch: F5, dur: 0.5, vel: 90}
      - {beat: 18.67, pitch: D5, dur: 0.33, vel: 80}
      # bar 6 Bdim7 — descending
      - {beat: 21.00, pitch: Ab5, dur: 0.5, vel: 88}
      - {beat: 21.67, pitch: F5, dur: 0.33, vel: 80}
      - {beat: 22.00, pitch: D5, dur: 0.5, vel: 84}
      - {beat: 22.67, pitch: B4, dur: 0.33, vel: 76}
      # bar 7 F7 → D7
      - {beat: 25.00, pitch: C5, dur: 0.5, vel: 86}
      - {beat: 25.67, pitch: A4, dur: 0.33, vel: 78}
      - {beat: 26.00, pitch: F4, dur: 0.5, vel: 82}
      - {beat: 27.00, pitch: F#4, dur: 0.5, vel: 84, art: accent}
      - {beat: 27.67, pitch: A4, dur: 0.33, vel: 78}
      - {beat: 28.00, pitch: C5, dur: 1.0, vel: 86, art: tenuto}
      # bar 8 Gm7 → C7 resolution
      - {beat: 29.00, pitch: Bb4, dur: 0.5, vel: 84}
      - {beat: 29.67, pitch: D5, dur: 0.33, vel: 78}
      - {beat: 30.00, pitch: F5, dur: 0.5, vel: 86}
      - {beat: 31.00, pitch: E5, dur: 0.5, vel: 82}
      - {beat: 31.67, pitch: D5, dur: 0.33, vel: 76}
      - {beat: 32.00, pitch: C5, dur: 1.5, vel: 84, art: tenuto}
      # bar 9-10 Bb7-F7 echo
      - {beat: 33.00, pitch: F5, dur: 0.5, vel: 84}
      - {beat: 33.67, pitch: D5, dur: 0.33, vel: 76}
      - {beat: 34.00, pitch: Bb4, dur: 0.5, vel: 80}
      - {beat: 37.00, pitch: A4, dur: 0.5, vel: 84}
      - {beat: 37.67, pitch: C5, dur: 0.33, vel: 78}
      - {beat: 38.00, pitch: F4, dur: 1.0, vel: 82}
      # bar 11 D7 — chromatic phrase
      - {beat: 41.00, pitch: F#5, dur: 0.5, vel: 88, art: accent}
      - {beat: 41.67, pitch: E5, dur: 0.33, vel: 80}
      - {beat: 42.00, pitch: D5, dur: 0.5, vel: 84}
      - {beat: 43.00, pitch: C5, dur: 0.5, vel: 82}
      - {beat: 43.67, pitch: A4, dur: 0.33, vel: 76}
      # bar 12 G7/C7 — final cadence to F1 (next chorus)
      - {beat: 45.00, pitch: G4, dur: 0.5, vel: 84}
      - {beat: 45.67, pitch: F4, dur: 0.33, vel: 78}
      - {beat: 46.00, pitch: E4, dur: 0.5, vel: 80}
      - {beat: 47.00, pitch: G4, dur: 0.5, vel: 82}
      - {beat: 48.00, pitch: F4, dur: 1.0, vel: 86, art: tenuto}

sections:
  - id: head-a1
    title: first chorus
    duration: 16s
    harmony: "F7 | Bb7 | F7 | Cm7 F7"
    scene: "head relaxed"
    variation: "statement"
    groove: swing_56
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.6}
          - {at: 100, value: 0.85}

  - id: head-a2
    title: second chorus
    duration: 16s
    harmony: "Bb7 | Bdim7 | F7 D7 | Gm7 C7"
    scene: "head glide"
    variation: "sequence-up"
    groove: swing_56
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.8}
          - {at: 100, value: 0.9}

  - id: bridge
    title: D7 lift
    duration: 8s
    harmony: "D7 | G7 C7"
    scene: "bridge lift"
    variation: "open-register"
    groove: swing_56
    substitutions:
      - {rule: tritone_sub, apply_to: V, probability: 0.6}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.85}
          - {at: 60, value: 1.0}
          - {at: 100, value: 0.9}

  - id: out-head
    title: bar stools empty
    duration: 16s
    harmony: "F7 | Bb7 | F7 | Cm7 F7"
    scene: "outro cadence"
    variation: "cadence"
    groove: swing_56
    substitutions:
      - {rule: deceptive, apply_to: V, probability: 0.6}
    automation:
      - param: expression
        breakpoints:
          - {at: 0, value: 0.85}
          - {at: 100, value: 0.3}

title: Smoke and Mirrors
description: |
  Smoky after-hours jazz vocal in the style of late-set torch songs. Female
  alto over piano trio. Bill Evans-style rootless voicings, walking bass,
  brushed ride cymbal.
style: jazz
substyle: vocal-ballad
listen_mode: hour-stream
render_engine: acestep
seed: 90022

key: Fmin
tempo: 76
total_duration: 3m20s
tags: [jazz, vocal, female alto, smoky, torch song, ballad, after-hours]

acestep:
  voice: smoky female alto, breathy chest voice with subtle vibrato, lounge-jazz delivery
  style: >
    After-hours jazz lounge. Female alto vocal up front with breathy
    intimacy and slight vibrato. Piano trio backing: Bill Evans-style
    rootless voicings (3-5-7-9 in left hand), walking upright bass that
    never repeats a pitch within a bar, brushed ride cymbal spang-a-lang.
    Hot tube saturation on the master. Small-club room sound, vocal sits
    close to the mic.
  tags: [jazz, vocal, female alto, smoky, torch song, ballad]
  scale: minor
  time_signature: 4/4
  inference_steps: 12
  motif: |
    Verse melody starts on the 5th, falls a sixth, climbs back with chromatic
    passing tones. Chorus is wider — leaps of a fourth and fifth, lands on the
    minor 9th for tension before resolution.
  lyrics: |
    [Verse]
    The neon outside spells your name in pink
    Through the window of this dive
    I order coffee, I should have asked for stronger
    Just to feel a little more alive

    [Chorus]
    Smoke and mirrors, that's what love is
    Smoke and mirrors, here and gone
    I keep waiting for a curtain
    That was never really drawn

    [Verse]
    The piano plays a song we used to dance to
    The brush on the snare is soft and slow
    The bartender knows the look I'm wearing
    He's seen it on every face in this row

    [Chorus]
    Smoke and mirrors, that's what love is
    Smoke and mirrors, here and gone
    I keep waiting for a curtain
    That was never really drawn

    [Bridge]
    Tell me you weren't a magician
    Tell me the trick was real
    Tell me I didn't imagine
    Every word, every feel

    [Chorus]
    Smoke and mirrors, that's what love is
    Smoke and mirrors, here and gone
    I keep waiting for a curtain
    That was never really drawn

    [Outro]
    Was never really drawn
  sections:
    - id: intro
      bars: 4
      description: piano alone, slow rubato chord, bass enters bar 3
      harmony: "Fm9 | Bbm7 | Eb13 | Abmaj7"
      dynamic: soft
    - id: verse1
      bars: 12
      description: vocal enters intimate, brushes come in, walking bass
      harmony: "Fm9 Bbm7 | Eb13 Abmaj7 | Dm7b5 G7b9 | Cm7"
      dynamic: building
    - id: chorus1
      bars: 8
      description: vocal opens, piano comping more rhythmically, ride accent
      harmony: "Bbm7 Eb13 | Abmaj7 Fm9 | Dm7b5 G7b9 | Cm7"
      dynamic: peak
    - id: verse2
      bars: 12
      description: pulls back, second verse, occasional sax fill between phrases
      harmony: "Fm9 Bbm7 | Eb13 Abmaj7 | Dm7b5 G7b9 | Cm7"
      dynamic: soft
    - id: chorus2
      bars: 8
      description: chorus repeats, vocal adds harmony on the title line
      harmony: "Bbm7 Eb13 | Abmaj7 Fm9 | Dm7b5 G7b9 | Cm7"
      dynamic: peak
    - id: bridge
      bars: 8
      description: harmonic shift up a half step, drums lay out, vocal exposed
      harmony: "Dbmaj7 | Bb7alt | Ebmaj7 | C7b9"
      dynamic: drop
    - id: chorus3
      bars: 8
      description: final chorus, full ensemble, climactic
      harmony: "Bbm7 Eb13 | Abmaj7 Fm9 | Dm7b5 G7b9 | Cm7"
      dynamic: peak
    - id: outro
      bars: 4
      description: tag the title line, rallentando, vocal fades
      harmony: "Fm9 | Cm7"
      dynamic: fade

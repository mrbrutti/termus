title: Sunday Cafe
description: |
  Indie chill-pop vocal track. Warm female mezzo with conversational
  phrasing, acoustic guitar fingerpicking, light brushed kit, gentle
  Rhodes pad, and a relaxed Sunday-afternoon feeling.
style: chill
substyle: vocal-indie-pop
listen_mode: hour-stream
render_engine: acestep
seed: 90044

key: Cmaj
tempo: 96
total_duration: 3m
tags: [chill, indie pop, vocal, female mezzo, acoustic, sunday afternoon]

acestep:
  voice: warm female mezzo, conversational phrasing, light vibrato on sustained notes
  style: >
    Indie chill-pop on a Sunday afternoon. Female mezzo vocal sits forward,
    warm and conversational. Acoustic guitar fingerpicks Travis-style
    eighth-note patterns. Soft brushed kit on a half-time backbeat. Rhodes
    pad sustains underneath the verses. Walking bass picks up in the chorus.
    Vocal harmony stacks on the chorus refrain. Light tape saturation, no
    over-compression. Recorded close-up so you can hear the room.
  tags: [chill, indie pop, vocal, female mezzo, acoustic guitar, brushed drums]
  scale: major
  time_signature: 4/4
  inference_steps: 10
  motif: |
    Verse melody is conversational, mostly stepwise in the middle of the
    voice. Chorus hook leaps a fifth up and lands on the major 7th before
    resolving — that's the singalong moment.
  lyrics: |
    [Verse]
    Sunday morning, slow and golden
    Coffee steams up on the sill
    You're still sleeping, sheets are tangled
    And the whole apartment's still

    [Verse]
    Out the window, neighbors waving
    Someone's playing piano scales
    Plans we made for going somewhere
    Got rewritten without details

    [Chorus]
    But it's alright, it's alright
    Nothing's burning, nothing's wrong
    Just us breathing, just us being
    Just a sunny afternoon song

    [Verse]
    Books we said we'd read together
    Stacked beside the radiator
    Bookmarks in them, marking promises
    We'll get back to a little later

    [Chorus]
    But it's alright, it's alright
    Nothing's burning, nothing's wrong
    Just us breathing, just us being
    Just a sunny afternoon song

    [Bridge]
    Maybe we don't have to chase the
    Loudest version of our lives
    Maybe Sunday's all the proof we
    Need that something here survives

    [Chorus]
    And it's alright, it's alright
    Nothing's burning, nothing's wrong
    Just us breathing, just us being
    Just a sunny afternoon song

    [Outro]
    Just a sunny afternoon song
  sections:
    - id: intro
      bars: 4
      description: acoustic guitar alone, fingerpicked, light dynamic
      harmony: "Cmaj7 | G/B | Am7 | Fmaj7"
      dynamic: soft
    - id: verse1
      bars: 8
      description: vocal enters intimate, Rhodes pad joins on bar 5, no drums yet
      harmony: "Cmaj7 G/B | Am7 Fmaj7 | Cmaj7 G/B | Am7 Fmaj7"
      dynamic: soft
    - id: verse2
      bars: 8
      description: brushed kit and walking bass enter, dynamic builds slightly
      harmony: "Cmaj7 G/B | Am7 Fmaj7 | Cmaj7 G/B | Am7 Fmaj7"
      dynamic: building
    - id: chorus1
      bars: 8
      description: full ensemble, vocal harmony stacks on the refrain
      harmony: "Fmaj7 Cmaj7 | G Am7 | Fmaj7 Cmaj7 | G Cmaj7"
      dynamic: peak
    - id: verse3
      bars: 8
      description: pulls back to verse dynamic, vocal continues
      harmony: "Cmaj7 G/B | Am7 Fmaj7 | Cmaj7 G/B | Am7 Fmaj7"
      dynamic: soft
    - id: chorus2
      bars: 8
      description: chorus repeats with full harmony, slight melodic variation on top line
      harmony: "Fmaj7 Cmaj7 | G Am7 | Fmaj7 Cmaj7 | G Cmaj7"
      dynamic: peak
    - id: bridge
      bars: 8
      description: harmonic shift, drums drop out, acoustic and vocal exposed
      harmony: "Dm7 | G7 | Em7 | Am7 | Dm7 | G7 | Cmaj7 | G7"
      dynamic: drop
    - id: chorus3
      bars: 8
      description: final chorus, full ensemble, climactic with vocal ad-libs
      harmony: "Fmaj7 Cmaj7 | G Am7 | Fmaj7 Cmaj7 | G Cmaj7"
      dynamic: peak
    - id: outro
      bars: 4
      description: vocal tags the title phrase, acoustic guitar resolves to tonic
      harmony: "Fmaj7 | Cmaj7"
      dynamic: fade

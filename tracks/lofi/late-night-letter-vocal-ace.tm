title: Late Night Letter
description: |
  Lofi vocal track. Intimate, slightly raspy male tenor over warm Rhodes,
  brushed kit, walking upright bass. A late-night confession to someone far away.
style: lofi
substyle: vocal-ballad
listen_mode: hour-stream
render_engine: acestep
seed: 90011

key: Dmin
tempo: 84
total_duration: 3m
tags: [lofi, vocal, male tenor, intimate, jazz-influenced, rainy night]

acestep:
  voice: raspy male tenor, intimate close-mic delivery, slight breath audible
  style: >
    Warm late-night lo-fi with a single male vocal up front. Fender Rhodes
    plays slow rootless voicings in D minor. Upright bass walks in quarters
    with chromatic approach to the next chord. Brushed kit on 2 and 4,
    sparse ghost notes. Vinyl crackle and subtle tape saturation underneath.
    The vocal sits forward in the mix, room reverb pulled back so the
    delivery feels close and confessional.
  tags: [lofi, vocal, male tenor, intimate, jazz-influenced, late night]
  scale: minor
  time_signature: 4/4
  inference_steps: 10
  motif: |
    Verse melody descends from the 5th, rests on the 3rd. Chorus melody
    climbs to the 9th and resolves down to the tonic. Same shape in both
    choruses, slight variation on the third repeat.
  lyrics: |
    [Verse]
    Three a.m., the city is quiet
    I'm still awake with the lamp on low
    Pen in hand but the page is empty
    Don't know exactly where these words should go

    [Chorus]
    If you read this in the morning
    Just remember I'm still here
    Counting hours, counting moments
    Wishing distance would disappear

    [Verse]
    Outside the rain is keeping rhythm
    Soft like brushes on a snare
    Every drop a small reminder
    Of all the things I wish that we could share

    [Chorus]
    If you read this in the morning
    Just remember I'm still here
    Counting hours, counting moments
    Wishing distance would disappear

    [Bridge]
    Maybe time will fold around us
    Maybe seasons turn our way
    Until then I'll write these letters
    Just to make it through the day

    [Outro]
    Just to make it through the day
  sections:
    - id: intro
      bars: 4
      description: rhodes and bass enter alone, no vocal, soft pad
      harmony: "Dm9 Gm7"
      dynamic: soft
    - id: verse1
      bars: 8
      description: vocal enters intimate, kit brushes come in halfway
      harmony: "Dm9 Gm7 | Bb6 A7"
      dynamic: building
    - id: chorus1
      bars: 8
      description: vocal opens up, full kit, chord stabs more present
      harmony: "Dm9 Gm7 | Bb6 A7"
      dynamic: peak
    - id: verse2
      bars: 8
      description: drop back to verse dynamic, vocal more textured
      harmony: "Dm9 Gm7 | Bb6 A7"
      dynamic: soft
    - id: chorus2
      bars: 8
      description: chorus repeats with slight melodic variation, harmony added on the second pass
      harmony: "Dm9 Gm7 | Bb6 A7"
      dynamic: peak
    - id: bridge
      bars: 8
      description: harmonic shift to relative major area, drums drop out, vocal floats
      harmony: "Bbmaj7 | F | Am7 | Dm7"
      dynamic: drop
    - id: outro
      bars: 4
      description: kit fades back in for a final tag, vocal sustains the title line, rallentando feel
      harmony: "Dm9"
      dynamic: fade

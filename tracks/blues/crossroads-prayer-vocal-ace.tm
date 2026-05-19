title: Crossroads Prayer
description: |
  Slow blues with a gravelly male baritone vocal. Slide guitar, harmonica
  interjections, walking bass shuffle, soulful organ pad.
style: blues
substyle: vocal-slow-blues
listen_mode: hour-stream
render_engine: acestep
seed: 90033

key: Emaj
tempo: 68
total_duration: 3m30s
tags: [blues, vocal, male baritone, slow blues, slide guitar, harmonica]

acestep:
  voice: gravelly male baritone, weathered timbre, prayer-like delivery with breath catches
  style: >
    Slow 12-bar blues at 68 BPM in E major. Gravelly male baritone vocal
    that sounds like he's lived it. Slide guitar fills between vocal lines,
    harmonica answers the slide guitar. Walking bass plays a shuffled
    quarter-note pulse with chromatic approach to each new chord.
    Hammond B3 organ pad sustains underneath everything. Brushed kit with
    snare on 2 and 4. Tape-saturated mix, no auto-tune, no vocal effects
    beyond a touch of plate reverb. Captured at a smoky bar near closing time.
  tags: [blues, vocal, male baritone, slow blues, slide guitar, harmonica, B3 organ]
  scale: major
  time_signature: 4/4
  inference_steps: 12
  motif: |
    Vocal phrases follow the classic blues call-and-response shape: two-bar
    line, two-bar instrumental answer (slide or harp). Melody hangs on the
    b3 (blue note), bends up to the major 3rd at phrase endings.
  lyrics: |
    [Verse]
    I went down to the crossroads
    Like the old man told me to
    I went down to the crossroads
    Like the old man told me to
    Said the answer's in the listenin'
    Not in anything you do

    [Verse]
    I stood there 'til the sun went down
    Watching shadows grow real long
    I stood there 'til the sun went down
    Watching shadows grow real long
    Heard the wind start playing
    Something close to a song

    [Chorus]
    Lord I'm tired, Lord I'm waitin'
    For a sign I'd understand
    Lord I'm tired, Lord I'm waitin'
    With my hat held in my hand

    [Verse]
    Maybe answers ain't for waitin'
    Maybe answers come in time
    Maybe answers ain't for waitin'
    Maybe answers come in time
    Like a slow train at the station
    Comin' down the line

    [Chorus]
    Lord I'm tired, Lord I'm waitin'
    For a sign I'd understand
    Lord I'm tired, Lord I'm waitin'
    With my hat held in my hand

    [Outro]
    With my hat held in my hand
  sections:
    - id: intro
      bars: 4
      description: slide guitar lays down a long bending phrase over organ pad, no vocal yet
      harmony: "E7 | E7 | A7 | E7"
      dynamic: soft
    - id: verse1
      bars: 12
      description: vocal enters, classic 12-bar blues form, slide fills between lines
      harmony: "E7 | A7 | E7 | E7 | A7 | A7 | E7 | E7 | B7 | A7 | E7 | B7"
      dynamic: building
    - id: verse2
      bars: 12
      description: harmonica takes the slide guitar's role, second verse, dynamics rise
      harmony: "E7 | A7 | E7 | E7 | A7 | A7 | E7 | E7 | B7 | A7 | E7 | B7"
      dynamic: building
    - id: chorus1
      bars: 8
      description: chorus form (8-bar bridge with IV emphasis), vocal more declarative
      harmony: "A7 | A7 | E7 | E7 | B7 | A7 | E7 | B7"
      dynamic: peak
    - id: verse3
      bars: 12
      description: third verse, full ensemble, slide and harp trade four-bar phrases
      harmony: "E7 | A7 | E7 | E7 | A7 | A7 | E7 | E7 | B7 | A7 | E7 | B7"
      dynamic: peak
    - id: chorus2
      bars: 8
      description: chorus repeats, vocal pushes higher on the third line
      harmony: "A7 | A7 | E7 | E7 | B7 | A7 | E7 | B7"
      dynamic: peak
    - id: outro
      bars: 4
      description: vocal tag on the last line, rit. into a long E7 chord with bends
      harmony: "E7 | E7 | E7 | E7"
      dynamic: fade

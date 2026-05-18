title: Deep Pad Breath
description: SP18 form-driven ambient long-form — C major, ambient_emerge_drift_recede. 60+80+40 = 180 bars @ 60 BPM = ~12m per pass.
style: ambient
mix_bus: ambient
listen_mode: hour-stream
seed: 38901
tags: [ambient, pad, brass, breath, deep, strings, choir, sp18]
key: Cmaj
tempo: 60
globals: {density: busy, brightness: warm, motion: slow, reverb: cathedral}

# SP18 form: ambient_emerge_drift_recede — three acts.
# emerge (60 bars), drift (80 bars), recede (40 bars). 180 bars @ 60 BPM = 12 minutes per pass.
form: ambient_emerge_drift_recede
total_duration: 12m

motif_library:
  drone_theme:
    pattern: "5 . . . . . . . | 7 . . . . . . . | 9 . . . . . . . | 5 . . . . . . ."
    description: "very slow rising arpeggio — one note per bar"
    bars: 4

roles:
  pad:
    family: pad
    tone: [soft, wide, deep]
    register: low
    prominence: anchor

  drone:
    family: pad
    voice: ambient_drone_deep
    tone: [warm, deep]
    register: low
    prominence: anchor

  strings:
    family: strings
    tone: [soft, warm]
    register: mid
    prominence: support

  choir:
    family: choir
    tone: [soft, airy]
    register: high
    prominence: air

  bell:
    family: bells
    tone: [glass, sparkle, soft]
    register: high
    prominence: air

  bass:
    family: bass
    tone: [soft, deep]
    register: low
    prominence: anchor

  lead:
    family: brass
    tone: [soft, airy]
    register: mid-high
    prominence: lead
    personality: brass_breath
    room: cathedral_large
    reverb_send_db: -6

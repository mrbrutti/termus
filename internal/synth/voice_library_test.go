package synth

import "testing"

func TestVoiceLibrary_HasAtLeast12Voices(t *testing.T) {
	voices := VoiceLibrary()
	if len(voices) < 12 {
		t.Errorf("voice library has %d entries; SP16 requires at least 12", len(voices))
	}
	seen := map[string]bool{}
	for _, v := range voices {
		if v.Name == "" {
			t.Errorf("voice with empty Name: %+v", v)
		}
		if seen[v.Name] {
			t.Errorf("duplicate voice name: %s", v.Name)
		}
		seen[v.Name] = true
	}
}

func TestVoiceByName_Lookup(t *testing.T) {
	cases := []string{
		"lofi_felt_piano",
		"lofi_rhodes_warm",
		"jazz_grand_piano",
		"jazz_upright_bass",
		"chill_pad_warm",
		"ambient_drone_choir",
		"bell_struck_bright",
	}
	for _, name := range cases {
		v := VoiceByName(name)
		if v == nil {
			t.Errorf("VoiceByName(%q) = nil; expected found", name)
			continue
		}
		if v.Name != name {
			t.Errorf("VoiceByName(%q).Name = %q", name, v.Name)
		}
	}
}

func TestVoiceByName_Missing(t *testing.T) {
	if v := VoiceByName("no_such_voice_definitely"); v != nil {
		t.Errorf("VoiceByName for missing voice returned %+v, expected nil", v)
	}
	if v := VoiceByName(""); v != nil {
		t.Errorf("VoiceByName(\"\") returned %+v, expected nil", v)
	}
}

func TestHzToMIDICutoff_Monotonic(t *testing.T) {
	// 100 < 1000 < 5000 should produce increasing CC74 values.
	a := HzToMIDICutoff(100)
	b := HzToMIDICutoff(1000)
	c := HzToMIDICutoff(5000)
	if !(a < b && b < c) {
		t.Errorf("HzToMIDICutoff not monotonic: 100=%d 1000=%d 5000=%d", a, b, c)
	}
}

package gen

import "testing"

func TestMixChainByName_Lookup(t *testing.T) {
	cases := []string{"piano_default", "bass_default", "drums_kick", "drums_snare", "drums_hat", "lead_default", "pad_default"}
	for _, name := range cases {
		c := MixChainByName(name)
		if c == nil {
			t.Errorf("MixChainByName(%q) = nil; expected entry", name)
		}
	}
	if c := MixChainByName("definitely_not_real"); c != nil {
		t.Errorf("MixChainByName for missing entry returned %+v", c)
	}
}

func TestDefaultChainForFamily_Drums(t *testing.T) {
	kick := DefaultChainForFamily("kick")
	if kick.Compress != "punchy" {
		t.Errorf("kick.Compress = %q want punchy", kick.Compress)
	}
	if kick.ReverbSend > 0.1 {
		t.Errorf("kick.ReverbSend = %f; should be <= 0.1", kick.ReverbSend)
	}
	hat := DefaultChainForFamily("hat")
	if hat.PanOffset == 0 {
		t.Errorf("hat.PanOffset = 0; want non-centre")
	}
	snare := DefaultChainForFamily("snare")
	if snare.ReverbSend <= kick.ReverbSend {
		t.Errorf("snare reverb (%f) should exceed kick reverb (%f)", snare.ReverbSend, kick.ReverbSend)
	}
}

func TestChainOverridePerField(t *testing.T) {
	base := DefaultChainForFamily("piano")
	originalCompress := base.Compress
	rev := 0.55
	tape := 2.5
	out := MergeChain(base, &rev, "", &tape, nil)
	if out.ReverbSend != 0.55 {
		t.Errorf("ReverbSend = %f want 0.55", out.ReverbSend)
	}
	if out.Compress != originalCompress {
		t.Errorf("Compress = %q want %q (unchanged)", out.Compress, originalCompress)
	}
	if out.TapeDriveDB != 2.5 {
		t.Errorf("TapeDriveDB = %f want 2.5", out.TapeDriveDB)
	}
	if out.PanOffset != base.PanOffset {
		t.Errorf("PanOffset = %f want %f (unchanged)", out.PanOffset, base.PanOffset)
	}
}

func TestMixChain_ReverbSendCC91(t *testing.T) {
	c := MixChain{ReverbSend: 0.5}
	if c.ReverbSendCC91() < 60 || c.ReverbSendCC91() > 70 {
		t.Errorf("ReverbSendCC91 for 0.5 = %d, want ~64", c.ReverbSendCC91())
	}
	c.ReverbSend = 0
	if c.ReverbSendCC91() != 0 {
		t.Errorf("ReverbSendCC91 for 0 = %d, want 0", c.ReverbSendCC91())
	}
}

func TestMixChain_PanCC10(t *testing.T) {
	c := MixChain{PanOffset: 0}
	if c.PanCC10() != 64 {
		t.Errorf("PanCC10 for 0 = %d, want 64", c.PanCC10())
	}
	c.PanOffset = -1
	if c.PanCC10() > 5 {
		t.Errorf("PanCC10 for -1 = %d, want near 0", c.PanCC10())
	}
	c.PanOffset = 1
	if c.PanCC10() < 122 {
		t.Errorf("PanCC10 for 1 = %d, want near 127", c.PanCC10())
	}
}

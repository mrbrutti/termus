package gen

import "testing"

func TestDensifyPhraseFillsRests(t *testing.T) {
	src := []int{0, 1, 0, 2, 0, 3}
	got := densifyPhrase(src, 0, []int{9, 7}, 2)
	if got[0] == 0 && got[2] == 0 && got[4] == 0 {
		t.Fatalf("expected at least some rests to fill: %v", got)
	}
}

func TestDensityPolicyPrefersSparseTexturesForAmbientFamily(t *testing.T) {
	profile := DefaultControlProfile()
	if got := densityPolicyFor("ambient", profile).SecondaryTextureFloor; got < 3 {
		t.Fatalf("ambient secondary texture floor = %d, want >= 3", got)
	}
	if got := densityPolicyFor("bells", profile).TextureExpressionBias; got >= 0 {
		t.Fatalf("bells expression bias = %d, want negative", got)
	}
	if got := densityPolicyFor("jazz", profile).LeadFillCount; got <= 0 {
		t.Fatalf("jazz lead fill count = %d, want positive", got)
	}
}


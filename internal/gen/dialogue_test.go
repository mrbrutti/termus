package gen

import "testing"

func TestJazzCompDialogueYieldsToLead(t *testing.T) {
	on := true
	j := &Jazz{
		section:    FormSection{Kind: FormA, LeadLevel: 1},
		saxOn:      &on,
		saxMotifs:  MotifMemory{A: []int{jazzPlanThird, jazzPlanResolveThird}},
		saxPlan:    []int{jazzPlanThird, jazzPlanResolveThird},
		accentAnd2: []bool{true},
		compLines:  map[int][]int{9: {72}},
	}
	if got := j.compAccentAnd2At(0); got != -1 {
		t.Fatalf("comp accent during active lead = %d, want rest", got)
	}
}

func TestChillDialogueSilencesAnswerLayersUnderLead(t *testing.T) {
	on := true
	c := &Chill{
		section:      FormSection{Kind: FormB, LeadLevel: 1},
		saxOn:        &on,
		saxMotifs:    MotifMemory{A: []int{chillPlanNinth}},
		saxPlan:      []int{chillPlanNinth},
		vibeMotifs:   MotifMemory{A: []int{chillPlanThird}},
		vibePlan:     []int{chillPlanThird},
		guitarMotifs: MotifMemory{A: []int{chillPlanNinth}},
		guitarPlan:   []int{chillPlanNinth},
	}
	if got := c.vibeDialogueCodeAt(0); got != chillPlanRest {
		t.Fatalf("vibe dialogue code = %d, want rest", got)
	}
	if got := c.guitarDialogueCodeAt(0); got != chillPlanRest {
		t.Fatalf("guitar dialogue code = %d, want rest", got)
	}
}

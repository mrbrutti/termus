package main

import (
	"testing"

	"github.com/mrbrutti/termus/internal/gen"
)

func TestScoreListeningResultRewardsMotionAndCadence(t *testing.T) {
	markers := []gen.ListeningMarker{
		{Label: "section:A", Sample: 0},
		{Label: "cadence:cadence", Sample: 100},
	}
	snapshots := []gen.DebugStatus{
		{Section: "intro", Chord: "I"},
		{Section: "A", Chord: "I"},
		{Section: "A'", Chord: "IV"},
		{Section: "B", Chord: "V"},
		{Section: "cadence", Chord: "I"},
	}

	score := scoreListeningResult(16, markers, snapshots)
	if score.Total <= 0 {
		t.Fatalf("score.Total = %.3f, want positive", score.Total)
	}
	if score.CadenceCount != 1 {
		t.Fatalf("score.CadenceCount = %d, want 1", score.CadenceCount)
	}
	if score.ChordChanges < 2 {
		t.Fatalf("score.ChordChanges = %d, want at least 2", score.ChordChanges)
	}
	if score.UniqueSections < 4 {
		t.Fatalf("score.UniqueSections = %d, want at least 4", score.UniqueSections)
	}
}

func TestRankResultsSortsByDescendingScore(t *testing.T) {
	results := []corpusResult{
		{Name: "b", Seed: 2, Score: listeningScore{Total: 0.6}},
		{Name: "a", Seed: 1, Score: listeningScore{Total: 0.8}},
		{Name: "c", Seed: 3, Score: listeningScore{Total: 0.8}},
	}
	rankResults(results)

	if results[0].Seed != 1 {
		t.Fatalf("first seed = %d, want 1", results[0].Seed)
	}
	if results[1].Seed != 3 {
		t.Fatalf("second seed = %d, want 3", results[1].Seed)
	}
	if results[0].Rank != 1 || results[1].Rank != 2 || results[2].Rank != 3 {
		t.Fatalf("ranks = %d,%d,%d, want 1,2,3", results[0].Rank, results[1].Rank, results[2].Rank)
	}
}

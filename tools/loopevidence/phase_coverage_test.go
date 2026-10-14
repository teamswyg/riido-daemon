package main

import "testing"

func TestPhaseCoverageRowsCountCompleteLoops(t *testing.T) {
	manifest := manifest{
		RequiredPhases: requiredPhases,
		Loops:          []loop{completeLoop("one"), completeLoop("two")},
	}
	rows := phaseCoverageRows(manifest)
	if len(rows) != len(requiredPhases) {
		t.Fatalf("phase rows = %d, want %d", len(rows), len(requiredPhases))
	}
	for _, row := range rows {
		if row.Count != len(manifest.Loops) {
			t.Fatalf("%s count = %d, want %d", row.Phase, row.Count, len(manifest.Loops))
		}
	}
}

func TestBuildEvidenceExposesPhaseCoverage(t *testing.T) {
	manifest := manifest{
		ID:             "loop",
		RequiredPhases: requiredPhases,
		LoopFiles:      []string{"loops/one.riido.json"},
		Loops:          []loop{completeLoop("one")},
	}
	got := buildEvidence(manifest, "loop.md", nil)
	if got.RegisteredLoopFileCount != 1 || len(got.PhaseCoverage) != len(requiredPhases) {
		t.Fatalf("evidence = %+v", got)
	}
	if got.ProblemCount != 0 || got.ProblemSummaries == nil {
		t.Fatalf("problem evidence = %+v", got)
	}
}

func completeLoop(id string) loop {
	return loop{
		ID:            id,
		Observation:   phase{Summary: "observe"},
		Hypothesis:    phase{Summary: "hypothesis"},
		Execution:     phase{Summary: "execute"},
		Evaluation:    phase{Summary: "evaluate"},
		Retrospective: phase{Summary: "retrospective"},
		Evidence:      []evidence{{Kind: "command", Ref: "go test ./tools/loopevidence", Proves: "test"}},
	}
}

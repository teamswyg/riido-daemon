package main

import "testing"

func TestManifestLoopBudgetRejectsNewDebt(t *testing.T) {
	loops := manifestLoopReport{
		Missing:       2,
		MissingGroups: []manifestGroupCount{{Group: "docs", Count: 2}},
	}
	budget := manifestLoopBudget{
		MaxMissing:        1,
		MaxMissingByGroup: map[string]int{"docs": 1},
	}
	problems := manifestLoopBudgetProblems(loops, budget)
	if len(problems) != 2 {
		t.Fatalf("problems = %#v", problems)
	}
}

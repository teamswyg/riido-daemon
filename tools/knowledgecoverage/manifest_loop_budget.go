package main

import "fmt"

func manifestLoopBudgetProblems(loops manifestLoopReport, budget manifestLoopBudget) []string {
	var problems []string
	if loops.Missing > budget.MaxMissing {
		problems = append(problems, fmt.Sprintf("manifest loop debt exceeds budget: %d > %d", loops.Missing, budget.MaxMissing))
	}
	for _, group := range loops.MissingGroups {
		limit := budget.MaxMissingByGroup[group.Group]
		if group.Count > limit {
			problems = append(problems, fmt.Sprintf("manifest loop debt for %s exceeds budget: %d > %d", group.Group, group.Count, limit))
		}
	}
	return problems
}

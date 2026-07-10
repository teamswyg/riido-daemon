package main

import "fmt"

func validateScenarios(m manifest) []string {
	var problems []string
	known := map[string]bool{}
	for _, scenario := range m.Scenarios {
		if scenario.ID == "" || scenario.Execution == "" || scenario.MaxDurationSeconds <= 0 ||
			scenario.EstimatedInputTokens < 0 || scenario.EstimatedOutputTokens < 0 {
			problems = append(problems, "scenario needs id, execution, non-negative tokens, and duration")
		}
		if known[scenario.ID] {
			problems = append(problems, fmt.Sprintf("duplicate scenario id %q", scenario.ID))
		}
		known[scenario.ID] = true
	}
	if len(known) == 0 {
		problems = append(problems, "scenarios must not be empty")
	}
	for _, provider := range m.Providers {
		for _, id := range provider.ScenarioIDs {
			if !known[id] {
				problems = append(problems, fmt.Sprintf("provider %q references unknown scenario %q", provider.ID, id))
			}
		}
	}
	return problems
}

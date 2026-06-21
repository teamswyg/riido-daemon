package main

func validateLoop(loop evidenceLoop) []string {
	var problems []string
	if loop.Observation == "" {
		problems = append(problems, "loop observation is required")
	}
	if loop.Hypothesis == "" {
		problems = append(problems, "loop hypothesis is required")
	}
	if loop.Execute == "" {
		problems = append(problems, "loop execute is required")
	}
	if loop.Evaluate == "" {
		problems = append(problems, "loop evaluate is required")
	}
	if loop.Retrospective == "" {
		problems = append(problems, "loop retrospective is required")
	}
	return problems
}

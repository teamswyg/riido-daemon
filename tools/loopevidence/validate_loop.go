package main

import "fmt"

func validateLoop(root string, item loop, seen map[string]bool) []string {
	var problems []string
	if item.ID == "" || item.Owner == "" {
		problems = append(problems, "loop id and owner are required")
	}
	if seen[item.ID] {
		problems = append(problems, fmt.Sprintf("duplicate loop id %q", item.ID))
	}
	seen[item.ID] = true
	for name, phase := range loopPhases(item) {
		problems = append(problems, validatePhase(root, item.ID, name, phase)...)
	}
	if len(item.Evidence) == 0 {
		problems = append(problems, fmt.Sprintf("%s must include evidence", item.ID))
	}
	for _, ev := range item.Evidence {
		problems = append(problems, validateEvidence(root, item.ID, ev)...)
	}
	return problems
}

func validatePhase(root, loopID, name string, p phase) []string {
	var problems []string
	if p.Summary == "" {
		problems = append(problems, fmt.Sprintf("%s.%s summary is required", loopID, name))
	}
	for _, artifact := range p.Artifacts {
		problems = append(problems, validatePath(root, loopID+"."+name, artifact)...)
	}
	return problems
}

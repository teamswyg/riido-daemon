package main

import "fmt"

func loopPhases(item loop) map[string]phase {
	return map[string]phase{
		"observe":       item.Observation,
		"hypothesis":    item.Hypothesis,
		"execute":       item.Execution,
		"evaluate":      item.Evaluation,
		"retrospective": item.Retrospective,
	}
}

func validateEvidence(root, loopID string, ev evidence) []string {
	var problems []string
	if ev.Kind == "" || ev.Ref == "" || ev.Proves == "" {
		return []string{fmt.Sprintf("%s evidence must include kind, ref, and proves", loopID)}
	}
	if ev.Kind != "command" {
		problems = append(problems, validatePath(root, loopID+".evidence", ev.Ref)...)
	}
	return problems
}

func validateGap(item gap) []string {
	if item.ID == "" || item.Owner == "" || item.CurrentHandling == "" || item.RequiredNextArtifact == "" {
		return []string{"gap must include id, owner, current_handling, and required_next_artifact"}
	}
	return nil
}

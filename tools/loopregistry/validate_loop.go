package main

import (
	"slices"
	"time"
)

func validateLoopEntry(item loopEntry) []string {
	var problems []string
	if item.ID == "" || item.Owner == "" || item.Kind == "" {
		problems = append(problems, "loop id/owner/kind are required")
	}
	for name, values := range map[string][]string{
		"observes":   item.Observes,
		"verifies":   item.Verifies,
		"evidence":   item.Evidence,
		"fails_when": item.FailsWhen,
	} {
		if len(values) == 0 {
			problems = append(problems, item.ID+" requires "+name)
		}
	}
	if _, err := time.ParseDuration(item.ExpiresAfter); err != nil {
		problems = append(problems, item.ID+" expires_after must be a Go duration")
	}
	return append(problems, validateGraph(item.ID, item.Graph)...)
}

func validateLoop(loop evidenceLoop, owner string) []string {
	values := []string{loop.Observation, loop.Hypothesis, loop.Execute, loop.Evaluate, loop.Retrospective}
	if slices.Contains(values, "") {
		return []string{owner + " must declare observation/hypothesis/execute/evaluate/retrospective"}
	}
	return nil
}

func validateGraph(owner string, graph evidenceGraph) []string {
	values := []string{
		graph.Observation, graph.Hypothesis, graph.Change,
		graph.Verifier, graph.Evidence, graph.Decision, graph.NextLoop,
	}
	if slices.Contains(values, "") {
		return []string{owner + " evidence_graph must close observation to next_loop"}
	}
	return nil
}

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
	problems = append(problems, validateCandidateMetadata(item)...)
	return append(problems, validateGraph(item.ID, item.Graph)...)
}

func validateCandidateMetadata(item loopEntry) []string {
	if item.Kind == "closed-loop" {
		return nil
	}
	if item.CandidateCreatedAt == "" || item.PromotionTarget == "" {
		return []string{item.ID + " non-closed loop requires candidate_created_at and promotion_target"}
	}
	if _, err := time.Parse("2006-01-02", item.CandidateCreatedAt); err != nil {
		return []string{item.ID + " candidate_created_at must be YYYY-MM-DD"}
	}
	return nil
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

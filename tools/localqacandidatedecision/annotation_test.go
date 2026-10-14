package main

import (
	"strings"
	"testing"
)

func TestCandidateAnnotationExplainsRequiredArtifacts(t *testing.T) {
	problem := candidateProblem{
		CandidateID:           "repair-contract.task.thread_message",
		Reason:                "candidate has no decision record",
		RequiredNextArtifacts: []string{"claim_binding", "verifier"},
		RecommendedAction:     "Add a decision record for this candidate.",
	}
	got := githubAnnotation(problem)
	for _, want := range []string{
		"::error file=docs/30-architecture/local-qa-candidate-decision.riido.json",
		"title=Local QA Candidate Decision",
		"repair-contract.task.thread_message",
		"required_next_artifacts=claim_binding,verifier",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("annotation missing %q in %s", want, got)
		}
	}
}

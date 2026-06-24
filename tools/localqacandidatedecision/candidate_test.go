package main

import "testing"

func TestCandidateDecisionCoversLocalQACandidate(t *testing.T) {
	dir := t.TempDir()
	path := writeCandidate(t, dir, `{"closed_loop_candidates":[`+
		`{"id":"close-x","required_next_artifacts":["claim_binding","verifier"]}]}`)
	result, err := verifyCandidateDecisions(".", testManifest(), path)
	if err != nil {
		t.Fatalf("verify candidate decisions: %v", err)
	}
	if result.CandidateCount != 1 || len(result.DecisionArtifacts) != 1 {
		t.Fatalf("result = %+v", result)
	}
}

func TestCandidateDecisionRejectsMissingDecision(t *testing.T) {
	dir := t.TempDir()
	path := writeCandidate(t, dir, `{"closed_loop_candidates":[`+
		`{"id":"missing","required_next_artifacts":["claim_binding"]}]}`)
	if _, err := verifyCandidateDecisions(".", testManifest(), path); err == nil {
		t.Fatal("expected missing decision to fail")
	}
}

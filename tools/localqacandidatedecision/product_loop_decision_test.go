package main

import "testing"

func TestCandidateDecisionAllowsProductLoopDecisionWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	path := writeCandidate(t, dir, `{"closed_loop_candidates":[]}`)
	m := testManifest()
	decision := testDecision("product-outcome-evidence-assignment_completion")
	decision.CandidateScope = "product_loop"
	m.Decisions = append(m.Decisions, decision)
	if _, err := verifyCandidateDecisions(".", m, path, "product_loop"); err != nil {
		t.Fatalf("verify product loop decision absence: %v", err)
	}
}

func TestCandidateDecisionAllowsPartialReductionProductLoopDecisionWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	path := writeCandidate(t, dir, `{"closed_loop_candidates":[]}`)
	m := testManifest()
	decision := testDecision("partial-reduction-candidate-aging")
	decision.CandidateScope = "product_loop"
	m.Decisions = append(m.Decisions, decision)
	result, err := verifyCandidateDecisions(".", m, path, "product_loop")
	if err != nil {
		t.Fatalf("verify partial reduction decision absence: %v", err)
	}
	if result.AllowedMissingCount != 1 || result.MatchedDecisionCount != 0 {
		t.Fatalf("result = %+v", result)
	}
}

func TestCandidateDecisionRejectsUnknownProductLoopDecisionWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	path := writeCandidate(t, dir, `{"closed_loop_candidates":[]}`)
	m := testManifest()
	decision := testDecision("product-only")
	decision.CandidateScope = "product_loop"
	m.Decisions = append(m.Decisions, decision)
	if _, err := verifyCandidateDecisions(".", m, path, "product_loop"); err == nil {
		t.Fatal("expected unknown product loop orphan to fail")
	}
}

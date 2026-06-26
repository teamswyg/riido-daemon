package main

import "testing"

func TestCandidateDecisionRejectsOrphanDecision(t *testing.T) {
	dir := t.TempDir()
	path := writeCandidate(t, dir, `{"closed_loop_candidates":[`+
		`{"id":"close-x","required_next_artifacts":["claim_binding"]}]}`)
	m := testManifest()
	m.Decisions = append(m.Decisions, testDecision("orphan"))
	if _, err := verifyCandidateDecisions(".", m, path, ""); err == nil {
		t.Fatal("expected orphan decision to fail")
	}
}

func TestCandidateDecisionAllowsLocalObservedDecisionWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	path := writeCandidate(t, dir, `{"closed_loop_candidates":[`+
		`{"id":"close-x","required_next_artifacts":["claim_binding"]}]}`)
	m := testManifest()
	decision := testDecision("local-only")
	decision.CandidateScope = "local_observed"
	m.Decisions = append(m.Decisions, decision)
	if _, err := verifyCandidateDecisions(".", m, path, ""); err != nil {
		t.Fatalf("verify local observed decision: %v", err)
	}
}

func TestCandidateDecisionAllowsProductLoopDecisionWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	path := writeCandidate(t, dir, `{"closed_loop_candidates":[]}`)
	m := testManifest()
	decision := testDecision("product-only")
	decision.CandidateScope = "product_loop"
	m.Decisions = append(m.Decisions, decision)
	if _, err := verifyCandidateDecisions(".", m, path, "product_loop"); err != nil {
		t.Fatalf("verify product loop decision absence: %v", err)
	}
}

func TestCandidateDecisionRejectsUnknownNextArtifact(t *testing.T) {
	dir := t.TempDir()
	path := writeCandidate(t, dir, `{"closed_loop_candidates":[`+
		`{"id":"close-x","required_next_artifacts":["verifier"]}]}`)
	if _, err := verifyCandidateDecisions(".", testManifest(), path, ""); err == nil {
		t.Fatal("expected unknown next artifact to fail")
	}
}

func TestCandidateDecisionRequiresReviewBy(t *testing.T) {
	decision := testDecision("close-x")
	decision.ReviewBy = ""
	if err := verifyDecision(decision); err == nil {
		t.Fatal("expected missing review_by to fail")
	}
}

package main

import "testing"

func TestCandidateDecisionCoversLocalQACandidate(t *testing.T) {
	dir := t.TempDir()
	path := writeCandidate(t, dir, `{"closed_loop_candidates":[`+
		`{"id":"close-x","required_next_artifacts":["claim_binding","verifier"],`+
		testEvidenceGraphJSON()+`}]}`)
	result, err := verifyCandidateDecisions(".", testManifest(), path, "")
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
		`{"id":"missing","required_next_artifacts":["claim_binding"],`+
		testEvidenceGraphJSON()+`}]}`)
	if _, err := verifyCandidateDecisions(".", testManifest(), path, ""); err == nil {
		t.Fatal("expected missing decision to fail")
	}
}

func TestCandidateDecisionAllowsFreshGeneratedCandidate(t *testing.T) {
	dir := t.TempDir()
	path := writeCandidate(t, dir, `{"closed_loop_candidates":[`+
		`{"id":"coverage.fresh","status":"candidate",`+
		testEvidenceGraphJSON()+`}]}`)
	m := testManifest()
	m.Decisions = nil
	result, err := verifyCandidateDecisions(".", m, path, "")
	if err != nil {
		t.Fatalf("verify generated candidate: %v", err)
	}
	if result.AllowedMissingCount != 1 || result.MatchedDecisionCount != 0 {
		t.Fatalf("result = %+v", result)
	}
}

func TestCandidateDecisionRejectsStaleGeneratedCandidate(t *testing.T) {
	dir := t.TempDir()
	path := writeCandidate(t, dir, `{"closed_loop_candidates":[`+
		`{"id":"coverage.stale","status":"stale","stale":true,`+
		testEvidenceGraphJSON()+`}]}`)
	m := testManifest()
	m.Decisions = nil
	if _, err := verifyCandidateDecisions(".", m, path, ""); err == nil {
		t.Fatal("expected stale generated candidate to require a decision")
	}
}

func TestCandidateDecisionAcceptsEmbeddedPromotionDecision(t *testing.T) {
	dir := t.TempDir()
	path := writeCandidate(t, dir, `{"closed_loop_candidates":[`+
		`{"id":"coverage.promoted","status":"promoted","promoted":true,`+
		testEvidenceGraphJSON()+`}]}`)
	m := testManifest()
	m.Decisions = nil
	result, err := verifyCandidateDecisions(".", m, path, "")
	if err != nil {
		t.Fatalf("verify embedded promotion: %v", err)
	}
	if result.MatchedDecisionCount != 1 || result.DecisionArtifacts[0].CandidateID != "coverage.promoted" {
		t.Fatalf("result = %+v", result)
	}
}

func TestCandidateDecisionRejectsMissingEvidenceGraph(t *testing.T) {
	dir := t.TempDir()
	path := writeCandidate(t, dir, `{"closed_loop_candidates":[`+
		`{"id":"close-x","required_next_artifacts":["claim_binding"]}]}`)
	if _, err := verifyCandidateDecisions(".", testManifest(), path, ""); err == nil {
		t.Fatal("expected missing evidence graph to fail")
	}
}

func TestCandidateDecisionScopeCoversProductLoopCandidate(t *testing.T) {
	dir := t.TempDir()
	path := writeCandidate(t, dir, `{"closed_loop_candidates":[`+
		`{"id":"product-x","required_next_artifacts":["scenario coverage row with passed or observed status"],`+
		testEvidenceGraphJSON()+`}]}`)
	m := testManifest()
	m.Decisions = append(m.Decisions, decisionRecord{
		CandidateID: "product-x", CandidateScope: "product_loop",
		Disposition: "triage_required", Priority: "P1", Owner: "product-qa-loop",
		NextLoop: "local-product-acceptance", NextArtifact: "scenario coverage row with passed or observed status",
		ReviewBy: "2026-12-31", Reason: "product loop outcome needs observed run evidence",
	})
	result, err := verifyCandidateDecisions(".", m, path, "product_loop")
	if err != nil {
		t.Fatalf("verify product loop decision: %v", err)
	}
	if result.CandidateScope != "product_loop" || len(result.DecisionArtifacts) != 1 {
		t.Fatalf("result = %+v", result)
	}
	if result.ScopeDecisionCount != 1 || result.MatchedDecisionCount != 1 {
		t.Fatalf("result = %+v", result)
	}
}

package main

import "testing"

func TestApplyClosedLoopCandidatesPromotesHarnessFindings(t *testing.T) {
	evidence := runEvidence{
		Artifacts: runArtifacts{
			ProviderEvidence: "provider.json",
			CoverageEvidence: "coverage.json",
		},
		OpenRepairs: []runRepair{{
			ProviderID: "claude",
			Class:      "provider_auth_required",
			Summary:    "login required",
		}},
		Steps: []stepEvidence{{
			ID:         "provider-integration",
			Status:     statusFailed,
			Command:    "go test ./internal/provider/claude",
			OutputTail: "permission bridge failed",
		}},
		Coverage: &runCoverage{Rows: []runCoverageRow{{
			ID:     "product.claude_approval",
			Title:  "Claude approval",
			Status: statusPartial,
			Detail: "staging approval flow not verified",
		}}},
	}

	evidence.ObservedAt = "2026-06-29T00:00:00Z"
	got := applyClosedLoopCandidates(evidence, nil)
	if got.CandidateSummary.Total != 3 || len(got.Candidates) != 3 {
		t.Fatalf("candidates=%+v summary=%+v", got.Candidates, got.CandidateSummary)
	}
	assertCandidate(t, got.Candidates, "harness-step.provider-integration")
	assertCandidate(t, got.Candidates, "open-repair.claude-provider-auth-required")
	assertCandidate(t, got.Candidates, "coverage.product-claude-approval")
}

func TestApplyClosedLoopCandidatesMarksStaleFromPriorEvidence(t *testing.T) {
	evidence := runEvidence{
		ObservedAt: "2026-06-29T00:00:00Z",
		Steps: []stepEvidence{{
			ID:     "provider-integration",
			Status: statusFailed,
		}},
	}
	previous := []closedLoopCandidate{{
		ID:              "harness-step.provider-integration",
		FirstObservedAt: "2026-06-25T00:00:00Z",
	}}

	got := applyClosedLoopCandidates(evidence, previous)
	if got.CandidateSummary.Stale != 1 {
		t.Fatalf("summary=%+v candidates=%+v", got.CandidateSummary, got.Candidates)
	}
	row := got.Candidates[0]
	if !row.Stale || row.Status != "stale" || row.AgeHours != 96 {
		t.Fatalf("candidate=%+v", row)
	}
}

func assertCandidate(t *testing.T, rows []closedLoopCandidate, id string) {
	t.Helper()
	for _, row := range rows {
		if row.ID == id && row.Status == "candidate" && row.StaleAfterHours > 0 &&
			row.FirstObservedAt != "" {
			return
		}
	}
	t.Fatalf("candidate %q missing from %+v", id, rows)
}

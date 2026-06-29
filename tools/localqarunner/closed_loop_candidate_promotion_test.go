package main

import "testing"

func TestApplyClosedLoopCandidatesMarksPromotedCandidates(t *testing.T) {
	evidence := runEvidence{
		ObservedAt: "2026-06-29T00:00:00Z",
		Coverage: &runCoverage{Rows: []runCoverageRow{{
			ID:     "local.qa.daily_freshness",
			Title:  "Daily freshness",
			Status: statusPartial,
		}}},
	}
	promotion := closedLoopPromotion{
		CandidateID: "coverage.local-qa-daily-freshness",
		LoopSource:  "docs/30-architecture/loop-engineering/local-qa-daily-trigger.riido.json",
		Verifier:    "go run ./tools/localqaschedule && go run ./tools/localqadashboard",
		Evidence:    "local.qa.daily_freshness",
		Decision:    "promoted_to_closed_loop",
	}

	got := applyClosedLoopCandidates(evidence, nil, []closedLoopPromotion{promotion})
	if got.CandidateSummary.Promoted != 1 {
		t.Fatalf("summary=%+v candidates=%+v", got.CandidateSummary, got.Candidates)
	}
	if got.CandidateSummary.Pending != 0 {
		t.Fatalf("summary=%+v", got.CandidateSummary)
	}
	row := got.Candidates[0]
	if !row.Promoted || row.Status != "promoted" || row.Graph.Decision != "promoted_to_closed_loop" {
		t.Fatalf("candidate=%+v", row)
	}
}

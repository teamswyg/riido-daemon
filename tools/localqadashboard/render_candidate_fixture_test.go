package main

func renderTestPendingCandidate() closedLoopCandidate {
	return closedLoopCandidate{
		ID:         "coverage.product",
		Source:     "coverage",
		Trigger:    "coverage_not_passed",
		Status:     "candidate",
		Summary:    "missing product outcome",
		NextAction: "promote to verifier",
		Graph: candidateEvidenceGraph{
			Decision: "candidate_for_promotion",
			NextLoop: "closed-loop.coverage-product",
		},
		AgeHours: 12,
		StaleAt:  "2026-06-25T00:00:00Z",
	}
}

func renderTestPromotedCandidate() closedLoopCandidate {
	return closedLoopCandidate{
		ID:         "coverage.local-qa-daily-freshness",
		Source:     "coverage",
		Trigger:    "coverage_not_passed",
		Status:     "promoted",
		Summary:    "daily freshness has a closed loop",
		NextAction: "watch loop registry",
		Promoted:   true,
		Graph: candidateEvidenceGraph{
			Decision: "promoted_to_closed_loop",
			NextLoop: "docs/30-architecture/loop-engineering/local-qa-daily-trigger.riido.json",
		},
	}
}

package main

func closedLoopSummaryFor(candidates []closedLoopCandidate) closedLoopSummary {
	summary := closedLoopSummary{
		Total:           len(candidates),
		StaleAfterHours: candidateStaleAfterHours,
	}
	for _, candidate := range candidates {
		if candidate.Promoted {
			summary.Promoted++
		}
		if candidate.Stale {
			summary.Stale++
		}
	}
	return summary
}

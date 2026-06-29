package main

func countAllowedMissingDecisions(
	decisions []decisionRecord,
	candidates []closedLoopCandidate,
) int {
	candidateByID := map[string]bool{}
	for _, item := range candidates {
		candidateByID[item.ID] = true
	}
	count := 0
	for _, decision := range decisions {
		if candidateByID[decision.CandidateID] {
			continue
		}
		if decisionAllowsMissingCandidate(decision) {
			count++
		}
	}
	return count
}

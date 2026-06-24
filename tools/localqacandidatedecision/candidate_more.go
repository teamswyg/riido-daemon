package main

import "fmt"

func decisionsByID(decisions []decisionRecord) map[string]decisionRecord {
	out := map[string]decisionRecord{}
	for _, decision := range decisions {
		out[decision.CandidateID] = decision
	}
	return out
}

func verifyNoOrphanDecisions(decisions []decisionRecord, candidates []closedLoopCandidate) error {
	candidateByID := map[string]bool{}
	for _, item := range candidates {
		candidateByID[item.ID] = true
	}
	for _, decision := range decisions {
		if !candidateByID[decision.CandidateID] {
			return fmt.Errorf("decision %s has no matching candidate", decision.CandidateID)
		}
	}
	return nil
}

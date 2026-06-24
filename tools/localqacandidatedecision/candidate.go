package main

import (
	"fmt"
	"slices"
)

func verifyCandidateDecisions(root string, m manifest, path string) (verifyResult, error) {
	candidate, err := loadCandidate(repoPath(root, path))
	if err != nil {
		return verifyResult{}, err
	}
	decisionByID := decisionsByID(m.Decisions)
	result := verifyResult{CandidateCount: len(candidate.ClosedLoops)}
	for _, item := range candidate.ClosedLoops {
		decision, ok := decisionByID[item.ID]
		if !ok {
			return result, fmt.Errorf("candidate %s has no decision record", item.ID)
		}
		if err := verifyDecisionNextArtifact(item, decision); err != nil {
			return result, err
		}
		result.DecisionIDs = append(result.DecisionIDs, item.ID)
		result.DecisionArtifacts = append(result.DecisionArtifacts, decisionArtifactEvidence{
			CandidateID: item.ID, NextArtifact: decision.NextArtifact,
		})
	}
	if err := verifyNoOrphanDecisions(m.Decisions, candidate.ClosedLoops); err != nil {
		return result, err
	}
	return result, nil
}

func verifyDecisionNextArtifact(candidate closedLoopCandidate, decision decisionRecord) error {
	if !slices.Contains(candidate.RequiredNextArtifacts, decision.NextArtifact) {
		return fmt.Errorf("candidate %s next_artifact %s is not required", candidate.ID, decision.NextArtifact)
	}
	return nil
}

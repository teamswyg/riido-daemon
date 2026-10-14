package main

func verifyCandidateDecisions(root string, m manifest, path, scope string) (verifyResult, error) {
	candidate, err := loadCandidate(repoPath(root, path))
	if err != nil {
		return verifyResult{}, err
	}
	decisions := scopedDecisions(m.Decisions, scope)
	decisionByID := decisionsByID(decisions)
	result := verifyResult{
		CandidateScope:      scope,
		CandidateCount:      len(candidate.ClosedLoops),
		ScopeDecisionCount:  len(decisions),
		AllowedMissingCount: countAllowedMissingDecisions(decisions, candidate.ClosedLoops),
	}
	var problems []candidateProblem
	for _, item := range candidate.ClosedLoops {
		if !item.Graph.complete() {
			problems = append(problems, missingEvidenceGraphProblem(item))
			continue
		}
		decision, ok := decisionByID[item.ID]
		if !ok {
			if item.hasEmbeddedDecision() {
				result.DecisionIDs = append(result.DecisionIDs, item.ID)
				result.DecisionArtifacts = append(result.DecisionArtifacts, item.embeddedDecisionArtifact())
				continue
			}
			if item.allowsMissingDecision() {
				result.AllowedMissingCount++
				continue
			}
			problems = append(problems, missingDecisionProblem(item))
			continue
		}
		if problem, ok := invalidArtifactProblem(item, decision); ok {
			problems = append(problems, problem)
			continue
		}
		result.DecisionIDs = append(result.DecisionIDs, item.ID)
		result.DecisionArtifacts = append(result.DecisionArtifacts, decisionArtifactEvidence{
			CandidateID: item.ID, NextArtifact: decision.NextArtifact,
		})
	}
	result.MatchedDecisionCount = len(result.DecisionIDs)
	problems = append(problems, orphanDecisionProblems(decisions, candidate.ClosedLoops)...)
	if len(problems) > 0 {
		return result, candidateDecisionError{Problems: problems}
	}
	return result, nil
}

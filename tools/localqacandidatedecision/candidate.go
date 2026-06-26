package main

func verifyCandidateDecisions(root string, m manifest, path, scope string) (verifyResult, error) {
	candidate, err := loadCandidate(repoPath(root, path))
	if err != nil {
		return verifyResult{}, err
	}
	decisions := scopedDecisions(m.Decisions, scope)
	decisionByID := decisionsByID(decisions)
	result := verifyResult{CandidateScope: scope, CandidateCount: len(candidate.ClosedLoops)}
	var problems []candidateProblem
	for _, item := range candidate.ClosedLoops {
		decision, ok := decisionByID[item.ID]
		if !ok {
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
	problems = append(problems, orphanDecisionProblems(decisions, candidate.ClosedLoops)...)
	if len(problems) > 0 {
		return result, candidateDecisionError{Problems: problems}
	}
	return result, nil
}

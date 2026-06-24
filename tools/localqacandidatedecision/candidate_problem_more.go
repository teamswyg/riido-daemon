package main

func missingDecisionProblem(candidate closedLoopCandidate) candidateProblem {
	return candidateProblem{
		CandidateID:           candidate.ID,
		Reason:                "candidate has no decision record",
		RequiredNextArtifacts: candidate.RequiredNextArtifacts,
		RecommendedAction:     "Add a decision record for this candidate.",
	}
}

func orphanDecisionProblem(decision decisionRecord) candidateProblem {
	return candidateProblem{
		CandidateID:          decision.CandidateID,
		Reason:               "decision has no matching current candidate",
		DecisionNextArtifact: decision.NextArtifact,
		RecommendedAction:    "Remove the orphan decision or provide matching candidate evidence.",
	}
}

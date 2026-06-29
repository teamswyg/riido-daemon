package main

func (c closedLoopCandidate) hasEmbeddedDecision() bool {
	return c.Promoted || c.Graph.Decision == "promoted_to_closed_loop"
}

func (c closedLoopCandidate) allowsMissingDecision() bool {
	if c.Stale || c.Status == "stale" || c.Status == "promoted" {
		return false
	}
	return len(c.RequiredNextArtifacts) == 0
}

func (c closedLoopCandidate) embeddedDecisionArtifact() decisionArtifactEvidence {
	next := c.Graph.NextLoop
	if next == "" {
		next = c.NextEvidence
	}
	return decisionArtifactEvidence{
		CandidateID:  c.ID,
		NextArtifact: next,
	}
}

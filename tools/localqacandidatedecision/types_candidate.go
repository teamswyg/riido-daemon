package main

type candidateEvidence struct {
	ClosedLoops []closedLoopCandidate `json:"closed_loop_candidates"`
}

type closedLoopCandidate struct {
	ID                    string   `json:"id"`
	Class                 string   `json:"class"`
	Reason                string   `json:"reason"`
	NextEvidence          string   `json:"next_evidence"`
	RequiredNextArtifacts []string `json:"required_next_artifacts"`
}

type verifyResult struct {
	CandidateScope    string                     `json:"candidate_scope,omitempty"`
	CandidateCount    int                        `json:"candidate_count"`
	DecisionCount     int                        `json:"decision_count"`
	DecisionIDs       []string                   `json:"decision_ids,omitempty"`
	DecisionArtifacts []decisionArtifactEvidence `json:"decision_artifacts,omitempty"`
}

type decisionArtifactEvidence struct {
	CandidateID  string `json:"candidate_id"`
	NextArtifact string `json:"next_artifact"`
}

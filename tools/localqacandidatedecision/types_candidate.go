package main

type candidateEvidence struct {
	ClosedLoops []closedLoopCandidate `json:"closed_loop_candidates"`
}

type closedLoopCandidate struct {
	ID                    string                  `json:"id"`
	Source                string                  `json:"source,omitempty"`
	Trigger               string                  `json:"trigger,omitempty"`
	Status                string                  `json:"status,omitempty"`
	Class                 string                  `json:"class"`
	Reason                string                  `json:"reason"`
	Summary               string                  `json:"summary,omitempty"`
	Evidence              string                  `json:"evidence,omitempty"`
	NextAction            string                  `json:"next_action,omitempty"`
	NextEvidence          string                  `json:"next_evidence"`
	RequiredNextArtifacts []string                `json:"required_next_artifacts"`
	Promoted              bool                    `json:"promoted,omitempty"`
	Stale                 bool                    `json:"stale,omitempty"`
	AgeHours              int                     `json:"age_hours,omitempty"`
	StaleAt               string                  `json:"stale_at,omitempty"`
	Graph                 closedLoopEvidenceGraph `json:"evidence_graph"`
}

type closedLoopEvidenceGraph struct {
	Observation string `json:"observation"`
	Hypothesis  string `json:"hypothesis"`
	Change      string `json:"change"`
	Verifier    string `json:"verifier"`
	Evidence    string `json:"evidence"`
	Decision    string `json:"decision"`
	NextLoop    string `json:"next_loop"`
}

type verifyResult struct {
	CandidateScope        string                     `json:"candidate_scope,omitempty"`
	CandidateCount        int                        `json:"candidate_count"`
	DecisionCount         int                        `json:"decision_count"`
	ManifestDecisionCount int                        `json:"manifest_decision_count,omitempty"`
	ScopeDecisionCount    int                        `json:"scope_decision_count,omitempty"`
	MatchedDecisionCount  int                        `json:"matched_decision_count,omitempty"`
	AllowedMissingCount   int                        `json:"allowed_missing_decision_count,omitempty"`
	DecisionIDs           []string                   `json:"decision_ids,omitempty"`
	DecisionArtifacts     []decisionArtifactEvidence `json:"decision_artifacts,omitempty"`
}

type decisionArtifactEvidence struct {
	CandidateID  string `json:"candidate_id"`
	NextArtifact string `json:"next_artifact"`
}

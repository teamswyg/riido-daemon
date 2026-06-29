package main

type closedLoopCandidate struct {
	ID              string                 `json:"id"`
	Source          string                 `json:"source"`
	Trigger         string                 `json:"trigger"`
	Status          string                 `json:"status"`
	Summary         string                 `json:"summary"`
	Evidence        string                 `json:"evidence,omitempty"`
	NextAction      string                 `json:"next_action"`
	Graph           candidateEvidenceGraph `json:"evidence_graph"`
	Stale           bool                   `json:"stale,omitempty"`
	FirstObservedAt string                 `json:"first_observed_at,omitempty"`
	AgeHours        int                    `json:"age_hours,omitempty"`
	StaleAt         string                 `json:"stale_at,omitempty"`
}

type candidateEvidenceGraph struct {
	Decision string `json:"decision"`
	NextLoop string `json:"next_loop"`
}

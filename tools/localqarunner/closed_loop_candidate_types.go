package main

type closedLoopSummary struct {
	Total           int `json:"total"`
	Pending         int `json:"pending"`
	Promoted        int `json:"promoted"`
	Stale           int `json:"stale"`
	StaleAfterHours int `json:"stale_after_hours"`
}

type closedLoopCandidate struct {
	ID              string                 `json:"id"`
	Source          string                 `json:"source"`
	Trigger         string                 `json:"trigger"`
	Status          string                 `json:"status"`
	Summary         string                 `json:"summary"`
	Evidence        string                 `json:"evidence,omitempty"`
	NextAction      string                 `json:"next_action"`
	Graph           candidateEvidenceGraph `json:"evidence_graph"`
	Promoted        bool                   `json:"promoted"`
	Stale           bool                   `json:"stale"`
	FirstObservedAt string                 `json:"first_observed_at"`
	LastObservedAt  string                 `json:"last_observed_at"`
	AgeHours        int                    `json:"age_hours"`
	StaleAt         string                 `json:"stale_at"`
	StaleAfterHours int                    `json:"stale_after_hours"`
}

type candidateEvidenceGraph struct {
	Observation string `json:"observation"`
	Hypothesis  string `json:"hypothesis"`
	Change      string `json:"change"`
	Verifier    string `json:"verifier"`
	Evidence    string `json:"evidence"`
	Decision    string `json:"decision"`
	NextLoop    string `json:"next_loop"`
}

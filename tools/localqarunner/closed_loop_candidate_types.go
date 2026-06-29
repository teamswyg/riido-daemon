package main

type closedLoopSummary struct {
	Total           int `json:"total"`
	Promoted        int `json:"promoted"`
	Stale           int `json:"stale"`
	StaleAfterHours int `json:"stale_after_hours"`
}

type closedLoopCandidate struct {
	ID              string `json:"id"`
	Source          string `json:"source"`
	Trigger         string `json:"trigger"`
	Status          string `json:"status"`
	Summary         string `json:"summary"`
	Evidence        string `json:"evidence,omitempty"`
	NextAction      string `json:"next_action"`
	Promoted        bool   `json:"promoted"`
	StaleAfterHours int    `json:"stale_after_hours"`
}

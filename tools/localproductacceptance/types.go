package main

import "time"

const (
	statusPassed  = "passed"
	statusFailed  = "failed"
	statusSkipped = "skipped"
)

type config struct {
	clientRoot  *string
	baseURL     *string
	workspaceID *string
	evidenceOut *string
	validFor    *time.Duration
	probeRoutes *bool
}

type evidenceFile struct {
	SchemaVersion string     `json:"schema_version"`
	ID            string     `json:"id"`
	ObservedAt    string     `json:"observed_at"`
	ExpiresAt     string     `json:"expires_at"`
	Status        string     `json:"status"`
	Scenarios     []scenario `json:"scenarios"`
}

type scenario struct {
	ID             string  `json:"id"`
	Status         string  `json:"status"`
	FailureSummary string  `json:"failure_summary,omitempty"`
	Screenshot     string  `json:"screenshot,omitempty"`
	Repair         *repair `json:"repair,omitempty"`
}

type repair struct {
	Class            string `json:"class"`
	Owner            string `json:"owner"`
	Mode             string `json:"mode"`
	Summary          string `json:"summary"`
	SuggestedCommand string `json:"suggested_command,omitempty"`
}

package main

type runDeploymentGate struct {
	Status                 string                     `json:"status"`
	RequiredCoverageStatus string                     `json:"required_coverage_status"`
	ObservedCoverageStatus string                     `json:"observed_coverage_status"`
	StrictCommand          string                     `json:"strict_command"`
	Blockers               []runDeploymentGateBlocker `json:"blockers,omitempty"`
}

type runDeploymentGateBlocker struct {
	Code    string `json:"code"`
	Summary string `json:"summary"`
	Count   int    `json:"count"`
}

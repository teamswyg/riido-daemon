package main

type localRunDeploymentGate struct {
	Status                 string                          `json:"status"`
	RequiredCoverageStatus string                          `json:"required_coverage_status"`
	ObservedCoverageStatus string                          `json:"observed_coverage_status"`
	StrictCommand          string                          `json:"strict_command"`
	Blockers               []localRunDeploymentGateBlocker `json:"blockers,omitempty"`
}

type localRunDeploymentGateBlocker struct {
	Code    string `json:"code"`
	Summary string `json:"summary"`
	Count   int    `json:"count"`
}

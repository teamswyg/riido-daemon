package main

import "encoding/json"

type localRunEvidenceWire struct {
	ObservedAt     string                  `json:"observed_at"`
	ExpiresAt      string                  `json:"expires_at"`
	Status         string                  `json:"status"`
	CoverageStatus string                  `json:"coverage_status"`
	ProviderStatus string                  `json:"provider_status,omitempty"`
	DeploymentGate localRunDeploymentGate  `json:"deployment_gate"`
	OpenRepairs    []repairEvidence        `json:"open_repairs,omitempty"`
	ClosedLoops    []localRunLoopCandidate `json:"product_closed_loop_candidates,omitempty"`
	Candidates     []closedLoopCandidate   `json:"closed_loop_candidates,omitempty"`
	Steps          []localRunStep          `json:"steps"`
}

func (run *localRunEvidence) UnmarshalJSON(data []byte) error {
	var wire localRunEvidenceWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	*run = localRunEvidence(wire)
	return run.applyLegacyClosedLoopCandidates(data)
}

package main

import "encoding/json"

type localRunEvidence struct {
	ObservedAt     string                  `json:"observed_at"`
	ExpiresAt      string                  `json:"expires_at"`
	Status         string                  `json:"status"`
	CoverageStatus string                  `json:"coverage_status"`
	ProviderStatus string                  `json:"provider_status,omitempty"`
	DeploymentGate localRunDeploymentGate  `json:"deployment_gate"`
	OpenRepairs    []repairEvidence        `json:"open_repairs,omitempty"`
	ClosedLoops    []localRunLoopCandidate `json:"closed_loop_candidates,omitempty"`
	Steps          []localRunStep          `json:"steps"`
}

type localRunStep struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func runEvidenceScenarios(path string) []externalScenario {
	run, ok := loadLocalRunEvidence(path)
	if !ok {
		return nil
	}
	return withScenarioSource([]externalScenario{s3PublishScenario(run.Steps)}, path, run.ExpiresAt)
}

func loadLocalRunEvidence(path string) (localRunEvidence, bool) {
	data, ok := readOptional(path)
	if !ok {
		return localRunEvidence{}, false
	}
	var run localRunEvidence
	if json.Unmarshal(data, &run) != nil {
		return localRunEvidence{}, false
	}
	return run, true
}

func s3PublishScenario(steps []localRunStep) externalScenario {
	seen := false
	status := statusPassed
	for _, step := range steps {
		if !isUploadStep(step.ID) {
			continue
		}
		seen = true
		if step.Status != statusPassed {
			status = statusFailed
		}
	}
	if !seen {
		return externalScenario{}
	}
	return externalScenario{ID: "local.qa.s3_publish", Status: status}
}

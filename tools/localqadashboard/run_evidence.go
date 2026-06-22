package main

import "encoding/json"

type localRunEvidence struct {
	Steps []localRunStep `json:"steps"`
}

type localRunStep struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func runEvidenceScenarios(path string) []externalScenario {
	data, ok := readOptional(path)
	if !ok {
		return nil
	}
	var run localRunEvidence
	if json.Unmarshal(data, &run) != nil {
		return nil
	}
	return []externalScenario{s3PublishScenario(run.Steps)}
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

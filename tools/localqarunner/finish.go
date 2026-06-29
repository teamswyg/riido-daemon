package main

import "fmt"

func appendStep(evidence *runEvidence, step stepEvidence) {
	evidence.Steps = append(evidence.Steps, step)
	if step.Status != statusPassed {
		evidence.Status = statusFailed
	}
}

func finishRun(root string, cfg config, evidence runEvidence) (string, error) {
	evidence = applyClosedLoopCandidates(evidence)
	evidence = applyDeploymentGate(evidence)
	evidence = applyStrictCoverage(evidence)
	path := runEvidenceAbs(root, cfg)
	if err := writeJSON(path, evidence); err != nil {
		return statusFailed, fmt.Errorf("write run evidence: %w", err)
	}
	return evidence.Status, nil
}

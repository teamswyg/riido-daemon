package main

import "fmt"

func appendStep(evidence *runEvidence, step stepEvidence) {
	evidence.Steps = append(evidence.Steps, step)
	if step.Status != statusPassed {
		evidence.Status = statusFailed
	}
}

func finishRun(root string, cfg config, evidence runEvidence) (string, error) {
	path := runEvidenceAbs(root, cfg)
	promotions := loadClosedLoopPromotions(outputPath(root, *cfg.promotionManifest))
	evidence = applyClosedLoopCandidates(evidence, evidence.PreviousCandidates, promotions)
	evidence = applyDeploymentGate(evidence)
	evidence = applyStrictCoverage(evidence)
	if err := writeJSON(path, evidence); err != nil {
		return statusFailed, fmt.Errorf("write run evidence: %w", err)
	}
	return evidence.Status, nil
}

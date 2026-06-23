package main

func runS3Phase(root string, cfg config, evidence runEvidence) (string, error) {
	if _, err := finishRun(root, cfg, evidence); err != nil {
		return statusFailed, err
	}
	runUploadSteps(root, cfg, &evidence)
	if _, err := finishRun(root, cfg, evidence); err != nil {
		return statusFailed, err
	}
	runFinalDashboardStep(root, cfg, &evidence)
	if err := applyCoverageEvidence(root, cfg, &evidence); err != nil {
		return statusFailed, err
	}
	if _, err := finishRun(root, cfg, evidence); err != nil {
		return statusFailed, err
	}
	if err := syncFinalDashboardArtifacts(root, cfg, &evidence); err != nil {
		evidence.Status = statusFailed
		_, _ = finishRun(root, cfg, evidence)
		return statusFailed, err
	}
	if _, err := finishRun(root, cfg, evidence); err != nil {
		return statusFailed, err
	}
	if err := syncFinalRunEvidence(root, cfg, evidence.ObservedAt); err != nil {
		evidence.Status = statusFailed
		_, _ = finishRun(root, cfg, evidence)
		return statusFailed, err
	}
	return finishRun(root, cfg, evidence)
}

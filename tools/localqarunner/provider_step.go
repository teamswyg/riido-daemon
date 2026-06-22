package main

func runProviderStep(root string, cfg config, evidence *runEvidence) string {
	args := []string{"run", *cfg.providerTool, "-repo", root, "-check-doc"}
	if *cfg.runIntegration {
		args = append(args, "-run-integration")
	}
	args = append(args,
		"-valid-for", cfg.validFor.String(),
		"-evidence-out", *cfg.providerEvidence,
	)
	step := runStep(root, "provider-evidence", "go", args...)
	appendStep(evidence, step)
	if step.Status != statusPassed {
		return step.Status
	}
	if err := applyProviderEvidence(root, cfg, evidence); err != nil {
		appendStep(evidence, providerReadFailure(err))
		return statusFailed
	}
	return step.Status
}

func providerReadFailure(err error) stepEvidence {
	return stepEvidence{
		ID:         "provider-evidence-aggregate",
		Status:     statusFailed,
		OutputTail: err.Error(),
	}
}

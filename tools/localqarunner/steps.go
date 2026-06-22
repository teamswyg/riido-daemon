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
	return step.Status
}

func runDashboardStep(root string, cfg config, evidence *runEvidence) {
	args := []string{
		"run", *cfg.dashboardTool,
		"-provider-evidence", *cfg.providerEvidence,
		"-run-evidence", *cfg.runEvidence,
		"-schedule-evidence", *cfg.scheduleEvidence,
		"-release-evidence", *cfg.releaseEvidence,
		"-coverage-manifest", *cfg.coverageManifest,
		"-out", *cfg.dashboardHTML,
	}
	if *cfg.productEvidence != "" {
		args = append(args, "-product-evidence", *cfg.productEvidence)
	}
	appendStep(evidence, runStep(root, "dashboard-render", "go", args...))
}

func runProductStep(root string, cfg config, evidence *runEvidence) string {
	args := []string{
		"run", *cfg.productTool,
		"-client-root", *cfg.clientRoot,
		"-base-url", *cfg.productBaseURL,
		"-workspace-id", *cfg.productWorkspace,
		"-screenshots", *cfg.productScreenshots,
		"-storage-state", *cfg.productStorage,
		"-agent-host", *cfg.productAgentHost,
		"-riido-api-host", *cfg.productRiidoHost,
		"-team-id", *cfg.productTeamID,
		"-valid-for", cfg.validFor.String(),
		"-evidence-out", *cfg.productEvidence,
		"-lab-out", *cfg.productLab,
	}
	args = appendProductTaskArgs(args, cfg)
	if *cfg.productBrowserE2E {
		args = append(args, "-browser-e2e")
	}
	if *cfg.productStartClient {
		args = append(args, "-start-client")
	}
	if !*cfg.productTaskFixture {
		args = append(args, "-create-task-fixture=false")
	}
	step := runStep(root, "product-acceptance", "go", args...)
	appendStep(evidence, step)
	return step.Status
}

func runUploadSteps(root string, cfg config, evidence *runEvidence) {
	stamp := timestampSlug(evidence.ObservedAt)
	for _, upload := range uploads(cfg, stamp) {
		args := []string{"s3", "cp", upload.source, upload.target}
		if upload.recursive {
			args = append(args, "--recursive")
		}
		args = append(args, "--cache-control", "no-store")
		appendStep(evidence, runStep(root, upload.id, "aws", args...))
	}
}

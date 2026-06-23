package main

func runDashboardStep(root string, cfg config, evidence *runEvidence) {
	runDashboardStepID(root, cfg, evidence, "dashboard-render")
}

func runFinalDashboardStep(root string, cfg config, evidence *runEvidence) {
	runDashboardStepID(root, cfg, evidence, "dashboard-render-final")
}

func runDashboardStepID(root string, cfg config, evidence *runEvidence, id string) {
	args := []string{
		"run", *cfg.dashboardTool,
		"-provider-evidence", *cfg.providerEvidence,
		"-run-evidence", *cfg.runEvidence,
		"-schedule-evidence", *cfg.scheduleEvidence,
		"-infra-evidence", *cfg.infraEvidence,
		"-release-evidence", *cfg.releaseEvidence,
		"-coverage-manifest", *cfg.coverageManifest,
		"-out", *cfg.dashboardHTML,
		"-coverage-out", *cfg.coverageEvidence,
	}
	if *cfg.productEvidence != "" {
		args = append(args, "-product-evidence", *cfg.productEvidence)
	}
	appendStep(evidence, runStep(root, id, "go", args...))
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
		"-manual-evidence-out", *cfg.manualEvidence,
		"-domain-cache", *cfg.domainCache,
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
